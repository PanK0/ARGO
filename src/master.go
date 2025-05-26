package main

import (
	"bufio"
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

func handleMaster(s network.Stream, ctx context.Context, thisNode host.Host, messageContainer *MessageContainer, topology *Topology, disjointPaths *DisjointPaths) error {
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
	} else if m.Content == mst_disconnect {
		disconnectNodes(ctx, thisNode, m.Source)
	} else if m.Content == mst_crc_exp {
		// Generate an ID for the message
		timestamp := time.Now().Unix()
		hasher := sha1.New()
		hasher.Write([]byte(fmt.Sprintf("%d", timestamp)))
		msgid := fmt.Sprintf("%x", hasher.Sum(nil))
		neighbourhood := topology.ctop.GetNeighbourhood(getNodeAddress(thisNode, ADDR_DEFAULT))
		var visitedSet []string
		var crc_message Message = 
		Message{
			ID: msgid, 
			Type: TYPE_CRC_EXP, 
			Sender: "", 
			Source: getNodeAddress(thisNode, ADDR_DEFAULT), 
			Target: "",
			Content: "",
			Neighbourhood: neighbourhood,
			Path: visitedSet,
		}
		sendCombinedRC(ctx, thisNode, crc_message, topology, disjointPaths)
	} else if m.Content == mst_graph {
		g := exp2_ConvertCTopToGraph(&topology.ctop)
		event := g.GraphToString()
		logEvent(thisNode.ID().String(), false, event)
		g.PrintGraph()
	} else if m.Content == mst_djp {
		event := disjointPaths.toString()
		logEvent(thisNode.ID().String(), PRINTOPTION, event)
	} else if m.Content == mst_log {
		timestamp := time.Now().Unix()
		hasher := sha1.New()
		hasher.Write([]byte(fmt.Sprintf("%d", timestamp)))
		msgid := fmt.Sprintf("%x", hasher.Sum(nil))
		var neighbourhood []string
		var visitedSet []string
		var log_master_message Message = 
		Message{
			ID: msgid, 
			Type: TYPE_MASTER, 
			Sender: getNodeAddress(thisNode, ADDR_DEFAULT), 
			Source: getNodeAddress(thisNode, ADDR_DEFAULT), 
			Target: "",
			Content: "",
			Neighbourhood: neighbourhood,
			Path: visitedSet,
		}
		sendLogToMaster(ctx, thisNode, log_master_message)
	} else {
		saveReceivedLog(m)
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

		if m.Content == mst_crc_exp {
			// sleep for 1 second to allow the message to be processed
			time.Sleep(1 * time.Second)
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

// Send this node's log file to the master node
func sendLogToMaster(ctx context.Context, thisNode host.Host, m Message) error {
    // Prepare log file path
    nodeID := addressToPrint(thisNode.ID().String(), NODE_PRINTLAST)
    logFile := fmt.Sprintf("%s/%s.log", LOGDIR, nodeID)

    // Open log file
    f, err := os.Open(logFile)
	if err != nil {
        return fmt.Errorf("failed to open log file: %v", err)
    }
    defer f.Close()

	// Transform the content of f into a string
	log_str, err := io.ReadAll(f)
	if err != nil {
		printError(err)
	}
	m.Content = string(log_str)

    // Parse master_address as multiaddr and get peer info
    maddr, err := multiaddr.NewMultiaddr(master_address)
    if err != nil {
        printError(err)
    }
    peerInfo, err := peer.AddrInfoFromP2pAddr(maddr)
    if err != nil {
        printError(err)
    }

    stream, err := openStream(ctx, thisNode, peerInfo.ID, PROTOCOL_MST)
    if err != nil {
        printError(err)
    }
    defer stream.Close()

	dataBytes, err := json.Marshal(m)
	if err != nil {
		printError(err)
	}
	message := fmt.Sprintf("%s\n", string(dataBytes))
	_, err = stream.Write([]byte(message))
	if err != nil {
		printError(err)
	}
    return nil
}