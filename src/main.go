package main

import (
	"context"
	"flag"
	"fmt"
	"os"
)

/*
	TO DO
	-	WHEN A NODE IS CLOSED, CLOSE THE CONNECTION WITH ALL ITS PEERS
*/

func main() {

	// Create a context
	ctx, cancel := context.WithCancel(context.Background())
	// Cancel this function to avoid content leak
	defer cancel()

	dest := flag.String("d", "", "Destination multiaddr string of the master node")
	mod := flag.String("m", "", "Start in auto mod. Must be followed by a valid -n value")
	nod := flag.String("n", "", "Replace node")
	help := flag.Bool("help", false, "Display help")

	flag.Parse()

	if *help {
		fmt.Printf("This program demonstrates a simple p2p chat application using libp2p\n\n")

		os.Exit(0)
	}

	// Create the proper data structs
	readMaxByzantines(BYZANTINE_CONFIG, &MAX_BYZANTINES)
	topology := NewTopology()
	receivedMessages := NewMessageContainer()
	deliveredMessages := NewMessageContainer()
	sentMessages := NewMessageContainer()
	disjointPaths := NewDisjointPaths()
	h := createNode()
	thisnode_address = getNodeAddress(h, ADDR_DEFAULT)

	if *mod == start_automatic && *nod != "" && *dest != "" {
		master_address = *dest
		ReplaceInCSV(topology_path, getNodeAddress(h, ADDR_DEFAULT), *nod)
		runNode_knownTopology(ctx, h, receivedMessages, deliveredMessages, sentMessages, disjointPaths, topology)
		connectNodes(ctx, h, master_address, topology)
		sendAddressToMaster(ctx, h, *nod)
		manageConsoleInput(ctx, h, receivedMessages, deliveredMessages, disjointPaths, topology)
	} else if *mod == start_automatic && *nod != "" {
		ReplaceInCSV(topology_path, getNodeAddress(h, ADDR_DEFAULT), *nod)
		runNode_knownTopology(ctx, h, receivedMessages, deliveredMessages, sentMessages, disjointPaths, topology)
		manageConsoleInput(ctx, h, receivedMessages, deliveredMessages, disjointPaths, topology)
	} else if *mod == "" && *dest != "" {
		master_address = *dest
		runNode(ctx, h, receivedMessages, deliveredMessages, sentMessages, disjointPaths, topology)
		connectNodes(ctx, h, master_address, topology)
		if *nod != "" {sendAddressToMaster(ctx, h, *nod)}
		manageConsoleInput(ctx, h, receivedMessages, deliveredMessages, disjointPaths, topology)
	} else {
		runNode(ctx, h, receivedMessages, deliveredMessages, sentMessages, disjointPaths, topology)
		manageConsoleInput(ctx, h, receivedMessages, deliveredMessages, disjointPaths, topology)
	}

	// Wait forever
	select {}
}
