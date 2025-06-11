package main

import "fmt"

/*
	A message container is a dictionnary that stores messages
	using the message ID as key and a list of objects of type Message as value.
	All Message object in a list must have the same ID
*/
type MessageContainer struct {
	messages map[string] []Message
}

// Return a new MessageContainer
func NewMessageContainer() *MessageContainer {
	return &MessageContainer {
		messages: make(map[string] []Message, 0),
	}
}

// Add an element to the MessageContainer
func (mc MessageContainer) Add(msg Message) {
	mc.messages[msg.ID] = append(mc.messages[msg.ID], msg)
}


// Get all the objects Message associated with an ID
func (mc MessageContainer) Get(msg_id string) []Message {
	return mc.messages[msg_id]
}

// Return all the messages
//lint:ignore U1000 Unused function for future use
func (mc MessageContainer) toString() string {
	msg := ""
	for k, v := range mc.messages {
		msg += fmt.Sprintf("%s - %s\n", k, v)
	}
	return msg
}

// Delete an element from a message container
func (mc MessageContainer) deleteElement(msg_id string) {
	delete(mc.messages, msg_id)
}

// RemoveMessage removes a specific message from the messages corresponding to its message ID.
func (mc *MessageContainer) RemoveMessage(msg Message) {
    msgs, exists := mc.messages[msg.ID]
    if !exists {
        return
    }
    newMsgs := make([]Message, 0, len(msgs))
	for _, m := range msgs {
		if !equalMessage(m, msg) {
			newMsgs = append(newMsgs, m)
		}
	}
    if len(newMsgs) > 0 {
        mc.messages[msg.ID] = newMsgs
    } else {
        delete(mc.messages, msg.ID)
    }
}

// Look for a node being in at least one path of at least one instance of msg_id
// Used for BFT in Explorer2
func (mc MessageContainer) lookInPaths(msg_id string, node_id string) bool {
	messages := mc.Get(msg_id)
	for _, m := range messages {
		if getNodeID(m.Sender) == node_id || getNodeID(m.Source) == node_id {
			return true
		}
		for _, p := range m.Path {
			if getNodeID(p) == node_id {
				return true
			}	
		}
	}
	return false
}


// Count node disjoint paths in relation to a single message
// Given a string msg_id, find it into the message container.
// For that message corresponding to a specific ID, build a graph 
// made of all the nodes that all the copies of that messages traversed in their paths.
// This function builds a graph taking into account the paths the messages took, NOT THE TOPOLOGY.
func (mc MessageContainer) countNodeDisjointPaths(msg_id string) int {
	// Get all the messages corresponding to the ID msg_id
	messages := mc.Get(msg_id)
	
	// Create a new graph
	graph := NewGraph()
	
	// Add all the nodes to the graph
	for _, msg := range messages {
		for i := 0; i < len(msg.Path) - 1; i++ {
			graph.AddEdge(msg.Path[i], msg.Path[i+1])
		}
	}
	//graph.PrintGraph()
	
	// Count the number of node disjoint paths
	source := messages[0].Sender
	target := messages[0].Target
	return graph.FordFulkerson(source, target)
}

// Function to compute disjoint paths by checking intersections
// This function is suited for BFT communication primitive and doesn't admit 
// the source and the last node in the path.
// If you want to make the function consider them, change 
// msg.path WITH msg.path[1 : len(msg.path)-1] 
// in the loop commented with BFT_PATHS
func (mc *MessageContainer) countNodeDisjointPaths_intersection(msg_id string) [][]string {
	var disjointPaths [][]string
	usedNodes := make(map[string]bool)

	// Retrieve messages corresponding to msg_id
	messages, exists := mc.messages[msg_id]
	if !exists {
		return nil // No messages found for this msg_id
	}

	// Iterate through each message and check for node-disjoint paths
	for _, msg := range messages {
		isDisjoint := true

		// Check if any node in the path is already used
		for _, node := range msg.Path { // BFT_PATHS
			if usedNodes[node] {
				isDisjoint = false
				break
			}
		}

		// If disjoint, add to result and mark nodes as used
		if isDisjoint {
			disjointPaths = append(disjointPaths, msg.Path)
			for _, node := range msg.Path { // BFT_PATHS
				usedNodes[node] = true
			}
		}
	}

	fmt.Printf("DISJOINTPATHS: %d\n", len(disjointPaths))

	return disjointPaths
}

// Count node disjoint paths in relation to a node ID
func (mc *MessageContainer) findDisjointPaths(node_id string, djp DisjointPaths) {
	
	for msgid, messages := range mc.messages {
		if messages[0].Source == node_id {
			disjointpaths := mc.countNodeDisjointPaths_intersection(msgid)
			for _, dp := range disjointpaths {
				djp.Add(node_id, dp)
			}
		}
	}
}