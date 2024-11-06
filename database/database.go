package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"morris-backend.com/main/services/helper"
)

func Initdb() {
	const (
		host     = "ep-icy-sound-a2pm6hga.eu-central-1.pg.koyeb.app"
		port     = 5432
		user     = "koyeb-adm"
		password = "ZPt1wnmgf9kE"
		dbname   = "koyebdb"
	)

	// Construct the connection string
	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=require", host, port, user, password, dbname)

	// Attempt to connect to the database
	var err error
	helper.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	if err = helper.DB.Ping(); err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}
	fmt.Println("Database connection established")

	// createTable := `CREATE TABLE IF NOT EXISTS parts (
	// id SERIAL PRIMARY KEY,
	// part_number VARCHAR(256),
	// remain_part_number VARCHAR(256),
	// part_description VARCHAR(256),
	// fg_wison_part_number VARCHAR(256),
	// super_ss_number VARCHAR(256),
	// weight VARCHAR(256),
	// coo VARCHAR(256),
	// hs_code VARCHAR(256)
	// )`

	// _, err = helper.DB.Exec(createTable)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println("Table created successfully")

}
