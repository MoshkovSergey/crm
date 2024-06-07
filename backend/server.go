package backend

import (
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

const dataFilePath = "../data/data.txt"

// ensureDataFileExists creates the data file if it does not exist or
// opens it if it already exists. It also ensures the file contains an
// empty JSON array.
func ensureDataFileExists() {
	// Check if the data file exists
	if _, err := os.Stat(dataFilePath); os.IsNotExist(err) {
		// If the file does not exist, create it
		_, err := os.Create(dataFilePath)
		if err != nil {
			panic(err)
		}
	} else if err != nil {
		// If there was an error checking the file existence, panic
		panic(err)
	} else {
		// If the file exists, open it in append mode for writing
		dataFile, err := os.OpenFile(dataFilePath, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}

		// Close the file when the function returns
		defer dataFile.Close()

		// If the file is empty, write an empty JSON array
		stat, err := dataFile.Stat()
		if err != nil {
			log.Fatal(err)
		}
		if stat.Size() == 0 {
			_, err = dataFile.Write([]byte("[]"))
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

// getData handles GET requests for the /data endpoint. It opens the data file, sends its contents to the client,
// and writes the client's request body to the data file.
// Parameters:
// - w: http.ResponseWriter - the response writer used to send the response to the client
// - r: *http.Request - the client's request
func getData(w http.ResponseWriter, r *http.Request) {
	// Open the data file
	dataFile, err := os.Open(dataFilePath)
	if err != nil {
		// If there was an error opening the file, log the error and send an internal server error response
		log.Fatal("file open on get", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Internal server error")
		return
	}
	defer dataFile.Close() // Close the file when the function returns

	// Send the data file's contents to the client
	http.ServeContent(w, r, dataFilePath, time.Now(), dataFile)

	// Write the client's request body to the data file
	_, err = io.Copy(dataFile, r.Body)
	if err != nil {
		// If there was an error writing to the file, log the error and send an internal server error response
		log.Fatal("copy from request", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Internal server error")
		return
	}
}


// postData handles POST requests for the /data endpoint. It opens the data file in write-only mode and truncates the
// file. It then copies the contents of the request body to the data file.
// Parameters:
// - w: http.ResponseWriter - the response writer used to send the response to the client
// - r: *http.Request - the client's request
func postData(w http.ResponseWriter, r *http.Request) {

	// Open the data file in write-only mode and truncate the file.
	// If the file doesn't exist, it will be created.
	dataFile, err := os.OpenFile(dataFilePath, os.O_WRONLY|os.O_TRUNC, 0644)

	// If there was an error opening the file, log the error and send an internal server error response
	if err != nil {
		log.Println("file open on post func", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Internal server error")
		return
	}

	// Close the file when the function returns
	defer dataFile.Close()

	// Copy the contents of the request body to the data file
	_, err = io.Copy(dataFile, r.Body)

	// If there was an error writing to the file, log the error and send an internal server error response
	if err != nil {
		log.Println("copy from request", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Internal server error")
		return
	}
}

// homePage handles GET requests for the root endpoint ("/"). It opens the "index.html" file in the "static" directory,
// reads its contents, and sends them to the client.
// Parameters:
// - w: http.ResponseWriter - the response writer used to send the response to the client
// - r: *http.Request - the client's request
func homePage(w http.ResponseWriter, r *http.Request) {
	// Open the "index.html" file in read-only mode.
	indexFile, err := os.Open("../static/index.html")
	if err != nil {
		// If there was an error opening the file, log the error and send an internal server error response.
		log.Fatal("file open on get", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Internal server error")
		return
	}
	defer indexFile.Close() // Close the file when the function returns.

	// Copy the contents of the file to the response writer.
	io.Copy(w, indexFile)

	// Serve the file with the given file name, modification time, and content.
	http.ServeContent(w, r, "index.html", time.Now(), indexFile)
}


// Start starts the HTTP server that handles requests for the application.
// It creates a new router, sets up the routes, and starts the server.
func Start() {
	// Ensure that the data file exists.
	ensureDataFileExists()

	// Create a new router.
	r := mux.NewRouter()

	// Create a new server.
	srv := &http.Server{
		Handler: r, // Set the router as the server's handler.
		Addr:    "127.0.0.1:8080", // Set the server's address.
		// Enforce timeouts for the server to avoid resource leaks.
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	// Define the routes.

	// Handle GET requests for the root endpoint ("/").
	r.HandleFunc("/", homePage)

	// Handle GET requests for the "/data" endpoint.
	r.Methods("GET").Path("/data").HandlerFunc(getData)

	// Handle POST requests for the "/data" endpoint.
	r.Methods("POST").Path("/data").HandlerFunc(postData)

	// Handle all requests for the "/static/" prefix.
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("../static"))))

	// Log the server's address and start the server.
	log.Printf("Server listening on %s\n", srv.Addr)
	log.Fatal(srv.ListenAndServe())

	// Older code, removed.
	// http.HandleFunc("/data", getData)
	// http.HandleFunc("/", homePage)
	// http.ListenAndServe(":8080", nil)
}
