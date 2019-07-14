package dao

import (
	"context"
	"database/sql"
	"errors"
	"github.com/vishalkuo/bimap"
	dic "msq.ai/db/postgres/dictionaries"
)

const loadExchangesSql = "SELECT id, name FROM exchange"
const loadDirectionsSql = "SELECT id, value FROM direction"
const loadOrderTypesSql = "SELECT id, type FROM order_type"
const loadTimeInForceSql = "SELECT id, type FROM time_in_force"
const loadExecutionTypesSql = "SELECT id, type FROM execution_type"
const loadExecutionStatusSql = "SELECT id, value FROM execution_status"

func InsertCommand(db *sql.DB,
	exchangeId int16,
	instrument string,
	directionId int16,
	orderTypeId int16,
	limitPrice float32,
	amount float32,
	executionTypeId int16,
	refPositionIdVal string,
	accountId int64) (int64, error) {

	// TODO

	return 1, nil
}

func LoadDictionaries(db *sql.DB) (*dic.Dictionaries, error) {

	exchanges, err := loadExchanges(db)

	if err != nil {
		return nil, err
	}

	if exchanges.Size() == 0 {
		return nil, errors.New("exchanges dictionary is empty")
	}

	directions, err := loadDirections(db)

	if err != nil {
		return nil, err
	}

	if directions.Size() == 0 {
		return nil, errors.New("directions dictionary is empty")
	}

	orderTypes, err := loadOrderTypes(db)

	if err != nil {
		return nil, err
	}

	if orderTypes.Size() == 0 {
		return nil, errors.New("orderTypes dictionary is empty")
	}

	timeInForce, err := loadTimeInForce(db)

	if err != nil {
		return nil, err
	}

	if timeInForce.Size() == 0 {
		return nil, errors.New("timeInForce dictionary is empty")
	}

	executionTypes, err := loadExecutionTypes(db)

	if err != nil {
		return nil, err
	}

	if executionTypes.Size() == 0 {
		return nil, errors.New("executionTypes dictionary is empty")
	}

	executionStatuses, err := loadExecutionStatuses(db)

	if err != nil {
		return nil, err
	}

	if executionStatuses.Size() == 0 {
		return nil, errors.New("executionStatuses dictionary is empty")
	}

	return dic.NewDictionaries(exchanges, directions, orderTypes, timeInForce, executionTypes, executionStatuses), nil
}

func loadExchanges(db *sql.DB) (*bimap.BiMap, error) {
	return loadDictionary(db, loadExchangesSql)
}

func loadDirections(db *sql.DB) (*bimap.BiMap, error) {
	return loadDictionary(db, loadDirectionsSql)
}

func loadOrderTypes(db *sql.DB) (*bimap.BiMap, error) {
	return loadDictionary(db, loadOrderTypesSql)
}

func loadTimeInForce(db *sql.DB) (*bimap.BiMap, error) {
	return loadDictionary(db, loadTimeInForceSql)
}

func loadExecutionTypes(db *sql.DB) (*bimap.BiMap, error) {
	return loadDictionary(db, loadExecutionTypesSql)
}

func loadExecutionStatuses(db *sql.DB) (*bimap.BiMap, error) {
	return loadDictionary(db, loadExecutionStatusSql)
}

func loadDictionary(db *sql.DB, sqlValue string) (*bimap.BiMap, error) {

	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: true})

	if err != nil {
		return nil, err
	}

	rows, err := tx.Query(sqlValue)

	if err != nil {
		return nil, err
	}

	var (
		id   int16
		name string
	)

	biMap := bimap.NewBiMap()

	for rows.Next() {

		err = rows.Scan(&id, &name)

		if err != nil {
			return nil, err
		}

		biMap.Insert(id, name)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	err = rows.Close()

	if err != nil {
		return nil, err
	}

	err = tx.Commit()

	if err != nil {
		return nil, err
	}

	biMap.MakeImmutable()

	return biMap, nil
}
