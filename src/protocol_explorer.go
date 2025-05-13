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

// Handle stream for EXPLORER protocol
// In Nesterenko, Tixeuil, Discovering Network Topology in the Presence of Byzantine Faults, 2009
// This is the accept action
// WARNING: Explorer protocol has been proven to be wrong

func handleExplorer(s network.Stream, ctx context.Context, thisNode host.Host, top *Topology, messageContainer *MessageContainer) error {

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
	
	// If m.sender is in cTop, then add it to uTop
	if !top.ctop.checkInCTop(m.Source) {
		// If part1 Else If part 2
		if !top.utop.checkInUTop(m.Source, m.Neighbourhood) {
			// Add the sender to the visited set of m and send to all peers
			m.Path = append(m.Path, m.Sender)
			sendExplorer(ctx, thisNode, m)
		} else if top.utop.checkInUTop(m.Source, m.Neighbourhood) && 
					len(Intersect(top.utop.tuples[m.Source][1], append(m.Path, m.Sender))) >= (2*MAX_BYZANTINES)+1 {

			// Add m to cTop
			top.ctop.AddNeighbourhood(m.Source, m.Neighbourhood)
			fmt.Printf("Explorer - Neighbourhood of %s added to the cTop\n\n", addressToPrint(m.Source, NODE_PRINTLAST) )
			// empty the visited set and send m to all peers
			m.Path = make([]string, 0)
			sendExplorer(ctx, thisNode, m)
		}	
	}
	top.utop.AddElement(m.Source, m.Neighbourhood, m.Path)
	fmt.Printf("\nExplorer - node %s wants to share its topology\n", addressToPrint(m.Source, NODE_PRINTLAST))
	fmt.Printf("Explorer Message ID: %s\n", m.ID)
	fmt.Printf("Explorer - Neighbourhood of %s added to the uTop\n\n", addressToPrint(m.Source, NODE_PRINTLAST) )
	messageContainer.Add(m)
	printShell()	
	
	return nil
}
	
	

// Send an explorer message
// WARNING: Explorer protocol has been proven to be wrong
func sendExplorer(ctx context.Context, thisNode host.Host, exp_msg Message) {
	exp_msg.Sender = getNodeAddress(thisNode, ADDR_DEFAULT)
	dataBytes, err := json.Marshal(exp_msg)
	if err != nil {
		printError(err)
	}
	msg := string(dataBytes)

	// Cycle through the peers connected to the current node
	for _, p := range thisNode.Network().Peers() {

		// open a new stream
		stream, err := openStream(ctx, thisNode, p, PROTOCOL_EXP)
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