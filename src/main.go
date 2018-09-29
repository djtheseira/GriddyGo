package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// Request struct contains data that will be sent
// by the user
type Request struct {
	Value string `json:"value"`
}

// Response struct contains data that will be returned
// to the user
type Response struct {
	Msg string `json:"msg,omitempty"`
}

// Create a struct to contain server fields
type server struct {
	httpServer *http.Server
	listener   net.Listener
}

// Data strig to be returned when called by the API
var Data string
var port = "3000"

// Function to create the server at a certain address
func (s *server) ListenAndServe() error {
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
func (s *server) shutdown() error {
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

// TODO:
// - Write a function to connect to the db

// Function for display a home page message
func homePage(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "DT Griddy Go Server")
}

// TODO
// Modify this function to get a value from the db.
// Function to handle getting data
func handleGetData(w http.ResponseWriter) {
	response := &Response{Msg: Data}
	json, err := json.Marshal(response)
	if err != nil {
		log.Println("json marshal error", err)
	}
	log.Println("response", string(json))
	io.WriteString(w, string(json))
}

// TODO
// Modify this function to write a value to the db.
// Function to handle posting data
func handlePostData(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Error reading request body", err)
		return
	}

	var request Request
	err = json.Unmarshal(body, &request)
	if err != nil {
		log.Println("json unmarshal error", err)
		return
	}
	log.Println(request.Value)

	Data = request.Value
	io.WriteString(w, Data)
}

// TODO
// Modify this function to delete a value from the db.
// Function to handle deleting data
func handleDeleteData() {
	Data = ""
}

// Function that handles all data requests for /data
func handleDataRequest(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/data" {
		http.Error(w, "404 Not Found", http.StatusNotFound)
		return
	}

	switch r.Method {
	case "GET":
		handleGetData(w)
	case "POST":
		handlePostData(w, r)
	case "DELETE":
		handleDeleteData()
	default:
		fmt.Fprintf(w, "Sorry, only DELETE, GET, and POST methods are supported.")
	}
}

// Function to create a new server
func newServer(port string) *server {
	mux := http.NewServeMux()

	mux.HandleFunc("/", homePage)
	mux.HandleFunc("/data", handleDataRequest)
	httpServer := &http.Server{Addr: ":" + port, Handler: mux}
	fmt.Println("Server Listening")
	return &server{httpServer: httpServer}
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

	server := newServer(port)
	server.ListenAndServe()
	defer byebye()

	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Printf("Signal: %s\n", sig)
		// Graceful sutdown
		server.shutdown()
		done <- true
	}()

	fmt.Println("Ctrl-C to interrupt...")
	<-done
	fmt.Println("Exiting...")

}
