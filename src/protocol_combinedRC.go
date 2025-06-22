package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
)

// Handle stream for CombinedRC protocol
func handleCombinedRC(s network.Stream, ctx context.Context, thisNode host.Host, top *Topology, 
					messageContainer *MessageContainer, deliveredMessages *MessageContainer, sentMessages *MessageContainer,
					disjointPaths *DisjointPaths) error {

	defer s.Close()

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

	// Apply byzantine modifications
	// returns true if byzantine is type 2 [drop messages], so this function must be stopped
	// returns false otherwise and applies changes to the message
	if applyByzantine(thisNode, &m) {return nil}

	if m.Type == TYPE_CRC_EXP {
		err = receive_EXP2(ctx, thisNode, &m, top, messageContainer, deliveredMessages)
		if err != nil {
			printError(err)
		}
	} else if m.Type == TYPE_CRC_ROU {
		err = receive_ROU(ctx, thisNode, &m, top, messageContainer, disjointPaths)
		if err != nil {
			printError(err)
		}
	} else if m.Type == TYPE_CRC_CNT {
		err = receive_CNT(ctx, thisNode, &m, top, messageContainer, disjointPaths)
		if err != nil {
			printError(err)
		}
	}

	return nil
}

// Send function for CombinedRC protocol
func sendCombinedRC(ctx context.Context, thisNode host.Host, m Message, top *Topology, disjointPaths *DisjointPaths) {

	if m.Type == TYPE_CRC_EXP {
		sendEXP2(ctx, thisNode, m)
	} else if m.Type == TYPE_CRC_ROU {
		send_CRC_ROU(ctx, thisNode, m, top, disjointPaths)
	} else if m.Type == TYPE_CRC_CNT {
		send_CRC_CNT(ctx, thisNode, m, top, disjointPaths)
	} else {
		fmt.Println("Set correct CRC type")
	}
	
}