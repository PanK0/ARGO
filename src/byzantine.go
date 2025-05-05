package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

var bz Byzantine

// Byzantine type represents various Byzantine faults in a network
/*
Type 1 = dealys actions
Type 2 = doesn't respond and doesn't relays messages
Type 3 = alters information
*/
type Byzantine struct {
	Type1      	bool         	// First fault type
	Type2      	bool         	// Second fault type
	Type3      	bool          	// Third fault type
	Delay      	time.Duration 	// Delay in milliseconds
	DropRate   	float64      	// Packet drop rate (between 0 and 1)
	Alterations string			// Description of message alterations
}


// LoadByzantineConfig reads the byzantine.config file and loads data into a Byzantine struct
func LoadByzantineConfig(config_filename string) (Byzantine, error) {
	bz := Byzantine{} // Default values

	// Open the config file
	file, err := os.Open(config_filename)
	if err != nil {
		return bz, fmt.Errorf("error opening config file: %v", err)
	}
	defer file.Close()

	// Read file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Ignore empty lines and comments
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split key-value pairs
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return bz, fmt.Errorf("invalid config line: %s", line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Parse values based on keys
		switch key {
		case "Type1":
			bz.Type1, err = strconv.ParseBool(value)
		case "Type2":
			bz.Type2, err = strconv.ParseBool(value)
		case "Type3":
			bz.Type3, err = strconv.ParseBool(value)
		case "Delay":
			delayMs, err := strconv.Atoi(value)
			if err == nil {
				bz.Delay = time.Duration(delayMs) * time.Millisecond
			}
		case "DropRate":
			bz.DropRate, err = strconv.ParseFloat(value, 64)
		case "Alterations":
			bz.Alterations = value
		default:
			fmt.Printf("Warning: Unknown config key '%s'\n", key)
		}

		if err != nil {
			return bz, fmt.Errorf("error parsing key '%s': %v", key, err)
		}
	}

	// Check for scanning errors
	if err := scanner.Err(); err != nil {
		return bz, fmt.Errorf("error reading config file: %v", err)
	}

	return bz, nil
}
