package surreal

import (
	"encoding/json"
	"fmt"
	"github.com/terawatthour/surreal-go/rpc"
)

type DB struct {
	conn    Connection
	options *Options
}

type QueryError struct {
	QueryNo int
	Message string
}

type QueryErrors []QueryError

func (q QueryErrors) Error() string {
	var s string
	for _, e := range q {
		s += fmt.Sprintf("query %d failed with error: `%s`; ", e.QueryNo, e.Message)
	}
	return s
}

type Map map[string]any

// Use sets the namespace and database name for the current connection. Should be called after the connection is
// established, but before any queries are sent.
func (db *DB) Use(namespace, databaseName string) error {
	_, err := db.conn.Send("use", []any{namespace, databaseName})
	return err
}

// Let binds an identifier to a value. The value may be used in subsequent queries.
func (db *DB) Let(identifier string, value any) error {
	_, err := db.conn.Send("let", []any{identifier, value})
	return err
}

// Query sends a query (or multiple semicolon separated queries) to the database and returns the result (results).
// The result is decoded into the scanDestinations. If there are multiple queries, the results are decoded into the
// corresponding scanDestinations. `vars` is a map of variables that are used to bind the query (or queries).
func (db *DB) Query(query string, vars Map, scanDestinations ...any) error {
	raw, err := db.conn.Send("query", []any{query, vars})
	if err != nil {
		return err
	}

	var rawQueryResult rpc.RawResult
	if err := json.Unmarshal(raw, &rawQueryResult); err != nil {
		return fmt.Errorf("failed to decode result: %s", err)
	}

	var errors QueryErrors
	for i, row := range rawQueryResult {
		if !row.OK {
			errors = append(errors, QueryError{i, string(row.Result)})
		}
	}

	if len(errors) > 0 {
		return errors
	}

	for i := 0; i < len(scanDestinations) && i < len(rawQueryResult); i++ {
		if err := json.Unmarshal(rawQueryResult[i].Result, scanDestinations[i]); err != nil {
			return fmt.Errorf("failed to decode result of %d query: %s", i, err)
		}
	}

	return nil
}

// Select performs a select query and decodes the results into the destination. May target a single record or all
// records in a table. Returns error if id is not a table name and there is no row found.
func (db *DB) Select(id string, destination any) error {
	raw, err := db.conn.Send("select", []any{id})
	if err != nil {
		return err
	}

	if string(raw) == "null" {
		return fmt.Errorf("record not found")
	}

	if err := json.Unmarshal(raw, destination); err != nil {
		return fmt.Errorf("failed to decode result: %s", err)
	}

	return nil
}

// Delete deletes a record, or all records, from a table, then decodes the rows into the destination, if provided.
// Panics if more than one destination is provided.
func (db *DB) Delete(id string, destination ...any) error {
	raw, err := db.conn.Send("delete", []any{id})
	if err != nil {
		return err
	}

	if len(destination) > 1 {
		panic("expected at most 1 destination")
	}

	if len(destination) > 0 {
		if err := json.Unmarshal(raw, destination[0]); err != nil {
			return fmt.Errorf("failed to decode result: %s", err)
		}
	}

	return nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}
