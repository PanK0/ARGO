package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
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

	message := fmt.Sprintf("%s\n", msg)
	//fmt.Printf("Sending message...")
	_, err = stream.Write([]byte(message))
	if err != nil {
		printError(err)
	}
}


// Handle a broadcast stream
// This is actually a Byzantine Reliable Broadcast, granted by DolevU protocol
// See @ Thesis Farina, 2021, CPT 6.1, Algorithm 1, PDF pg 42/142 
// + PDF pg 43/142, CPT 6.1.2 - DolevU Message Complexity - for performance analysis
func handleBroadcast(s network.Stream, ctx context.Context, thisNode host.Host, messageContainer *MessageContainer) error {
	buf := bufio.NewReader(s)
	message, err := buf.ReadString('\n')
	if err != nil {
		return err
	}

	// Transform the message into a json object with all the information
	var m Message
	err = json.Unmarshal([]byte(message), &m)
	if err != nil {
		printError(err)
	}

	// Byzantine checking
	if byzantine_status {
		// If byzantine is of Type 1, then sleep for bz.Delay milliseconds
		if bz.Type1 {
			event := fmt.Sprintf("byzantine %s - delay of %s ms", m.Content, bz.Delay)
				logEvent(thisNode.ID().String(), PRINTOPTION, event)
			time.Sleep(bz.Delay)
		}
		// If byzantine is of Type 2, then drop the message with bz.Droprate probability
		if bz.Type2 {
			if (rand.Float64() < bz.DropRate) {
				event := fmt.Sprintf("byzantine %s - Message from %s dropped", m.Content, addressToPrint(m.Sender, NODE_PRINTLAST))
				logEvent(thisNode.ID().String(), PRINTOPTION, event)
				return nil
			}
		}
	}

	// Add sender node to the path.
	// The PATH is represented as an array of strings
	// If this is the target node, then append the target to the path
	newPath := append(m.Path, m.Sender)
	if (m.Target == getNodeAddress(thisNode, ADDR_DEFAULT)) {
		newPath = append(newPath, m.Target)
	}
	m.Path = newPath

	// add the message to the dedicated data struct
	//receivedMessages.Add(string(new_message))
	messageContainer.Add(m)

	new_message, err := json.Marshal(m)
	if err != nil {
		fmt.Println("Error marshalling data while handling a message:", err)
	}

	printMessage(string(new_message))

	// If we are on the target node do not forward the message
	if (m.Target != getNodeAddress(thisNode, ADDR_DEFAULT)) {
		sendBroadcast(ctx, thisNode, string(new_message))
	}

	return nil
}

// Send a Broadcast message
// This is actually a Byzantine Reliable Broadcast, granted by DolevU protocol
// See @ Thesis Farina, 2021, CPT 6.1, Algorithm 1, PDF pg 42/142 
// + PDF pg 43/142, CPT 6.1.2 - DolevU Message Complexity - for performance analysis
func sendBroadcast(ctx context.Context, thisNode host.Host, msg string) {

	// Get the json object from content, where is stored all the information about the sender, the source and the destination
	var m Message
	err := json.Unmarshal([]byte(msg), &m)
	if err != nil {
		printError(err)
	}

	// Cycle through the peers connected to the current node
	for _, p := range thisNode.Network().Peers() {

		// If the peer p is already in the path of the message, then do not forward the message 
		// then create a stream with p and send the message

		if (contains(m.Path, p.String())) {
			fmt.Println("Do not forward on node ", p)
			printShell()
		} else {	
			stream, err := openStream(ctx, thisNode, p, PROTOCOL_NAB)
			if err != nil {
				printError(err)
			}

			// Change the sender into the content: the sender node is now this node, that can be different from the sourceNode
			m.Sender = getNodeAddress(thisNode, ADDR_DEFAULT)

			dataBytes, err := json.Marshal(m)
			if err != nil {
				fmt.Println("Error marshalling data while sending a message:", err)
			}

			// Lock the function (critical section)
			streamMutex.Lock()

			// Write the message on the stream 
			message := fmt.Sprintf("%s\n", string(dataBytes))
			_, err = stream.Write([]byte(message))
			if err != nil {
				printError(err)
			}

			// Unlock before exiting
			streamMutex.Unlock()
		}
		
	} 	
}


// Handler for Detector protocol
func handleDetector(s network.Stream, ctx context.Context, h host.Host, top *Topology, messageContainer *MessageContainer) error {

	// Read the buffer and extract the message
	buf := bufio.NewReader(s)
	message, err := buf.ReadString('\n')
	if err != nil {
		printError(err)
	}

	// Transform the message into a json object with all the information
	var m Message
	err = json.Unmarshal([]byte(message), &m)
	if err != nil {
		printError(err)
	}

	// Byzantine checking
	if byzantine_status {
		// If byzantine is of Type 1, then sleep for bz.Delay milliseconds
		if bz.Type1 {
			event := fmt.Sprintf("byzantine %s - delay of %s ms", m.Content, bz.Delay)
				logEvent(h.ID().String(), PRINTOPTION, event)
			time.Sleep(bz.Delay)
		}
		// If byzantine is of Type 2, then drop the message with bz.Droprate probability
		if bz.Type2 {
			if (rand.Float64() < bz.DropRate) {
				event := fmt.Sprintf("byzantine %s - Message from %s dropped", m.Content, addressToPrint(m.Sender, NODE_PRINTLAST))
				logEvent(h.ID().String(), PRINTOPTION, event)
				return nil
			}
		}
	}
	
	// BYZANTINE DETECTION
	// WARNING: DETECTOR protocol only works on STATIC NETWORKS, with FIXED TOPOLOGY.
	// This means that if the networktopology changes, Detector will dectect a byzantine process even if there is none.
	// To avoid this, the network must be static and the topology must be fixed.
	// more @ Discovering Network Topology in the Presence of Byzantine Faults - 2009 - Nesterenko, Tixeuil - cpt. 6.3
	// Check whether m.Source is in the cTop AND m.Neighbourhood is different from the cTop's node neighbourhood
	if (top.ctop.checkInCTop(m.Source) && compareLists(top.ctop.GetNeighbourhood(m.Source), m.Neighbourhood) == -1) {
		fmt.Println("checkInCTop")
		printByzantineAlert()
		return nil
	}

	// Connectivity check
	// Create a temporary graph with the known neighbourhood of the source node and the topology of this node
	// Remember to add an edge between every pair of unexplored nodes, as in the paper
	temp_ctop := top.ctop.DeepCopy()
	temp_ctop.AddNeighbourhood(m.Source, m.Neighbourhood)
	// Add an edge between every pair of unexplored nodes
	var unexplored []string
	for n := range temp_ctop.tuples {
		for _, m := range temp_ctop.GetNeighbourhood(n) {
			if !temp_ctop.checkInCTop(m) {
				unexplored = append(unexplored, m)
			}
		}
	}
	for n := range unexplored {
		temp_ctop.AddNeighbourhood(unexplored[n], unexplored)
	}

	g := ConvertCTopToGraph(temp_ctop)
	connectivity := g.nodeConnectivity()

	if connectivity < MAX_BYZANTINES +1 {
		fmt.Printf("\nDetector - node %s wants to share its topology\n", addressToPrint(m.Source, NODE_PRINTLAST))
		messageContainer.Add(m)
		printByzantineAlert()
		return nil
	}

	// If the m.source node is not in the cTop, add it to the cTop with its neighbourhood
	if (!top.ctop.checkInCTop(m.Source)) {
		top.ctop.AddNeighbourhood(m.Source, m.Neighbourhood)
		fmt.Printf("\nDetector - node %s wants to share its topology\n", addressToPrint(m.Source, NODE_PRINTLAST))
		fmt.Printf("Detector Message ID: %s\n", m.ID)
		fmt.Printf("Detector - Neighbourhood of %s added to the cTop\n\n", addressToPrint(m.Source, NODE_PRINTLAST) )
		messageContainer.Add(m)
		printShell()
		sendDetector(ctx, h, m)
	}

	return nil
}


// Send a detector message
func sendDetector(ctx context.Context, thisNode host.Host, det_msg Message) {

	dataBytes, err := json.Marshal(det_msg)
	if err != nil {
		printError(err)
	}
	msg := string(dataBytes)

	// Cycle through the peers connected to the current node
	for _, p := range thisNode.Network().Peers() {

		// open a new stream
		stream, err := openStream(ctx, thisNode, p, PROTOCOL_DET)
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
}


// Handle stream for EXPLORER protocol
// In Nesterenko, Tixeuil, Discovering Network Topology in the Presence of Byzantine Faults, 2009
// This is the accept action
// WARNING: Explorer protocol has been proven to be wrong

func handleExplorer(s network.Stream, ctx context.Context, thisNode host.Host, top *Topology, messageContainer *MessageContainer) error {

	// Read the buffer and extract the message
	buf := bufio.NewReader(s)
	message, err := buf.ReadString('\n')
	if err != nil {
		printError(err)
	}

	// Transform the message into a json object with all the information
	var m Message
	err = json.Unmarshal([]byte(message), &m)
	if err != nil {
		printError(err)
	}

	// Byzantine checking
	if byzantine_status {
		// If byzantine is of Type 1, then sleep for bz.Delay milliseconds
		if bz.Type1 {
			event := fmt.Sprintf("byzantine %s - delay of %s ms", m.Content, bz.Delay)
				logEvent(thisNode.ID().String(), PRINTOPTION, event)
			time.Sleep(bz.Delay)
		}
		// If byzantine is of Type 2, then drop the message with bz.Droprate probability
		if bz.Type2 {
			if (rand.Float64() < bz.DropRate) {
				event := fmt.Sprintf("byzantine %s - Message from %s dropped", m.Content, addressToPrint(m.Sender, NODE_PRINTLAST))
				logEvent(thisNode.ID().String(), PRINTOPTION, event)
				return nil
			}
		}
	}
	
	// If m.sender is in cTop, then add it to uTop
	if !top.ctop.checkInCTop(m.Source) {
		// If part1 Else If part 2
		if !top.utop.checkInUTop(m.Source, m.Neighbourhood) {
			// Add the sender to the visited set of m and send to all peers
			m.Path = append(m.Path, m.Sender)
			sendExplorer(ctx, thisNode, m)
		} else if top.utop.checkInUTop(m.Source, m.Neighbourhood) && 
					len(Intersect(top.utop.tuples[m.Source][1], append(m.Path, m.Sender))) >= (2*MAX_BYZANTINES)+1 {

			// Add m to cTop
			top.ctop.AddNeighbourhood(m.Source, m.Neighbourhood)
			fmt.Printf("Explorer - Neighbourhood of %s added to the cTop\n\n", addressToPrint(m.Source, NODE_PRINTLAST) )
			// empty the visited set and send m to all peers
			m.Path = make([]string, 0)
			sendExplorer(ctx, thisNode, m)
		}	
	}
	top.utop.AddElement(m.Source, m.Neighbourhood, m.Path)
	fmt.Printf("\nExplorer - node %s wants to share its topology\n", addressToPrint(m.Source, NODE_PRINTLAST))
	fmt.Printf("Explorer Message ID: %s\n", m.ID)
	fmt.Printf("Explorer - Neighbourhood of %s added to the uTop\n\n", addressToPrint(m.Source, NODE_PRINTLAST) )
	messageContainer.Add(m)
	printShell()	
	
	return nil
}
	
	

// Send an explorer message
// WARNING: Explorer protocol has been proven to be wrong
func sendExplorer(ctx context.Context, thisNode host.Host, exp_msg Message) {
	exp_msg.Sender = getNodeAddress(thisNode, ADDR_DEFAULT)
	dataBytes, err := json.Marshal(exp_msg)
	if err != nil {
		printError(err)
	}
	msg := string(dataBytes)

	// Cycle through the peers connected to the current node
	for _, p := range thisNode.Network().Peers() {

		// open a new stream
		stream, err := openStream(ctx, thisNode, p, PROTOCOL_EXP)
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
}


// Handle stream for EXPLORER2 protocol
// described @ `Tractable Reliable Communication in Compromised Networks, Giovanni Farina - cpt. 9.3, 9.4`
//lint:ignore U1000 Unused function for future use
func handleExplorer2(s network.Stream, ctx context.Context, thisNode host.Host, top *Topology, 
					messageContainer *MessageContainer, deliveredMessages *MessageContainer) error {

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

	m.Content = time.Now().Format("05.00000")
	event := fmt.Sprintf("handle %s - Handling message from %s", m.Content, addressToPrint(m.Sender, NODE_PRINTLAST))
	logEvent(thisNode.ID().String(), PRINTOPTION, event)

	// TO DO: When the local neighbourhood changes, 
	// relay all received messages with m.ID to the new neighbours

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

	// Byzantine checking
	if byzantine_status {
		// If byzantine is of Type 1, then sleep for bz.Delay milliseconds
		if bz.Type1 {
			event := fmt.Sprintf("byzantine %s - delay of %s ms", m.Content, bz.Delay)
			logEvent(thisNode.ID().String(), PRINTOPTION, event)
			time.Sleep(bz.Delay)
		}
		// If byzantine is of Type 2, then drop the message with bz.Droprate probability
		if bz.Type2 {
			if (rand.Float64() < bz.DropRate) {
				event := fmt.Sprintf("byzantine %s - Message from %s dropped", m.Content, addressToPrint(m.Sender, NODE_PRINTLAST))
				logEvent(thisNode.ID().String(), PRINTOPTION, event)
				return nil
			}
		}
		// If byzantine is of Type 3, then remove one random element from the neighbourhood or path
		if bz.Type3 {
			if bz.Alterations == BYZ_NEIGHBOURHOOD {
				if len(m.Neighbourhood) > 0 {
					// Remove a random element from the neighbourhood
					rand.Seed(time.Now().UnixNano())
					index := rand.Intn(len(m.Neighbourhood))
					m.Neighbourhood = append(m.Neighbourhood[:index], m.Neighbourhood[index+1:]...)
					event := fmt.Sprintf("byzantine %s - Message from %s altered. Removed %s from neighbourhood.", m.Content, addressToPrint(m.Sender, NODE_PRINTLAST), addressToPrint(m.Neighbourhood[index], NODE_PRINTLAST))
					logEvent(thisNode.ID().String(), PRINTOPTION, event)
				} 
			} else if bz.Alterations == BYZ_PATH {
				if len(m.Path) > 0 {
					// Remove a random element from the path
					rand.Seed(time.Now().UnixNano())
					index := rand.Intn(len(m.Path))
					m.Path = append(m.Path[:index], m.Path[index+1:]...)
					event := fmt.Sprintf("byzantine %s - Message from %s altered. Removed %s from path.", m.Content, addressToPrint(m.Sender, NODE_PRINTLAST), addressToPrint(m.Path[index], NODE_PRINTLAST))
					logEvent(thisNode.ID().String(), PRINTOPTION, event)
				}
		}
		}
	}


	// Check whether m is in deliveredMessages
	// Modification 4
	if len(deliveredMessages.Get(m.ID)) == 0 {
		// if m arrived with a void path, means that m.Sender delivered the message
		// Add the message m to the message container
		m.Target = getNodeAddress(thisNode, ADDR_DEFAULT)
		messageContainer.Add(m)
		/*
		event := fmt.Sprintf("handle %s - Message from %s added to message container", m.Content, addressToPrint(m.Sender, NODE_PRINTLAST))
		logEvent(thisNode.ID().String(), PRINTOPTION, event)
		old_sender := m.Sender
		*/

		// Check whether m.Source == m.Sender
		// Modification 1
		if m.Source == m.Sender {
			// DELIVER AND RELAY
			BFT_deliver_and_relay(ctx, thisNode, *messageContainer, *deliveredMessages, m, top)
		} else if len(messageContainer.countNodeDisjointPaths_intersection(m.ID)) > MAX_BYZANTINES  {
			// Count the node disjoint paths of a message with a certain msg_id
			// if n_disj_paths > MAX_BYZANTINES then deliver and relay
			BFT_deliver_and_relay(ctx, thisNode, *messageContainer, *deliveredMessages, m, top)
	
		} else {
			// Send the message to all the nodes who never ever received the message
			
			// Forward the message
			for _, p := range thisNode.Network().Peers() {
				// Only forward the message if p is not in m.path 
				// or if it doesen't exist in any of the paths of the instances of m.ID that are present in messageContainer
				if !messageContainer.lookInPaths(m.ID, p.String()) && !contains(m.Path, p.String()){
					// send the message
					/*
					event = fmt.Sprintf("handle %s - Node %s in no paths: send message sent by %s", m.Content, addressToPrint(p.String(), NODE_PRINTLAST), addressToPrint(old_sender, NODE_PRINTLAST))
					logEvent(thisNode.ID().String(), PRINTOPTION, event)
					*/
					m.Sender = getNodeAddress(thisNode, ADDR_DEFAULT)
					send(ctx, thisNode, p, m, PROTOCOL_EXP2)					
				} 
			}
		}
	} else {
		
		deliveredMessages.Add(m)
		
		BFT_deliver(*messageContainer, *deliveredMessages, m, top)
		event := fmt.Sprintf("handle %s - delivering message from %s", m.Content, addressToPrint(m.Sender, NODE_PRINTLAST))
		logEvent(thisNode.ID().String(), PRINTOPTION, event)
	}
	return nil
}

// Send an EXPLORER2 message
// described @ `Tractable Reliable Communication in Compromised Networks, Giovanni Farina - cpt. 9.3, 9.4`
//lint:ignore U1000 Unused function for future use
func sendExplorer2(ctx context.Context, thisNode host.Host, exp_msg Message) {
	
	// Add the sender
	exp_msg.Sender = getNodeAddress(thisNode, ADDR_DEFAULT)

	/*
	// Add the sender (thisNode) to the visited set (path)
	// Only if the path contains no elements
	// OR the last element of the path is not this node
	if 	len(exp_msg.Path) == 0 || 
		(len(exp_msg.Path) > 0 && exp_msg.Path[len(exp_msg.Path)-1] != getNodeAddress(thisNode, ADDR_DEFAULT)) {
		exp_msg.Path = append(exp_msg.Path, getNodeAddress(thisNode, ADDR_DEFAULT))
	}
		*/

	dataBytes, err := json.Marshal(exp_msg)
	if err != nil {
		printError(err)
	}
	msg := string(dataBytes)

	// Cycle through the peers connected to the current node
	for _, p := range thisNode.Network().Peers() {

		// If the peer p is already in the path of the message, then do not forward the message 
		// then open a stream with p and send the message
		if (contains(exp_msg.Path, p.String())) {
			printShell()
		} else {
			stream, err := openStream(ctx, thisNode, p, PROTOCOL_EXP2)
			if err != nil {
				printError(err)
			}

			message := fmt.Sprintf("%s\n", msg)

			// Write the message on the stream
			_, err = stream.Write([]byte(message))
			if err != nil {
				printError(err)
			}
			event := fmt.Sprintf("sendExp2 %s - Forwarded message from %s to node %s", exp_msg.Content, addressToPrint(exp_msg.Sender, NODE_PRINTLAST), addressToPrint(p.String(), NODE_PRINTLAST))
			logEvent(thisNode.ID().String(), PRINTOPTION, event)
		}
	}
}

// Delivery function for BFT
func BFT_deliver(messageContainer MessageContainer, deliveredMessages MessageContainer, m Message, top *Topology) {
	// Add the message to the delivered messages
	messages := messageContainer.Get(m.ID)

	// Topology update
	// If m.Source is not in the cTop, add it to the cTop with its neighbourhood
	// If m.Source is already in cTop and m.Neighbourhood is a super set of the current registered neighbourhood in cTop, then substitute the neighbourhood
	// If m.Source is already in cTop and m.Neighbourhood is a subset of the current registered neighbourhood in cTop, then move m.Source and its neighbourhood from cTop to uTop
	if !top.ctop.checkInCTop(m.Source) {
		top.ctop.AddNeighbourhood(m.Source, m.Neighbourhood)
	} else if top.ctop.checkInCTop(m.Source) &&
		isSubSet(top.ctop.GetNeighbourhood(m.Source), m.Neighbourhood) == 0 {
		top.ctop.AddNeighbourhood(m.Source, m.Neighbourhood)
	} else if top.ctop.checkInCTop(m.Source) &&
		isSubSet(m.Neighbourhood, top.ctop.GetNeighbourhood(m.Source)) == 0 {
		top.utop.AddElement(m.Source, top.ctop.GetNeighbourhood(m.Source), m.Path)
		top.ctop.RemoveElement(m.Source)
	}

	// Add m to delivered messages
	for _, m := range messages {
		deliveredMessages.Add(m)
	}
	// Delete the message from the messageContainer
	messageContainer.deleteElement(m.ID)
}

// Delivery function for BFT
func BFT_deliver_and_relay(ctx context.Context, thisNode host.Host, 
							messageContainer MessageContainer, deliveredMessages MessageContainer,
							m Message, top *Topology) {

	BFT_deliver(messageContainer, deliveredMessages, m, top)
	event := fmt.Sprintf("deliver %s - Message sent by %s delivered!", m.Content, addressToPrint(m.Sender, NODE_PRINTLAST))
	logEvent(thisNode.ID().String(), PRINTOPTION, event)
	// Remove the neighbourhood
	// Modification 2
	m.Path = []string{}

	// Update the sender
	old_sender := m.Sender
	m.Sender = getNodeAddress(thisNode, ADDR_DEFAULT)

	dataBytes, err := json.Marshal(m)
	if err != nil {
		printError(err)
	}
	msg := string(dataBytes)

	// Forward the message with void visitedset (path) set to all the peers that do not appear in any path of the visited set
	for _, p := range thisNode.Network().Peers() {
		
		if !deliveredMessages.lookInPaths(m.ID, p.String()) {
			stream, err := openStream(ctx, thisNode, p, PROTOCOL_EXP2)
			if err != nil {
				printError(err)
			}

			message := fmt.Sprintf("%s\n", msg)

			// Write the message on the stream
			_, err = stream.Write([]byte(message))
			if err != nil {
				printError(err)
			}
			event := fmt.Sprintf("delandrelay %s - Forward message from %s on node %s", m.Content, addressToPrint(old_sender, NODE_PRINTLAST), addressToPrint(p.String(), NODE_PRINTLAST) )
			logEvent(thisNode.ID().String(), PRINTOPTION, event)
		}
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