package memc_process

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/bradfitz/gomemcache/memcache"
	//"github.com/EzrAzuri/memc/configs" // Not implemented yet
)

const devilsBytes int = 62 // what is this
const MAX_CHUNK_SIZE = 1024*1024 - devilsBytes

var Memcached_Server string

func Retrieve_Process(f, m_server, outFile string, debug bool) ([]byte, error) {
	// get number of required chunks for the file
	filehash, num, errGet := Getter(m_server, f, debug)

	if errGet != nil {
		return []byte{}, errGet
	}

	if debug == true {
		fmt.Printf("Retrieve_Process ->\n")
		fmt.Printf("Key: %s\n", f)
		fmt.Printf("Chunks: %v\n", num)
		fmt.Printf("Filehash: %x\n", string(filehash))
	}

	// Open file
	file, errCreate := os.Create(outFile)

	if errCreate != nil {
		return []byte{}, errCreate
	}

	defer file.Close()

	// Reconstitute
	for i := 1; i <= int(num); i++ {
		chunk_key := f + "-" + strconv.Itoa(i)

		// Get single chunk
		chunk_item, _, err := Getter(m_server, chunk_key, debug)

		if err != nil {
			return []byte{}, err
		}

		// Write file
		n2, werr := file.Write(chunk_item)

		if debug == true {
			fmt.Printf("chunk_key: %s\n", chunk_key)
			fmt.Printf("\tchunk: %v", i)
			fmt.Printf("\tBytes written: %d\n", n2)

			if werr != nil {
				return []byte{}, werr
			}
		}
	}

	// Read newly created file
	data, err_read := ioutil.ReadFile(outFile)

	if err_read != nil {
		return []byte{}, err_read
	}

	// Delete file
	err_remove := os.Remove(outFile)

	if err_remove != nil {
		fmt.Printf("error: %v\n", err_remove)
	}

	// Hash the file and output results
	newHash := sha256.Sum256(data)

	if debug == true {
		fmt.Printf("%s: sha256sum: %x\n", outFile, newHash)
	}

	compare_result := bytes.Compare(filehash[:], newHash[:])

	var err error

	if compare_result != 0 {
		err = errors.New("hash mismatch")
	}

	return data, err
}

func Store_Process(f, mserver string, max int64, debug, force bool) ([32]byte, error) {
	buffer_size := MAX_CHUNK_SIZE - len(f)

	if _, err := os.Stat(f); os.IsNotExist(err) {
		// file does not exist
		return [32]byte{}, err
	}

	// Get number of required chunks for file
	num, shasum, err := Num_Chunks(f, buffer_size, max, debug)

	if err != nil {
		return [32]byte{}, err
	}

	if debug == true {
		fmt.Printf("Store_Process ->\n")
		fmt.Printf("\tKey: %s\n", f)
		fmt.Printf("\tValue: %x\n", shasum)
		fmt.Printf("chunks: %v\n", num)
		fmt.Printf("sha256sum: %x\n", shasum)
		fmt.Printf("Setting item:\n")
	}

	// Set key named after file with value of shasum
	errSetterFile := Setter(mserver, f, shasum[:], uint32(num), 0, debug, force)

	if errSetterFile != nil {
		return [32]byte{}, errSetterFile
	}

	// Open file
	file, errOpen := os.Open(f)

	if errOpen != nil {
		return [32]byte{}, errOpen
	}

	defer file.Close()

	buffer := make([]byte, buffer_size)

	var i int = 1

	for {
		bytesread, err := file.Read(buffer)

		if err != nil {
			if err != io.EOF {
				return [32]byte{}, err
			}
			break
		}
		buff := buffer[:bytesread]

		// Set file contents
		fileKey := f + "-" + strconv.Itoa(i)

		if debug == true {
			fmt.Printf("\tChunk: %v\n", i)
			fmt.Printf("\tBytes read: %v", bytesread)
			fmt.Printf("\tKey: %v\n", fileKey)
		}

		errSet := Setter(mserver, fileKey, buff, 0, 0, debug, force)

		if errSet != nil {
			return [32]byte{}, err
		}

		i++
	}

	fmt.Printf("key: %s\n", f)
	fmt.Printf("sha256sum: %x\n", shasum)
	return shasum, err
}

// For setting mcache values
func Setter(mserver, key string, val []byte, fla uint32, exp int32, debug, force bool) error {
	mc := memcache.New(mserver)

	// Check for pre-existing key
	_, _, errGet := Getter(mserver, key, debug)

	if errGet == nil && force != true {
		return errors.New("key exists")
	}

	// Set key
	err := mc.Set(&memcache.Item{Key: key, Value: val, Flags: fla, Expiration: exp})

	if debug == true {
		fmt.Printf("SETTER> %v\n", err)
	}

	return err
}

func Getter(mserver, key string, debug bool) ([]byte, uint32, error) {
	mc := memcache.New(mserver)

	// Get key
	if debug == true {
		fmt.Printf("Get key -> %s\n", key)
	}

	myitem, err := mc.Get(key)

	if err != nil {
		return []byte{}, 0, err
	}

	return myitem.Value, myitem.Flags, err
}

func Server_Check(Memcached_Server string) error {
	fmt.Println("memc_process->server_check->")

	// mc := memcache.New("10.0.0.1:11211", "10.0.0.2:11211", "10.0.0.3:11211")
	mc := memcache.New(Memcached_Server)

	// Check connection to memcached server
	fmt.Printf("Ping memcached server\n")

	errPing := mc.Ping()

	if errPing != nil {
		fmt.Printf("Ping failed!\n")
		fmt.Printf("ERROR: %v", errPing)
	}

	// Set key
	key_in := "foo"
	value_in := "bar"

	fmt.Printf("Set item\n")
	fmt.Printf("Set key -> %s\tvalue: %s\n", key_in, value_in)
	mc.Set(&memcache.Item{Key: key_in, Value: []byte(value_in)})

	// Get key
	fmt.Printf("Key key ->\n")
	myitem, err := mc.Get("foo")

	if err != nil {
		fmt.Printf("ERROR: %v", err)
	}

	fmt.Printf("Item: %v\n", myitem)
	fmt.Printf("Key: %v\n", myitem.Key)
	fmt.Printf("Value: %v\n", string(myitem.Value))
	fmt.Printf("Flags: %v\n", myitem.Flags)
	fmt.Printf("Expiration: %v\n", myitem.Expiration)

	return err
}

func Num_Chunks(file_name string, chunk_size int, max int64, debug bool) (int, [32]byte, error) {
	size_bytes, errFS := File_Size(file_name)

	if errFS != nil {
		return 0, [32]byte{}, errFS
	}

	// Empty file check
	if size_bytes == 0 {
		return 0, [32]byte{}, errors.New("Invalid file (empty file)")
	}

	// Max file size check
	if size_bytes > max {
		fmt.Printf("Max size: %d\n", max)
		errMsg := fmt.Sprintf("ERROR: File too large: %d\n", size_bytes)
		return 0, [32]byte{}, errors.New(errMsg)
	}

	data, err := ioutil.ReadFile(file_name)

	if err != nil {
		return 0, [32]byte{}, err
	}

	fileSHA256 := sha256.Sum256(data)

	// Calculate the number of chunks (1 MiB each)
	float_chunks := float64(size_bytes) / float64(chunk_size)

	int_num_chunks := int(float_chunks)

	if float_chunks > float64(int_num_chunks) {
		int_num_chunks++
	}

	if debug == true {
		fmt.Printf("File: %s\n", file_name)
		fmt.Printf("Size (bytes): %d\n", size_bytes)
		fmt.Printf("SHA-256: %x\n", fileSHA256)
		fmt.Printf("Chunks (1 MiB) Float: %f\n", float_chunks)
		fmt.Printf("Chunks (1 MiB) Int: %d\n", int_num_chunks)
	}

	return int_num_chunks, fileSHA256, err
}

// Provide help usage message and exits
func Helper(msg string) {
	if msg != "" {
		fmt.Printf("%s\n\n", msg)
	}

	fmt.Println("Store file in memcached")
	fmt.Println("Supply file name i.e. /path/to/myfile.txt")
	flag.PrintDefaults()
	os.Exit(1)
}

// Check file size
func File_Size(f string) (int64, error) {
	file, err := os.Stat(f)
	return file.Size(), err
}

// Info handler displays http header
func Info(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s %s %s\n", r.Method, r.URL, r.Proto)

	for k, v := range r.Header {
		fmt.Fprintf(w, "Header[%q] = %q\n", k, v)
	}

	fmt.Fprintf(w, "Host = %q\n", r.Host)
	fmt.Fprintf(w, "RemoteAddr = %q\n", r.RemoteAddr)
	fmt.Fprintf(w, "Memcached_Server = %q\n", Memcached_Server)

	if err := r.ParseForm(); err != nil {
		log.Print(err)
	}

	for k, v := range r.Form {
		fmt.Fprintf(w, "Form[%q] = %q\n", k, v)
	}
}
