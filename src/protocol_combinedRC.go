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

// Function to manage a CNT message
//lint:ignore U1000 Unused function for future use
func receive_CNT(ctx context.Context, thisNode host.Host, m *Message, top *Topology, messageContainer *MessageContainer, disjointPaths *DisjointPaths) error {
	messageContainer.Add(*m)
	/*
	event := fmt.Sprintf("receive_CRC_CNT - msg from %s added to MessageContainer", addressToPrint(m.Sender, NODE_PRINTLAST))
	logEvent(thisNode.ID().String(), PRINTOPTION, event)
	*/
	
	if m.Target == getNodeAddress(thisNode, ADDR_DEFAULT) {
		event := fmt.Sprintf("receive_CNT - Content from %s received from %s!", addressToPrint(m.Source, NODE_PRINTLAST), addressToPrint(m.Sender, NODE_PRINTLAST))
		logEvent(thisNode.ID().String(), PRINTOPTION, event)
		fmt.Print(msgToString(*m))
	} else {
		// Forward this message to the next node in the path
		thisPeer, idx := findElement(m.Path, getNodeAddress(thisNode, ADDR_DEFAULT))
		old_sender := m.Sender
		m.Sender = thisPeer
		
		if idx+1 >= len(m.Path) {
			event := fmt.Sprintf("receive_CNT - Invalid path index: idx+1=%d, len(m.Path)=%d", idx+1, len(m.Path))
			logEvent(thisNode.ID().String(), PRINTOPTION, event)
			return nil
		}
		// Turn the destination into a multiaddr
		peer_maddr, err := multiaddr.NewMultiaddr(m.Path[idx+1])
		if err != nil {
			printError(err)
		}

		// Extract the peer ID from the multiaddr
		peer_info, err := peer.AddrInfoFromP2pAddr(peer_maddr)
		if err != nil {
			printError(err)
		}

		stream, err := openStream(ctx, thisNode, peer_info.ID, PROTOCOL_CRC)
		if err != nil {
			printError(err)
		}

		dataBytes, err := json.Marshal(m)
		if err != nil {
			printError(err)
		}
		msg := string(dataBytes)
		msg += "\n" 

		// Write the message on the stream
		_, err = stream.Write([]byte(msg))
		if err != nil {
			printError(err)
		}

		event := fmt.Sprintf("receive_CNT - Content from %s forwarded to %s",addressToPrint(old_sender, NODE_PRINTLAST), addressToPrint(m.Path[idx+1], NODE_PRINTLAST))
		logEvent(thisNode.ID().String(), PRINTOPTION, event)

	}

	return nil
}

// Function to manage a ROU message
func receive_ROU(ctx context.Context, thisNode host.Host, m *Message, top *Topology, messageContainer *MessageContainer, disjointPaths *DisjointPaths) error {
	disjointPaths.mu.Lock()
	defer disjointPaths.mu.Unlock()
	messageContainer.Add(*m)
	/*
	event := fmt.Sprintf("receive_CRC_ROU - msg from %s added to MessageContainer", addressToPrint(m.Sender, NODE_PRINTLAST))
	logEvent(thisNode.ID().String(), PRINTOPTION, event)
	*/

	if m.Target == getNodeAddress(thisNode, ADDR_DEFAULT) {
		// reverse the path and add it into DisjointPaths
		for i, j := 0, len(m.Path)-1; i < j; i, j = i+1, j-1 {
            m.Path[i], m.Path[j] = m.Path[j], m.Path[i]
        }
		disjointPaths.Add(m.Source, m.Path)
		event := fmt.Sprintf("receive_ROU - Route from %s added to DJP for %s", addressToPrint(m.Sender, NODE_PRINTLAST), addressToPrint(m.Source, NODE_PRINTLAST))
		logEvent(thisNode.ID().String(), PRINTOPTION, event)
	} else {
		// Forward this message to the next node in the path
		thisPeer, idx := findElement(m.Path, getNodeAddress(thisNode, ADDR_DEFAULT))
		old_sender := m.Sender
		m.Sender = thisPeer

		if idx+1 >= len(m.Path) {
			event := fmt.Sprintf("receive_ROU - Invalid path index: idx+1=%d, len(m.Path)=%d", idx+1, len(m.Path))
			logEvent(thisNode.ID().String(), PRINTOPTION, event)
			return nil
		}

		// Turn the destination into a multiaddr
		peer_maddr, err := multiaddr.NewMultiaddr(m.Path[idx+1])
		if err != nil {
			printError(err)
		}

		// Extract the peer ID from the multiaddr
		peer_info, err := peer.AddrInfoFromP2pAddr(peer_maddr)
		if err != nil {
			printError(err)
		}

		stream, err := openStream(ctx, thisNode, peer_info.ID, PROTOCOL_CRC)
		if err != nil {
			printError(err)
		}

		dataBytes, err := json.Marshal(m)
		if err != nil {
			printError(err)
		}
		msg := string(dataBytes)
		msg += "\n" 

		// Write the message on the stream
		_, err = stream.Write([]byte(msg))
		if err != nil {
			printError(err)
		}

		event := fmt.Sprintf("receive_ROU - Route from %s forwarded to %s", addressToPrint(old_sender, NODE_PRINTLAST), addressToPrint(m.Path[idx+1], NODE_PRINTLAST))
		logEvent(thisNode.ID().String(), PRINTOPTION, event)

	}

	return nil
}


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
		err = receive_EXP(ctx, thisNode, &m, top, messageContainer, deliveredMessages)
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

// Send function for CombinedRC ROU messages
func send_CRC_ROU(ctx context.Context, thisNode host.Host, m Message, top *Topology, disjointPaths *DisjointPaths) {

	// Add the sender
	m.Sender = getNodeAddress(thisNode, ADDR_DEFAULT)
	m.Neighbourhood = []string{}

	// Create a graph
	g := generateGraph(top, mod_graph_byz)
	//g.PrintGraph()

	// Find Disjoint Paths
	disjointPaths.MergeDP(g.GetDisjointPaths(m.Target, m.Source))

	// Send routed messages to target node
	for _, path := range disjointPaths.paths[m.Target] {

		m.Path = path
		if len(path) <= 1 {
			event := fmt.Sprintf("send_CRC_ROU - Invalid path length: %d (need at least 2)", len(path))
			logEvent(thisNode.ID().String(), PRINTOPTION, event)
			continue
		}

		// Turn the destination into a multiaddr
		peer_maddr, err := multiaddr.NewMultiaddr(path[1])
		if err != nil {
			printError(err)
			continue
		}

		// Extract the peer ID from the multiaddr
		peer_info, err := peer.AddrInfoFromP2pAddr(peer_maddr)
		if err != nil {
			printError(err)
			continue
		}

		stream, err := openStream(ctx, thisNode, peer_info.ID, PROTOCOL_CRC)
		if err != nil {
			printError(err)
			continue
		}

		dataBytes, err := json.Marshal(m)
		if err != nil {
			printError(err)
		}
		msg := string(dataBytes) 
		msg += "\n"

		// Write the message on the stream
		_, err = stream.Write([]byte(msg))
		if err != nil {
			printError(err)
		}

		event := fmt.Sprintf("send_ROU - Route sent to %s for %s", addressToPrint(path[1], NODE_PRINTLAST), addressToPrint(m.Target, NODE_PRINTLAST))
		logEvent(thisNode.ID().String(), PRINTOPTION, event)

	}
}


// Send function for CombinedRC CNT messages
func send_CRC_CNT(ctx context.Context, thisNode host.Host, m Message, top *Topology, disjointPaths *DisjointPaths) {

	// Add the sender
	m.Sender = getNodeAddress(thisNode, ADDR_DEFAULT)
	m.Neighbourhood = []string{}

	// Send routed messages to target node
	for _, path := range disjointPaths.paths[m.Target] {
		if len(path) <= 1 {
			event := fmt.Sprintf("send_CRC_CNT - Invalid path length: %d (need at least 2)", len(path))
			logEvent(thisNode.ID().String(), PRINTOPTION, event)
			continue
		}

		m.Path = path

		// Turn the destination into a multiaddr
		peer_maddr, err := multiaddr.NewMultiaddr(path[1])
		if err != nil {
			printError(err)
			continue
		}

		// Extract the peer ID from the multiaddr
		peer_info, err := peer.AddrInfoFromP2pAddr(peer_maddr)
		if err != nil {
			printError(err)
			continue
		}

		stream, err := openStream(ctx, thisNode, peer_info.ID, PROTOCOL_CRC)
		if err != nil {
			printError(err)
			continue
		}

		dataBytes, err := json.Marshal(m)
		if err != nil {
			printError(err)
		}
		msg := string(dataBytes) 
		msg += "\n"

		// Write the message on the stream
		_, err = stream.Write([]byte(msg))
		if err != nil {
			printError(err)
		}

		event := fmt.Sprintf("send_CNT - Content sent to %s for %s", addressToPrint(path[1], NODE_PRINTLAST), addressToPrint(m.Target, NODE_PRINTLAST))
		logEvent(thisNode.ID().String(), PRINTOPTION, event)

	}
}

// Send function for CombinedRC protocol
func sendCombinedRC(ctx context.Context, thisNode host.Host, m Message, top *Topology, disjointPaths *DisjointPaths) {

	if m.Type == TYPE_CRC_EXP {
		sendExplorer2(ctx, thisNode, m, PROTOCOL_CRC)
	} else if m.Type == TYPE_CRC_ROU {
		send_CRC_ROU(ctx, thisNode, m, top, disjointPaths)
	} else if m.Type == TYPE_CRC_CNT {
		send_CRC_CNT(ctx, thisNode, m, top, disjointPaths)
	} else {
		fmt.Println("Set correct CRC type")
	}
	
}