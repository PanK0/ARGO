package main

import (
	"context"

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
					messageContainer *MessageContainer, deliveredMessages *MessageContainer, sentMessages *MessageContainer) error {

	return nil
}

// Send function for CombinedRC protocol
//lint:ignore U1000 Unused function for future use
func sendCombinedRC_exp2(ctx context.Context, thisNode host.Host, exp_msg Message, sentMessages *MessageContainer) {
	sendExplorer2(ctx, thisNode, exp_msg)
	sentMessages.Add(exp_msg)
	
}

