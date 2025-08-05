package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
)

// function to manage an EXP2 message
func receive_EXP2(ctx context.Context, thisNode host.Host, m *Message, top *Topology,
		messageContainer *MessageContainer, deliveredMessages *MessageContainer) error {
	
	m.Content = time.Now().Format("05.00000")

	inst := addressToPrint(m.Sender, NODE_PRINTLAST)
	m.InstanceID += "_"+inst

	event := fmt.Sprintf("receive_EXP2 %s - Handling message from %s", m.ID[len(m.ID)-5:], addressToPrint(m.Sender, NODE_PRINTLAST))
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
	event = fmt.Sprintf("BFT_execution %s - performed in time %f seconds", m.ID[len(m.ID)-5:], timestamp_end.Sub(timestamp_start).Seconds())
	logEvent(thisNode.ID().String(), false, event)
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

			event := fmt.Sprintf("send_EXP2 %s - Forwarded message from %s to node %s", exp_msg.ID[len(exp_msg.ID)-5:], addressToPrint(exp_msg.Sender, NODE_PRINTLAST), addressToPrint(p.String(), NODE_PRINTLAST))
			logEvent(thisNode.ID().String(), PRINTOPTION, event)
		}
	}
}

// Helper function to handle common delivery logic
func manageDelivery(messageContainer MessageContainer, deliveredMessages MessageContainer, m Message, top *Topology) bool {
    // Add the message to the delivered messages
    messages := messageContainer.Get(m.ID)


	if !top.ctop.checkInCTop(m.Source) {
		// Add m.Source to cTop
		top.ctop.AddNeighbourhood(m.Source, m.Neighbourhood)
		event := fmt.Sprintf("manageDelivery %s - Node %s added to cTop", m.ID[len(m.ID)-5:], addressToPrint(m.Source, NODE_PRINTLAST))
		logEvent(top.nodeID, PRINTOPTION, event)
	} else if top.ctop.checkInCTop(m.Source) &&
			isSubSet(top.ctop.GetNeighbourhood(m.Source), m.Neighbourhood) == 0 {
		
		top.ctop.AddNeighbourhood(m.Source, m.Neighbourhood)
		event := fmt.Sprintf("manageDelivery %s - Neighbourhood updated for node %s", m.ID[len(m.ID)-5:], addressToPrint(m.Source, NODE_PRINTLAST))
		logEvent(top.nodeID, PRINTOPTION, event)
	
	} else if top.ctop.checkInCTop(m.Source) &&
			isSubSet(m.Neighbourhood, top.ctop.GetNeighbourhood(m.Source)) == 0 {
		
		event := fmt.Sprintf("manageDelivery %s - Non consistent information from node %s", m.ID[len(m.ID)-5:], addressToPrint(m.Sender, NODE_PRINTLAST))
		logEvent(top.nodeID, PRINTOPTION, event)
		return false
	}
	
    // Add the message to delivered messages
    for _, msg := range messages {
		deliveredMessages.Add(msg)
	}  

    // Remove the message from the message container
    messageContainer.deleteElement(m.ID)

	
	return true
}

// Delivery function for BFT
func BFT_deliver(messageContainer MessageContainer, deliveredMessages MessageContainer, m Message, top *Topology) bool {
    // Handle the common delivery logic
    return manageDelivery(messageContainer, deliveredMessages, m, top)
}

// Delivery and relay function for BFT
func BFT_deliver_and_relay(ctx context.Context, thisNode host.Host,
    messageContainer MessageContainer, deliveredMessages MessageContainer,
    m Message, top *Topology) {

    // Deliver the message
    del := BFT_deliver(messageContainer, deliveredMessages, m, top)
		
    // Log the delivery event
    event := fmt.Sprintf("deliver_EXP2 %s - Message sent by %s delivered? %t", m.ID[len(m.ID)-5:], addressToPrint(m.Sender, NODE_PRINTLAST), del)
    logEvent(thisNode.ID().String(), PRINTOPTION, event)

	if !del {return}

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

    // Modification 3: relay the message to peers not in any path of the delivered messages
    for _, p := range thisNode.Network().Peers() {

		if p.String() == extractPeerIDFromMultiaddr(master_address) {
			continue // Do not send the message to the master node
		}

        if !deliveredMessages.lookInPaths(m.ID, p.String()) {
            stream, err := openStream(ctx, thisNode, p, PROTOCOL_CRC)
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

            event := fmt.Sprintf("delandrelay_EXP2 %s - Forward message from %s on node %s", m.ID[len(m.ID)-5:], addressToPrint(old_sender, NODE_PRINTLAST), addressToPrint(p.String(), NODE_PRINTLAST))
            logEvent(thisNode.ID().String(), PRINTOPTION, event)
        }
    }
}