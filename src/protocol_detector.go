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

// Handler for Detector protocol
func handleDetector(s network.Stream, ctx context.Context, h host.Host, top *Topology, messageContainer *MessageContainer) error {

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
				logEvent(h.ID().String(), PRINTOPTION, event)
			time.Sleep(bz.Delay)
		}
		// If byzantine is of Type 2, then drop the message with bz.Droprate probability
		if bz.Type2 {
			if (rand.Float64() < bz.DropRate) {
				event := fmt.Sprintf("byzantine %s - Message from %s dropped", m.Content, addressToPrint(m.Sender, NODE_PRINTLAST))
				logEvent(h.ID().String(), PRINTOPTION, event)
				return nil
			}
		}
	}
	
	// BYZANTINE DETECTION
	// WARNING: DETECTOR protocol only works on STATIC NETWORKS, with FIXED TOPOLOGY.
	// This means that if the network topology changes, Detector will dectect a byzantine process even if there is none.
	// To avoid this, the network must be static and the topology must be fixed.
	// more @ Discovering Network Topology in the Presence of Byzantine Faults - 2009 - Nesterenko, Tixeuil - cpt. 6.3
	// Check whether m.Source is in the cTop AND m.Neighbourhood is different from the cTop's node neighbourhood
	if (top.ctop.checkInCTop(m.Source) && compareLists(top.ctop.GetNeighbourhood(m.Source), m.Neighbourhood) == -1) {
		fmt.Println("checkInCTop")
		printByzantineAlert()
		return nil
	}

	// Connectivity check
	// Create a temporary graph with the known neighbourhood of the source node and the topology of this node
	// Remember to add an edge between every pair of unexplored nodes, as in the paper
	temp_ctop := top.ctop.DeepCopy()
	temp_ctop.AddNeighbourhood(m.Source, m.Neighbourhood)
	// Add an edge between every pair of unexplored nodes
	var unexplored []string
	for n := range temp_ctop.tuples {
		for _, m := range temp_ctop.GetNeighbourhood(n) {
			if !temp_ctop.checkInCTop(m) {
				unexplored = append(unexplored, m)
			}
		}
	}
	for n := range unexplored {
		temp_ctop.AddNeighbourhood(unexplored[n], unexplored)
	}

	g := ConvertCTopToGraph(temp_ctop)
	connectivity := g.nodeConnectivity()

	if connectivity < MAX_BYZANTINES +1 {
		fmt.Printf("\nDetector - node %s wants to share its topology\n", addressToPrint(m.Source, NODE_PRINTLAST))
		messageContainer.Add(m)
		printByzantineAlert()
		return nil
	}

	// If the m.source node is not in the cTop, add it to the cTop with its neighbourhood
	if (!top.ctop.checkInCTop(m.Source)) {
		top.ctop.AddNeighbourhood(m.Source, m.Neighbourhood)
		fmt.Printf("\nDetector - node %s wants to share its topology\n", addressToPrint(m.Source, NODE_PRINTLAST))
		fmt.Printf("Detector Message ID: %s\n", m.ID)
		fmt.Printf("Detector - Neighbourhood of %s added to the cTop\n\n", addressToPrint(m.Source, NODE_PRINTLAST) )
		messageContainer.Add(m)
		printShell()
		sendDetector(ctx, h, m)
	}

	return nil
}


// Send a detector message
func sendDetector(ctx context.Context, thisNode host.Host, det_msg Message) {

	dataBytes, err := json.Marshal(det_msg)
	if err != nil {
		printError(err)
	}
	msg := string(dataBytes)

	// Cycle through the peers connected to the current node
	for _, p := range thisNode.Network().Peers() {

		// open a new stream
		stream, err := openStream(ctx, thisNode, p, PROTOCOL_DET)
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