package sqlite_utils

import (
	"database/sql"
	"errors"
	"fmt"
	_ "modernc.org/sqlite"
	"strings"
)

type Table_row struct {
	Target                      string
	Vhost                       string
	Baseline_response_body_md5  string
	Spoofed_response_body_md5   string
	Spoofed_request_status_code int
}

func QuoteString(s string) string {
	// simple SQL-escaping of single-quotes
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

func Open_database_interface(database_directory string) (*sql.DB, error) {
	database_interface, database_interfaceError := sql.Open("sqlite", database_directory)
	return database_interface, database_interfaceError
}

func AddRowToTable(database_interface *sql.DB, table_name string, table_row Table_row) error {

	TableExistAndCreateQuery := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS enumerated_vhosts(
	    target TEXT NOT NULL,
		vhost TEXT NOT NULL,
		baseline_response_body_md5 TEXT NOT NULL,
		spoofed_response_body_md5 TEXT NOT NULL,
		spoofed_request_status_code INT NOT NULL
	);`)
	_, db_table_err := database_interface.Exec(TableExistAndCreateQuery)
	if db_table_err != nil {
		return errors.New("An error occurred while creating the database table || Error: " + db_table_err.Error())
	}

	add_row_query := fmt.Sprintf(
		"INSERT INTO %s("+
			"target, vhost, baseline_response_body_md5, spoofed_response_body_md5, spoofed_request_status_code"+
			") VALUES (%s, %s, %s, %s, %d);",
		table_name,
		QuoteString(table_row.Target),
		QuoteString(table_row.Vhost),
		QuoteString(table_row.Baseline_response_body_md5),
		QuoteString(table_row.Spoofed_response_body_md5),
		table_row.Spoofed_request_status_code,
	)
	//fmt.Println(add_row_query) // DEBUG
	_, db_row_err := database_interface.Exec(add_row_query)
	if db_row_err != nil {
		return errors.New(fmt.Sprintf("An error occurred while adding row using query: %s || Error: %s", add_row_query, db_row_err.Error()))
	}
	return nil
}

func Close_database_interface(database_interface *sql.DB) error {
	close_db_err := database_interface.Close()
	if close_db_err != nil {
		return errors.New("An error occurred while closing the database interface: " + close_db_err.Error())
	}
	return nil
}
