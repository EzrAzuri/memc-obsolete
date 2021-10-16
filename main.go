package main

import (
	"flag"

	"github.com/EzrAzuri/memc/src/memc_process"
	//"./src/configs"
)

func main() {
	var MAX_FILE_SIZE int64 = 1024 * 1024 * 50 // 50 MiB
	var file_store, file_retrieve, file_path, memcachedServer string
	var test, debug, force bool

	flag.StringVar(&file_store, "s", "", "File to store")
	flag.StringVar(&file_retrieve, "r", "", "File to retrieve")
	flag.StringVar(&file_path, "p", "", "Path to the file")
	flag.StringVar(&memcachedServer, "m", "localhost:11211", "memcached_server:port")
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

	// for what?
	if file_store == "" && file_path == "" && test == false {
		memc_process.Helper("")
	}
}
