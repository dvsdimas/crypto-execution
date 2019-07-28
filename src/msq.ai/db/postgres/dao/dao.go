package dao

import (
	"context"
	"database/sql"
	"github.com/go-errors/errors"
	"github.com/lib/pq"
	"github.com/vishalkuo/bimap"
	"msq.ai/data/cmd"
	dic "msq.ai/db/postgres/dictionaries"
	"time"
)

const duplicateKeyValueViolates = "23505"

const loadExchangesSql = "SELECT id, name FROM exchange"
const loadDirectionsSql = "SELECT id, value FROM direction"
const loadOrderTypesSql = "SELECT id, type FROM order_type"
const loadTimeInForceSql = "SELECT id, type FROM time_in_force"
const loadExecutionTypesSql = "SELECT id, type FROM execution_type"
const loadExecutionStatusSql = "SELECT id, value FROM execution_status"

const getCommandIdByFingerPrintSql = "SELECT id FROM execution WHERE finger_print = $1"

const insertCommandSql = "INSERT INTO execution (exchange_id, instrument_name, direction_id, order_type_id, limit_price, time_in_force_id, " +
	"amount, status_id, execution_type_id, execute_till_time, ref_position_id, update_timestamp, account_id, api_key, secret_key, " +
	"finger_print) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16) RETURNING id"

const insertCommandHistorySql = "INSERT INTO execution_history (execution_id, status_from_id, status_to_id, timestamp, description) " +
	"VALUES ($1, $2, $3, $4, $5)"

const selectCommandSql = "SELECT id, exchange_id, instrument_name, direction_id, order_type_id, limit_price, amount, " +
	"status_id, connector_id, execution_type_id,execute_till_time, ref_position_id, time_in_force_id, update_timestamp, account_id, " +
	"api_key, secret_key, finger_print FROM execution"

const loadCommandByIdSql = selectCommandSql + " WHERE id = $1"

const tryGetCommandForExecutionSql = selectCommandSql + " WHERE exchange_id = $1 AND status_id = $2 AND connector_id ISNULL " +
	"AND execute_till_time > $3 FOR UPDATE LIMIT $4"

const updateCommandStatusByIdSql = "UPDATE execution SET status_id = $1, connector_id = $2, update_timestamp = $3 WHERE id = $4"

func FinishExecution(db *sql.DB, executionId int64, connectorId int16, currentStatusId int16, newStatusId int16, description string) error {

	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: false})

	if err != nil {
		return errors.New(err)
	}

	stmt, err := tx.Prepare(updateCommandStatusByIdSql)

	if err != nil {
		_ = tx.Rollback()
		return errors.New(err)
	}

	now := time.Now()

	_, err = stmt.Exec(newStatusId, connectorId, now, executionId)

	if err != nil {
		_ = stmt.Close()
		_ = tx.Rollback()
		return errors.New(err)
	}

	err = stmt.Close()

	if err != nil {
		_ = tx.Rollback()
		return errors.New(err)
	}

	stmt, err = tx.Prepare(insertCommandHistorySql)

	if err != nil {
		_ = tx.Rollback()
		return errors.New(err)
	}

	_, err = stmt.Exec(executionId, currentStatusId, newStatusId, now, nullString(description))

	if err != nil {
		_ = stmt.Close()
		_ = tx.Rollback()
		return errors.New(err)
	}

	err = stmt.Close()

	if err != nil {
		_ = tx.Rollback()
		return errors.New(err)
	}

	// TODO insert order

	err = tx.Commit()

	if err != nil {
		return errors.New(err)
	}

	return nil
}

func TryGetCommandForExecution(db *sql.DB, exchangeId int16, conId int16, validTimeTo time.Time, statusCreatedId int16,
	statusExecutingId int16, limit int16) (*cmd.Command, error) {

	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: false})

	if err != nil {
		return nil, errors.New(err)
	}

	stmt, err := tx.Prepare(tryGetCommandForExecutionSql)

	if err != nil {
		_ = tx.Rollback()
		return nil, errors.New(err)
	}

	row := stmt.QueryRow(exchangeId, statusCreatedId, validTimeTo, limit)

	var (
		limitPrice    sql.NullFloat64
		connectorId   sql.NullInt64
		refPositionId sql.NullString

		command cmd.Command
	)

	err = row.Scan(&command.Id, &command.ExchangeId, &command.InstrumentName, &command.DirectionId, &command.OrderTypeId,
		&limitPrice, &command.Amount, &command.StatusId, &connectorId, &command.ExecutionTypeId, &command.ExecuteTillTime,
		&refPositionId, &command.TimeInForceId, &command.UpdateTimestamp, &command.AccountId, &command.ApiKey, &command.SecretKey,
		&command.FingerPrint)

	if err != nil {

		_ = stmt.Close()
		_ = tx.Rollback()

		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			return nil, errors.New(err)
		}
	}

	if limitPrice.Valid {
		command.LimitPrice = limitPrice.Float64
	} else {
		command.LimitPrice = -1
	}

	if connectorId.Valid {
		command.ConnectorId = connectorId.Int64
	} else {
		command.ConnectorId = -1
	}

	if refPositionId.Valid {
		command.RefPositionId = refPositionId.String
	} else {
		command.RefPositionId = ""
	}

	err = stmt.Close()

	if err != nil {
		_ = tx.Rollback()
		return nil, errors.New(err)
	}

	stmt, err = tx.Prepare(updateCommandStatusByIdSql)

	if err != nil {
		_ = tx.Rollback()
		return nil, errors.New(err)
	}

	now := time.Now()

	_, err = stmt.Exec(statusExecutingId, conId, now, command.Id)

	if err != nil {
		_ = stmt.Close()
		_ = tx.Rollback()
		return nil, errors.New(err)
	}

	err = stmt.Close()

	if err != nil {
		_ = tx.Rollback()
		return nil, errors.New(err)
	}

	stmt, err = tx.Prepare(insertCommandHistorySql)

	if err != nil {
		_ = tx.Rollback()
		return nil, errors.New(err)
	}

	_, err = stmt.Exec(command.Id, statusCreatedId, statusExecutingId, now, sql.NullString{Valid: false})

	if err != nil {
		_ = stmt.Close()
		_ = tx.Rollback()
		return nil, errors.New(err)
	}

	err = stmt.Close()

	if err != nil {
		_ = tx.Rollback()
		return nil, errors.New(err)
	}

	err = tx.Commit()

	if err != nil {
		return nil, errors.New(err)
	}

	command.ConnectorId = int64(conId)
	command.StatusId = statusExecutingId

	return &command, nil
}

func LoadCommandById(db *sql.DB, id int64) (*cmd.Command, error) {

	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: true})

	if err != nil {
		return nil, errors.New(err)
	}

	stmt, err := tx.Prepare(loadCommandByIdSql)

	if err != nil {
		_ = tx.Rollback()
		return nil, errors.New(err)
	}

	row := stmt.QueryRow(id)

	var (
		limitPrice    sql.NullFloat64
		connectorId   sql.NullInt64
		refPositionId sql.NullString

		command cmd.Command
	)

	err = row.Scan(&command.Id, &command.ExchangeId, &command.InstrumentName, &command.DirectionId, &command.OrderTypeId,
		&limitPrice, &command.Amount, &command.StatusId, &connectorId, &command.ExecutionTypeId, &command.ExecuteTillTime,
		&refPositionId, &command.TimeInForceId, &command.UpdateTimestamp, &command.AccountId, &command.ApiKey, &command.SecretKey,
		&command.FingerPrint)

	if err != nil {
		_ = stmt.Close()
		_ = tx.Rollback()

		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			return nil, errors.New(err)
		}
	}

	if limitPrice.Valid {
		command.LimitPrice = limitPrice.Float64
	} else {
		command.LimitPrice = -1
	}

	if connectorId.Valid {
		command.ConnectorId = connectorId.Int64
	} else {
		command.ConnectorId = -1
	}

	if refPositionId.Valid {
		command.RefPositionId = refPositionId.String
	} else {
		command.RefPositionId = ""
	}

	err = stmt.Close()

	if err != nil {
		_ = tx.Rollback()
		return nil, errors.New(err)
	}

	err = tx.Commit()

	if err != nil {
		return nil, errors.New(err)
	}

	return &command, nil
}

func nullString(s string) sql.NullString {

	if len(s) == 0 {
		return sql.NullString{Valid: false}
	}

	return sql.NullString{String: s, Valid: true}
}

func nullLimitPrice(limitPrice float64) sql.NullFloat64 {

	if limitPrice < 0 {
		return sql.NullFloat64{Valid: false}
	}

	return sql.NullFloat64{Float64: limitPrice, Valid: true}
}

func getCommandIdByFingerPrint(db *sql.DB, fingerPrint string) (int64, error) {

	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: true})

	if err != nil {
		return -1, errors.New(err)
	}

	stmt, err := tx.Prepare(getCommandIdByFingerPrintSql)

	if err != nil {
		_ = tx.Rollback()
		return -1, errors.New(err)
	}

	row := stmt.QueryRow(fingerPrint)

	var id int64

	err = row.Scan(&id)

	if err != nil {
		_ = stmt.Close()
		_ = tx.Rollback()

		return -1, errors.New(err)
	}

	err = stmt.Close()

	if err != nil {
		_ = tx.Rollback()
		return -1, errors.New(err)
	}

	err = tx.Commit()

	if err != nil {
		return -1, errors.New(err)
	}

	return id, nil
}

func InsertCommand(db *sql.DB, exchangeId int16, instrument string, directionId int16, orderTypeId int16, limitPrice float64, timeInForceId int16,
	amount float64, statusId int16, executionTypeId int16, future time.Time, refPositionIdVal string, now time.Time, accountId int64,
	apiKey string, secretKey string, fingerPrint string) (int64, error) {

	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: false})

	if err != nil {
		return -1, errors.New(err)
	}

	stmt, err := tx.Prepare(insertCommandSql)

	if err != nil {
		_ = tx.Rollback()
		return -1, errors.New(err)
	}

	row := stmt.QueryRow(exchangeId, instrument, directionId, orderTypeId, nullLimitPrice(limitPrice), timeInForceId, amount, statusId,
		executionTypeId, future, nullString(refPositionIdVal), now, accountId, apiKey, secretKey, fingerPrint)

	var id int64

	err = row.Scan(&id)

	if err != nil {
		_ = stmt.Close()
		_ = tx.Rollback()

		pqErr := err.(*pq.Error)

		if pqErr.Code == duplicateKeyValueViolates {
			return getCommandIdByFingerPrint(db, fingerPrint)
		}

		return -1, errors.New(err)
	}

	err = stmt.Close()

	if err != nil {
		_ = tx.Rollback()
		return -1, errors.New(err)
	}

	stmt, err = tx.Prepare(insertCommandHistorySql)

	if err != nil {
		_ = tx.Rollback()
		return -1, errors.New(err)
	}

	_, err = stmt.Exec(id, statusId, statusId, now, sql.NullString{Valid: false})

	if err != nil {
		_ = stmt.Close()
		_ = tx.Rollback()
		return -1, errors.New(err)
	}

	err = stmt.Close()

	if err != nil {
		_ = tx.Rollback()
		return -1, errors.New(err)
	}

	err = tx.Commit()

	if err != nil {
		return -1, errors.New(err)
	}

	return id, nil
}

func LoadDictionaries(db *sql.DB) (*dic.Dictionaries, error) {

	exchanges, err := loadExchanges(db)

	if err != nil {
		return nil, errors.New(err)
	}

	if exchanges.Size() == 0 {
		return nil, errors.Errorf("Inconsistent DB schema! 'exchanges' dictionary is empty")
	}

	directions, err := loadDirections(db)

	if err != nil {
		return nil, errors.New(err)
	}

	if directions.Size() == 0 {
		return nil, errors.Errorf("Inconsistent DB schema! 'directions' dictionary is empty")
	}

	orderTypes, err := loadOrderTypes(db)

	if err != nil {
		return nil, errors.New(err)
	}

	if orderTypes.Size() == 0 {
		return nil, errors.Errorf("Inconsistent DB schema! 'orderTypes' dictionary is empty")
	}

	timeInForce, err := loadTimeInForce(db)

	if err != nil {
		return nil, errors.New(err)
	}

	if timeInForce.Size() == 0 {
		return nil, errors.Errorf("Inconsistent DB schema! 'timeInForce' dictionary is empty")
	}

	executionTypes, err := loadExecutionTypes(db)

	if err != nil {
		return nil, errors.New(err)
	}

	if executionTypes.Size() == 0 {
		return nil, errors.Errorf("Inconsistent DB schema! 'executionTypes' dictionary is empty")
	}

	executionStatuses, err := loadExecutionStatuses(db)

	if err != nil {
		return nil, errors.New(err)
	}

	if executionStatuses.Size() == 0 {
		return nil, errors.Errorf("Inconsistent DB schema! 'executionStatuses' dictionary is empty")
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
		return nil, errors.New(err)
	}

	rows, err := tx.Query(sqlValue)

	if err != nil {
		_ = tx.Rollback()
		return nil, errors.New(err)
	}

	var (
		id   int16
		name string
	)

	biMap := bimap.NewBiMap()

	for rows.Next() {

		err = rows.Scan(&id, &name)

		if err != nil {
			_ = rows.Close()
			_ = tx.Rollback()
			return nil, errors.New(err)
		}

		biMap.Insert(id, name)
	}

	if err = rows.Err(); err != nil {
		_ = rows.Close()
		_ = tx.Rollback()
		return nil, errors.New(err)
	}

	err = rows.Close()

	if err != nil {
		_ = tx.Rollback()
		return nil, errors.New(err)
	}

	err = tx.Commit()

	if err != nil {
		return nil, errors.New(err)
	}

	biMap.MakeImmutable()

	return biMap, nil
}
