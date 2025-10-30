package file_utils

import (
	"bufio"
	"errors"
	"os"
)

// Read_lines reads a whole file into memory and returns a slice of its lines.
// https://stackoverflow.com/questions/5884154/read-text-file-into-string-array-and-write
func Read_lines(path string) ([]string, error) {
	file, file_open_err := os.Open(path)
	if file_open_err != nil {
		return nil, file_open_err
	}

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	file_close_err := file.Close()
	if file_close_err != nil {
		return nil, errors.New("An error occurred while closing file: " + path + " || Error: " + file_close_err.Error())
	}

	return lines, scanner.Err()
}
