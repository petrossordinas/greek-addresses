package main

import (
	"gr_addresses/database"
	"gr_addresses/router"
	"log"
	"os"
)

func main() {
	logFile := initLogger()
	defer logFile.Close()
	database.InitDatabase()
	router.InitRouter(":9000")
}

func initLogger() *os.File {
	logFile, err := os.OpenFile("gr_addresses.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(logFile)
	return logFile
}
