// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package safesql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"time"
)

func Drivers() []string { return sql.Drivers() }

func Register(name string, driver driver.Driver) { sql.Register(name, driver) }

type ColumnType = sql.ColumnType
type DBStats = sql.DBStats
type IsolationLevel = sql.IsolationLevel
type NamedArg = sql.NamedArg
type NullBool = sql.NullBool
type NullFloat64 = sql.NullFloat64
type NullInt32 = sql.NullInt32
type NullInt64 = sql.NullInt64
type NullString = sql.NullString
type NullTime = sql.NullTime
type Out = sql.Out
type RawBytes = sql.RawBytes
type Result = sql.Result
type Row = sql.Row
type Rows = sql.Rows
type Scanner = sql.Scanner
type Stmt = sql.Stmt
type TxOptions = sql.TxOptions

// Conn behaves as the standard SQL package one, with the exception that it does not implement the `Raw` method for security reasons.
type Conn struct {
	c *sql.Conn
}

/*
func (c *Conn) Raw(f func(driverConn interface{}) error) (err error) {
	// Dangerous to expose
}
*/
func (c Conn) BeginTx(ctx context.Context, opts *TxOptions) (Tx, error) {
	t, err := c.c.BeginTx(ctx, opts)
	return Tx{t}, err
}
func (c Conn) Close() error { return c.c.Close() }
func (c Conn) ExecContext(ctx context.Context, query TrustedSQLString, args ...interface{}) (Result, error) {
	return c.c.ExecContext(ctx, query.s, args)
}
func (c Conn) PingContext(ctx context.Context) error { return c.c.PingContext(ctx) }
func (c Conn) PrepareContext(ctx context.Context, query TrustedSQLString) (*Stmt, error) {
	return c.c.PrepareContext(ctx, query.s)
}
func (c Conn) QueryContext(ctx context.Context, query TrustedSQLString, args ...interface{}) (*Rows, error) {
	return c.c.QueryContext(ctx, query.s, args)
}
func (c Conn) QueryRowContext(ctx context.Context, query TrustedSQLString, args ...interface{}) *Row {
	return c.c.QueryRowContext(ctx, query.s, args)
}

// DB behaves as the standard SQL package one, with the exception that it does not implement the `Driver` method for security reasons.
type DB struct {
	db *sql.DB
}

func Open(driverName, dataSourceName string) (DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	return DB{db}, err
}
func OpenDB(c driver.Connector) DB { return DB{sql.OpenDB(c)} }
func (db DB) Begin() (Tx, error) {
	t, err := db.db.Begin()
	return Tx{t}, err
}
func (db DB) BeginTx(ctx context.Context, opts *TxOptions) (Tx, error) {
	t, err := db.db.BeginTx(ctx, opts)
	return Tx{t}, err
}
func (db DB) Close() error { return db.db.Close() }
func (db DB) Conn(ctx context.Context) (Conn, error) {
	c, err := db.db.Conn(ctx)
	return Conn{c}, err
}
func (db DB) Exec(query TrustedSQLString, args ...interface{}) (Result, error) {
	return db.db.Exec(query.s, args)
}
func (db DB) ExecContext(ctx context.Context, query TrustedSQLString, args ...interface{}) (Result, error) {
	return db.db.ExecContext(ctx, query.s, args)
}
func (db DB) Ping() error                                   { return db.db.Ping() }
func (db DB) PingContext(ctx context.Context) error         { return db.db.PingContext(ctx) }
func (db DB) Prepare(query TrustedSQLString) (*Stmt, error) { return db.db.Prepare(query.s) }
func (db DB) PrepareContext(ctx context.Context, query TrustedSQLString) (*Stmt, error) {
	return db.db.PrepareContext(ctx, query.s)
}
func (db DB) Query(query TrustedSQLString, args ...interface{}) (*Rows, error) {
	return db.db.Query(query.s, args)
}
func (db DB) QueryContext(ctx context.Context, query TrustedSQLString, args ...interface{}) (*Rows, error) {
	return db.db.QueryContext(ctx, query.s, args)
}
func (db DB) QueryRow(query TrustedSQLString, args ...interface{}) *Row {
	return db.db.QueryRow(query.s, args)
}
func (db DB) QueryRowContext(ctx context.Context, query TrustedSQLString, args ...interface{}) *Row {
	return db.db.QueryRowContext(ctx, query.s, args)
}
func (db DB) SetConnMaxIdleTime(d time.Duration) { db.db.SetConnMaxIdleTime(d) }
func (db DB) SetConnMaxLifetime(d time.Duration) { db.db.SetConnMaxLifetime(d) }
func (db DB) SetMaxIdleConns(n int)              { db.db.SetMaxIdleConns(n) }
func (db DB) SetMaxOpenConns(n int)              { db.db.SetMaxOpenConns(n) }
func (db DB) Stats() DBStats                     { return db.db.Stats() }

type Tx struct {
	tx *sql.Tx
}

func (tx Tx) Commit() error { return tx.tx.Commit() }
func (tx Tx) Exec(query TrustedSQLString, args ...interface{}) (Result, error) {
	return tx.tx.Exec(query.s, args)
}
func (tx Tx) ExecContext(ctx context.Context, query TrustedSQLString, args ...interface{}) (Result, error) {
	return tx.tx.ExecContext(ctx, query.s, args)
}
func (tx Tx) Prepare(query TrustedSQLString) (*Stmt, error) { return tx.tx.Prepare(query.s) }
func (tx Tx) PrepareContext(ctx context.Context, query TrustedSQLString) (*Stmt, error) {
	return tx.tx.PrepareContext(ctx, query.s)
}
func (tx Tx) Query(query TrustedSQLString, args ...interface{}) (*Rows, error) {
	return tx.tx.Query(query.s, args)
}
func (tx Tx) QueryContext(ctx context.Context, query TrustedSQLString, args ...interface{}) (*Rows, error) {
	return tx.tx.QueryContext(ctx, query.s, args)
}
func (tx Tx) QueryRow(query TrustedSQLString, args ...interface{}) *Row {
	return tx.tx.QueryRow(query.s, args)
}
func (tx Tx) QueryRowContext(ctx context.Context, query TrustedSQLString, args ...interface{}) *Row {
	return tx.tx.QueryRowContext(ctx, query.s, args)
}
func (tx Tx) Rollback() error                                   { return tx.tx.Rollback() }
func (tx Tx) Stmt(stmt *Stmt) *Stmt                             { return tx.tx.Stmt(stmt) }
func (tx Tx) StmtContext(ctx context.Context, stmt *Stmt) *Stmt { return tx.tx.StmtContext(ctx, stmt) }
