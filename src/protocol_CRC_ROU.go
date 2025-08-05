package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

// Function to manage a ROU message
func receive_ROU(ctx context.Context, thisNode host.Host, m *Message, top *Topology, messageContainer *MessageContainer, disjointPaths *DisjointPaths) error {
	messageContainer.Add(*m)

	if m.Target == getNodeAddress(thisNode, ADDR_DEFAULT) {
		// reverse the path and add it into DisjointPaths
		for i, j := 0, len(m.Path)-1; i < j; i, j = i+1, j-1 {
            m.Path[i], m.Path[j] = m.Path[j], m.Path[i]
        }
		disjointPaths.Add(m.Source, m.Path)
		event := fmt.Sprintf("receive_ROU %s - Route from %s added to DJP for %s",m.ID[len(m.ID)-5:], addressToPrint(m.Sender, NODE_PRINTLAST), addressToPrint(m.Source, NODE_PRINTLAST))
		logEvent(thisNode.ID().String(), PRINTOPTION, event)
	} else {
		// Forward this message to the next node in the path
		thisPeer, idx := findElement(m.Path, getNodeAddress(thisNode, ADDR_DEFAULT))
		old_sender := m.Sender
		m.Sender = thisPeer

		if idx+1 >= len(m.Path) {
			event := fmt.Sprintf("receive_ROU %s - Invalid path index: idx+1=%d, len(m.Path)=%d", m.ID[len(m.ID)-5:], idx+1, len(m.Path))
			logEvent(thisNode.ID().String(), PRINTOPTION, event)
			return nil
		}

		// Turn the destination into a multiaddr
		peer_maddr, err := multiaddr.NewMultiaddr(m.Path[idx+1])
		if err != nil {
			printError(err)
			return err
		}

		// Extract the peer ID from the multiaddr
		peer_info, err := peer.AddrInfoFromP2pAddr(peer_maddr)
		if err != nil {
			printError(err)
			return err
		}

		stream, err := openStream(ctx, thisNode, peer_info.ID, PROTOCOL_CRC)
		if err != nil {
			printError(err)
			return err
		}

		dataBytes, err := json.Marshal(m)
		if err != nil {
			printError(err)
			return err
		}
		msg := string(dataBytes)
		msg += "\n" 

		// Write the message on the stream
		_, err = stream.Write([]byte(msg))
		if err != nil {
			printError(err)
			return err
		}

		event := fmt.Sprintf("receive_ROU %s - Route from %s forwarded to %s", m.ID[len(m.ID)-5:], addressToPrint(old_sender, NODE_PRINTLAST), addressToPrint(m.Path[idx+1], NODE_PRINTLAST))
		logEvent(thisNode.ID().String(), PRINTOPTION, event)

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
	disjointPaths.MergeDP(g.GetDisjointPaths(m.Source, m.Target))
	fmt.Println(disjointPaths.toString())

	// Send routed messages to target node
	for _, path := range disjointPaths.paths[m.Target] {

		m.Path = path
		if len(path) <= 1 {
			event := fmt.Sprintf("send_CRC_ROU %s - Invalid path length: %d (need at least 2)", m.ID[len(m.ID)-5:], len(path))
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

		event := fmt.Sprintf("send_ROU %s - Route sent to %s for %s", m.ID[len(m.ID)-5:], addressToPrint(path[1], NODE_PRINTLAST), addressToPrint(m.Target, NODE_PRINTLAST))
		logEvent(thisNode.ID().String(), PRINTOPTION, event)

	}
}