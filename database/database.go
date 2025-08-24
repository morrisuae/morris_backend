package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"morris-backend.com/main/services/helper"
)

func Initdb() {
	// const (
	// 	host     = "ep-lingering-brook-a255dk8x.eu-central-1.pg.koyeb.app"
	// 	port     = 5432
	// 	user     = "koyeb-adm"
	// 	password = "1ITJl8VcnfHy"
	// 	dbname   = "koyebdb"
	// )

	const (
		host     = "ep-icy-forest-a25cphqe.eu-central-1.pg.koyeb.app"
		port     = 5432
		user     = "koyeb-adm"
		password = "npg_v1Nj6WPEehwS"
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

	// createTable := `CREATE TABLE IF NOT EXISTS enquiries (
	// id SERIAL PRIMARY KEY,
	// name TEXT,
	// email TEXT,
	// phone TEXT,
	// enquiry TEXT,
	// attachment TEXT,
	// created_date TIMESTAMP
	// )`

	// createTable := `CREATE TABLE IF NOT EXISTS banners (
	// id SERIAL PRIMARY KEY,
	// image TEXT,
	// title TEXT,
	// created_date TIMESTAMP
	// )`

	// _, err = helper.DB.Exec(createTable)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println("Table created successfully")

	// // Rename table
	// oldTableName := "category"
	// newTableName := "part"
	// query := fmt.Sprintf("ALTER TABLE %s RENAME TO %s;", oldTableName, newTableName)

	// _, err = helper.DB.Exec(query)
	// if err != nil {
	// 	log.Fatalf("Error renaming table: %v", err)
	// }

	// fmt.Println("Table renamed successfully")

	// // Table and column details
	// tableName := "parts"
	// columnName := "image"
	// columnType := "TEXT" // Change to desired type, e.g., VARCHAR(100), INT, etc.

	// // Construct the ALTER TABLE query
	// query := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", tableName, columnName, columnType)

	// // Execute the query
	// _, err = helper.DB.Exec(query)
	// if err != nil {
	// 	log.Fatalf("Failed to add column: %v", err)
	// }

	// log.Println("Column added successfully!")

	// Step 1: Update NULL values to a default value (e.g., empty string)
	// updateQuery := `UPDATE parts SET id = '' WHERE sub_category IS NULL;`
	// _, err = helper.DB.Exec(updateQuery)
	// if err != nil {
	// 	log.Fatalf("Failed to update NULL values: %v", err)
	// }
	// fmt.Println("NULL values in 'sub_category' column updated successfully.")

	// 	// Step 2: Set the column to NOT NULL
	// 	alterQuery := `ALTER TABLE parts ALTER COLUMN id SET DEFAULT nextval('parts_id_seq');`
	// 	_, err = helper.DB.Exec(alterQuery)
	// 	if err != nil {
	// 		log.Fatalf("Failed to execute query: %v", err)
	// 	}
	// 	fmt.Println("Column 'id' has been successfully updated to auto-increment.")

}
