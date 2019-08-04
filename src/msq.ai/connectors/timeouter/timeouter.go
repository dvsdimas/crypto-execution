package timeouter

import (
	"fmt"
	"github.com/go-errors/errors"
	log "github.com/sirupsen/logrus"
	"msq.ai/constants"
	"msq.ai/data/cmd"
	"msq.ai/db/postgres/dao"
	dic "msq.ai/db/postgres/dictionaries"
	pgh "msq.ai/db/postgres/helper"
	"time"
)

func RunTimeOuter(dburl string, dictionaries *dic.Dictionaries) {

	ctxLog := log.WithFields(log.Fields{"id": "TimeOuter"})

	ctxLog.Info("TimeOuter is going to start")

	if len(dburl) < 1 {
		ctxLog.Fatal("dburl is empty !")
	}

	if dictionaries == nil {
		ctxLog.Fatal("dictionaries is nil !")
		return
	}

	logErrWithST := func(msg string, err error) {
		ctxLog.WithField("stacktrace", fmt.Sprintf("%+v", err.(*errors.Error).ErrorStack())).Error(msg)
	}

	db, err := pgh.GetDbByUrl(dburl)

	if err != nil {
		logErrWithST("Cannot connect to DB with URL ["+dburl+"] ", err)
	}

	db.SetMaxIdleConns(1)
	db.SetMaxOpenConns(1)
	db.SetConnMaxLifetime(time.Hour)

	statusCreatedId := dictionaries.ExecutionStatuses().GetIdByName(constants.ExecutionStatusCreatedName)
	statusTimedOutId := dictionaries.ExecutionStatuses().GetIdByName(constants.ExecutionStatusTimedOutName)

	finishStaleCommands := func(baseLine time.Time) (*[]*cmd.Command, error) {
		return dao.FinishStaleCommands(db, statusCreatedId, statusTimedOutId, baseLine, 10)
	}

	go func() {

		var cmds *[]*cmd.Command

		for {

			time.Sleep(1 * time.Second)

			for {

				cmds, err = finishStaleCommands(time.Now())

				if err != nil {
					logErrWithST("tryGetStaleCommands error ", err)
					time.Sleep(10 * time.Second)
					break
				}

				if cmds == nil {
					break
				}

				// TODO notify
			}
		}
	}()

}
