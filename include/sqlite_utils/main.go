package sqlite_utils

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	_ "modernc.org/sqlite"
	"strings"
)

type TableColumn struct {
	Type string
	Name string
}

type TableRowValue struct {
	Value string
	Type  string
}

type DatabaseTableStruct struct {
	Name    string
	Columns []string
}

type TableRow struct {
	Target                 string
	Vhost                  string
	BaselineRequestBodyMD5 string
	SpoofedRequestBodyMD5  string
}

func QuoteString(s string) string {
	// simple SQL-escaping of single-quotes
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

func InitializeDatabaseInterface(DatabaseDirectory string) (*sql.DB, error) {
	DatabaseInterface, DatabaseInterfaceError := sql.Open("sqlite", DatabaseDirectory)
	return DatabaseInterface, DatabaseInterfaceError
}

func AddRowToTable(DatabaseInterface *sql.DB, TableName string, TableRow TableRow) error {

	TableExistAndCreateQuery := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS enumerated_vhosts(
	    target TEXT NOT NULL,
		vhost TEXT NOT NULL,
		baseline_request_body_md5 TEXT NOT NULL,
		spoofed_request_body_md5 TEXT NOT NULL,
	);`)
	DatabaseInterface.Exec(TableExistAndCreateQuery)

	AddRowQuery := fmt.Sprintf(
		"INSERT INTO %s("+
			"target, vhost, baseline_request_body_md5, spoofed_request_body_md5"+
			") VALUES (%s, %s, %s, %s);",
		TableName,
		QuoteString(TableRow.Target),
		QuoteString(TableRow.Vhost),
		QuoteString(TableRow.BaselineRequestBodyMD5),
		QuoteString(TableRow.SpoofedRequestBodyMD5),
	)
	//fmt.Println(AddRowQuery) // DEBUG
	_, err := DatabaseInterface.Exec(AddRowQuery)
	if err != nil {
		return errors.New("An error occurred while adding row using query: " + AddRowQuery + "\n" + err.Error())
	}
	return nil
}

func CloseDatabaseInterface(DatabaseInterface *sql.DB) {
	Error := DatabaseInterface.Close()
	if Error != nil {
		log.Panic(Error)
	}
}
