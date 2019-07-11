package helper

import (
	"database/sql"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"msq.ai/constants"
)

func CloseDb(db *sql.DB) {

	if db == nil {
		return
	}

	err := db.Close()

	if err != nil {
		log.Error("CloseDb error ", err)
	}
}

func GetDbByUrl(url string) (*sql.DB, error) {

	log.Trace("Try connect to DB ...")

	db, err := sql.Open(constants.DbName, url)

	if err != nil {
		return nil, err
	}

	err = db.Ping()

	if err != nil {
		return nil, err
	}

	log.Trace("Successfully connected to DB")

	return db, nil
}

func CheckDbUrl(url string) error {

	log.Trace("Try to check DB connection ...")

	db, err := sql.Open(constants.DbName, url)

	if err != nil {
		return err
	}

	err = db.Ping()

	if err != nil {
		return err
	}

	err = db.Close()

	if err != nil {
		return err
	}

	log.Trace("Successfully connected to DB")

	return nil
}
