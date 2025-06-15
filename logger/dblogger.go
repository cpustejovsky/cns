package logger

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/lib/pq"
)

type PostgresTransactionLogger struct {
	events chan<- Event
	errors <-chan error
	wg     *sync.WaitGroup
	db     *sql.DB
}

type PostgresDbParams struct {
	Host     string
	DBName   string
	User     string
	Password string
}

const createQuery = `create table transactions (
		sequence      bigserial primary key,
		event_type    smallint,
		key	      text,
		value         text
	  );`

const insert = `INSERT INTO transactions
			(event_type, key, value)
			VALUES ($1, $2, $3);`

const query = "SELECT sequence, event_type, key, value FROM transactions"

func NewPostgresTransactionLogger(param PostgresDbParams) (*PostgresTransactionLogger, error) {
	connStr := fmt.Sprintf("host=%s dbname=%s user=%s password=%s sslmode=disable",
		param.Host, param.DBName, param.User, param.Password)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create db value with connection string '%s': %w", connStr, err)
	}

	err = db.Ping() // Test the databases connection
	if err != nil {
		return nil, fmt.Errorf("failed to opendb connection: %w", err)
	}

	ptl := &PostgresTransactionLogger{db: db, wg: &sync.WaitGroup{}}

	if err = ptl.createTable(); err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}
	return ptl, nil
}

func (l *PostgresTransactionLogger) createTable() error {
	_, err := l.db.Exec(createQuery)
	return err

}
func (l *PostgresTransactionLogger) WriteDelete(key string) {
	l.events <- Event{
		EventType: EventDelete,
		Key:       key,
	}
}
func (l *PostgresTransactionLogger) WritePut(key, value string) {
	l.events <- Event{
		EventType: EventPut,
		Key:       key,
		Value:     value,
	}
}
func (l *PostgresTransactionLogger) Err() <-chan error {
	return l.errors
}
func (l *PostgresTransactionLogger) ReadEvents() (<-chan Event, <-chan error) {
	outEvent := make(chan Event)
	outError := make(chan error, 1)

	go func() {
		var e Event

		defer close(outEvent)
		defer close(outError)

		rows, err := l.db.Query(query)
		if err != nil {
			outError <- fmt.Errorf("sql query error: %w", err)
		}
		defer rows.Close()
		for rows.Next() {

			err = rows.Scan(&e.Sequence, &e.EventType, &e.Key, &e.Value)
			if err != nil {
				outError <- fmt.Errorf("err reading row: %w", err)
				return
			}
			outEvent <- e
		}

		if err := rows.Err(); err != nil {
			outError <- fmt.Errorf("transaction log read failure: %w", err)
		}
	}()

	return outEvent, outError
}
func (l *PostgresTransactionLogger) Run() {
	//TODO: make sure this is hiding deadlocks
	events := make(chan Event, 16)
	l.events = events

	errors := make(chan error, 1)
	l.errors = errors

	go func() {
		for e := range events {
			_, err := l.db.Exec(insert, e.EventType, e.Key, e.Value)
			if err != nil {
				errors <- err
			}
			l.wg.Done()
		}
	}()
}
