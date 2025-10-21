package file_utils

import (
	"bufio"
	"os"
)

// readLines reads a whole file into memory  and returns a slice of its lines.
// https://stackoverflow.com/questions/5884154/read-text-file-into-string-array-and-write
func Read_lines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
