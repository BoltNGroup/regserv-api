package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	epp "github.com/BoltNGroup/go-epp"
	"github.com/bitly/go-simplejson"
	"github.com/gorilla/mux"
	"github.com/likexian/whois-go"
	"github.com/likexian/whois-parser-go"
)

type Configuration struct {
	EPPAddress  string
	EPPUsername string
	EPPPassword string
	WhoisServer string
	HTTPPort    string
}

var configuration Configuration

func main() {
	file, _ := os.Open("conf.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	configuration = Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
		fmt.Println("error:", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/status", GetStatus).Methods("GET")
	router.HandleFunc("/domain/{domain}/availability", GetDomainAvailability).Methods("GET")
	router.HandleFunc("/domain/{domain}/whois", GetDomainWhois).Methods("GET")
	router.HandleFunc("/domain/{domain}/whois/raw", GetDomainWhoisRaw).Methods("GET")
	http.ListenAndServe(":"+configuration.HTTPPort, router)
}

// Get API Status
func GetStatus(w http.ResponseWriter, r *http.Request) {
	json := simplejson.New()
	json.Set("status", "ok")
	json.Set("EPPAddress", configuration.EPPAddress)

	payload, err := json.MarshalJSON()
	if err != nil {
		log.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(payload)
}

// Get Domain Availability
func GetDomainAvailability(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	json := simplejson.New()
	json.Set("domain", params["domain"])

	tconn, err := tls.Dial("tcp", configuration.EPPAddress, nil)
	if err != nil {
		log.Println(err)
	}
	conn, err := epp.NewConn(tconn)
	if err != nil {
		log.Println(err)
	}
	err = conn.Login(configuration.EPPUsername, configuration.EPPPassword, "")
	if err != nil {
		log.Println(err)
	}
	dcr, err := conn.CheckDomain(params["domain"])
	if err != nil {
		log.Println(err)
	}
	if err != nil {
		log.Println(err)
	}
	av := make(map[string]bool)
	for _, v := range dcr.Checks {
		av[v.Domain] = v.Available
		if v.Available {
			json.Set("available", "true")
		} else {
			json.Set("available", "false")
		}
	}
	for _, v := range dcr.Charges {
		if v.Category == "premium" {
			json.Set("premium", "true")
		} else {
			json.Set("premium", "false")
		}

	}

	payload, err := json.MarshalJSON()

	w.Header().Set("Content-Type", "application/json")
	w.Write(payload)
}

// Get Parsed Domain Whois Data
func GetDomainWhois(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	json := simplejson.New()
	result, err := whois.Whois(params["domain"], configuration.WhoisServer)

	parsed, err := whois_parser.Parse(result)

	json.Set("Registrar", parsed)

	payload, err := json.MarshalJSON()
	if err != nil {
		log.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(payload)
}

// Get Raw Domain Whois Data
func GetDomainWhoisRaw(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	result, _ := whois.Whois(params["domain"], configuration.WhoisServer)

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(result))
}
