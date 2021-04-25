package database

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DBConn Connection to database
var DBConn *gorm.DB

// InitDatabase initializes the database
func InitDatabase() {
	var err error
	fmt.Println("Attempting to connect to database...")
	DBConn, err = gorm.Open(sqlite.Open("db/gr_addresses.db"), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database db/gr_addresses.db")
	}
	fmt.Println("Connection opened to SQLite3 database db/gr_addresses.db")
}
