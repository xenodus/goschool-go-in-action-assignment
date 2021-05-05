package main

import (
	"assignment4/clinic"
	"assignment4/web"
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func main() {

	db, err := sql.Open("mysql", clinic.DbConnection())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	web.StartHttpServer(db)
}
