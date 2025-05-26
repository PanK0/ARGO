package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

// Send a message from a node to another
func sendMessage(ctx context.Context, thisNode host.Host, msg string) {

	// Transform the message into a json object with all the information
	var m Message
	err := json.Unmarshal([]byte(msg), &m)
	if err != nil {
		printError(err)
	}

	// Turn the destination into a multiaddr.
	targetNode_maddr, err := multiaddr.NewMultiaddr(m.Target)
	if err != nil {
		printError(err)
	}

	// Extract the peer ID from the multiaddr.
	targetNode_info, err := peer.AddrInfoFromP2pAddr(targetNode_maddr)
	if err != nil {
		printError(err)
	}

	// Check if the target peer is the same as the current node
    if targetNode_info.ID == thisNode.ID() {
        fmt.Println("Warning: Attempted to send a message to self. Ignoring.")
        return
    }

/*
	!! PROBLEM !!
	When opening a new stream, old ones are no more used.
	If we open a lot of streams (517?) the program explodes because it can't handle all.
	So I created a provisional function openStream() that closes an already existent stream
	and returns a new one.

	stream, err := thisNode.NewStream(ctx, targetNode_info.ID, PROTOCOL_CHAT)
*/
	stream, err := openStream(ctx, thisNode, targetNode_info.ID, PROTOCOL_CHAT)
	if err != nil {
		printError(err)
	}

	defer stream.Close()

	message := fmt.Sprintf("%s\n", msg)
	//fmt.Printf("Sending message...")
	_, err = stream.Write([]byte(message))
	if err != nil {
		printError(err)
	}
}

// Function to handle the stream
func handleStream(s network.Stream, messageContainer *MessageContainer) error {
	buf := bufio.NewReader(s)
	message, err := buf.ReadString('\n')
	if err != nil {
		return err
	}

	//connection := s.Conn()
	//remotePeer := connection.RemotePeer().String()

	// Transform the message into a json object with all the information
	var m Message
	err = json.Unmarshal([]byte(message), &m)
	if err != nil {
		printError(err)
	}

	// add the message to the dedicated data struct
	//receivedMessages.Add(message)
	messageContainer.Add(m)

	printMessage(message)

	return nil
}