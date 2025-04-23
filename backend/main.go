package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type EmailJob struct {
	Email      string `json:"email"`
	RetryCount int    `json:"retry_count"`
}

var (
	queue       []EmailJob
	queueMutex  sync.Mutex
	sentCount   int
	failedCount int
)

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if r.Method == "OPTIONS" {
			return
		}

		next.ServeHTTP(w, r)
	}
}

func main() {
	// Start email worker
	go processQueue()

	http.HandleFunc("/upload", enableCORS(uploadHandler))
	http.HandleFunc("/queue", enableCORS(queueHandler))
	http.HandleFunc("/stats", enableCORS(statsHandler))

	fmt.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("emails")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		http.Error(w, "Invalid CSV", http.StatusBadRequest)
		return
	}

	queueMutex.Lock()
	defer queueMutex.Unlock()

	for _, record := range records[1:] { // Skip header
		if len(record) > 0 {
			queue = append(queue, EmailJob{Email: record[0], RetryCount: 0})
		}
	}

	w.Write([]byte(fmt.Sprintf("Đã thêm %d email vào queue", len(records)-1)))
}

func processQueue() {
	for {
		queueMutex.Lock()
		if len(queue) > 0 {
			job := queue[0]
			queue = queue[1:]

			// Simulate email sending
			if mockSendEmail(job.Email) {
				sentCount++
			} else {
				if job.RetryCount < 3 {
					job.RetryCount++
					// Exponential backoff
					time.AfterFunc(time.Duration(2^job.RetryCount)*time.Second, func() {
						queueMutex.Lock()
						queue = append(queue, job)
						queueMutex.Unlock()
					})
				} else {
					failedCount++
				}
			}
		}
		queueMutex.Unlock()
		time.Sleep(100 * time.Millisecond) // Batch processing
	}
}

func mockSendEmail(email string) bool {
	// Simulate 90% success rate
	if time.Now().UnixNano()%10 < 9 {
		log.Printf("Sent to %s", email)
		return true
	}
	log.Printf("Failed to send %s", email)
	return false
}

func queueHandler(w http.ResponseWriter, r *http.Request) {
	queueMutex.Lock()
	defer queueMutex.Unlock()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"queue_length": len(queue),
		"pending":      queue,
	})
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sent":   sentCount,
		"failed": failedCount,
	})
}
