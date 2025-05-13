package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
)

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