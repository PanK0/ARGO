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
	"github.com/libp2p/go-libp2p/core/protocol"
)

// function to manage an EXP2 message
func receive_EXP(ctx context.Context, thisNode host.Host, m *Message, top *Topology,
		messageContainer *MessageContainer, deliveredMessages *MessageContainer) error {
	
	m.Content = time.Now().Format("05.00000")
	event := fmt.Sprintf("receive_EXP2 %s - Handling message from %s", m.Content, addressToPrint(m.Sender, NODE_PRINTLAST))
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
					removed := m.Neighbourhood[index]
					m.Neighbourhood = append(m.Neighbourhood[:index], m.Neighbourhood[index+1:]...)
					event := fmt.Sprintf("byzantine %s - Message from %s altered. Removed %s from neighbourhood.", m.Content, addressToPrint(m.Sender, NODE_PRINTLAST), addressToPrint(removed, NODE_PRINTLAST))
					logEvent(thisNode.ID().String(), PRINTOPTION, event)
				} 
			} else if bz.Alterations == BYZ_PATH {
				if len(m.Path) > 0 {
					// Remove a random element from the path
					rand.Seed(time.Now().UnixNano())
					index := rand.Intn(len(m.Path))
					removed := m.Path[index]
					m.Path = append(m.Path[:index], m.Path[index+1:]...)
					event := fmt.Sprintf("byzantine %s - Message from %s altered. Removed %s from path.", m.Content, addressToPrint(m.Sender, NODE_PRINTLAST), addressToPrint(removed, NODE_PRINTLAST))
					logEvent(thisNode.ID().String(), PRINTOPTION, event)
				}
			} else if bz.Alterations == BYZ_SWAP_PATH {
				if len(m.Path) > 1 {
					rand.Seed(time.Now().UnixNano())
					i := rand.Intn(len(m.Path))
					j := rand.Intn(len(m.Path))
					for j == i {
						j = rand.Intn(len(m.Path))
					}
					m.Path[i], m.Path[j] = m.Path[j], m.Path[i]
					event := fmt.Sprintf("byzantine %s - Message from %s altered. Swapped %s and %s in path.", m.Content, addressToPrint(m.Sender, NODE_PRINTLAST), addressToPrint(m.Path[j], NODE_PRINTLAST), addressToPrint(m.Path[i], NODE_PRINTLAST))
					logEvent(thisNode.ID().String(), PRINTOPTION, event)
				}
			}
		}
	}
	// Check whether m is in deliveredMessages
	// Modification 4
	timestamp_start := time.Now()
	if len(deliveredMessages.Get(m.ID)) == 0 {
		// if m arrived with a void path, means that m.Sender delivered the message
		m.Target = getNodeAddress(thisNode, ADDR_DEFAULT)
		messageContainer.Add(*m)

		// Modification 1
		if m.Source == m.Sender && m.Path[0] == m.Source {
			BFT_deliver_and_relay(ctx, thisNode, *messageContainer, *deliveredMessages, *m, top)
		} else if len(messageContainer.countNodeDisjointPaths_intersection(m.ID)) > MAX_BYZANTINES  {
			BFT_deliver_and_relay(ctx, thisNode, *messageContainer, *deliveredMessages, *m, top)
		} else {
			// Send the message to all the nodes who never ever received the message
			for _, p := range thisNode.Network().Peers() {

				if p.String() == extractPeerIDFromMultiaddr(master_address) {
					continue // Do not send the message to the master node
				}

				// Only forward the message if p is not in m.path or if it doesen't exist in any of the paths of the instances of m.ID that are present in messageContainer
				if !messageContainer.lookInPaths(m.ID, p.String()) && !contains(m.Path, p.String()) {
					m.Sender = getNodeAddress(thisNode, ADDR_DEFAULT)
					send(ctx, thisNode, p, *m, PROTOCOL_EXP2)					
				} 
			}
		}
	} else {		
		deliveredMessages.Add(*m)		
		BFT_deliver(*messageContainer, *deliveredMessages, *m, top)
		event := fmt.Sprintf("receive_del_EXP2 %s - delivering message from %s", m.Content, addressToPrint(m.Source, NODE_PRINTLAST))
		logEvent(thisNode.ID().String(), PRINTOPTION, event)
		/*
		Che succede se rimuovo la delivery dell'else?
		I messaggi non vengono inoltrati correttamente e gli altri nodi non hanno la possibilità di calcolare i disjoint paths
		*/
	}
	timestamp_end := time.Now()
	event = fmt.Sprintf("BFT_execution %s - performed in time %f seconds", m.Content, timestamp_end.Sub(timestamp_start).Seconds())
	logEvent(thisNode.ID().String(), false, event)
	return nil
}

// Handle stream for EXPLORER2 protocol
// described @ `Tractable Reliable Communication in Compromised Networks, Giovanni Farina - cpt. 9.3, 9.4`
//lint:ignore U1000 Unused function for future use
func handleExplorer2(s network.Stream, ctx context.Context, thisNode host.Host, top *Topology, 
					messageContainer *MessageContainer, deliveredMessages *MessageContainer) error {

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

	err = receive_EXP(ctx, thisNode, &m, top, messageContainer, deliveredMessages)
	if err != nil {
		printError(err)
	}

	return nil
}

// Send an EXPLORER2 message
// described @ `Tractable Reliable Communication in Compromised Networks, Giovanni Farina - cpt. 9.3, 9.4`
//lint:ignore U1000 Unused function for future use
func sendExplorer2(ctx context.Context, thisNode host.Host, exp_msg Message, PROTOCOL protocol.ID) {
	
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
			stream, err := openStream(ctx, thisNode, p, PROTOCOL)
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

// Helper function to handle common delivery logic
func manageDelivery(messageContainer MessageContainer, deliveredMessages MessageContainer, m Message, top *Topology) {
    // Add the message to the delivered messages
    messages := messageContainer.Get(m.ID)

    // Update the topology
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

    // Add the message to delivered messages
    for _, msg := range messages {
        deliveredMessages.Add(msg)
    }

    // Remove the message from the message container
    messageContainer.deleteElement(m.ID)
}

// Delivery function for BFT
func BFT_deliver(messageContainer MessageContainer, deliveredMessages MessageContainer, m Message, top *Topology) {
    // Handle the common delivery logic
    manageDelivery(messageContainer, deliveredMessages, m, top)
}

// Delivery and relay function for BFT
func BFT_deliver_and_relay(ctx context.Context, thisNode host.Host,
    messageContainer MessageContainer, deliveredMessages MessageContainer,
    m Message, top *Topology) {

    // Deliver the message
    BFT_deliver(messageContainer, deliveredMessages, m, top)

    // Log the delivery event
    event := fmt.Sprintf("deliver_EXP2 %s - Message sent by %s delivered!", m.Content, addressToPrint(m.Sender, NODE_PRINTLAST))
    logEvent(thisNode.ID().String(), PRINTOPTION, event)

    // Prepare the message for relaying
    m.Path = []string{} // Clear the path
    old_sender := m.Sender
    m.Sender = getNodeAddress(thisNode, ADDR_DEFAULT)

    dataBytes, err := json.Marshal(m)
    if err != nil {
        printError(err)
        return
    }
    msg := string(dataBytes)

    // Relay the message to peers not in any path of the delivered messages
    for _, p := range thisNode.Network().Peers() {

		if p.String() == extractPeerIDFromMultiaddr(master_address) {
			continue // Do not send the message to the master node
		}

        if !deliveredMessages.lookInPaths(m.ID, p.String()) {
            stream, err := openStream(ctx, thisNode, p, PROTOCOL_EXP2)
            if err != nil {
                printError(err)
                continue
            }

            message := fmt.Sprintf("%s\n", msg)
            _, err = stream.Write([]byte(message))
            if err != nil {
                printError(err)
            }
            stream.Close()

            event := fmt.Sprintf("delandrelay_EXP2 %s - Forward message from %s on node %s", m.Content, addressToPrint(old_sender, NODE_PRINTLAST), addressToPrint(p.String(), NODE_PRINTLAST))
            logEvent(thisNode.ID().String(), PRINTOPTION, event)
        }
    }
}