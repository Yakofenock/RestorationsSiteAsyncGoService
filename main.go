package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

var WAIT = 7
var PROBABILITY = 1
var SECRET_KEY = "gfdswaqASFGHGFD"

// For sending to the same channel (always succes):
type Response struct {
	Message string `json:"message"`
}

// Request structure:
type Request struct {
	ID     int    `json:"id"`
	Status string `json:"status"`
}

// Checks and decodes required fields:
func RequestParser(w http.ResponseWriter, r *http.Request) {
	var requestData Request

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&requestData)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	paymentID := strconv.Itoa(requestData.ID)
	if requestData.ID == 0 {
		http.Error(w, "id field not applied", http.StatusForbidden)
		return
	}

	status := requestData.Status
	if requestData.Status == "" {
		http.Error(w, "status field not applied", http.StatusForbidden)
		return
	}

	go BackendRequester(paymentID, status)

	response := Response{Message: "OK"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Sends request to django with given pribability and paraemters:
func BackendRequester(requestID string, status string) {
	fmt.Println("Got request, waiting for ", WAIT, " seconds...")
	time.Sleep(time.Duration(WAIT) * time.Second)

	payload := map[string]interface{}{
		"key":    SECRET_KEY,
		"id":     requestID,
		"status": status,
	}

	reqData, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}

	req, err := http.NewRequest("POST", "http://127.0.0.1:8000/api/restorations_api/v1/asinc_pay_service/", bytes.NewBuffer(reqData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		fmt.Println("Error decoding response:", err)
		return
	}

	fmt.Println("Response:", response)
}

func main() {
	r := http.NewServeMux()
	r.HandleFunc("/pay", RequestParser)

	server := http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	fmt.Println("Server is running on :8080")
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println("Error:", err)
	}
}
