package main

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"vhost-scout/include/banner_utils"
	"vhost-scout/include/file_utils"
	"vhost-scout/include/input_utils"
	"vhost-scout/include/request_utils"
	"vhost-scout/include/sqlite_utils"
)

type t_vhost struct {
	target                      string
	vhost                       string
	baseline_response_body_md5  string
	spoofed_response_body_md5   string
	spoofed_request_status_code int
}

func gen_random_string(string_length int) string {

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

func gen_response_body_md5(response_body io.ReadCloser) (string, error) {
	response_body_md5_hash := md5.New()
	if _, err := io.Copy(response_body_md5_hash, response_body); err != nil {
		return "", errors.New("An error occurred while generating md5 hash of response body: " + err.Error())
	}
	return hex.EncodeToString(response_body_md5_hash.Sum(nil)), nil
}

func send_request_with_spoofed_host_header(target string, vhost string) (string, http.Response, error) {

	// ----| Build request so we can spoof Host header
	spoofed_req, err := http.NewRequest("GET", target, nil)
	if err != nil {
		return "", http.Response{}, errors.New("Error occurred while attempting to build request to: " + target + "with Host header: " + vhost + "\n" + err.Error())
	}

	// ----| Set Request Headers
	spoofed_req_headers := request_utils.Generate_random_request_headers() // Generate Random Request Headers
	spoofed_req.Header = spoofed_req_headers
	// Set additional headers here as needed
	spoofed_req.Host = vhost // Spoof host header

	// ----| Make request with spoofed Host header
	resp_to_spoofed_req, err := http.DefaultClient.Do(spoofed_req)
	if err != nil {
		return "", http.Response{}, errors.New("An error occurred while making a spoofed request to: " + target + " with Host header: " + vhost + "\n" + err.Error())
	}

	// ----| Generate md5 hash from baseline_resp body
	resp_to_spoofed_req_md5_hash, err := gen_response_body_md5(resp_to_spoofed_req.Body)
	if err != nil {
		return "", http.Response{}, errors.New("Error occurred while attempting to generate md5 hash of the baseline response body while processing target: " + target)
	}
	return resp_to_spoofed_req_md5_hash, *resp_to_spoofed_req, nil
}

func process_target(target string, vhosts_list []string) ([]t_vhost, error) {

	// ----| Shuffle vhosts list to avoid basic defences
	rand.Shuffle(len(vhosts_list), func(i, j int) {
		vhosts_list[i], vhosts_list[j] = vhosts_list[j], vhosts_list[i]
	})

	// ----| Make initial request to target with random host header to establish baseline response to requests to non-existent vhosts
	baseline_resp_md5_hash, _, err := send_request_with_spoofed_host_header(target, gen_random_string(rand.Intn(10))+".com") // Send baseline request with a spoofed Host header set to a random 1 to 10 letter string followed by .com
	if err != nil {
		return nil, errors.New("Error occurred while attempting to make baseline request to: " + target + "with Host header: " + target + "\n" + err.Error())
	}

	var enumerated_vhosts []t_vhost
	for _, vhost := range vhosts_list {

		// ----| Send request with spoofed Host header
		spoofed_req_md5_hash, spoofed_request_interface, spoofed_req_err := send_request_with_spoofed_host_header(target, vhost)
		if spoofed_req_err != nil {
			return nil, errors.New("Error occurred while attempting to send spoofed request to: " + target + "with Host header: " + target + "\n" + spoofed_req_err.Error())
		}

		// fmt.Printf("baseline request hash: %s\n", baseline_resp_md5_hash) // DEBUG
		// fmt.Printf("spoofed request hash: %s\n", spoofed_req_md5_hash)    // DEBUG

		if spoofed_req_md5_hash != baseline_resp_md5_hash {

			fmt.Printf("  > %s (Status Code: %d)\n\n", vhost, spoofed_request_interface.StatusCode)

			vhost_information := t_vhost{
				target:                      target,
				vhost:                       vhost,
				baseline_response_body_md5:  baseline_resp_md5_hash,
				spoofed_response_body_md5:   spoofed_req_md5_hash,
				spoofed_request_status_code: spoofed_request_interface.StatusCode,
			}
			//fmt.Println(vhost_information) // DEBUG
			enumerated_vhosts = append(enumerated_vhosts, vhost_information)
		}
	}
	return enumerated_vhosts, nil
}

func add_enumerated_vhosts_to_db(enumerated_vhosts []t_vhost) {

	// ----| Ensure there are vhost to add to db
	if len(enumerated_vhosts) == 0 {
		fmt.Println("  > No vhosts were enumerated\n")
		return
	}

	// ----| Open database interface
	database_interface, err := sqlite_utils.Open_database_interface("db.sqlite")
	if err != nil {
		panic(fmt.Errorf("error initializing database interface: %w", err))
	}

	for _, vhost_information := range enumerated_vhosts {

		// ----| Build row
		db_row := sqlite_utils.TableRow{
			vhost_information.target,
			vhost_information.vhost,
			vhost_information.baseline_response_body_md5,
			vhost_information.spoofed_response_body_md5,
			vhost_information.spoofed_request_status_code,
		}

		// ----| Insert row into table
		err = sqlite_utils.AddRowToTable(database_interface, "enumerated_vhosts", db_row)
		if err != nil {
			panic(fmt.Errorf("error adding enumerated vhost table: %w", err))
		}
	}

	// ----| Close database interface
	sqlite_utils.Close_database_interface(database_interface)
}

func print_banner(targets []string, vhosts_list []string) {

	var banner_art []string
	banner_art = append(banner_art, banner_utils.Guy_pointing)
	banner_art = append(banner_art, banner_utils.Patric)

	fmt.Println(banner_art[rand.Intn(len(banner_art))])

	fmt.Println("▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁")

	fmt.Println("\n")

	if len(targets) > 10 {
		fmt.Println("Targets: " + strings.Join(targets[:10], ", ") + ", " + strconv.Itoa(len(targets)-10) + " more targets")
	} else {
		fmt.Println("Targets: " + strings.Join(targets, ", "))
	}

	print("\n")

	if len(vhosts_list) > 10 {
		fmt.Println("VHosts: " + strings.Join(vhosts_list[:10], ", ") + ", " + strconv.Itoa(len(vhosts_list)-10) + " more vhosts")
	} else {
		fmt.Println("VHosts: " + strings.Join(vhosts_list, ", "))
	}

	print("\n")
}

func run(targets_from_file_or_target_url string, vhosts_lists_path string) {

	var targets_list []string
	if input_utils.IsDomainOrURL(targets_from_file_or_target_url) == false {
		// Means targets_list_path is a file
		// ----| Load targets from file
		targets_from_file, err := file_utils.Read_lines(targets_from_file_or_target_url)
		if err != nil {
			panic(fmt.Errorf("error reading targets list: %w", err))
		}
		targets_list = targets_from_file
	} else {
		// Means targets_list_path is a url
		targets_list = append(targets_list, targets_from_file_or_target_url)
	}

	// ----| Load vhosts from file
	vhosts_list, err := file_utils.Read_lines(vhosts_lists_path)
	if err != nil {
		panic(fmt.Errorf("an error occured while attempting to read: %w", err))
	}

	fmt.Println("▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁")

	// ----| Print banner
	print_banner(targets_list, vhosts_list)

	fmt.Println("▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁")

	for _, target := range targets_list {

		fmt.Printf("\n\n> Starting VHost Enumeration On Target: %s\n\n", target)
		enumerated_vhosts, err := process_target(target, vhosts_list)
		if err != nil {
			panic(fmt.Errorf("error processing target: %w", err))
		}

		if len(enumerated_vhosts) != 0 {
			fmt.Printf("  > Adding enumerated vhosts to database\n\n")
			add_enumerated_vhosts_to_db(enumerated_vhosts)
		} else {
			fmt.Println("  > No vhosts were enumerated")
		}
		fmt.Printf("  > Finished VHost Enumeration On Target: %s\n\n", target)
	}
}

func main() {

	/*
		targets_from_file_or_target_url := "example-targets.txt"
		vhosts_lists_path := "example-vhosts.txt"
		run(targets_from_file_or_target_url, vhosts_lists_path)
	*/

	if len(os.Args) != 3 || os.Args[1] == "help" || os.Args[1] == "-h" || os.Args[1] == "--help" {
		fmt.Println("=====| vhost-scout |=====\n")
		fmt.Printf("Takes a two inputs the first is either files one containing a of list targets (host to scan for vhosts on) or a singuler target url the other input takes a list of vhosts to scan each host for.\n\n")
		fmt.Printf("Usage: %s (<target.tld> || <targets.txt>) <vhosts.txt> \n", os.Args[0])
		os.Exit(0)
	}

	targets_from_file_or_target_url := os.Args[1]
	vhosts_lists_path := os.Args[2]
	run(targets_from_file_or_target_url, vhosts_lists_path)
}
