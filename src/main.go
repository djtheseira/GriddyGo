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
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/lib/pq"

	_ "github.com/lib/pq"
)

// Request struct contains data that will be sent
// by the user
type Request struct {
	Key   int    `json:"key"`
	T1Key int    `json:"t1key"`
	Value string `json:"value"`
}

// Result struct contains the results
type Result struct {
	Key       int        `json:"key"`
	Value     string     `json:"value"`
	T2Results []T2Result `json:"children"`
}

// T2Result struct contains the results connected
// to the T1 Key
type T2Result struct {
	Key        int    `json:"key"`
	Value      string `json:"value"`
	CreateDate string `json:"createdate"`
}

// Response struct contains data that will be returned
// to the user
type Response struct {
	Results []Result `json:"results"`
}

// MessageResponse struct contains data that returns status message
type MessageResponse struct {
	Msg string `json:"msg,omitempty"`
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
var results []Result

var serverport = "3000"
var filename = "config/config.json"

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

func createT2Results(t2String string) []T2Result {
	t2Array := strings.Split(t2String, "&&")
	t2Results := []T2Result{}
	if len(t2Array) > 0 && len(t2Array[0]) > 0 {
		t2Results = make([]T2Result, len(t2Array))
		for i := range t2Array {
			t2RowSplit := strings.Split(t2Array[i], ",")
			key, iErr := strconv.Atoi(t2RowSplit[0])
			if iErr != nil {
				continue
			}
			createdate, tErr := time.Parse("2006-01-02 15:04:05.000000", t2RowSplit[2])
			if tErr != nil {
				log.Println("Error with date parse", tErr)
				continue
			}

			t2Results[i] = T2Result{Key: key, Value: t2RowSplit[1], CreateDate: createdate.Format("2006-01-02")}
		}
	}
	return t2Results
}

// Function for display a home page message
func homePage(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "DT Griddy Go Server")
}

// Function to handle getting data
func handleGetData(w http.ResponseWriter, r *http.Request) {
	msg := ""
	hKey := r.Header.Get("key")
	isOk := 1
	dbErr := s.db.Ping()
	if dbErr != nil {
		msg = "Server Error"
		isOk = 0
	}

	// If header "key" is passed in, only get that value, otherwise, get all values
	// .....safe
	if isOk == 1 {
		if len(hKey) > 0 {
			query := `SELECT t1.key AS T1Key, t1.value AS T1Value, 
					COALESCE(string_agg(t2.key || ',' || t2.value || ',' || createdate, '&&'), '')
				FROM griddy.t1
				LEFT JOIN griddy.t2 ON (t2.t1key = t1.key)
				WHERE t1.key = $1
				GROUP BY t1.key, t1.value;`
			var t1Key int
			var t1Value string
			var t2String string

			row := s.db.QueryRow(query, hKey)
			switch err := row.Scan(&t1Key, &t1Value, &t2String); err {
			case sql.ErrNoRows:
				msg = ""
				results = make([]Result, 0)
			case nil:
				results = []Result{Result{Key: t1Key, Value: t1Value, T2Results: createT2Results(t2String)}}
			default:
				msg = "Error getting data"
				log.Println("Error getting data", err)
			}
		} else {
			query := `SELECT t1.key AS T1Key, t1.value AS T1Value, 
					COALESCE(string_agg(t2.key || ',' || t2.value || ',' || createdate, '&&'), '')
				FROM griddy.t1
				LEFT JOIN griddy.t2 ON (t2.t1key = t1.key)
				GROUP BY t1.key, t1.value;`
			countQuery := `SELECT COUNT(*) FROM griddy.t1;`
			var count int
			row := s.db.QueryRow(countQuery)
			cErr := row.Scan(&count)
			if cErr != nil {
				log.Println("Error counting", cErr)
			}
			rows, err := s.db.Query(query)
			if err != nil {
				log.Println("Error setting data", err)
			}
			defer rows.Close()
			results = make([]Result, count)
			place := 0
			for rows.Next() {
				var t1Key int
				var t1Value string
				var t2String string
				err = rows.Scan(&t1Key, &t1Value, &t2String)
				if err != nil {
					msg = "Error getting data"
					log.Println("Error setting data", err)
				}
				results[place] = Result{Key: t1Key, Value: t1Value, T2Results: createT2Results(t2String)}
				place++
			}
		}
	}

	var gJSON []byte

	if len(msg) > 0 {
		response := &MessageResponse{Msg: msg}
		gJSON, _ = json.Marshal(response)
	} else {
		response := &Response{Results: results}
		gJSON, _ = json.Marshal(response)
	}
	io.WriteString(w, string(gJSON))
}

// Function to handle posting data
func handlePostData(w http.ResponseWriter, r *http.Request) {
	msg := ""
	isOk := 1
	dbErr := s.db.Ping()
	if dbErr != nil {
		msg = "Server Error"
		isOk = 0
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Error reading request body", err)
		msg += "Server Error"
		isOk = 0
	}

	var request Request
	err = json.Unmarshal(body, &request)
	if err != nil {
		msg += "Error parsing the request, please try again"
		log.Println("json unmarshal error", err)
		isOk = 0
	}

	if len(request.Value) > 0 && isOk == 1 {
		table := "T1"
		query := `INSERT INTO griddy.t1(value) VALUES($1) RETURNING key, value;`
		if request.T1Key > 0 {
			countQuery := `SELECT COUNT(*) FROM griddy.t1 WHERE key = $1`
			var count int
			row := s.db.QueryRow(countQuery, request.T1Key)
			err = row.Scan(&count)
			if err != nil {
				log.Println("Error counting", err)
			}
			if count > 0 {
				query = `INSERT INTO griddy.t2(t1key, value) VALUES($1,$2) RETURNING key, value;`
			} else {
				msg = fmt.Sprintf("No row exists with the key: %d.", request.T1Key)
				isOk = 0
			}
			table = "T2"
		}
		if isOk == 1 {
			key := 0
			value := ""
			if request.T1Key > 0 {
				err = s.db.QueryRow(query, request.T1Key, request.Value).Scan(&key, &value)
			} else {
				err = s.db.QueryRow(query, request.Value).Scan(&key, &value)
			}

			if err != nil {
				if pgerr, ok := err.(*pq.Error); ok {
					if pgerr.Code == "23505" {
						msg = "That value is already taken in the database, " +
							"please insert a unique value."
					}
				} else {
					msg = "Error inserting into db"
				}
				log.Println("Error inserting into db", err)
			} else {
				msg = fmt.Sprintf("Successfully added '%s' with Key of: %d into %s", value, key, table)
			}
		}
	} else if isOk != 0 {
		msg = "No value to insert into the Database, please pass a value in."
	}

	response := &MessageResponse{Msg: msg}
	json, _ := json.Marshal(response)
	io.WriteString(w, string(json))
}

// Function to handle deleting data
func handleDeleteData(w http.ResponseWriter, r *http.Request) {
	msg := ""
	isOk := 1
	dbErr := s.db.Ping()
	if dbErr != nil {
		msg = "Server Error"
		isOk = 0
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Error reading request body", err)
		if len(msg) > 0 {
			msg += "\n"
		}
		msg += "Server Error"
		isOk = 0
	}

	var request Request
	err = json.Unmarshal(body, &request)
	if err != nil {
		if len(msg) > 0 {
			msg += "\n"
		}
		if err.Error() == "json: cannot unmarshal string into Go struct field Request.key of type int" {
			msg += "Error with your request: The key is of an invalid type."
		} else {
			msg += "Error parsing the request, please try again"
		}
		log.Println("json unmarshal error", err)
	}

	if (request.Key > 0 || request.T1Key > 0) && isOk == 1 {
		table := "T1"
		key := request.Key
		if request.Key == 0 && request.T1Key > 0 {
			table = "T2"
			key = request.T1Key
		}
		countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM griddy.%s WHERE key = $1;`, table)
		var count int
		row := s.db.QueryRow(countQuery, key)
		cErr := row.Scan(&count)
		if cErr != nil {
			log.Println("Error counting", cErr)
		}
		if count == 0 {
			msg = fmt.Sprintf("There is no key that matches %d in table %s", key, table)
		} else {
			query := fmt.Sprintf(`DELETE FROM griddy.%s WHERE key = $1;`, table)
			_, err = s.db.Exec(query, key)

			if err != nil {
				msg = "Error deleting from db"
				log.Println("Error deleting value from database", err)
			} else {
				msg = fmt.Sprintf("Key deleted successfully from table %s", table)
			}
		}
	} else if isOk != 0 {
		if len(msg) == 0 {
			msg = "No key given to delete from database, please pass in a key to delete."
		}
	}

	response := &MessageResponse{Msg: msg}
	json, _ := json.Marshal(response)
	io.WriteString(w, string(json))
}

// Function that handles all data requests for /data
func handleDataRequest(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/data" {
		http.Error(w, "404 Not Found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		handleGetData(w, r)
	case "POST":
		handlePostData(w, r)
	case "DELETE":
		handleDeleteData(w, r)
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
