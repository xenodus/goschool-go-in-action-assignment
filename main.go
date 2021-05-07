// Package main kickstarts the application by calling StartHttpServer function from web package.
// It also creates a pooled mysql database connection that is passed to StartHttpServer and used in the rest of the program.
package main

import (
	"assignment4/clinic"
	"assignment4/web"
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {

	// The connection pool to be used by the application's life cycle.
	// Defering the closing of connection to when the application ends.
	db, err := sql.Open("mysql", clinic.DbConnection())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	web.StartHttpServer(db)
}
