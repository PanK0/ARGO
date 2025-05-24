package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// mutex for log
var logMutex sync.Mutex

// ReplaceInCSV replaces all occurrences of a given string, that is a given node in the .csv file,
// in the CSV file with node_id and saves the file.
// ReplaceInCSV replaces all occurrences of `s` in the CSV file with `nodeID`
// only if the content of the cell exactly matches `s`.
func ReplaceInCSV(filePath, nodeID, nodeID_csv string) {
	// Open the CSV file
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	// Parse the CSV file
	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read CSV file: %v", err)
	}

	// Iterate over each cell and replace if the content exactly matches `s`
	for i := range rows {
		for j := range rows[i] {
			if rows[i][j] == nodeID_csv {
				rows[i][j] = nodeID
			}
		}
	}

	// Write the updated data back to the file
	file, err = os.Create(filePath)
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	err = writer.WriteAll(rows)
	if err != nil {
		log.Fatalf("Failed to write to CSV file: %v", err)
	}

	fmt.Println("File updated successfully!")
}


// Creates a log for what happens in the node
func logEvent(nodeID string, printoption bool, event string) {
	logMutex.Lock() // Ensure only one write at a time
	defer logMutex.Unlock()

	nodeID = addressToPrint(nodeID, NODE_PRINTLAST)

	// Ensure the logs directory exists
	if _, err := os.Stat(LOGDIR); os.IsNotExist(err) {
		err := os.Mkdir(LOGDIR, 0755) // Create /logs directory
		if err != nil {
			fmt.Println("Error creating log directory:", err)
			return
		}
	}

	// Define log file path
	logFile := fmt.Sprintf("%s/%s.log", LOGDIR, nodeID)

	// Open or create the log file
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening log file:", err)
		return
	}
	defer f.Close()

	// Format the log entry
	timestamp := time.Now().Format("15:04:05.00000")
	logMessage := fmt.Sprintf("[%s] [%s] %s\n", timestamp, nodeID, event)

	// Write log entry to file
	_, err = f.WriteString(logMessage)
	if err != nil {
		fmt.Println("Error writing to log file:", err)
	}

	if printoption {
		logMessage = fmt.Sprintf("%s[%s%s%s]%s %s[%s%s%s]%s %s\n", color_info, RESET, timestamp, color_info, RESET, color_info, RESET, nodeID, color_info, RESET, event)
		fmt.Print(logMessage)
	}
}
