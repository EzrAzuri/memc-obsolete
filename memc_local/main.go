package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/EzrAzuri/memc/memc_process"
)

func main() {
	var MAX_FILE_SIZE int64 = 1024 * 1024 * 50 // 50 MiB
	var file_store, file_retrieve, file_path, memcached_server string
	var test, debug, force bool

	flag.StringVar(&file_store, "s", "", "File to store")
	flag.StringVar(&file_retrieve, "r", "", "File to retrieve")
	flag.StringVar(&file_path, "p", "", "Path to the file")
	flag.StringVar(&memcached_server, "m", "localhost:11211", "memcached_server:port")
	flag.BoolVar(&debug, "d", false, "Debug mode")
	flag.BoolVar(&test, "t", false, "Check memcached server")
	flag.BoolVar(&force, "f", false, "Force memcached key overwrite")
	flag.Parse()

	if test == true && (file_store != "" || file_path != "") {
		memc_process.Helper("Must supply test as single argument (-t).")
	}

	if file_store != "" && file_path != "" {
		memc_process.Helper("Must supply file as argument (-p or -g).")
	}

	if file_store == "" && file_path == "" && test == false {
		memc_process.Helper("")
	}

	if file_store != "" {
		if _, err := memc_process.Store_Process(file_store, memcached_server, MAX_FILE_SIZE, debug, force); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	}

	if file_retrieve != "" {
		file_data, err := memc_process.Retrieve_Process(file_path, memcached_server, file_retrieve, debug)

		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("%s", file_data)
	}

	if test == true {
		if err := memc_process.Server_Check(memcached_server); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	}
}
