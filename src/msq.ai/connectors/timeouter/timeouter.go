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

func RunTimeOuter(dbUrl string, dictionaries *dic.Dictionaries, redisUrl string) {

	ctxLog := log.WithFields(log.Fields{"id": "TimeOuter"})

	ctxLog.Info("TimeOuter is going to start")

	if len(dbUrl) < 1 {
		ctxLog.Fatal("dbUrl is empty !")
	}

	if dictionaries == nil {
		ctxLog.Fatal("dictionaries is nil !")
		return
	}

	logErrWithST := func(msg string, err error) {
		ctxLog.WithField("stacktrace", fmt.Sprintf("%+v", err.(*errors.Error).ErrorStack())).Error(msg)
	}

	db, err := pgh.GetDbByUrl(dbUrl)

	if err != nil {
		logErrWithST("Cannot connect to DB with URL ["+dbUrl+"] ", err)
	}

	db.SetMaxIdleConns(1)
	db.SetMaxOpenConns(1)

	statusCreatedId := dictionaries.ExecutionStatuses().GetIdByName(constants.ExecutionStatusCreatedName)
	statusTimedOutId := dictionaries.ExecutionStatuses().GetIdByName(constants.ExecutionStatusTimedOutName)

	finishStaleCommands := func(baseLine time.Time) *[]*cmd.Command {

		cmds, err := dao.FinishStaleCommands(db, statusCreatedId, statusTimedOutId, baseLine, 10)

		if err != nil {
			logErrWithST("tryGetStaleCommands error ", err)
			time.Sleep(constants.DbErrorSleepTime)
			return nil
		}

		return cmds
	}

	go func() {

		var cmds *[]*cmd.Command

		for {

			time.Sleep(1 * time.Second)

			for {

				cmds = finishStaleCommands(time.Now())

				if cmds == nil {
					break
				}

				for _, c := range *cmds {
					ctxLog.Trace("Finished stale command", c)
				}

				// TODO notify
			}
		}
	}()

}
