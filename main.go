package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

const maxConcurrentRequests = 100

var sem = make(chan struct{}, maxConcurrentRequests)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	http.HandleFunc("/webhook", webhookHandler)
	http.HandleFunc("/createPo", webhookHandlerPurchaseInvoice)
	log.Fatal(http.ListenAndServe(":8088", nil))
}

func processRequest(destinationURL string, data map[string]interface{}, w http.ResponseWriter) {
	defer func() { <-sem }() // Release the semaphore when the function returns

	// Encode the JSON data
	jsonData, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "Error encoding JSON data", http.StatusInternalServerError)
		return
	}

	// Create a new POST request with the JSON data
	newRequest, err := http.NewRequest(http.MethodPost, destinationURL, bytes.NewBuffer(jsonData))
	if err != nil {
		http.Error(w, "Error creating request", http.StatusInternalServerError)
		return
	}

	// Set the Content-Type header
	newRequest.Header.Set("Content-Type", "application/json")

	// Send the POST request
	client := &http.Client{}
	resp, err := client.Do(newRequest)
	if err != nil {
		http.Error(w, "Error sending request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Log the response status, headers, and body for debugging purposes
	log.Println("Response status:", resp.Status)
	log.Println("Response headers:", resp.Header)
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Error reading response body", http.StatusInternalServerError)
		return
	}
	log.Println("Response body:", string(respBody))

	// Respond to the caller with a 200 OK status and a JSON message
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "OK"})
}

func webhookHandlerPurchaseInvoice(w http.ResponseWriter, r *http.Request) {
	destinationURL := os.Getenv("PURCHASE_INVOICE_URL")
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read and decode the request body
	var data map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&data)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, "Error decoding request body", http.StatusInternalServerError)
		return
	}

	sem <- struct{}{} // Acquire the semaphore
	go processRequest(destinationURL, data, w)
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	destinationURL := os.Getenv("SUPPLIER_URL")

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read and decode the request body
	var data map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&data)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, "Error decoding request body", http.StatusInternalServerError)
		return
	}

	sem <- struct{}{} // Acquire the semaphore
	go processRequest(destinationURL, data, w)
}
