package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"os"
)

func init() {

	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})

	log.SetOutput(os.Stdout)
	log.SetLevel(log.TraceLevel)
}

func main() {

	log.Info("Execution is going to start")

	pwd, err := os.Getwd()

	if err != nil {
		log.Fatal("Getwd error", err)
	}

	log.Trace("Current folder is [" + pwd + "]")

	connStr := "postgres://msq:pwd@localhost:5432/msq"

	db, err := sql.Open("postgres", connStr)

	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()

	if err != nil {
		log.Fatal(err)
	}

	err = db.Close()

	if err != nil {
		log.Fatal(err)
	}

}
