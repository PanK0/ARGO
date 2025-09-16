package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
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

// Save the content of a Message as a log file named r_log_<Source>.log in the logs directory
func saveReceivedLog(m Message) error {
    // Ensure the logs directory exists
    if _, err := os.Stat(LOGDIR); os.IsNotExist(err) {
        err := os.Mkdir(LOGDIR, 0755)
        if err != nil {
            return fmt.Errorf("Error creating log directory: %v", err)
        }
    }

    // Prepare the filename using the source (sanitize if needed)
    filename := fmt.Sprintf("r_log_%s.log", addressToPrint(m.Source, NODE_PRINTLAST))
    logPath := filepath.Join(LOGDIR, filename)

    // Write the content to the file
    f, err := os.Create(logPath)
    if err != nil {
        return fmt.Errorf("Error creating log file: %v", err)
    }
    defer f.Close()

    _, err = f.WriteString(m.Content)
    if err != nil {
        return fmt.Errorf("Error writing to log file: %v", err)
    }

    return nil
}

// Save the content of a Message as topology file
func saveReceivedTop(m Message) error {
    // Prepare the filename using the source (sanitize if needed)
	filename := topology_path

    // Write the content to the file
    f, err := os.Create(filename)
    if err != nil {
        printError(err)
    }
    defer f.Close()

    _, err = f.WriteString(m.Content)
    if err != nil {
        printError(err)
    }

    return nil
}

// Reset all the data structures and byzantines
func totalReset(h host.Host, messageContainer *MessageContainer, deliveredMessages *MessageContainer, disjointPaths *DisjointPaths, topology *Topology) {

	// Reset data structs
	messageContainer.Reset()
	deliveredMessages.Reset()
	disjointPaths.Reset()
	topology.Reset()

	// Reset byzantine status
	if byzantine_status {
		event := fmt.Sprintf("byzantine - Node %s is no more a byzantine", addressToPrint(h.ID().String(), NODE_PRINTLAST))
		logEvent(h.ID().String(), PRINTOPTION, event)
		color_info = GREEN
	}
	byzantine_status = false


	// Parse master_address as multiaddr and get peer info
    maddr, err := multiaddr.NewMultiaddr(master_address)
    if err != nil {
        printError(err)
    }
    master_peerInfo, err := peer.AddrInfoFromP2pAddr(maddr)
    if err != nil {
        printError(err)
    }

	// Disconnect from peers
	peers := h.Network().Peers()
    for _, peerID := range peers {
		if peerID != master_peerInfo.ID {
			err := h.Network().ClosePeer(peerID)
			if err != nil {
				fmt.Printf("Error disconnecting from peer "+peerID.String()+"\n")
			} else {
				fmt.Printf("Disconnected from peer "+peerID.String()+"\n")
			}
		}        
    }

	fmt.Println("Node reset: DONE")
}