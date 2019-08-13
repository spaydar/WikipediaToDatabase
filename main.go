package main

import (
	"bufio"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"io/ioutil"
	"net/http"
	"os"
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

func queryAndCache(db *sql.DB, infile string) error {
	// open input file and defer closing until main function returns
	file, err := os.Open(infile)
	if err != nil {
		return err
	}
	defer file.Close()

	url := "a"
	// create a scanner to read input file
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		city := scanner.Text()
		go func() {
			var cityval string
			var urlval string
			rec := db.QueryRow("SELECT city, url FROM cities WHERE city = $1;", city)
			switch err := rec.Scan(&cityval, &urlval); err {
			case sql.ErrNoRows:
				fmt.Println("Record not in database. Creating record")
				_, err := db.Exec("INSERT INTO cities (city, url) VALUES ($1, $2)", city, url)
				if err != nil {
					panic(err)
				}
				url = url + string(url[len(url)-1]+1)
			case nil:
				fmt.Println(cityval, urlval)
			default:
				panic(err)
			}
		}()
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error closing scanner: %s\n", err)
		return err
	}
	time.Sleep(5 * time.Second)
	// FIXME -- need to implement channels so that all goroutines finish


	resp, err := http.Get("https://en.wikipedia.org/api/rest_v1/page/html/Paris")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(len(data))


	return nil
}

func main() {
	/*
	// start PostgreSQL database and defer closing until main function returns
	db, err := startDatabase()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = queryAndCache(db, os.Args[1])
	if err != nil {
		panic(err)
	}

	 */
	pdf, err := os.Open("Paris.pdf")
	if err != nil {
		panic(err)
	}
	defer pdf.Close()

	service, err := getService()
	if err != nil {
		panic(err)
	}
	dir, err := createDir(service, "Dynamic Signal", "root")
	if err != nil {
		panic(err)
	}
	file, err := createFile(service, "paris.pdf", "application/pdf", pdf, dir.Id)
	if err != nil {
		panic(err)
	}
	fmt.Printf("File '%s' successfully uploaded in '%s' directory", file.Name,  dir.Name)
}


