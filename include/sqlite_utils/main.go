package sqlite_utils

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	_ "modernc.org/sqlite"
	"strings"
)

type TableRow struct {
	Target                   string
	Vhost                    string
	BaselineRequestBodyMD5   string
	SpoofedRequestBodyMD5    string
	SpoofedRequestStatusCode int
}

func QuoteString(s string) string {
	// simple SQL-escaping of single-quotes
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

func Open_database_interface(DatabaseDirectory string) (*sql.DB, error) {
	DatabaseInterface, DatabaseInterfaceError := sql.Open("sqlite", DatabaseDirectory)
	return DatabaseInterface, DatabaseInterfaceError
}

func AddRowToTable(DatabaseInterface *sql.DB, TableName string, TableRow TableRow) error {

	TableExistAndCreateQuery := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS enumerated_vhosts(
	    target TEXT NOT NULL,
		vhost TEXT NOT NULL,
		baseline_response_body_md5 TEXT NOT NULL,
		spoofed_response_body_md5 TEXT NOT NULL,
		spoofed_request_status_code INT NOT NULL
	);`)
	_, db_table_err := DatabaseInterface.Exec(TableExistAndCreateQuery)
	if db_table_err != nil {
		return errors.New("Error creating database table")
	}

	AddRowQuery := fmt.Sprintf(
		"INSERT INTO %s("+
			"target, vhost, baseline_response_body_md5, spoofed_response_body_md5, spoofed_request_status_code"+
			") VALUES (%s, %s, %s, %s, %d);",
		TableName,
		QuoteString(TableRow.Target),
		QuoteString(TableRow.Vhost),
		QuoteString(TableRow.BaselineRequestBodyMD5),
		QuoteString(TableRow.SpoofedRequestBodyMD5),
		TableRow.SpoofedRequestStatusCode,
	)
	//fmt.Println(AddRowQuery) // DEBUG
	_, db_row_err := DatabaseInterface.Exec(AddRowQuery)
	if db_row_err != nil {
		return errors.New("An error occurred while adding row using query: " + AddRowQuery + "\n" + db_row_err.Error())
	}
	return nil
}

func Close_database_interface(DatabaseInterface *sql.DB) {
	Error := DatabaseInterface.Close()
	if Error != nil {
		log.Panic(Error)
	}
}
