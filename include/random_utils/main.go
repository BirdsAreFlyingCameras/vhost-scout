package random_utils

import "math/rand"

func Gen_random_string(string_length int) string {

	// ----| Ensure string length is not 0
	if string_length <= 0 {
		string_length = 7
	}

	// ----| Generate random string
	const letters = "abcdefghijklmnopqrstuvwxyz"
	random_string := ""
	for range string_length {
		random_string += string(letters[rand.Intn(len(letters))])
	}
	return random_string
}
