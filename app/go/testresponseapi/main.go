package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type testResponse struct {
	Value  string `json:"value"`
}

func returnTestValue(w http.ResponseWriter, r *http.Request){
	json.NewEncoder(w).Encode(testResponse{"Hello world!"})
}

func main(){
	port := ":10000"
	log.Printf("Starting test service, Listening to: %s", port)
	http.HandleFunc("/api/get/testvalue", returnTestValue)
	log.Fatal(http.ListenAndServe(port, nil))
}