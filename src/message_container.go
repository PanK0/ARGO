package main

import (
	"fmt"
	"time"
)

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

// GetDisjointPathsMinCut finds the maximum set of node-disjoint paths
// among all Message.Path of messages[msg_id], using Edmonds–Karp (BFS).
func (mc *MessageContainer) GetDisjointPathsEdmondKarp(msg_id string) [][]string {
    timestamp_start := time.Now()
    messages := mc.Get(msg_id)
    if len(messages) == 0 {
        return nil
    }
    // build the connectivity graph from all paths
    g := NewGraph()
    for _, msg := range messages {
        for i := 0; i < len(msg.Path)-1; i++ {
            g.AddEdge(msg.Path[i], msg.Path[i+1])
        }
    }
    source := messages[0].Sender
    sink   := messages[0].Target

    // build residual graph: bool capacity = true/false
    residual := make(map[string]map[string]bool, len(g.nodes))
    for u := range g.nodes {
        residual[u] = make(map[string]bool, len(g.adjList[u]))
        for _, v := range g.adjList[u] {
            residual[u][v] = true
        }
    }

    usedNodes := make(map[string]bool)    // to enforce node-disjoint (except source/sink)
    parent    := make(map[string]string)  // to reconstruct BFS augmenting path

    var result [][]string
    for {
        // --- BFS to find an augmenting path that avoids usedNodes ---
        for k := range parent {
            delete(parent, k)
        }
        visited := make(map[string]bool, len(g.nodes))
        queue   := []string{source}
        visited[source] = true

        found := false
        for len(queue) > 0 && !found {
            u := queue[0]; queue = queue[1:]
            for _, v := range g.adjList[u] {
                if !residual[u][v] || visited[v] {
                    continue
                }
                // skip any intermediate node already used
                if v != source && v != sink && usedNodes[v] {
                    continue
                }
                parent[v] = u
                if v == sink {
                    found = true
                    break
                }
                visited[v] = true
                queue = append(queue, v)
            }
        }
        if !found {
            break
        }

        // --- reconstruct the path & update residual capacities ---
        path := []string{}
        for v := sink; v != source; v = parent[v] {
            u := parent[v]
            residual[u][v] = false  // consume forward edge
            residual[v][u] = true   // add reverse edge
            path = append([]string{v}, path...)
        }
        path = append([]string{source}, path...)

        // mark intermediate nodes as used
        for _, n := range path {
            if n != source && n != sink {
                usedNodes[n] = true
            }
        }

        result = append(result, path)
    }

    timestamp_end := time.Now()
	event := fmt.Sprintf("DJP COUNT: %d - performed in time %f seconds", len(result), timestamp_end.Sub(timestamp_start).Seconds())
	logEvent(addressToPrint(thisnode_address, NODE_PRINTLAST), PRINTOPTION, event)

    return result
}


// GetDisjointPathsBrute tries every subset of message paths and
// returns the largest node-disjoint collection (NP-complete approach)
// runs in O(2^m * m * l) with m paths of length up to l
func (mc *MessageContainer) GetDisjointPathsBrute(msg_id string) [][]string {
    timestamp_start := time.Now()
    messages := mc.Get(msg_id)
    n := len(messages)
    if n == 0 {
        return nil
    }

    var best [][]string

    // iterate all non-empty subsets via bitmask
    for mask := 1; mask < (1 << n); mask++ {
        used := make(map[string]bool)
        var candidate [][]string
        ok := true

        for i := 0; i < n; i++ {
            if mask&(1<<i) == 0 {
                continue
            }
            path := messages[i].Path
            // check node-disjointness
            for _, node := range path {
                if used[node] {
                    ok = false
                    break
                }
            }
            if !ok {
                break
            }
            // accept this path
            candidate = append(candidate, path)
            for _, node := range path {
                used[node] = true
            }
        }

        // if valid and larger than previous best, keep it
        if ok && len(candidate) > len(best) {
            best = candidate
        }
    }

    timestamp_end := time.Now()
	event := fmt.Sprintf("DJP COUNT - performed in time %f seconds", timestamp_end.Sub(timestamp_start).Seconds())
	logEvent(addressToPrint(thisnode_address, NODE_PRINTLAST), PRINTOPTION, event)

    return best
}
