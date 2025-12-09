package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"vhost-scout/include/banner_utils"
	"vhost-scout/include/file_utils"
	"vhost-scout/include/input_utils"
	"vhost-scout/include/random_utils"
	"vhost-scout/include/request_utils"
	"vhost-scout/include/sqlite_utils"
)

type t_target_that_encountered_error struct {
	target string
	error  error
}

type t_vhost struct {
	target                      string
	vhost                       string
	baseline_response_body_md5  string
	spoofed_response_body_md5   string
	spoofed_request_status_code int
}

func process_target(target string, vhosts_list []string) ([]t_vhost, error) {

	// ----| Shuffle vhosts list to avoid basic defences
	rand.Shuffle(len(vhosts_list), func(i, j int) {
		vhosts_list[i], vhosts_list[j] = vhosts_list[j], vhosts_list[i]
	})

	// ----| Make initial request to target with random host header to establish baseline response to requests to non-existent vhosts
	baseline_resp_md5_hash, _, baseline_req_err := request_utils.Send_request_with_spoofed_host_header(target, random_utils.Gen_random_string(rand.Intn(10))+".com") // Send baseline request with a spoofed Host header set to a random 1 to 10 letter string followed by .com
	if baseline_req_err != nil {
		return nil, errors.New("Error occurred while attempting to make baseline request to: " + target + " with Host header: " + target + "\n" + baseline_req_err.Error())
	}

	var enumerated_vhosts []t_vhost
	for _, vhost := range vhosts_list {

		// ----| Send request with spoofed Host header
		spoofed_req_md5_hash, spoofed_request_interface, spoofed_req_err := request_utils.Send_request_with_spoofed_host_header(target, vhost)
		if spoofed_req_err != nil {
			return nil, errors.New("Error occurred while attempting to send spoofed request to: " + target + "with Host header: " + target + "\n" + spoofed_req_err.Error())
		}

		if spoofed_req_md5_hash != baseline_resp_md5_hash {

			switch {
			case strings.HasPrefix(strconv.Itoa(spoofed_request_interface.StatusCode), "2"):
				fmt.Printf("  > %s %s", vhost, color.GreenString("(Status Code: %d)\n\n", spoofed_request_interface.StatusCode))
			case strings.HasPrefix(strconv.Itoa(spoofed_request_interface.StatusCode), "3"):
				fmt.Printf("  > %s %s", vhost, color.YellowString("(Status Code: %d)\n\n", spoofed_request_interface.StatusCode))
			case strings.HasPrefix(strconv.Itoa(spoofed_request_interface.StatusCode), "4") || strings.HasPrefix(strconv.Itoa(spoofed_request_interface.StatusCode), "5"):
				fmt.Printf("  > %s %s", vhost, color.RedString("(Status Code: %d)\n\n", spoofed_request_interface.StatusCode))
			default:
				fmt.Printf("  > %s %s", vhost, color.RedString("(Status Code: %d)\n\n", spoofed_request_interface.StatusCode))
			}

			vhost_information := t_vhost{
				target:                      target,
				vhost:                       vhost,
				baseline_response_body_md5:  baseline_resp_md5_hash,
				spoofed_response_body_md5:   spoofed_req_md5_hash,
				spoofed_request_status_code: spoofed_request_interface.StatusCode,
			}

			enumerated_vhosts = append(enumerated_vhosts, vhost_information)
		}

		sleep_time := rand.Intn(3) // n will be between 0 and 3
		time.Sleep(time.Duration(sleep_time) * time.Second)
	}
	return enumerated_vhosts, nil
}

func add_enumerated_vhosts_to_db(enumerated_vhosts []t_vhost) error {

	// ----| Ensure there are vhost to add to db
	if len(enumerated_vhosts) == 0 {
		fmt.Println("  > No vhosts were enumerated\n")
		return nil
	}

	// ----| Open database interface
	database_interface, open_db_interface_err := sqlite_utils.Open_database_interface("db.sqlite")
	if open_db_interface_err != nil {
		return errors.New("An error occurred while initializing the database interface || Error: " + open_db_interface_err.Error())
	}

	for _, vhost_information := range enumerated_vhosts {

		// ----| Build row
		db_row := sqlite_utils.Table_row{
			Target:                      vhost_information.target,
			Vhost:                       vhost_information.vhost,
			Baseline_response_body_md5:  vhost_information.baseline_response_body_md5,
			Spoofed_response_body_md5:   vhost_information.spoofed_response_body_md5,
			Spoofed_request_status_code: vhost_information.spoofed_request_status_code,
		}

		// ----| Insert row into table
		add_row_to_table_err := sqlite_utils.AddRowToTable(database_interface, "enumerated_vhosts", db_row)
		if add_row_to_table_err != nil {
			return errors.New("An error occurred while adding row to enumerated vhosts db table || Error: " + add_row_to_table_err.Error())
		}
	}

	// ----| Close database interface
	db_close_err := sqlite_utils.Close_database_interface(database_interface)
	if db_close_err != nil {
		return db_close_err
	}
	return nil
}

func run(targets_file_path_or_target_url string, vhosts_lists_path string, allow_insecure_requests string) error {

	if strings.ToLower(allow_insecure_requests) == "true" {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // Configure http to allow insecure requests
	}

	var targets_list []string
	if input_utils.IsDomainOrURL(targets_file_path_or_target_url) == false { // Means targets_list_path is a file
		// ----| Load targets from file
		targets_from_file, file_read_err := file_utils.Read_lines(targets_file_path_or_target_url)
		if file_read_err != nil {
			return errors.New(fmt.Sprintf("An error occurred while attempting to read targets from file: %s || Error: %s", targets_file_path_or_target_url, file_read_err.Error()))
		}
		targets_list = targets_from_file
	} else { // Means targets_list_path is a url
		targets_list = append(targets_list, targets_file_path_or_target_url)
	}

	// ----| Load vhosts from file
	vhosts_list, file_read_err := file_utils.Read_lines(vhosts_lists_path)
	if file_read_err != nil {
		return errors.New(fmt.Sprintf("An error occurred while attempting to read vhosts from file: %s || Error: %s", vhosts_lists_path, file_read_err.Error()))
	}

	fmt.Println("▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁")

	// ----| Print banner
	banner_utils.Print_banner(targets_list, vhosts_list)

	fmt.Println("▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁")

	var targets_that_errored []t_target_that_encountered_error
	for _, target := range targets_list {

		fmt.Printf("\n\n> Starting VHost Enumeration On: %s\n\n", target)
		enumerated_vhosts, target_processing_err := process_target(target, vhosts_list)
		if target_processing_err != nil {
			fmt.Printf("> An error occured while processing target: %s || Error: %s", target, target_processing_err.Error())
			targets_that_errored = append(targets_that_errored, t_target_that_encountered_error{target, target_processing_err})
			continue
		}

		if len(enumerated_vhosts) != 0 {
			fmt.Printf("  > Adding enumerated vhosts to database\n\n")
			err := add_enumerated_vhosts_to_db(enumerated_vhosts)
			if err != nil {
				fmt.Printf("> An error occurred while adding enumerated vhosts on target: %s to the db. || Error: %s\n", target, err.Error())
				targets_that_errored = append(targets_that_errored, t_target_that_encountered_error{target, err})
				continue
			}

		} else {
			fmt.Println("  > No vhosts were enumerated\n")
		}

		fmt.Printf("  > Finished VHost Enumeration On Target: %s\n\n", target)

		sleep_time := rand.Intn(10) // n will be between 0 and 10
		fmt.Printf("  > Sleeping %d seconds...\n", sleep_time)
		time.Sleep(time.Duration(sleep_time) * time.Second)
	}

	if len(targets_that_errored) != 0 {
		fmt.Println("▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁")
		fmt.Println("> Targets that encountered an error during scanning")
		for _, target_that_encountered_error := range targets_that_errored {
			fmt.Println("  > " + target_that_encountered_error.target + " || Error: " + target_that_encountered_error.error.Error())
		}
	} else {
		fmt.Println("\n\n> All targets were enumerated successfully")
	}
	return nil
}

func main() {

	/*
		targets_file_path_or_target_url := "example-targets.txt"
		vhosts_lists_path := "example-vhosts.txt"
		run(targets_file_path_or_target_url, vhosts_lists_path)
	*/

	if len(os.Args) != 3 || os.Args[1] == "help" || os.Args[1] == "-h" || os.Args[1] == "--help" {
		fmt.Print(banner_utils.Cat + "\n")
		fmt.Println("▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁\n\n")
		fmt.Print("vhost-scout probes one or more targets to discover virtual-host (vhost) names.\n\n")
		fmt.Printf("Usage: %s (<target ip> | <targets-ips.txt>) <vhosts.txt> \n", os.Args[0])
		os.Exit(0)
	}

	targets_file_path_or_target_url := os.Args[1]
	vhosts_lists_path := os.Args[2]
	allow_insecure_requests := os.Args[3]
	err := run(targets_file_path_or_target_url, vhosts_lists_path, allow_insecure_requests)
	if err != nil {
		panic("An error occurred while running the program" + err.Error())
		return
	}
}
