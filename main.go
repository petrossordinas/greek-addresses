package main

import (
	"flag"
	"gr_addresses/database"
	"gr_addresses/router"
	"log"
	"os"

	"github.com/joho/godotenv"
)

const defaultPort = "9013"

func main() {
	portFlag := flag.String("port", "", "port to listen on, e.g. 9013 (overrides PORT env var)")
	flag.Parse()

	logFile := initLogger()
	defer logFile.Close()
	godotenv.Load()

	database.InitDatabase()
	router.InitRouter(":" + resolvePort(*portFlag))
}

func resolvePort(portFlag string) string {
	if portFlag != "" {
		return portFlag
	}
	if port := os.Getenv("PORT"); port != "" {
		return port
	}
	return defaultPort
}

func initLogger() *os.File {
	logFile, err := os.OpenFile("gr_addresses.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(logFile)
	return logFile
}
