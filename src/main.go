package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	_ "github.com/lib/pq"
)

// Request struct contains data that will be sent
// by the user
type Request struct {
	Value string `json:"value"`
}

// Response struct contains data that will be returned
// to the user
type Response struct {
	Msg string `json:"msg"`
}

// Create a struct to contain server fields
type server struct {
	httpServer *http.Server
	listener   net.Listener
	db         *sql.DB
}

type configuration struct {
	Port     int
	Host     string
	Dbname   string
	Username string
	Password string
}

// global server for everyone to access...probably bad idea?
var s server

// Data string to be returned when called by the API
var Data string
var serverport = "3000"
var filename = "src/config/config.json"

var badResponse = Response{Msg: "Something went wrong, try again later..."}

func (config *configuration) CreateConfig() error {
	filepath, _ := filepath.Abs(filename)
	file, err := os.Open(filepath)
	if err != nil {
		fmt.Println("Error creating config", err)
		return err
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		fmt.Println("Error decoding config", err)
		return err
	}
	return nil
}

// Function to create the server at a certain address
func listenAndServe() error {
	listener, err := net.Listen("tcp", s.httpServer.Addr)
	// If any errors occur, end and return the error for analysis
	if err != nil {
		return err
	}

	// Store listener in the server struct
	s.listener = listener

	// Allow concurrent services to run
	go s.httpServer.Serve(s.listener)
	fmt.Println("Server listening")
	return nil
}

// Function that will shutdown the server
func shutdown() error {
	// If listener isnt empty, close it, and check for errors
	if s.listener != nil {
		err := s.listener.Close()
		s.listener = nil
		if err != nil {
			return err
		}
	}
	fmt.Println("Shutting down server")
	return nil
}

// Function to create the connection to the db with
// provided connection string
func createDBConnection(psqlInfo string) error {
	fmt.Printf("info %s", psqlInfo)
	tempDb, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		fmt.Println("error opening db connection", err)
		return err
	}
	s.db = tempDb

	// Test the connection with a simple ping to make
	// sure the connection is valid
	err = s.db.Ping()
	if err != nil {
		fmt.Println("error closing db connection", err)
		return err
	}
	fmt.Println("DB Connection Successful")
	return nil
}

func closeDBConnection() error {
	if s.db != nil {
		s.db.Close()
		fmt.Println("Close DB Connection Successful")
	}
	return nil
}

// Function for display a home page message
func homePage(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "DT Griddy Go Server")
}

// TODO
// Modify this function to get a value from the db.
// Function to handle getting data
func handleGetData(w http.ResponseWriter) {
	dbErr := s.db.Ping()
	if dbErr != nil {
		dbjson, _ := json.Marshal(badResponse)
		io.WriteString(w, string(dbjson))
		return
	}
	if Data == "" {
		Data = ""
	}
	response := &Response{Msg: Data}
	rjson, err := json.Marshal(response)
	if err != nil {
		ejson, _ := json.Marshal(badResponse)
		io.WriteString(w, string(ejson))
		log.Println("json marshal error", err)
	} else {
		io.WriteString(w, string(rjson))
	}
}

// TODO
// Modify this function to write a value to the db.
// Function to handle posting data
func handlePostData(w http.ResponseWriter, r *http.Request) {
	dbErr := s.db.Ping()
	if dbErr != nil {
		dbjson, _ := json.Marshal(badResponse)
		io.WriteString(w, string(dbjson))
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Error reading request body", err)
		return
	}

	var request Request
	err = json.Unmarshal(body, &request)
	if err != nil {
		json, _ := json.Marshal(badResponse)
		io.WriteString(w, string(json))
		log.Println("json unmarshal error", err)
		return
	}
	Data = request.Value
	json, _ := json.Marshal(&Response{Msg: "Successfully added data"})
	io.WriteString(w, string(json))
}

// TODO
// Modify this function to delete a value from the db.
// Function to handle deleting data
func handleDeleteData(w http.ResponseWriter) {
	dbErr := s.db.Ping()
	if dbErr != nil {
		dbjson, _ := json.Marshal(badResponse)
		io.WriteString(w, string(dbjson))
		return
	}
	Data = ""
	json, _ := json.Marshal(&Response{Msg: "Data Deleted Successfully"})
	io.WriteString(w, string(json))
}

// Function that handles all data requests for /data
func handleDataRequest(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/data" {
		http.Error(w, "404 Not Found", http.StatusNotFound)
		return
	}

	r.Header.Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		handleGetData(w)
	case "POST":
		handlePostData(w, r)
	case "DELETE":
		handleDeleteData(w)
	default:
		fmt.Fprintf(w, "Sorry, only DELETE, GET, and POST methods are supported.")
	}
}

// Function to create a new server
func newServer(port string) server {
	mux := http.NewServeMux()

	mux.HandleFunc("/", homePage)
	mux.HandleFunc("/data", handleDataRequest)
	httpServer := &http.Server{Addr: ":" + port, Handler: mux}
	return server{httpServer: httpServer}
}

// Main function that runs the show
func main() {

	// Channel to receive unix signals
	sigs := make(chan os.Signal, 1)

	// Channel to receive a confirmation on interrupt
	done := make(chan bool, 1)

	// Channel that receives SIGINT, SIGTERM signals
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	byebye := func() {
		fmt.Println("This is our goodbye")
	}

	s = newServer(serverport)
	err := listenAndServe()
	if err != nil {
		done <- true
	}

	// Create config
	config := &configuration{}
	// Get the config setup
	err = config.CreateConfig()
	if err != nil {
		done <- true
	}

	// Create the psql connection string
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.Username, config.Password, config.Dbname)

	err = createDBConnection(psqlInfo)
	if err != nil {
		done <- true
	}
	defer byebye()

	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Printf("Signal: %s\n", sig)
		// Graceful shutdown
		closeDBConnection()
		shutdown()
		done <- true
	}()

	fmt.Println("Ctrl-C to interrupt...")
	<-done
	fmt.Println("Exiting...")
}
