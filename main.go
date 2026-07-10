package main

import (
	"flag"
	"gr_addresses/database"
	"gr_addresses/router"
	"log"
	"os"

	"github.com/joho/godotenv"
)

const (
	defaultPort   = "9013"
	defaultDBPath = "db/gr_addresses.db"
)

func main() {
	portFlag := flag.String("port", "", "port to listen on, e.g. 9013 (overrides PORT env var)")
	dbFlag := flag.String("db", "", "path to the sqlite database file (overrides DB_PATH env var)")
	flag.Parse()

	logFile := initLogger()
	defer logFile.Close()
	godotenv.Load()

	database.InitDatabase(resolve(*dbFlag, "DB_PATH", defaultDBPath))
	router.InitRouter(":" + resolve(*portFlag, "PORT", defaultPort))
}

// resolve returns flagVal if set, otherwise the envVar environment variable
// if set, otherwise defaultVal.
func resolve(flagVal, envVar, defaultVal string) string {
	if flagVal != "" {
		return flagVal
	}
	if v := os.Getenv(envVar); v != "" {
		return v
	}
	return defaultVal
}

func initLogger() *os.File {
	logFile, err := os.OpenFile("gr_addresses.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(logFile)
	return logFile
}
