package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

// function to manage an EXP2 message
func receive_EXP2(ctx context.Context, thisNode host.Host, m *Message, top *Topology,
		messageContainer *MessageContainer, deliveredMessages *MessageContainer) error {
	
	m.Content = time.Now().Format("05.00000")

	inst := addressToPrint(m.Sender, NODE_PRINTLAST)
	m.InstanceID += "_"+inst

	event := fmt.Sprintf("receive_EXP2 %s - Handling message from %s", m.Content, addressToPrint(m.Sender, NODE_PRINTLAST))
	logEvent(thisNode.ID().String(), PRINTOPTION, event)

	// TO DO: When the local neighbourhood changes, 
	// relay all received messages with m.ID to the new neighbours

	// Add the sender to the path
	m.Path = append(m.Path, m.Sender)

	marshalledMessage, err := json.Marshal(m)
	if err != nil {
		printError(err)
		return nil
	}
	printMessage(string(marshalledMessage))
	
	// Lock to ensure only one execution at a time
	explorer2Mutex.Lock()
	defer explorer2Mutex.Unlock()

	// Start counting BFT Logics
	timestamp_start := time.Now()

	// Modification 4: check whether m is in deliveredMessages
	if len(deliveredMessages.Get(m.ID)) == 0 {
		m.Target = getNodeAddress(thisNode, ADDR_DEFAULT)
		messageContainer.Add(*m)

		// Modification 1: check whether source is equal to sender
		if m.Source == m.Sender && len(m.Path) == 1 && m.Path[0] == m.Source {
			BFT_deliver_and_relay(ctx, thisNode, *messageContainer, *deliveredMessages, *m, top)
		} else if len(messageContainer.GetDisjointPathsBrute(m.ID)) > MAX_BYZANTINES  {
			BFT_deliver_and_relay(ctx, thisNode, *messageContainer, *deliveredMessages, *m, top)
		} else {
			// Send the message to all the nodes who never ever received the message
			for _, p := range thisNode.Network().Peers() {
				if p.String() == extractPeerIDFromMultiaddr(master_address) {continue}

				// Only forward the message if p is not in m.path or if it doesen't exist in any of the paths of the instances of m.ID that are present in messageContainer
				if !messageContainer.lookInPaths(m.ID, p.String()) && !contains(m.Path, p.String()) {
					m.Sender = getNodeAddress(thisNode, ADDR_DEFAULT)
					send(ctx, thisNode, p, *m, PROTOCOL_CRC)					
				} 
			}
		}
	} else {
		// Enters this if the message has already been delivered by the node
		del := BFT_deliver(*messageContainer, *deliveredMessages, *m, top)
		if  del {
			deliveredMessages.Add(*m)
			messageContainer.RemoveMessage(*m)
		} else {
			messageContainer.Add(*m)
		}
		/*
		event := fmt.Sprintf("receive_del_EXP2 %s - Message coming from %s, source %s delivered? %t", m.Content,addressToPrint(m.Sender, NODE_PRINTLAST), addressToPrint(m.Source, NODE_PRINTLAST), del)
		logEvent(thisNode.ID().String(), false, event)
		*/
	}

	timestamp_end := time.Now()
	event = fmt.Sprintf("BFT_execution %s - performed in time %f seconds", m.Content, timestamp_end.Sub(timestamp_start).Seconds())
	logEvent(thisNode.ID().String(), false, event)
	return nil
}

// Function to manage a CNT message
func receive_CNT(ctx context.Context, thisNode host.Host, m *Message, top *Topology, messageContainer *MessageContainer, disjointPaths *DisjointPaths) error {
	messageContainer.Add(*m)
	
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
	messageContainer.Add(*m)

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

func sendEXP2(ctx context.Context, thisNode host.Host, exp_msg Message) {
	
	// Add the sender
	exp_msg.Sender = getNodeAddress(thisNode, ADDR_DEFAULT)
	dataBytes, err := json.Marshal(exp_msg)
	if err != nil {
		printError(err)
	}
	msg := string(dataBytes)

	// Cycle through the peers connected to the current node
	for _, p := range thisNode.Network().Peers() {

		if p.String() == extractPeerIDFromMultiaddr(master_address) {
			continue // Do not send the message to the master node
		}

		// If the peer p is already in the path of the message, then do not forward the message 
		// then open a stream with p and send the message
		if (contains(exp_msg.Path, p.String())) {
			printShell()
		} else {
			stream, err := openStream(ctx, thisNode, p, PROTOCOL_CRC)
			if err != nil {
				printError(err)
			}

			message := fmt.Sprintf("%s\n", msg)

			// Write the message on the stream
			_, err = stream.Write([]byte(message))
			if err != nil {
				printError(err)
			}
			stream.Close() // <-- chiudi sempre lo stream dopo la scrittura

			event := fmt.Sprintf("send_EXP2 %s - Forwarded message from %s to node %s", exp_msg.Content, addressToPrint(exp_msg.Sender, NODE_PRINTLAST), addressToPrint(p.String(), NODE_PRINTLAST))
			logEvent(thisNode.ID().String(), PRINTOPTION, event)
		}
	}
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
		sendEXP2(ctx, thisNode, m)
	} else if m.Type == TYPE_CRC_ROU {
		send_CRC_ROU(ctx, thisNode, m, top, disjointPaths)
	} else if m.Type == TYPE_CRC_CNT {
		send_CRC_CNT(ctx, thisNode, m, top, disjointPaths)
	} else {
		fmt.Println("Set correct CRC type")
	}
	
}