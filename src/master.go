package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
)

func handleMaster(s network.Stream, ctx context.Context, thisNode host.Host, messageContainer *MessageContainer, topology *Topology) error {
	// Read the buffer and extract the message
	buf := bufio.NewReader(s)
	message, err := buf.ReadString('\n')
	if err != nil {
		printError(err)
	}

	// Transform the message into a json object with all the information
	var m Message
	err = json.Unmarshal([]byte(message), &m)
	if err != nil {
		printError(err)
	}

	if m.Content == mst_top_acquire {
		acquireTopology(thisNode, topology)
		fmt.Println("Topology acquired")
		fmt.Println(topology.ctop.toString())
	} else if m.Content == mst_top_load {
		topology_graph := LoadGraphFromCSV(topology_path)
		topology.ctop.loadNeigh(topology_graph, getNodeAddress(thisNode, ADDR_DEFAULT))
		// topology.ctop = *loadCTop(topology_graph) // Uncomment this to laod the whole topology
		fmt.Println(topology.ctop.toString())	
	} else if m.Content == mst_connectall {
		connectAllNodes(ctx, thisNode, topology)
	}

	fmt.Printf("\n%s_> %s", GREEN, RESET)	
	return nil
}


// Send master message
func sendMaster(ctx context.Context, thisNode host.Host, m Message) {
	m.Sender = getNodeAddress(thisNode, ADDR_DEFAULT)
	dataBytes, err := json.Marshal(m)
	if err != nil {
		printError(err)
	}
	msg := string(dataBytes)

	// Cycle through the peers connected to the current node
	for _, p := range thisNode.Network().Peers() {

		// open a new stream
		stream, err := openStream(ctx, thisNode, p, PROTOCOL_MST)
		if err != nil {
			printError(err)
		}

		message := fmt.Sprintf("%s\n", msg)

		// Write the message on the stream
		_, err = stream.Write([]byte(message))
		if err != nil {
			printError(err)
		}
	}
}

// Manage MASTER message
//lint:ignore U1000 Unused function for future use
func manageMaster(commands string) {
	commands_list := parseCommandString(commands)
	for c := range commands_list {
		fmt.Printf("Command flag: %s\nCommand: %s\n\n", c, commands_list[c])
	}
}
// parseCommandString extracts flag-value pairs while handling multi-word values
////lint:ignore U1000 Unused function for future use 
func parseCommandString(input string) map[string]string {
	result := make(map[string]string)

	// Regular expression to match flags (-command_X) and capture their values
	re := regexp.MustCompile(`-(\S+)\s+([^-\n]+)`) // Ensures the value is NOT another flag

	matches := re.FindAllStringSubmatch(input, -1)

	for _, match := range matches {
		flag := "-" + match[1]                     // Reconstruct the full flag
		value := strings.TrimSpace(match[2])       // Remove leading/trailing spaces
		result[flag] = value                        // Store in map
	}

	return result
}