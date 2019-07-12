package dao

import (
	"context"
	"database/sql"
	"github.com/vishalkuo/bimap"
)

const loadExchangesSql = "SELECT id, name  FROM exchange"
const loadDirectionsSql = "SELECT id, value FROM direction"
const loadOrderTypesSql = "SELECT id, type  FROM order_type"
const loadTimeInForceSql = "SELECT id, type  FROM time_in_force"
const loadExecutionTypesSql = "SELECT id, type  FROM execution_type"
const loadExecutionStatusSql = "SELECT id, value FROM execution_status"

func LoadExchanges(db *sql.DB) (*bimap.BiMap, error) {
	return loadDictionary(db, loadExchangesSql)
}

func LoadDirections(db *sql.DB) (*bimap.BiMap, error) {
	return loadDictionary(db, loadDirectionsSql)
}

func LoadOrderTypes(db *sql.DB) (*bimap.BiMap, error) {
	return loadDictionary(db, loadOrderTypesSql)
}

func LoadTimeInForce(db *sql.DB) (*bimap.BiMap, error) {
	return loadDictionary(db, loadTimeInForceSql)
}

func LoadExecutionTypes(db *sql.DB) (*bimap.BiMap, error) {
	return loadDictionary(db, loadExecutionTypesSql)
}

func LoadExecutionStatuses(db *sql.DB) (*bimap.BiMap, error) {
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

	defer rows.Close()

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

	err = tx.Commit()

	if err != nil {
		return nil, err
	}

	biMap.MakeImmutable()

	return biMap, nil
}
