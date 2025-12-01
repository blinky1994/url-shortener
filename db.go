package main

import (
	"database/sql"

	// We use a pure-Go SQLite driver so the project builds on Windows
	// without requiring a separate C toolchain (cgo). The driver package
	// registers itself with database/sql when imported for side effects.
	_ "github.com/glebarez/sqlite"
)

// DB is the global database handle used throughout the application.
// It's initialized in the init() function below so handlers can use it.
var DB *sql.DB

// init runs automatically before main() and prepares the database.
// Tasks performed here:
// - open (or create) the SQLite database file
// - ensure the `links` table exists with the expected schema
func init() {
	var err error

	// glebarez/sqlite registers the driver name "sqlite". The DSN uses
	// `file:...` URI form and sets a couple of pragmas:
	// - busy_timeout=5000: wait up to 5s if the DB is locked
	// - journal_mode=WAL: use write-ahead logging for better concurrency
	DB, err = sql.Open("sqlite", "file:shortener.db?_pragma=busy_timeout=5000&_pragma=journal_mode=WAL")
	if err != nil {
		// On failure to open the DB, panic so the program doesn't run in a
		// broken state. For production code you'd want more graceful
		// handling and logging.
		panic(err)
	}

	// Create the `links` table if it doesn't already exist. Fields:
	// - id: the short identifier (TEXT) and primary key
	// - target: the original URL (TEXT)
	// - clicks: simple integer counter for number of redirects
	// - created_at: timestamp of when the row was created
	_, err = DB.Exec(`CREATE TABLE IF NOT EXISTS links (
		id TEXT PRIMARY KEY,
		target TEXT NOT NULL,
		clicks INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		// If the schema creation fails, panic as above.
		panic(err)
	}
}
