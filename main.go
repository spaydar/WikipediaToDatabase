package WikipediaToDatabase

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "sample"
	password = "password"
	dbname   = "dynamic_signal"
)

func main() error {
	// specify information to open PostgreSQL database
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s " +
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// open database defer closing it until main function returns
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return err
	}
	defer db.Close()

	// create a connection to the database
	err = db.Ping()
	if err != nil {
		return err
	}
	fmt.Println("Connection to database successful")
	
	return nil
}
