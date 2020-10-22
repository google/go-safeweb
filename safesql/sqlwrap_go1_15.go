// +build go1.15

package safesql

import "time"

// File to hold stuff that is only available in Go 1.15+

// SetConnMaxIdleTime is a tiny wrapper for https://pkg.go.dev/sql#DB.SetConnMaxIdleTime
func (db DB) SetConnMaxIdleTime(d time.Duration) {
  db.db.SetConnMaxIdleTime(d)
}
