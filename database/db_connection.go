package database

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DBConn Connection to database
var DBConn *gorm.DB

// InitDatabase initializes the database at the given sqlite file path
func InitDatabase(path string) {
	var err error
	fmt.Println("Attempting to connect to database...")
	DBConn, err = gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database " + path)
	}
	fmt.Println("Connection opened to SQLite3 database " + path)
}
