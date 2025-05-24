package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
)

// Handle a broadcast stream
// This is actually a Byzantine Reliable Broadcast, granted by DolevU protocol
// See @ Thesis Farina, 2021, CPT 6.1, Algorithm 1, PDF pg 42/142
// + PDF pg 43/142, CPT 6.1.2 - DolevU Message Complexity - for performance analysis
func handleBroadcast(s network.Stream, ctx context.Context, thisNode host.Host, messageContainer *MessageContainer) error {
	buf := bufio.NewReader(s)
	message, err := buf.ReadString('\n')
	if err != nil {
		return err
	}

	// Transform the message into a json object with all the information
	var m Message
	err = json.Unmarshal([]byte(message), &m)
	if err != nil {
		printError(err)
	}

	// Byzantine checking
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
				return nil
			}
		}
	}

	// Add sender node to the path.
	// The PATH is represented as an array of strings
	// If this is the target node, then append the target to the path
	newPath := append(m.Path, m.Sender)
	if (m.Target == getNodeAddress(thisNode, ADDR_DEFAULT)) {
		newPath = append(newPath, m.Target)
	}
	m.Path = newPath

	// add the message to the dedicated data struct
	//receivedMessages.Add(string(new_message))
	messageContainer.Add(m)

	new_message, err := json.Marshal(m)
	if err != nil {
		fmt.Println("Error marshalling data while handling a message:", err)
	}

	printMessage(string(new_message))

	// If we are on the target node do not forward the message
	if (m.Target != getNodeAddress(thisNode, ADDR_DEFAULT)) {
		sendBroadcast(ctx, thisNode, string(new_message))
	}

	return nil
}

// Send a Broadcast message
// This is actually a Byzantine Reliable Broadcast, granted by DolevU protocol
// See @ Thesis Farina, 2021, CPT 6.1, Algorithm 1, PDF pg 42/142 
// + PDF pg 43/142, CPT 6.1.2 - DolevU Message Complexity - for performance analysis
func sendBroadcast(ctx context.Context, thisNode host.Host, msg string) {
    // Get the json object from content, where is stored all the information about the sender, the source and the destination
    var m Message
    err := json.Unmarshal([]byte(msg), &m)
    if err != nil {
        printError(err)
        return
    }

    // Cycle through the peers connected to the current node
    for _, p := range thisNode.Network().Peers() {
        // If the peer p is already in the path of the message, then do not forward the message
        if contains(m.Path, p.String()) {
            fmt.Println("Do not forward on node ", p)
            printShell()
            continue
        }

        // Open a stream with the peer
        stream, err := openStream(ctx, thisNode, p, PROTOCOL_NAB)
        if err != nil {
            printError(err)
            continue
        }

        // Ensure the stream is closed after use
        defer stream.Close()

        // Change the sender into the content: the sender node is now this node
        m.Sender = getNodeAddress(thisNode, ADDR_DEFAULT)

        dataBytes, err := json.Marshal(m)
        if err != nil {
            fmt.Println("Error marshalling data while sending a message:", err)
            continue
        }

        // Lock the function (critical section)
        streamMutex.Lock()

        // Write the message on the stream
        message := fmt.Sprintf("%s\n", string(dataBytes))
        _, err = stream.Write([]byte(message))
        if err != nil {
            printError(err)
        }

        // Unlock before exiting
        streamMutex.Unlock()
    }
}