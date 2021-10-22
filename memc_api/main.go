package main

// To upload a file: localhost:8080/upload
// To download a file: localhost:8080/download

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/EzrAzuri/memc/memc_process"
)

var MAX_FILE_SIZE int64 = 1024 * 1024 * 50 // 50 MiB
var debug bool
var file_retrieve string
var Memcached_Server string

func main() {
	flag.StringVar(&file_retrieve, "o", "out_file.dat", "Output file for retrieval")
	flag.StringVar(&Memcached_Server, "s", os.Getenv("MEMCACHED_SERVER_URL"), "Memcached_Server:port")
	flag.BoolVar(&debug, "d", false, "Debug mode")
	flag.Parse()

	if Memcached_Server == "" {
		Memcached_Server = "localhost11211"
	}

	memc_process.Memcached_Server = Memcached_Server
	dir, err := os.Getwd()

	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	http.HandleFunc("/upload", upload_handler)   // Display upload form
	http.HandleFunc("/receive", receive_handler) // Handle incoming file

	http.HandleFunc("/download", download_handler) // Display download form (from Memcached key)
	http.HandleFunc("/retrieve", retrieve_handler) // Retrieve Memcached key

	http.HandleFunc("/test", test_handler)      // Handle test
	http.HandleFunc("/info", memc_process.Info) // Handle test
	http.Handle("/", http.FileServer(http.Dir(dir)))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Return HTML upload form
func upload_handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		fmt.Fprintf(w,
			`<html>
			<head>
				<title>Memc Upload<title>
			<head>
			<body>
				<h3>Memc - Upload a file to Memcached<h3>
				<form action="/receive" method="post" enctype="multipart/form-data">
					<label for="file">Filename:</label><br>
					<input type="file" name="file" id="file"><br><br>
					<input type="submit" name="submit" value="Submit">
				</form>
			</body>
		</html>`,
		)
	}
}

// Accept the file and saves it to the current working directory
func receive_handler(w http.ResponseWriter, r *http.Request) {
	// Take in the post input id file
	file, header, err := r.FormFile("file")

	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	defer file.Close()

	uploaded_file := header.Filename
	out, err := os.Create(uploaded_file)

	if err != nil {
		fmt.Fprintf(w, "Unable to create the file for writing. Check your write access privilege")
		return
	}

	defer out.Close()

	// Write the content from POST to the file
	_, err = io.Copy(out, file)

	if err != nil {
		fmt.Fprintln(w, err)
	}

	fmt.Fprintf(w, "Memcached_Server: %s\n", Memcached_Server)
	fmt.Fprintf(w, "Storing file in memcache: %s\n", uploaded_file)
	sha, errStore := memc_process.Store_Process(uploaded_file, Memcached_Server, MAX_FILE_SIZE, debug, true)

	if errStore != nil {
		fmt.Printf("Error: file store process to memcached failed\n")
		return
	}

	// Delete file
	errRemove := os.Remove(uploaded_file)

	if errRemove != nil {
		fmt.Printf("Error: %v\n", errRemove)
	}

	fmt.Fprintf(w, "Key: %s\n", uploaded_file)
	fmt.Fprintf(w, "sha256sum: %x\n", sha)
}

func download_handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		fmt.Fprintf(w, `
			<html>
				<head>
					<title>Memc Download</title>
				</head>
				<body>
					<h2>Request Memcached Key</h2>
					<form action="/retrieve" method="post">
						<label for="key">memcache key:</label><br>
						<input type="text" name="key" id="key"><br>
					</form>
				</body>
			</html>
		`)
	}
}

// accept the memcache key request
func retrieve_handler(w http.ResponseWriter, r *http.Request) {
	// Take in POST input values
	requested_key := r.FormValue("key")

	data, err := memc_process.Retrieve_Process(requested_key, Memcached_Server, file_retrieve, debug)

	if err != nil {
		fmt.Printf("Error: file retrieve process from memcached failed\n")
		return
	}

	fmt.Fprintf(w, "%s", data)
}

func test_handler(w http.ResponseWriter, r *http.Request) {
	msg := "Memc"
	w.Write([]byte(msg))
	w.Write([]byte("\n"))
}
