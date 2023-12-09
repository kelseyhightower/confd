package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var testData map[string]interface{}

// Fnv1a64 returns a 64 bit hash of the given data using the FNV-1a hashing
// algorithm.  Golang's libraries natively support this hashing, but I want
// something simpler.
func Fnv1a64(data []byte) uint64 {
	var hash uint64 = 14695981039346656037
	for _, d := range data {
		hash = (hash ^ uint64(d)) * 1099511628211
	}
	return hash
}

func locate(path string, data interface{}) (string, int) {
	for strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	//log.Printf("Searching for %s in %v", path, data)
	if len(path) == 0 {
		blob, err := json.Marshal(data)
		if err != nil {
			return err.Error(), http.StatusPreconditionFailed
		}
		return string(blob), http.StatusOK
	}
	i := strings.Index(path, "/")
	directory, ok := data.(map[string]interface{})
	if !ok {
		log.Printf("Found type %T in answers JSON where a map was expected",
			data)
		return "Error in answers JSON", http.StatusPreconditionFailed
	}
	if i == -1 {
		entry, ok := directory[path]
		if !ok {
			return "Not Found", http.StatusNotFound
		}

		switch v := entry.(type) {
		case string:
			// file
			blob, err := json.Marshal(v)
			if err != nil {
				return "Error JSON encoding data", http.StatusInternalServerError
			}
			return string(blob), http.StatusOK
		case map[string]interface{}:
			// directory entry
			return locate("", v)
		case []string:
			// tags
			return locate("", v)
		default:
			log.Printf("Found unsupported type %T in answers JSON", v)
			return "Error in answers JSON", http.StatusPreconditionFailed
		}
	}

	fields := strings.SplitN(path, "/", 2)
	entry, ok := directory[fields[0]]
	if !ok {
		return "Not Found", http.StatusNotFound
	}

	if len(fields) > 1 {
		return locate(fields[1], entry)
	}
	return locate("", entry)
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		log.Println("405: Method Not Allowed")
		return
	}
	if r.Header.Get("Metadata-Flavor") != "Google" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		log.Println("402: Forbidden: Incorrect Metadata-Flavor header")
		return
	}

	// these will be the zero value ("") if not present
	r.ParseForm()
	etag := r.Form.Get("last_etag")
	wait := r.Form.Get("wait_for_change")

	if etag != "0" && wait == "true" {
		// Pretend it takes a couple seconds to generate a "change"
		time.Sleep(2 * time.Second)
	}
	result, status := locate(r.URL.Path, testData)
	if status == http.StatusOK {
		etag = fmt.Sprintf("%x", Fnv1a64([]byte(result)))
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("ETag", etag)
	} else {
		w.Header().Set("Content-Type", "text/plain")
	}
	w.WriteHeader(status)
	fmt.Fprintln(w, result)
	log.Printf("Status Code %d; Path: %s", status, r.URL.Path)
}

func main() {
	answers := flag.String("answers", "test.json",
		"JSON map of HTTP endpoint to mocked response")
	port := flag.String("port", "8001",
		"Port to listen on")

	flag.Parse()
	blob, err := ioutil.ReadFile(*answers)
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}
	err = json.Unmarshal(blob, &testData)
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}

	log.Println("Starting mock Google metadata server...")
	http.HandleFunc("/", handler)
	if strings.ContainsRune(*port, ':') {
		http.ListenAndServe(*port, nil)
	} else {
		http.ListenAndServe(":"+*port, nil)
	}
}
