package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
)

// Function to manage a CNT message
//lint:ignore U1000 Unused function for future use
func receive_CNT() {

}

// Function to manage a ROU message
//lint:ignore U1000 Unused function for future use
func receive_ROU() {

}


// Handle stream for CombinedRC protocol
//lint:ignore U1000 Unused function for future use
func handleCombinedRC(s network.Stream, ctx context.Context, thisNode host.Host, top *Topology, 
					messageContainer *MessageContainer, deliveredMessages *MessageContainer, sentMessages *MessageContainer,
					disjointPaths *DisjointPaths) error {

	// Read the buffer and extract the message
	buf := bufio.NewReader(s)
	message, err := buf.ReadString('\n')
	if err != nil {
		printError(err)
	}

	// Transform the message
	var m Message
	err = json.Unmarshal([]byte(message), &m)
	if err != nil {
		printError(err)
	}

	if m.Type == TYPE_CRC_EXP {
		err = receive_EXP(ctx, thisNode, &m, top, messageContainer, deliveredMessages)
		if err != nil {
			printError(err)
		}
	}



	return nil
}

// Send function for CombinedRC ROU messages
//lint:ignore U1000 Unused function for future use
func send_CRC_ROU(ctx context.Context, thisNode host.Host, m Message, top *Topology, disjointPaths *DisjointPaths) {

	// Add the sender
	m.Sender = getNodeAddress(thisNode, ADDR_DEFAULT)

	// Create a graph
	g := ConvertCTopToGraph(&top.ctop)
	g.PrintGraph()

	// Find Disjoint Paths
	disjointPaths.MergeDP(g.GetDisjointPaths(m.Target, m.Source))
	disjointPaths.PrintDP()

	// Fill the content
	m.Content = convertListToString(disjointPaths.paths[m.Target])
	fmt.Println(m.Content)

	/*
	// Send the message
	dataBytes, err := json.Marshal(m)
	if err != nil {
		printError(err)
	}
	msg := string(dataBytes)
	*/	

}

// Send function for CombinedRC protocol
func sendCombinedRC(ctx context.Context, thisNode host.Host, m Message, top *Topology, disjointPaths *DisjointPaths) {

	// if m.type == TYPE_CRC_CNT 
	//		check whether exists a dps between (thisNode, m.target) in DisjointPaths structure
	// 		and send a routed message among all the paths for dp(thisNode, m.target) 
	// else if m.type == TYPE_CRC_EXP 
	//		sendExplorer2()
	// else if m.type == TYPE_CRC_ROU
	//		send the computed dps between (thisNode, m.target) in DisjointPaths structure

	if m.Type == TYPE_CRC_EXP {
		sendExplorer2(ctx, thisNode, m, PROTOCOL_CRC)
	} else if m.Type == TYPE_CRC_ROU {
		send_CRC_ROU(ctx, thisNode, m, top, disjointPaths)
	}
	

}

