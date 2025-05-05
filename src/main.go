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
	
	dest 	:= flag.String("d", "", "Destination multiaddr string of the master node")
	mod 	:= flag.String("m", "", "Start in auto mod. Must be followed by a valid -n value")
	nod		:= flag.String("n", "", "Replace node")
	help 	:= flag.Bool("help", false, "Display help")

	flag.Parse()

	if *help {
		fmt.Printf("This program demonstrates a simple p2p chat application using libp2p\n\n")

		os.Exit(0)
	}

	// Create the proper data structs
	topology := NewTopology()
	receivedMessages := NewMessageContainer()
	deliveredMessages := NewMessageContainer()
	h := createNode()


	if *mod == start_automatic && *nod != "" && *dest != "" {
		master_address = *dest
		ReplaceInCSV(topology_path, getNodeAddress(h, ADDR_DEFAULT), *nod)
		runNode_knownTopology(ctx, h, receivedMessages, deliveredMessages, topology)
		connectNodes(ctx, h, *dest, topology)
		manageConsoleInput(ctx, h, receivedMessages, deliveredMessages, topology)
	} else if *mod == start_automatic && *nod != "" {
		ReplaceInCSV(topology_path, getNodeAddress(h, ADDR_DEFAULT), *nod)
		runNode_knownTopology(ctx, h, receivedMessages, deliveredMessages, topology)
		manageConsoleInput(ctx, h, receivedMessages, deliveredMessages, topology)
	} else if *mod == "" && *dest != "" {
		master_address = *dest
		runNode(ctx, h, receivedMessages, deliveredMessages, topology)
		connectNodes(ctx, h, *dest, topology)
		manageConsoleInput(ctx, h, receivedMessages, deliveredMessages, topology)
	} else {
		runNode(ctx, h, receivedMessages, deliveredMessages, topology)
		manageConsoleInput(ctx, h, receivedMessages, deliveredMessages, topology)
	}

	
	// Wait forever
	select {}
}


