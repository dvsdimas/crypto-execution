package helper

import (
	"database/sql"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"msq.ai/constants"
)

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
