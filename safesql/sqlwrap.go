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

// Drivers is a tiny wrapper for https://pkg.go.dev/sql#Drivers
func Drivers() []string { return sql.Drivers() }

// Register is a tiny wrapper for https://pkg.go.dev/sql#Register
func Register(name string, driver driver.Driver) { sql.Register(name, driver) }

// ColumnType is a tiny wrapper for https://pkg.go.dev/sql#ColumnType
type ColumnType = sql.ColumnType

// DBStats is a tiny wrapper for https://pkg.go.dev/sql#DBStats
type DBStats = sql.DBStats

// IsolationLevel is a tiny wrapper for https://pkg.go.dev/sql#IsolationLevel
type IsolationLevel = sql.IsolationLevel

// NamedArg is a tiny wrapper for https://pkg.go.dev/sql#NamedArg
type NamedArg = sql.NamedArg

// NullBool is a tiny wrapper for https://pkg.go.dev/sql#NullBool
type NullBool = sql.NullBool

// NullFloat64 is a tiny wrapper for https://pkg.go.dev/sql#NullFloat64
type NullFloat64 = sql.NullFloat64

// NullInt32 is a tiny wrapper for https://pkg.go.dev/sql#NullInt32
type NullInt32 = sql.NullInt32

// NullInt64 is a tiny wrapper for https://pkg.go.dev/sql#NullInt64
type NullInt64 = sql.NullInt64

// NullString is a tiny wrapper for https://pkg.go.dev/sql#NullString
type NullString = sql.NullString

// NullTime is a tiny wrapper for https://pkg.go.dev/sql#NullTime
type NullTime = sql.NullTime

// Out is a tiny wrapper for https://pkg.go.dev/sql#Out
type Out = sql.Out

// RawBytes is a tiny wrapper for https://pkg.go.dev/sql#RawBytes
type RawBytes = sql.RawBytes

// Result is a tiny wrapper for https://pkg.go.dev/sql#Result
type Result = sql.Result

// Row is a tiny wrapper for https://pkg.go.dev/sql#Row
type Row = sql.Row

// Rows is a tiny wrapper for https://pkg.go.dev/sql#Rows
type Rows = sql.Rows

// Scanner is a tiny wrapper for https://pkg.go.dev/sql#Scanner
type Scanner = sql.Scanner

// Stmt is a tiny wrapper for https://pkg.go.dev/sql#Stmt
type Stmt = sql.Stmt

// TxOptions is a tiny wrapper for https://pkg.go.dev/sql#TxOptions
type TxOptions = sql.TxOptions

// Conn behaves as the standard SQL package one, with the exception that it does not implement the `Raw` method for security reasons.
// Please see https://pkg.go.dev/sql#Conn
type Conn struct {
	c *sql.Conn
}

// Begin is a tiny wrapper for https://pkg.go.dev/sql#Conn.Begin
func (c Conn) BeginTx(ctx context.Context, opts *TxOptions) (Tx, error) {
	t, err := c.c.BeginTx(ctx, opts)
	return Tx{t}, err
}

// Close is a tiny wrapper for https://pkg.go.dev/sql#Conn.Close
func (c Conn) Close() error {
	return c.c.Close()
}

// ExecContext is a tiny wrapper for https://pkg.go.dev/sql#Conn.ExecContext
func (c Conn) ExecContext(ctx context.Context, query TrustedSQLString, args ...interface{}) (Result, error) {
	return c.c.ExecContext(ctx, query.s, args...)
}

// PingContext is a tiny wrapper for https://pkg.go.dev/sql#Conn.PingContext
func (c Conn) PingContext(ctx context.Context) error {
	return c.c.PingContext(ctx)
}

// PrepareContext is a tiny wrapper for https://pkg.go.dev/sql#Conn.PrepareContext
func (c Conn) PrepareContext(ctx context.Context, query TrustedSQLString) (*Stmt, error) {
	return c.c.PrepareContext(ctx, query.s)
}

// QueryContext is a tiny wrapper for https://pkg.go.dev/sql#Conn.QueryContext
func (c Conn) QueryContext(ctx context.Context, query TrustedSQLString, args ...interface{}) (*Rows, error) {
	return c.c.QueryContext(ctx, query.s, args...)
}

// QueryRowContext is a tiny wrapper for https://pkg.go.dev/sql#Conn.QueryRowContext
func (c Conn) QueryRowContext(ctx context.Context, query TrustedSQLString, args ...interface{}) *Row {
	return c.c.QueryRowContext(ctx, query.s, args...)
}

// DB behaves as the standard SQL package one, with the exception that it does not implement the `Driver` method for security reasons.
// Please see https://pkg.go.dev/sql#DB
type DB struct {
	db *sql.DB
}

// Open is a tiny wrapper for https://pkg.go.dev/sql#Open
func Open(driverName, dataSourceName string) (DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	return DB{db}, err
}

// OpenDB is a tiny wrapper for https://pkg.go.dev/sql#OpenDB
func OpenDB(c driver.Connector) DB { return DB{sql.OpenDB(c)} }

// Begin is a tiny wrapper for https://pkg.go.dev/sql#DB.Begin
func (db DB) Begin() (Tx, error) {
	t, err := db.db.Begin()
	return Tx{t}, err
}

// BeginTx is a tiny wrapper for https://pkg.go.dev/sql#DB.BeginTx
func (db DB) BeginTx(ctx context.Context, opts *TxOptions) (Tx, error) {
	t, err := db.db.BeginTx(ctx, opts)
	return Tx{t}, err
}

// Close is a tiny wrapper for https://pkg.go.dev/sql#DB.Close
func (db DB) Close() error {
	return db.db.Close()
}

// Conn is a tiny wrapper for https://pkg.go.dev/sql#DB.Conn
func (db DB) Conn(ctx context.Context) (Conn, error) {
	c, err := db.db.Conn(ctx)
	return Conn{c}, err
}

// Exec is a tiny wrapper for https://pkg.go.dev/sql#DB.Exec
func (db DB) Exec(query TrustedSQLString, args ...interface{}) (Result, error) {
	return db.db.Exec(query.s, args...)
}

// ExecContext is a tiny wrapper for https://pkg.go.dev/sql#DB.ExecContext
func (db DB) ExecContext(ctx context.Context, query TrustedSQLString, args ...interface{}) (Result, error) {
	return db.db.ExecContext(ctx, query.s, args...)
}

// Ping is a tiny wrapper for https://pkg.go.dev/sql#DB.Ping
func (db DB) Ping() error {
	return db.db.Ping()
}

// PingContext is a tiny wrapper for https://pkg.go.dev/sql#DB.PingContext
func (db DB) PingContext(ctx context.Context) error {
	return db.db.PingContext(ctx)
}

// Prepare is a tiny wrapper for https://pkg.go.dev/sql#DB.Prepare
func (db DB) Prepare(query TrustedSQLString) (*Stmt, error) {
	return db.db.Prepare(query.s)
}

// PrepareContext is a tiny wrapper for https://pkg.go.dev/sql#DB.PrepareContext
func (db DB) PrepareContext(ctx context.Context, query TrustedSQLString) (*Stmt, error) {
	return db.db.PrepareContext(ctx, query.s)
}

// Query is a tiny wrapper for https://pkg.go.dev/sql#DB.Query
func (db DB) Query(query TrustedSQLString, args ...interface{}) (*Rows, error) {
	return db.db.Query(query.s, args...)
}

// QueryContext is a tiny wrapper for https://pkg.go.dev/sql#DB.QueryContext
func (db DB) QueryContext(ctx context.Context, query TrustedSQLString, args ...interface{}) (*Rows, error) {
	return db.db.QueryContext(ctx, query.s, args...)
}

// QueryRow is a tiny wrapper for https://pkg.go.dev/sql#DB.QueryRow
func (db DB) QueryRow(query TrustedSQLString, args ...interface{}) *Row {
	return db.db.QueryRow(query.s, args...)
}

// QueryRowContext is a tiny wrapper for https://pkg.go.dev/sql#DB.QueryRowContext
func (db DB) QueryRowContext(ctx context.Context, query TrustedSQLString, args ...interface{}) *Row {
	return db.db.QueryRowContext(ctx, query.s, args...)
}

// SetConnMaxLifetime is a tiny wrapper for https://pkg.go.dev/sql#DB.SetConnMaxLifetime
func (db DB) SetConnMaxLifetime(d time.Duration) {
	db.db.SetConnMaxLifetime(d)
}

// SetMaxIdleConns is a tiny wrapper for https://pkg.go.dev/sql#DB.SetMaxIdleConns
func (db DB) SetMaxIdleConns(n int) {
	db.db.SetMaxIdleConns(n)
}

// SetMaxOpenConns is a tiny wrapper for https://pkg.go.dev/sql#DB.SetMaxOpenConns
func (db DB) SetMaxOpenConns(n int) {
	db.db.SetMaxOpenConns(n)
}

// Stats is a tiny wrapper for https://pkg.go.dev/sql#DB.Stats
func (db DB) Stats() DBStats {
	return db.db.Stats()
}

// Tx is a tiny wrapper for https://pkg.go.dev/sql#Tx
type Tx struct {
	tx *sql.Tx
}

// Commit is a tiny wrapper for https://pkg.go.dev/sql#Tx.Commit
func (tx Tx) Commit() error { return tx.tx.Commit() }

// Exec is a tiny wrapper for https://pkg.go.dev/sql#Tx.Exec
func (tx Tx) Exec(query TrustedSQLString, args ...interface{}) (Result, error) {
	return tx.tx.Exec(query.s, args...)
}

// ExecContext is a tiny wrapper for https://pkg.go.dev/sql#Tx.ExecContext
func (tx Tx) ExecContext(ctx context.Context, query TrustedSQLString, args ...interface{}) (Result, error) {
	return tx.tx.ExecContext(ctx, query.s, args...)
}

// Prepare is a tiny wrapper for https://pkg.go.dev/sql#Tx.Prepare
func (tx Tx) Prepare(query TrustedSQLString) (*Stmt, error) { return tx.tx.Prepare(query.s) }

// PrepareContext is a tiny wrapper for https://pkg.go.dev/sql#Tx.PrepareContext
func (tx Tx) PrepareContext(ctx context.Context, query TrustedSQLString) (*Stmt, error) {
	return tx.tx.PrepareContext(ctx, query.s)
}

// Query is a tiny wrapper for https://pkg.go.dev/sql#Tx.Query
func (tx Tx) Query(query TrustedSQLString, args ...interface{}) (*Rows, error) {
	return tx.tx.Query(query.s, args...)
}

// QueryContext is a tiny wrapper for https://pkg.go.dev/sql#Tx.QueryContext
func (tx Tx) QueryContext(ctx context.Context, query TrustedSQLString, args ...interface{}) (*Rows, error) {
	return tx.tx.QueryContext(ctx, query.s, args...)
}

// QueryRow is a tiny wrapper for https://pkg.go.dev/sql#Tx.QueryRow
func (tx Tx) QueryRow(query TrustedSQLString, args ...interface{}) *Row {
	return tx.tx.QueryRow(query.s, args...)
}

// QueryRowContext is a tiny wrapper for https://pkg.go.dev/sql#Tx.QueryRowContext
func (tx Tx) QueryRowContext(ctx context.Context, query TrustedSQLString, args ...interface{}) *Row {
	return tx.tx.QueryRowContext(ctx, query.s, args...)
}

// Rollback is a tiny wrapper for https://pkg.go.dev/sql#Tx.Rollback
func (tx Tx) Rollback() error {
	return tx.tx.Rollback()
}

// Stmt is a tiny wrapper for https://pkg.go.dev/sql#Tx.Stmt
func (tx Tx) Stmt(stmt *Stmt) *Stmt {
	return tx.tx.Stmt(stmt)
}

// StmtContext is a tiny wrapper for https://pkg.go.dev/sql#Tx.StmtContext
func (tx Tx) StmtContext(ctx context.Context, stmt *Stmt) *Stmt {
	return tx.tx.StmtContext(ctx, stmt)
}
