package main

import (
	"bufio"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"google.golang.org/api/drive/v3"
	"net/http"
	"os"
	"strconv"
	"time"
)

// data needed to access dynamic_signal database in PostgreSQL
const (
	host     = "localhost"
	port     = 5432
	user     = "sample"
	password = "password"
	dbname   = "dynamic_signal"
)

// this function starts the PostgreSQL database and connects to it
func startDatabase() (*sql.DB, error) {
	// specify information to open PostgreSQL database
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s " +
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// open database
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	// create a connection to the database
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	fmt.Println("Connection to database successful")
	return db, nil
}

func queryAndCache(db *sql.DB, service *drive.Service, file *os.File) error {
	// create a scanner to read content of input file
	scanner := bufio.NewScanner(file)

	// scan first line of input file to determine how many cities will be requested. verify >= 1
	scanner.Scan()
	temp, err := strconv.Atoi(scanner.Text())
	if err != nil {
		return err
	}
	if temp < 1 {
		return fmt.Errorf("Invalid number of input cities: must be at least 1")
		// later handle if input number inconsistent with input
	}
	numCities := uint(temp)

	// create file for successful requests
	urlsfile, err := os.Create("urls.txt")
	if err != nil {
		return err
	}
	defer urlsfile.Close()

	// create a file for failed requests
	errorsfile, err := os.Create("errors.txt")
	if err != nil {
		return err
	}
	defer errorsfile.Close()

	// variable to count number of queries requested
	var citiesRequested uint = 0

	// variable to track if Drive folder files have been created
	//var directoryCreated bool = false
	dir, err := createDir(service, "Cities", "root")
	if err != nil {
		panic(err)
	}

	// struct for sending city name and url/error source over channels
	type Pair struct {
		City string
		Val string
	}

	// channels used to write to output files. Channels implement locking
	urlsChan := make(chan *Pair)
	errorsChan := make(chan *Pair)

	// loop and query until end of file or error
	go func() {
		for scanner.Scan() {
			city := scanner.Text()
			go func() {
				var cityval string
				var urlval string
				// Go's database package implements prepared statements to avoid SQL injection
				rec := db.QueryRow("SELECT city, url FROM cities WHERE city = $1;", city)
				switch err := rec.Scan(&cityval, &urlval); err {
				case sql.ErrNoRows:
					resp, err := http.Get(fmt.Sprintf("https://en.wikipedia.org/api/rest_v1/page/pdf/%s", city))
					if err != nil {
						errorsChan <- &Pair{city, "Wikipedia"} ; return
					}
					defer resp.Body.Close()
					file, err := createFile(service, fmt.Sprintf("%s.pdf", city),
						"application/pdf", resp.Body, dir.Id)
					if err != nil {
						errorsChan <- &Pair{city, "Google Drive"} ; return
					}
					_, err = db.Exec("INSERT INTO cities (city, url) VALUES ($1, $2)", city, file.WebViewLink)
					if err != nil {
						errorsChan <- &Pair{city, "database"} ; return
					}
					urlsChan <- &Pair{city, file.WebViewLink}
				case nil:
					urlsChan <- &Pair{cityval, urlval}
				default:
					panic(err)
				}
			}()
			// Do not make more than 200 requests/sec as per Wikipedia's REST API
			time.Sleep(time.Second / 200)
		}
	}()
	for {
		select {
		case cityUrlPair := <-urlsChan:
			_, err := urlsfile.WriteString(fmt.Sprintf("%s: %s\n", cityUrlPair.City, cityUrlPair.Val))
			if err != nil { return err }
			// increment number of Requests made
			citiesRequested++
			if numCities == citiesRequested {
				if err := scanner.Err(); err != nil {
					fmt.Printf("Error closing scanner: %s\n", err) ; return err
				}
				return nil
			}
		case errormsg := <-errorsChan:
			// error was in adding to database
			if errormsg.Val != "Wikipedia" {
				_, err := urlsfile.WriteString(fmt.Sprintf("Error while adding %s to %s\n",
					errormsg.City, errormsg.Val))
				if err != nil { return err }
			} else {
				_, err := urlsfile.WriteString(fmt.Sprintf("Error while requesting %s from %s\n",
					errormsg.City, errormsg.Val))
				if err != nil {
					return err
				}
			}
			// increment number of Requests made
			citiesRequested++
			if numCities == citiesRequested {
				if err := scanner.Err(); err != nil {
					fmt.Printf("Error closing scanner: %s\n", err) ; return err
				}
				return nil
			}
		}
	}

}

func main() {
	// start PostgreSQL database and defer closing until main function returns
	db, err := startDatabase()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// initiate access to Google Drive API
	service, err := getService()
	if err != nil {
		panic(err)
	}

	// open input file containing city names and defer closing until main function returns
	infile, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer infile.Close()

	// call function that scans input file, attempts to concurrently query Google Drive links
	// of city Wikipedia Google Drive docs from database. if items not in database, create docs
	// by querying Wikipedia using its REST API and send to Google Drive API to create Doc
	// cache the resulting link in the database
	err = queryAndCache(db, service, infile)
	if err != nil {
		panic(err)
	}
}


