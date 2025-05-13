package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

// For critical section
var streamMutex sync.Mutex
var explorer2Mutex sync.Mutex

// Checks in the list of streams of the connection between the current node and the target node.
// If a stream exists, return that stream.
// Else, create a new one.
// !! WARNING : this function closes already existant streams and opens a new one.
// !! this is made to avoid to reach the limit of streams for each connection
func openStream(ctx context.Context, thisNode host.Host, targetNode_info peer.ID, protocol protocol.ID) (network.Stream, error) {
	
	// Lock the function (critical section)
	streamMutex.Lock()
	// Unlock before exiting
	defer streamMutex.Unlock()

	// List of connections of the current node
	connections := thisNode.Network().Conns()
	
	// Cycle through the connections of the node until you find the one
	// where this node is connected with targetNode
	for _, c := range connections {

		if c.RemotePeer().String() == targetNode_info.String() {
			// IF there are no streams in the connection, open a new stream.
			// ELSE close the existent one and open a new stream
			if len(c.GetStreams()) == 0 {
				stream, err := thisNode.NewStream(ctx, targetNode_info, protocol)
				return stream, err
			} else {
				for _, s := range c.GetStreams() {
					if s.Conn().RemotePeer() == targetNode_info {
						s.Close()
						s, err := thisNode.NewStream(ctx, targetNode_info, protocol)
						return s, err
					}
				}
			}

		}
	}
	stream, err := thisNode.NewStream(ctx, targetNode_info, protocol)
	return stream, err
} 

// Generic send function, to send a message to a single given peer
func send(ctx context.Context, thisNode host.Host, targetNode peer.ID, m Message, protocol protocol.ID) {

	dataBytes, err := json.Marshal(m)
	if err != nil {
		printError(err)
	}
	msg := string(dataBytes)

	// Open stream
	stream, err := openStream(ctx, thisNode, targetNode, protocol)
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

// Delivery Function for routed networks
// see @ Boosting the efficiency of byzantine-tolerant reliable communication - 2020, cpt 4.1
// Delivery is rooted to manageConsoleInput(), since the deliveredMessages data struct is given as a parameter
// in the main to function manageConsoleInput()
func dolevR_deliver(messageContainer *MessageContainer, msg_id string, deliveredMessages *MessageContainer) int { // {
	
	n := messageContainer.countNodeDisjointPaths(msg_id)
	fmt.Println("Message ID: ", msg_id)
	fmt.Println("Number of node disjoint paths: ", n)
	fmt.Println("Number of admissible Byzantine nodes: ", MAX_BYZANTINES)	
	
	// If n > 2 * MAX_BYZANTINES, then the message is delivered
	// Delivery conditions of Dolev_R are specified in the paper indicated above
	if n > 2 * MAX_BYZANTINES {
		// Add all the messages corresponding to msg_id to deliveredMessages
		messages := messageContainer.Get(msg_id)
		for _, m := range messages {
			deliveredMessages.Add(m)
		}
		// Delete the message from the messageContainer
		messageContainer.deleteElement(msg_id)
		fmt.Println("Message delivered")
	} else {
		fmt.Println("Message not delivered")
	}

	fmt.Println()
	return n
}