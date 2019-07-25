package helper

import (
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
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
		return nil, errors.WithStack(err)
	}

	err = db.Ping()

	if err != nil {
		_ = db.Close()
		return nil, errors.WithStack(err)
	}

	log.Trace("Successfully connected to DB")

	return db, nil
}

func CheckDbUrl(url string) error {

	log.Trace("Try to check DB connection ...")

	db, err := sql.Open(constants.DbName, url)

	if err != nil {
		return errors.WithStack(err)
	}

	err = db.Ping()

	if err != nil {
		return errors.WithStack(err)
	}

	err = db.Close()

	if err != nil {
		return errors.WithStack(err)
	}

	log.Trace("Successfully connected to DB")

	return nil
}
