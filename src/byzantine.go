package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
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

// Read the number of byzantines
func readMaxByzantines(config_filename string, MAX_BYZANTINES *int) error {
	// Open the config file
	file, err := os.Open(config_filename)
	if err != nil {
		return fmt.Errorf("error opening config file: %v", err)
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
			return fmt.Errorf("invalid config line: %s", line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if key == "MAX_BYZANTINES" {
			*MAX_BYZANTINES, err = strconv.Atoi(value)
		}
	}
	return nil
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

// SwapTwoRandom swaps the position of two random strings in a slice.
// If the slice has fewer than 2 elements, it does nothing.
func SwapTwoRandom(list []string) []string {
    if len(list) < 2 {
        return list
    }
    rand.Seed(time.Now().UnixNano())
    i := rand.Intn(len(list))
    j := rand.Intn(len(list))
    for j == i {
        j = rand.Intn(len(list))
    }
    list[i], list[j] = list[j], list[i]
    return list
}

// Applies the bizantine changes on the message
// Returns true if bz.Type2 == true, so that the message can be dropped in the main function
func applyByzantine(thisNode host.Host, m *Message) bool {
	if byzantine_status {
		// If byzantine is of Type 1, then sleep for bz.Delay milliseconds
		if bz.Type1 {
			event := fmt.Sprintf("byzantine %s - delay of %s ms", m.Content, bz.Delay)
			logEvent(thisNode.ID().String(), PRINTOPTION, event)
			time.Sleep(bz.Delay)
		}
		// If byzantine is of Type 2, then drop the message with bz.Droprate probability
		if bz.Type2 {
			if (rand.Float64() < bz.DropRate) {
				event := fmt.Sprintf("byzantine %s - Message from %s dropped", m.Content, addressToPrint(m.Sender, NODE_PRINTLAST))
				logEvent(thisNode.ID().String(), PRINTOPTION, event)
				return bz.Type2
			}
		}
		// If byzantine is of Type 3, then remove one random element from the neighbourhood or path
		if bz.Type3 {
			if bz.Alterations == BYZ_NEIGHBOURHOOD {
				if len(m.Neighbourhood) > 0 {
					// Remove a random element from the neighbourhood
					rand.Seed(time.Now().UnixNano())
					index := rand.Intn(len(m.Neighbourhood))
					removed := m.Neighbourhood[index]
					m.Neighbourhood = append(m.Neighbourhood[:index], m.Neighbourhood[index+1:]...)
					event := fmt.Sprintf("byzantine %s - Message from %s altered. Removed %s from neighbourhood.", m.Content, addressToPrint(m.Sender, NODE_PRINTLAST), addressToPrint(removed, NODE_PRINTLAST))
					logEvent(thisNode.ID().String(), PRINTOPTION, event)
				} 
			} else if bz.Alterations == BYZ_PATH {
				if len(m.Path) > 0 {
					// Remove a random element from the path
					rand.Seed(time.Now().UnixNano())
					index := rand.Intn(len(m.Path))
					removed := m.Path[index]
					m.Path = append(m.Path[:index], m.Path[index+1:]...)
					event := fmt.Sprintf("byzantine %s - Message from %s altered. Removed %s from path.", m.Content, addressToPrint(m.Sender, NODE_PRINTLAST), addressToPrint(removed, NODE_PRINTLAST))
					logEvent(thisNode.ID().String(), PRINTOPTION, event)
				}
			} else if bz.Alterations == BYZ_SWAP_PATH {
				if len(m.Path) > 1 {
					rand.Seed(time.Now().UnixNano())
					i := rand.Intn(len(m.Path))
					j := rand.Intn(len(m.Path))
					for j == i {
						j = rand.Intn(len(m.Path))
					}
					m.Path[i], m.Path[j] = m.Path[j], m.Path[i]
					event := fmt.Sprintf("byzantine %s - Message from %s altered. Swapped %s and %s in path.", m.Content, addressToPrint(m.Sender, NODE_PRINTLAST), addressToPrint(m.Path[j], NODE_PRINTLAST), addressToPrint(m.Path[i], NODE_PRINTLAST))
					logEvent(thisNode.ID().String(), PRINTOPTION, event)
				}
			} else if bz.Alterations == BYZ_ALTER_ID {
				if len(m.ID) > 0 {
					m.ID = m.ID[:len(m.ID)-1]
					event := fmt.Sprintf("byzantine %s - ID of message from %s altered. ", m.Content, addressToPrint(m.Sender, NODE_PRINTLAST))
					logEvent(thisNode.ID().String(), PRINTOPTION, event)
				}
			}
		}
	}
	return false
}