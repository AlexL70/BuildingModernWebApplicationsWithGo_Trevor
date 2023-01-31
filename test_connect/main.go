package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	// connect to a database
	// you need to define os variable called POSTGRES_URL as shown below:
	// "host=localhost port=5432 dbname=<your_db_name> user=<your_user_name> password=<your_password>"
	conn, err := sql.Open("pgx", os.Getenv("POSTGRES_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()
	log.Println("Connected to Postgres!")
	// test connection
	err = conn.Ping()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to ping the database: %v\n", err)
		os.Exit(1)
	}
	log.Println("Pinged succesfully!")

	// get rows from table
	err = getAllRows(conn)
	if err != nil {
		log.Fatal(err)
	}
	// insert a row
	query := `insert into users(first_name, last_name) values($1, $2)`
	_, err = conn.Exec(query, "John", "Lennon")
	if err != nil {
		log.Fatal(err)
	}
	// get rows from table again
	err = getAllRows(conn)
	if err != nil {
		log.Fatal(err)
	}
	// update a row
	stmt := `update users set first_name = $1 where first_name = $2`
	_, err = conn.Exec(stmt, "Julian", "John")
	if err != nil {
		log.Fatal(err)
	}
	// get rows from table again
	err = getAllRows(conn)
	if err != nil {
		log.Fatal(err)
	}
	// get one row by id
	query = `select id, first_name, last_name from users where id = $1`
	var fistName, lastName string
	var id int
	row := conn.QueryRow(query, 1)
	err = row.Scan(&id, &fistName, &lastName)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Record got by id: %d, %s, %s\n", id, fistName, lastName)
	// delete a row
	stmt = `delete from users where first_name = $1`
	_, err = conn.Exec(stmt, "Julian")
	if err != nil {
		log.Fatal(err)
	}
	// get rows again
	err = getAllRows(conn)
	if err != nil {
		log.Fatal(err)
	}
}

func getAllRows(conn *sql.DB) error {
	rows, err := conn.Query("select id, first_name, last_name from users")
	if err != nil {
		return err
	}
	defer rows.Close()

	var firstName, lastName string
	var id int
	fmt.Println("------------------------------")
	for rows.Next() {
		err := rows.Scan(&id, &firstName, &lastName)
		if err != nil {
			return err
		}
		fmt.Println(id, firstName, lastName)
	}
	fmt.Println("------------------------------")
	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}
