package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"crypto/rand"
	"encoding/hex"

	"github.com/joho/godotenv"
)

// init funct
func init() {
	// Load the .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

const maxConcurrentRequests = 100

var sem = make(chan struct{}, maxConcurrentRequests)
var linkToURLMap = make(map[string]string)

// var urlMap = make(map[string]string)

func main() {
	_, err := loadURLsFromFile("urls.json")
	if err != nil {
		log.Println("Error loading URLs from file:", err)
	}

	http.HandleFunc("/webhook", webhookHandler)
	http.HandleFunc("/createPo", webhookHandlerPurchaseInvoice)
	http.HandleFunc("/add-url", addURLHandler)
	http.HandleFunc("/forward/", forwardHandler)
	log.Fatal(http.ListenAndServe(":8088", nil))

	// Load the URLs from the JSON file
	// Load the URLs from the JSON file

}

func processRequest(destinationURL string, data map[string]interface{}, w http.ResponseWriter) {
	defer func() { <-sem }() // Release the semaphore when the function returns

	// Encode the JSON data
	jsonData, err := json.Marshal([]interface{}{data})
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

	// Check the Content-Type header
	contentType := r.Header.Get("Content-Type")

	var data map[string]interface{}

	if strings.Contains(contentType, "application/json") {
		// Read and decode the JSON request body
		err := json.NewDecoder(r.Body).Decode(&data)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, "Error decoding request body", http.StatusInternalServerError)
			return
		}
	} else {
		// Assume the request body is form data and parse it
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Error parsing form data", http.StatusInternalServerError)
			return
		}

		// Convert the form data to a map[string]interface{}
		data = make(map[string]interface{})
		for key, values := range r.Form {
			if len(values) > 0 {
				data[key] = values[0]
			}
		}
	}

	sem <- struct{}{} // Acquire the semaphore
	go processRequest(destinationURL, data, w)
}

func addURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read and decode the request body
	var requestData map[string]string
	err := json.NewDecoder(r.Body).Decode(&requestData)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, "Error decoding request body", http.StatusInternalServerError)
		return
	}

	url := requestData["url"]
	if url == "" {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	// Generate a unique link
	uniqueLink := generateUniqueLink()

	// Store the mapping
	linkToURLMap[uniqueLink] = url

	// Save the updated linkToURLMap to the JSON file
	err = saveURLsToFile("urls.json", linkToURLMap)
	if err != nil {
		http.Error(w, "Error saving URLs to file", http.StatusInternalServerError)
		return
	}

	// Generate the full URL
	fullURL := os.Getenv("HOST") + "/forward/" + uniqueLink

	// Respond to the caller with a 200 OK status and a JSON message
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"unique_link": uniqueLink, "full_url": fullURL, "message": "URL added"})
}

func forwardHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed", http.StatusMethodNotAllowed)
		return
	}

	uniqueLink := strings.TrimPrefix(r.URL.Path, "/forward/")
	destinationURL, ok := linkToURLMap[uniqueLink]
	if !ok {
		http.Error(w, "Invalid unique link", http.StatusNotFound)
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

func generateUniqueLink() string {
	key := make([]byte, 8)
	rand.Read(key)
	return hex.EncodeToString(key)
}

func saveURLsToFile(filename string, urlMap map[string]string) error {
	urlsJSON, err := json.Marshal(urlMap)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filename, urlsJSON, 0644)
	if err != nil {
		return err
	}

	return nil
}

func loadURLsFromFile(filename string) (map[string]string, error) {
	urlsJSON, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var urlMap map[string]string
	err = json.Unmarshal(urlsJSON, &urlMap)
	if err != nil {
		return nil, err
	}

	return urlMap, nil
}
