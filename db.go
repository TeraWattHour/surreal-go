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

// Unset removes an identifier from the current session.
func (db *DB) Unset(identifier string) error {
	_, err := db.conn.Send("unset", []any{identifier})
	return err
}

func (db *DB) SignIn(args AuthArgs) error {
	_, err := db.conn.Send("signin", []any{args})
	return err
}

func (db *DB) SignUp(args AuthArgs) error {
	_, err := db.conn.Send("signup", []any{args})
	return err
}

func (db *DB) Authenticate(token string) error {
	_, err := db.conn.Send("authenticate", []any{token})
	return err
}

func (db *DB) Invalidate() error {
	_, err := db.conn.Send("invalidate", nil)
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

// Create creates a record in a table, then decodes the row into the destination, if provided.
// Destination may be either a pointer to a slice or a pointer to a single record (struct, map).
func (db *DB) Create(table string, data any, destination ...any) error {
	raw, err := db.conn.Send("create", []any{table, data})
	if err != nil {
		return err
	}

	if len(destination) != 0 {
		return autoScan(raw, destination[0])
	}

	return nil
}

// Insert inserts a record, or multiple records, into a table, then decodes the rows into the destination, if provided.
// Destination may be either a pointer to a slice or a pointer to a single record (struct, map).
func (db *DB) Insert(table string, data any, destination ...any) error {
	raw, err := db.conn.Send("insert", []any{table, data})
	if err != nil {
		return err
	}

	if len(destination) != 0 {
		return autoScan(raw, destination[0])
	}

	return nil
}

func (db *DB) Relate(from any, thing string, to any, data any, destination ...any) error {
	raw, err := db.conn.Send("relate", []any{from, thing, to, data})
	if err != nil {
		return err
	}

	if len(destination) != 0 {
		return autoScan(raw, destination[0])
	}

	return nil
}

func (db *DB) Update(id string, data any, destination ...any) error {
	raw, err := db.conn.Send("update", []any{id, data})
	if err != nil {
		return err
	}

	if len(destination) != 0 {
		return autoScan(raw, destination[0])
	}

	return nil
}

func (db *DB) Upsert(id string, data any, destination ...any) error {
	raw, err := db.conn.Send("upsert", []any{id, data})
	if err != nil {
		return err
	}

	if len(destination) != 0 {
		return autoScan(raw, destination[0])
	}

	return nil
}

func (db *DB) Live(id string, callback func(notification rpc.LiveNotification), diff bool) (string, error) {
	raw, err := db.conn.Send("live", []any{id, diff})
	if err != nil {
		return "", err
	}

	if raw[0] == '"' {
		id := string(raw[1 : len(raw)-1])

		db.conn.RegisterLiveCallback(id, callback)

		return id, nil
	}

	return "", fmt.Errorf("failed to start live query")
}

func (db *DB) Kill(id string) error {
	_, err := db.conn.Send("kill", []any{id})
	return err
}

func (db *DB) Patch(id string, diff []Diff, destination ...any) error {
	raw, err := db.conn.Send("patch", []any{id, diff})
	if err != nil {
		return err
	}

	if len(destination) != 0 {
		return autoScan(raw, destination[0])
	}

	return nil
}

func (db *DB) Merge(id string, data any, destination ...any) error {
	raw, err := db.conn.Send("merge", []any{id, data})
	if err != nil {
		return err
	}

	if len(destination) != 0 {
		return autoScan(raw, destination[0])
	}

	return nil
}

// Delete deletes a record, or all records, from a table, then decodes the rows into the destination, if provided.
func (db *DB) Delete(id string, destination ...any) error {
	raw, err := db.conn.Send("delete", []any{id})
	if err != nil {
		return err
	}

	if len(destination) != 0 {
		return autoScan(raw, destination[0])
	}

	return nil
}

// Info retrieves information about the current scope(!) user.
func (db *DB) Info(destination any) error {
	raw, err := db.conn.Send("info", []any{})
	if err != nil {
		return err
	}

	return json.Unmarshal(raw, destination)
}

func (db *DB) Ping() error {
	_, err := db.conn.Send("ping", []any{})
	return err
}

// Version retrieves the version of the database.
func (db *DB) Version() (string, error) {
	raw, err := db.conn.Send("version", []any{})
	return string(raw[1 : len(raw)-1]), err
}

// Close closes the connection to the database.
func (db *DB) Close() error {
	return db.conn.Close()
}

type Diff struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value any    `json:"value"`
}

type AuthArgs struct {
	Namespace string `json:"NS"`
	Database  string `json:"DB"`
	Scope     string `json:"SC,omitempty"`
	Access    string `json:"AC,omitempty"`
	Other     Map    `json:"-"`
}

func (s AuthArgs) MarshalJSON() ([]byte, error) {
	s.Other["NS"] = s.Namespace
	s.Other["DB"] = s.Database
	if s.Scope != "" {
		s.Other["SC"] = s.Scope
	} else if s.Access != "" {
		s.Other["AC"] = s.Access
	}

	return json.Marshal(s.Other)
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
