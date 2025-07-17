package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/multiformats/go-multiaddr"
)

// Gets a string with the complete address of a node
// In the h.Addrs() array, the element in pos 3 (= ADDR_LAN_POS) is the LAN address
// and the element in pos 0 (=ADDR_LB_POS) is the loopback address
// The address is returned in the format "<ADDRESS>/p2p/<PEER_ID>"
func getNodeAddress(h host.Host, ADDRESS_TYPE string) string {
	if ADDRESS_TYPE == ADDR_LAN {
		return fmt.Sprintf("%s/p2p/%s", h.Addrs()[ADDR_LAN_POS], h.ID())
	} else if ADDRESS_TYPE == ADDR_LOOPBACK {
		return fmt.Sprintf("%s/p2p/%s", h.Addrs()[ADDR_LB_POS], h.ID())
	}
	return ""
	/*
	for _, addr := range h.Addrs() {
		if addr.Protocols()[0].Code == multiaddr.P_IP4 && addr.Protocols()[1].Code == multiaddr.P_TCP {
			return fmt.Sprintf("%s/p2p/%s", addr, h.ID())
		}
	}
	return ""
	*/
}

// Extract Peer ID from a string Multiaddress
func extractPeerIDFromMultiaddr(addr string) string {
    parts := strings.Split(addr, "/p2p/")
    if len(parts) == 2 {
        return parts[1]
    }
    return addr // fallback: return as is if not a multiaddr
}


// Creates a node
func createNode() host.Host {
	node, err := libp2p.New()
	if err != nil {
		printError(err)		
	}

	return node
}


// Retrieves the stream endpoints and returns their addresses
// This doesn't work well since senderAddrs[] doesn't have a fix value for the addresses I want,
// But may get be useful in the future
//lint:ignore U1000 Unused function for future use
func getStreamEndpoints(s network.Stream, h host.Host, ADDRESS_TYPE string) string {
	// Get the sender peer
	conn := s.Conn()
	senderID := conn.RemotePeer()

	// Retrieve the addresses of the sender from the peerstore
	senderAddrs := h.Peerstore().Addrs(senderID)
	for i, a := range(senderAddrs) {
		fmt.Printf("%d - %s\n", i, a)
	}

	if ADDRESS_TYPE == ADDR_LAN {
		return fmt.Sprintf("%s/p2p/%s",senderAddrs[ADDR_LAN_POS], senderID)
	} else if ADDRESS_TYPE == ADDR_LOOPBACK {
		return fmt.Sprintf("%s/p2p/%s", senderAddrs[ADDR_LB_POS], senderID)
	}
	return ""
}

// Run the node by setting the stream handler 
func runNode(ctx context.Context, h host.Host, messageContainer *MessageContainer, 
			deliveredMessages *MessageContainer, sentMessages *MessageContainer, 
			disjointPaths *DisjointPaths, topology *Topology) {
	fmt.Println("Running node: ", getNodeAddress(h, ADDR_DEFAULT))
	fmt.Println("Tolerating max byzantines: ", MAX_BYZANTINES)
	topology.nodeID = getNodeAddress(h, ADDR_DEFAULT)

	// Set stream handler for direct messages
	h.SetStreamHandler(PROTOCOL_CHAT, func (s network.Stream)  {
		//fmt.Println("/chat/1.0.0 stream created for node ", getNodeAddress(h))
		err := handleStream(s, messageContainer)
		if err != nil {
			s.Reset()
		} else {
			s.Close()
		}
	})

	// Set stream handler for Naive Broadcast messages
	h.SetStreamHandler(PROTOCOL_NAB, func (s network.Stream)  {
		err := handleBroadcast(s, ctx, h, messageContainer)
		if err != nil {
			s.Reset()
		} else {
			s.Close()
		}
	})

	// Set stream handler for Detector messages
	h.SetStreamHandler(PROTOCOL_DET, func (s network.Stream)  {
		err := handleDetector(s, ctx, h, topology, messageContainer)
		if err != nil {
			s.Reset()
		} else {
			s.Close()
		}
	})

	// Set stream handler for master-slave messages
	h.SetStreamHandler(PROTOCOL_MST, func (s network.Stream)  {
		err := handleMaster(s, ctx, h, messageContainer, deliveredMessages, topology, disjointPaths)
		if err != nil {
			s.Reset()
		} else {
			s.Close()
		}
	})
	
	// Set stream handler for combinedRC messages
	h.SetStreamHandler(PROTOCOL_CRC, func (s network.Stream)  {
		err := handleCombinedRC(s, ctx, h, topology, messageContainer, deliveredMessages, sentMessages, disjointPaths)
		if err != nil {
			s.Reset()
		} else {
			s.Close()
		}
	})

	printStartMessage(h, mod_help_prot)
	printNodeInfo(h)
}


// Run the node by setting the stream handler 
func runNode_knownTopology(ctx context.Context, h host.Host, messageContainer *MessageContainer, 
						deliveredMessages *MessageContainer, sentMessages *MessageContainer,
						disjointPaths *DisjointPaths, topology *Topology) {
	fmt.Println("Running node: ", getNodeAddress(h, ADDR_DEFAULT))
	fmt.Println("Tolerating max byzantines: ", MAX_BYZANTINES)
	topology.nodeID = getNodeAddress(h, ADDR_DEFAULT)

	// Set stream handler for direct messages
	h.SetStreamHandler(PROTOCOL_CHAT, func (s network.Stream)  {
		//fmt.Println("/chat/1.0.0 stream created for node ", getNodeAddress(h))
		err := handleStream(s, messageContainer)
		if err != nil {
			s.Reset()
		} else {
			s.Close()
		}
	})

	// Set stream handler for Naive Broadcast messages
	h.SetStreamHandler(PROTOCOL_NAB, func (s network.Stream)  {
		err := handleBroadcast(s, ctx, h, messageContainer)
		if err != nil {
			s.Reset()
		} else {
			s.Close()
		}
	})

	// Set stream handler for Detector messages
	h.SetStreamHandler(PROTOCOL_DET, func (s network.Stream)  {
		err := handleDetector(s, ctx, h, topology, messageContainer)
		if err != nil {
			s.Reset()
		} else {
			s.Close()
		}
	})

	// Set stream handler for master-slave messages
	h.SetStreamHandler(PROTOCOL_MST, func (s network.Stream)  {
		err := handleMaster(s, ctx, h, messageContainer, deliveredMessages, topology, disjointPaths)
		if err != nil {
			s.Reset()
		} else {
			s.Close()
		}
	})

	// Set stream handler for combinedRC messages
	h.SetStreamHandler(PROTOCOL_CRC, func (s network.Stream)  {
		err := handleCombinedRC(s, ctx, h, topology, messageContainer, deliveredMessages, sentMessages, disjointPaths)
		if err != nil {
			s.Reset()
		} else {
			s.Close()
		}
	})

	// Load the neighbourhood in cTop from a file
	topology_graph := LoadGraphFromCSV(topology_path)
	topology.ctop.loadNeigh(topology_graph, getNodeAddress(h, ADDR_DEFAULT))

	printStartMessage(h, mod_help_prot)
	printNodeInfo(h)
}

// Connects two nodes
func connectNodes(ctx context.Context, sourceNode host.Host, targetNode_address string, topology *Topology) {
	
	// Turn the destination into a multiaddr.
	targetNode_maddr, err := multiaddr.NewMultiaddr(targetNode_address)
	if err != nil {
		printError(err)		
	}

	// Extract the peer ID from the multiaddr.
	targetNode_info, err := peer.AddrInfoFromP2pAddr(targetNode_maddr)
	if err != nil {
		printError(err)		
	}

	// Add the destination's peer multiaddress in the peerstore.
	// This will be used during connection and stream creation by libp2p.
	sourceNode.Peerstore().AddAddrs(targetNode_info.ID, targetNode_info.Addrs, peerstore.PermanentAddrTTL)

	// Connect the source node with the other node
	err = sourceNode.Connect(ctx, *targetNode_info)
	if err != nil {
		printError(err)
	}

	// Add the connection to the topology
	if targetNode_address != master_address {
		topology.ctop.AddNeighbour(getNodeAddress(sourceNode, ADDR_DEFAULT), targetNode_address)
	}

	// Print the mischief
	printResult := fmt.Sprintf("Connection established between \n - Node %s \n - Node %s\n", getNodeAddress(sourceNode, ADDR_DEFAULT), targetNode_address)
	fmt.Println(printResult)

}


// Disconnects two nodes
func disconnectNodes(ctx context.Context, sourceNode host.Host, targetNode_address string) {
	// Turn the destination into a multiaddr.
	targetNode_maddr, err := multiaddr.NewMultiaddr(targetNode_address)
	if err != nil {
		printError(err)		
	}

	// Extract the peer ID from the multiaddr.
	targetNode_info, err := peer.AddrInfoFromP2pAddr(targetNode_maddr)
	if err != nil {
		printError(err)		
	}

	// Disconnect the source node with the other node
	err = sourceNode.Network().ClosePeer(targetNode_info.ID)
	if err != nil {
		printError(err)
	}

	printResult := fmt.Sprintf("Connection closed between \n - Node %s \n - Node %s\n", getNodeAddress(sourceNode, ADDR_DEFAULT), targetNode_address)
	fmt.Println(printResult)

}

// Connects this node with all the nodes in the topology
func connectAllNodes(ctx context.Context, sourceNode host.Host, topology *Topology) {
	for _, node := range topology.ctop.GetNeighbourhood(getNodeAddress(sourceNode, ADDR_DEFAULT)) {
		connectNodes(ctx, sourceNode, node, topology)
	}
}


// Counts the connections of a node
//lint:ignore U1000 Unused function for future use
func countNodePeers(sourceNode host.Host) int {
	return len(sourceNode.Network().Peers())
}

// Acquire the topology from the network and load it in the cTop
func acquireTopology(h host.Host, topology *Topology) {
	peers := h.Network().Peers()
    for _, peer := range peers {
        // Obtain the ip4 tcp address from the peer
        var peer_address string		
        for _, addr := range h.Peerstore().Addrs(peer) {
            if addr.Protocols()[0].Code == multiaddr.P_IP4 && addr.Protocols()[1].Code == multiaddr.P_TCP {
                peer_address = fmt.Sprintf("%s/p2p/%s", addr, peer)
                break
            }
        }

        // Add the connection to the topology. Do not add the master address
        if len(peer_address) > 0 && peer_address != master_address {
            topology.ctop.AddNeighbour(getNodeAddress(h, ADDR_DEFAULT), peer_address)
        }
    }
}

// Test function
func test() {
	fmt.Println("Hello World")
}