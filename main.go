package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/GeertJohan/go.rice"
	"github.com/gorilla/mux"
)

var bookingCounter = 0

type Booking struct {
	ID             int    `json:"id"`
	StationAddress string `json:"station"`
}

// Structs for json communication.
type NewBooking struct {
	DroneAddress   string `json:"droneID"`
	StationAddress string `json:"stationID"`
}

type DroneStatus struct {
	Status  string `json:"status"`
	Station string `json:"station"`
}

var bookingsMap = make(map[string][]Booking)

var bookings []Booking
var drones []DroneStatus

type handler func(w http.ResponseWriter, r *http.Request)

func basicAuth(pass handler) handler {

	return func(w http.ResponseWriter, r *http.Request) {

		auth := strings.SplitN(r.Header.Get("Authorization"), " ", 2)

		if len(auth) != 2 || auth[0] != "Basic" {
			http.Error(w, "authorization failed", http.StatusUnauthorized)
			return
		}

		payload, _ := base64.StdEncoding.DecodeString(auth[1])
		pair := strings.SplitN(string(payload), ":", 2)

		if len(pair) != 2 || !validate(pair[0], pair[1]) {
			http.Error(w, "authorization failed", http.StatusUnauthorized)
			return
		}

		pass(w, r)
	}
}

func validate(username, password string) bool {
	if username == "test" && password == "test" {
		return true
	}
	return false
}

// POST <- api/book/new {DroneAddress: WalletAddress, StationAddress: WalletAddress}
// GET ->  api/drone/bookings/:id [] ? [{id: booking_id, station: station_address}]
// POST <- api/drone/status/:id {ConnectedStation: address ? '', Idle: Boolean}

func handlerNewBooking(w http.ResponseWriter, r *http.Request) {
	fmt.Println("NEW BOOKING")
	decoder := json.NewDecoder(r.Body)
	var b NewBooking
	err := decoder.Decode(&b)
	if err != nil {
		panic(err)
	}
	booking := Booking{ID: bookingCounter, StationAddress: b.StationAddress}
	// Add booking to store
	bookings = append(bookings, booking)
	bookingsMap[b.DroneAddress] = append(bookingsMap[b.DroneAddress], booking)
	bookingCounter++
	defer r.Body.Close()
	log.Println("Testing booking for drone:" + b.DroneAddress + " station: " + b.StationAddress)
}

func handlerDroneStatus(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Called drone status")
	params := mux.Vars(r)
	id := params["id"]
	decoder := json.NewDecoder(r.Body)
	var d DroneStatus
	err := decoder.Decode(&d)
	if err != nil {
		panic(err)
	}
	fmt.Println("Parsed json")
	// w.WriteHeader(http.StatusOK)
	//drones = append(drones, d)
	defer r.Body.Close()
	log.Println("Storing status for drone " + id + " station " + d.Station + " status" + d.Status)
}

func handlerGetBookings(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	arr, found := bookingsMap[params["id"]]
	fmt.Println(params["id"])
	if found == false {
		fmt.Fprintf(w, "[]")
		return
	}
	json, err := json.Marshal(arr)
	if err != nil {
		panic(err)
	}
	//w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Radnom request: %s", r.URL.Path)
	//w http.ResponseWriter, r *http.Request
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/")
	fmt.Fprintf(w, "Hi there, I love %s!, %s", r.URL.Path[1:], id)
}

func main() {
	r := mux.NewRouter()
	port := flag.String("p", "8100", "port to serve on")
	directory := flag.String("d", "./web/dist", "the directory of static file to host")
	flag.Parse()
	fmt.Println("Hello from DroneChain server")
	r.HandleFunc("/api/drone/status/{id}", handlerDroneStatus).Methods("POST")
	r.HandleFunc("/api/booking/new", handlerNewBooking)          // .Methods("POST")
	r.HandleFunc("/api/drone/bookings/{id}", handlerGetBookings) // .Methods("GET")
	r.HandleFunc("/api/status/{id}", handlerDroneStatus).Methods("POST")
	r.PathPrefix("/").Handler(http.FileServer(rice.MustFindBox("web/dist").HTTPBox()))

	http.Handle("/", r)
	log.Printf("Serving %s on HTTP port: %s\n", *directory, *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
