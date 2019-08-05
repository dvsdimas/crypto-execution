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

const timedOutWithOutExecutionDescription = "Timed out without trying to execute"

const loadExchangesSql = "SELECT id, name FROM exchange"
const loadDirectionsSql = "SELECT id, value FROM direction"
const loadOrderTypesSql = "SELECT id, type FROM order_type"
const loadTimeInForceSql = "SELECT id, type FROM time_in_force"
const loadExecutionTypesSql = "SELECT id, type FROM execution_type"
const loadExecutionStatusSql = "SELECT id, value FROM execution_status"

const getOrderByIdSql = "select id, external_order_id, execution_id, price, commission, commission_asset from orders where execution_id = $1"

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

const finishStaleCommandsSql = selectCommandSql + " WHERE status_id = $1 AND execute_till_time < $2 FOR UPDATE LIMIT $3"

const tryGetCommandForRecoverySql = selectCommandSql + " WHERE exchange_id = $1 AND status_id = $2 AND connector_id = $3 " +
	"AND update_timestamp < $4 FOR UPDATE LIMIT 1"

const updateCommandStatusByIdSql = "UPDATE execution SET status_id = $1, connector_id = $2, update_timestamp = $3 WHERE id = $4"

const updateCommandTimestampByIdSql = "UPDATE execution SET update_timestamp = $1 WHERE id = $2"

const insertNewOrderSql = "INSERT INTO orders (external_order_id, execution_id, price, commission, commission_asset) VALUES ($1, $2, $3, $4, $5)"

const insertNewBalanceSql = "INSERT INTO balances(execution_id, asset, free, locked) VALUES ($1, $2, $3, $4)"

func scanRowCommand(row *sql.Row, rows *sql.Rows) (*cmd.Command, error) {
	var (
		limitPrice    sql.NullFloat64
		connectorId   sql.NullInt64
		refPositionId sql.NullString

		command cmd.Command
	)

	var err error

	if row != nil {
		err = row.Scan(&command.Id, &command.ExchangeId, &command.InstrumentName, &command.DirectionId, &command.OrderTypeId,
			&limitPrice, &command.Amount, &command.StatusId, &connectorId, &command.ExecutionTypeId, &command.ExecuteTillTime,
			&refPositionId, &command.TimeInForceId, &command.UpdateTimestamp, &command.AccountId, &command.ApiKey, &command.SecretKey,
			&command.FingerPrint)
	} else {
		err = rows.Scan(&command.Id, &command.ExchangeId, &command.InstrumentName, &command.DirectionId, &command.OrderTypeId,
			&limitPrice, &command.Amount, &command.StatusId, &connectorId, &command.ExecutionTypeId, &command.ExecuteTillTime,
			&refPositionId, &command.TimeInForceId, &command.UpdateTimestamp, &command.AccountId, &command.ApiKey, &command.SecretKey,
			&command.FingerPrint)
	}

	if err != nil {
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

	return &command, nil
}

func FinishStaleCommands(db *sql.DB, statusCreatedId int16, statusTimedOutId int16, baseLine time.Time, limit int) (*[]*cmd.Command, error) {

	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: false})

	if err != nil {
		return nil, errors.New(err)
	}

	rows, err := tx.Query(finishStaleCommandsSql, statusCreatedId, baseLine, limit)

	if err != nil {
		_ = tx.Rollback()
		return nil, errors.New(err)
	}

	commands := make([]*cmd.Command, 0)

	for rows.Next() {

		command, err := scanRowCommand(nil, rows)

		if err != nil {
			_ = rows.Close()
			_ = tx.Rollback()
			return nil, err
		}

		commands = append(commands, command)
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

	for _, command := range commands {

		err = finishExecution(tx, command.Id, -1, statusCreatedId, statusTimedOutId, timedOutWithOutExecutionDescription, nil, nil)

		if err != nil {
			return nil, err
		}

		command.StatusId = statusTimedOutId
		command.UpdateTimestamp = time.Now()
	}

	err = tx.Commit()

	if err != nil {
		return nil, errors.New(err)
	}

	if len(commands) == 0 {
		return nil, nil
	}

	return &commands, nil
}

func TryGetCommandForRecovery(db *sql.DB, exchangeId int16, conId int16, statusExecutingId int16, baseLine time.Time) (*cmd.Command, error) {

	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: false})

	if err != nil {
		return nil, errors.New(err)
	}

	stmt, err := tx.Prepare(tryGetCommandForRecoverySql)

	if err != nil {
		_ = tx.Rollback()
		return nil, errors.New(err)
	}

	row := stmt.QueryRow(exchangeId, statusExecutingId, conId, baseLine)

	command, err := scanRowCommand(row, nil)

	if err != nil {
		_ = stmt.Close()
		_ = tx.Rollback()
		return nil, err
	}

	err = stmt.Close()

	if err != nil {
		_ = tx.Rollback()
		return nil, errors.New(err)
	}

	if command == nil {

		err = tx.Commit()

		if err != nil {
			return nil, errors.New(err)
		}

		return nil, nil
	}

	stmt, err = tx.Prepare(updateCommandTimestampByIdSql)

	if err != nil {
		_ = tx.Rollback()
		return nil, errors.New(err)
	}

	_, err = stmt.Exec(time.Now(), command.Id)

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

	return command, nil
}

func finishExecution(tx *sql.Tx, executionId int64, connectorId int16, currentStatusId int16, newStatusId int16, description string,
	order *cmd.Order, balances *[]cmd.Balance) error {

	stmt, err := tx.Prepare(updateCommandStatusByIdSql)

	if err != nil {
		return errors.New(err)
	}

	now := time.Now()

	_, err = stmt.Exec(newStatusId, nullInt64(int64(connectorId)), now, executionId)

	if err != nil {
		_ = stmt.Close()
		return errors.New(err)
	}

	err = stmt.Close()

	if err != nil {
		return errors.New(err)
	}

	stmt, err = tx.Prepare(insertCommandHistorySql)

	if err != nil {
		return errors.New(err)
	}

	_, err = stmt.Exec(executionId, currentStatusId, newStatusId, now, nullString(description))

	if err != nil {
		_ = stmt.Close()
		return errors.New(err)
	}

	err = stmt.Close()

	if err != nil {
		return errors.New(err)
	}

	if order != nil {

		stmt, err = tx.Prepare(insertNewOrderSql)

		if err != nil {
			return errors.New(err)
		}

		_, err = stmt.Exec(order.ExternalOrderId, order.ExecutionId, order.Price, order.Commission, order.CommissionAsset)

		if err != nil {
			_ = stmt.Close()
			return errors.New(err)
		}

		err = stmt.Close()

		if err != nil {
			return errors.New(err)
		}
	}

	if balances != nil && len(*balances) > 0 {

		stmt, err = tx.Prepare(insertNewBalanceSql)

		if err != nil {
			return errors.New(err)
		}

		for _, bal := range *balances {

			_, err = stmt.Exec(executionId, bal.Asset, bal.Free, bal.Locked)

			if err != nil {
				_ = stmt.Close()
				return errors.New(err)
			}
		}

		err = stmt.Close()

		if err != nil {
			return errors.New(err)
		}
	}

	return nil
}

func FinishExecution(db *sql.DB, executionId int64, connectorId int16, currentStatusId int16, newStatusId int16,
	description string, order *cmd.Order, balances *[]cmd.Balance) error {

	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: false})

	if err != nil {
		return errors.New(err)
	}

	err = finishExecution(tx, executionId, connectorId, currentStatusId, newStatusId, description, order, balances)

	if err != nil {
		_ = tx.Rollback()
		return err
	}

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

	command, err := scanRowCommand(row, nil)

	if err != nil {
		_ = stmt.Close()
		_ = tx.Rollback()
		return nil, err
	}

	err = stmt.Close()

	if err != nil {
		_ = tx.Rollback()
		return nil, errors.New(err)
	}

	if command == nil {

		err = tx.Commit()

		if err != nil {
			return nil, errors.New(err)
		}

		return nil, nil
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

	return command, nil
}

func LoadCommandById(db *sql.DB, id int64, statusCompletedId int16, orderTypeInfoId int16) (*cmd.Command, *cmd.Order, *[]*cmd.Balance, error) {

	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: true})

	if err != nil {
		return nil, nil, nil, errors.New(err)
	}

	stmt, err := tx.Prepare(loadCommandByIdSql)

	if err != nil {
		_ = tx.Rollback()
		return nil, nil, nil, errors.New(err)
	}

	row := stmt.QueryRow(id)

	command, err := scanRowCommand(row, nil)

	if err != nil {
		_ = stmt.Close()
		_ = tx.Rollback()
		return nil, nil, nil, err
	}

	err = stmt.Close()

	if err != nil {
		_ = tx.Rollback()
		return nil, nil, nil, errors.New(err)
	}

	var order *cmd.Order = nil
	var balances *[]*cmd.Balance = nil

	if statusCompletedId == command.StatusId {

		if orderTypeInfoId == command.OrderTypeId {

			// TODO load balances

		} else {

			stmt, err := tx.Prepare(getOrderByIdSql)

			if err != nil {
				_ = tx.Rollback()
				return nil, nil, nil, errors.New(err)
			}

			row := stmt.QueryRow(command.Id)

			order = &cmd.Order{}

			err = row.Scan(&order.Id, &order.ExternalOrderId, &order.ExecutionId, &order.Price, &order.Commission, &order.CommissionAsset)

			if err != nil {
				_ = stmt.Close()
				_ = tx.Rollback()
				return nil, nil, nil, err
			}

			err = stmt.Close()

			if err != nil {
				_ = tx.Rollback()
				return nil, nil, nil, errors.New(err)
			}
		}
	}

	err = tx.Commit()

	if err != nil {
		return nil, nil, nil, errors.New(err)
	}

	return command, order, balances, nil
}

func nullString(s string) sql.NullString {

	if len(s) == 0 {
		return sql.NullString{Valid: false}
	}

	return sql.NullString{String: s, Valid: true}
}

func nullFloat(value float64) sql.NullFloat64 {

	if value < 0 {
		return sql.NullFloat64{Valid: false}
	}

	return sql.NullFloat64{Float64: value, Valid: true}
}

func nullInt64(value int64) sql.NullInt64 {

	if value < 0 {
		return sql.NullInt64{Valid: false}
	}

	return sql.NullInt64{Int64: value, Valid: true}
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

	row := stmt.QueryRow(exchangeId, instrument, directionId, orderTypeId, nullFloat(limitPrice), timeInForceId, amount, statusId,
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
