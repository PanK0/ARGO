package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/libp2p/go-libp2p/core/host"
)

// Print shell newline
func printShell() {
	fmt.Printf("%s_> %s", color_info, RESET)
}

// Print an error in a fancy way
func printError(err error) {
	header := fmt.Sprintf("\n%s!!!!----- ERROR -----!!!!%s\n", RED, RESET)	
	fmt.Println(header)
	panic(err)
}

// Message to print when starting a node
func printStartMessage(h host.Host, help_mod string) {

	if help_mod == mod_help_def {
		header := fmt.Sprintf("\n%s------- HELP -------%s\n", color_info, RESET)
		footer := fmt.Sprintf("%s--------------------%s\n", color_info, RESET)
		fmt.Printf("%s", header)
		printHelp_node()
		printHelp_ProtocolInfo(h)
		printHelp_MessagesInfo()
		printHelp_NetworkInfo()
		printHelp()
		fmt.Printf("%s", footer)

	} else if help_mod == mod_help_msg {
		header := fmt.Sprintf("\n%s------- MESSAGES -------%s\n", color_info, RESET)
		footer := fmt.Sprintf("%s-------------------------%s\n", color_info, RESET)
		fmt.Printf("%s", header)
		printHelp_MessagesInfo()
		fmt.Printf("%s", footer)

	} else if help_mod == mod_help_net {
		header := fmt.Sprintf("\n%s------- NETWORK -------%s\n", color_info, RESET)
		footer := fmt.Sprintf("%s------------------------%s\n", color_info, RESET)
		fmt.Printf("%s", header)
		printHelp_NetworkInfo()
		fmt.Printf("%s", footer)

	} else if help_mod == mod_help_node {
		header := fmt.Sprintf("\n%s------- NODE -------%s\n", color_info, RESET)
		footer := fmt.Sprintf("%s---------------------%s\n", color_info, RESET)
		fmt.Printf("%s", header)
		printHelp_node()
		fmt.Printf("%s", footer)

	} else if help_mod == mod_help_prot {
		header := fmt.Sprintf("\n%s------- PROTOCOLS -------%s\n", color_info, RESET)
		footer := fmt.Sprintf("%s--------------------------%s\n", color_info, RESET)
		fmt.Printf("%s", header)
		printHelp_ProtocolInfo(h)
		fmt.Printf("%s", footer)

	} else if help_mod == mod_help_help {
		header := fmt.Sprintf("\n%s------- HELP -------%s\n", color_info, RESET)
		footer := fmt.Sprintf("%s---------------------%s\n", color_info, RESET)
		fmt.Printf("%s", header)
		printHelp()
		fmt.Printf("%s", footer)


	} else {

	}

}

// Print help related information
func printHelp() {

	help := fmt.Sprintf(
		"%sHELP: %s \n" +
		"\t%sPrint general help panel%s \n" +
		"\t-help \n\n" +
		"\t%sPrint node help panel%s \n" +
		"\t-help NODE \n\n" +
		"\t%sPrint protocols help panel%s \n" +
		"\t-help PROTOCOLS \n\n" +
		"\t%sPrint messages help panel%s \n" +
		"\t-help MSG \n\n" +
		"\t%sPrint network help panel%s \n" +
		"\t-help NETWORK \n\n" +
		"\t%sPrint help list help panel%s \n" +
		"\t-help HELP \n",
		color_info, RESET, color_info, RESET, color_info, RESET, color_info, RESET,
		color_info, RESET, color_info, RESET, color_info, RESET,
	)

	fmt.Printf("%s", help)
}

// Print node related information
func printHelp_node() {

	info := fmt.Sprintf(
		"%sINFO: %s \n" +
		"\t%sShow information about this node%s \n" +
		"\t-info \n",
		color_info, RESET, color_info, RESET,
	)

	byzantine := fmt.Sprintf(
		"%sBYZANTINE: %s \n" +
		"\t%sTurn this node into a byzantine%s \n" +
		"\t-byzantine \n",
		color_info, RESET, color_info, RESET,
	)

	fmt.Printf("%s", info)
	fmt.Printf("%s", byzantine)

}

// Print protocols related information
func printHelp_ProtocolInfo(h host.Host) {

	connect := fmt.Sprintf(
		"%sCONNECT:%s \n" +
		"\t%sConnect another node with this node%s \n" +
		"\t-connect %s \n\n" +
		"\t%sConnect this node to its neighbours in the topology%s \n" +
		"\t-connectall \n",
		color_info, RESET, color_info, RESET, getNodeAddress(h, ADDR_DEFAULT), color_info, RESET,
	)

	connect_desc := fmt.Sprintf(
		"\n\t%sRemember to connect nodes before running any protocol!%s \n",
		color_desc, RESET,
	)

	send := fmt.Sprintf(
		"%sSEND:%s	\n" + 
		"\t%sSend a message MESSAGE from another node to this node%s \n" +
		"\t-send %s -msg \"MESSAGE\" \n",
		color_info, RESET, color_info, RESET, getNodeAddress(h, ADDR_DEFAULT),
	)

	bcast := fmt.Sprintf(
		"%sBROADCAST:%s \n" +
		"\t%sSend a broadcast with message MESSAGE from any node to this node%s \n" +
		"\t-broadcast %s -msg \"MESSAGE\"\n",
		color_info, RESET, color_info, RESET, getNodeAddress(h, ADDR_DEFAULT),
	)

	detector := fmt.Sprintf(
		"%sDETECTOR:%s \n" +
		"\t%sRun detector protocol%s \n" +
		"\t-detector \n",
		color_info, RESET, color_info, RESET,
	)

	explorer := fmt.Sprintf(
		"%sEXPLORER:%s \n" +
		"\t%sRun explorer protocol%s \n" +
		"\t-explorer \n",
		color_info, RESET, color_info, RESET,
	)

	explorer_desc := fmt.Sprintf( 
		"\t%sWARNING: this protocol has been proven to not work%s \n",
		color_desc, RESET,
	)

	explorer2 := fmt.Sprintf(
		"%sEXPLORER2:%s \n" +
		"\t%sRun explorer2 protocol%s \n" +
		"\t-exp2 \n",
		color_info, RESET, color_info, RESET,
	)

	combinedRC := fmt.Sprintf(
		"%sCOMBINED RC:%s \n" +
		"\t%sRun CombinedRC exploration phase%s \n" +
		"\t-crc %s \n" +
		"\t%sRun CombinedRC route declaration phase%s \n" +
		"\t-crc %s %s\n",
		color_info, RESET, color_info, RESET, mod_crc_exp, color_info, RESET, mod_crc_rou, getNodeAddress(h, ADDR_DEFAULT),
	)

	fmt.Printf("%s", connect)
	fmt.Printf("%s", connect_desc)
	fmt.Printf("%s", send)
	fmt.Printf("%s", bcast)
	fmt.Printf("%s", detector)
	fmt.Printf("%s", explorer)
	fmt.Printf("%s", explorer_desc)
	fmt.Printf("%s", explorer2)
	fmt.Printf("%s", combinedRC)
}

// Print messages related information
func printHelp_MessagesInfo() {

	deliver := fmt.Sprintf(
		"%sDELIVER:%s \n" +
		"\t%sForce the delivery of a single message with ID <MSG_ID> from this node %s \n" +
		"\t-deliver <MSG_ID> \n\n" +
		"\t%sForce the delivery of all messages from this node %s \n" +
		"\t-deliver ALL \n",
		color_info, RESET, color_info, RESET, color_info, RESET,
	)

	deliver_desc := fmt.Sprintf(
		"\n\t%sThese delivery modes are performed using disjoint paths and \n\tmay conflict with the correctness of Explorer2 protocol%s \n",
		color_desc, RESET,
	)

	show := fmt.Sprintf(
		"%sSHOW:%s \n" +
		"\t%sShow all the received messages from this node %s \n" +
		"\t-show RCV \n\n" +
		"\t%sShow all the delivered messages from this node %s \n" +
		"\t-show DEL \n",
		color_info, RESET, color_info, RESET, color_info, RESET,
	)

	fmt.Printf("%s", deliver)
	fmt.Printf("%s", deliver_desc)
	fmt.Printf("%s", show)
}

// Print network and topology information
func printHelp_NetworkInfo() {

	network := fmt.Sprintf(
		"%sNETWORK:%s \n" +
		"\t%sShow network information about this node%s \n" +
		"\t-network \n",
		color_info, RESET, color_info, RESET,
	)

	topology := fmt.Sprintf(
		"%sTOPOLOGY: %s \n" +
		"\t%sShow cTop (Confirmed Topology)%s \n" +
		"\t-topology SHOW \n\n" +
		"\t%sShow both cTop and uTop%s \n" +
		"\t-toplogy WHOLE \n\n" +
		"\t%sAcquire topology from network information%s \n" +
		"\t-topology ACQUIRE \n\n" +
		"\t%sLoad cTop from configuration files%s \n" +
		"\t-topology LOAD \n\n" +
		"\t%sChange <NODE> in default topology file with this node's address%s \n" +
		"\t-topology FORCE <NODE>\n",
		color_info, RESET, color_info, RESET, color_info, RESET, color_info, RESET, color_info, RESET, color_info, RESET,
	)

	graph := fmt.Sprintf(
		"%sGRAPH: %s\n" +
		"\t%sBuild a graph based on topology information and byzantine constraints%s \n" +
		"\t-graph",
		color_info, RESET, color_info, RESET,
	)

	graph_desc := fmt.Sprintf(
		"\n\t%sThis command should be used in combination with Explorer2 protocol%s \n",
		color_desc, RESET,
	)

	fmt.Printf("%s", network)
	fmt.Printf("%s", topology)
	fmt.Printf("%s", graph)
	fmt.Printf("%s", graph_desc)

}


// Print node information
func printNodeInfo(h host.Host) {
	fmt.Printf("\n%s##### NODE INFORMATION #####%s\n", CYAN, RESET)
	fmt.Println("Node address:", getNodeAddress(h, ADDR_DEFAULT))//h.Addrs()[0])
    fmt.Println("host.ID:", h.ID())
	fmt.Println("Peers: ", h.Network().Peers())

	fmt.Printf("\n%sThis node's multiaddresses:%s\n", CYAN, RESET)
	fmt.Printf("	Loopback Address: %s\n", h.Addrs()[ADDR_LB_POS])
	fmt.Printf("	LAN Address: %s\n", h.Addrs()[ADDR_LAN_POS])
	
	/*
	// Print all multiaddresses
	for i, la := range h.Addrs() {
		fmt.Printf("	%d. %v\n",i, la)
	}
		*/

	fmt.Printf("%s############################%s\n", CYAN, RESET)
	fmt.Println()
}


// Function to print a message
func printMessage(message string) {
	
	var msg Message
	err := json.Unmarshal([]byte(message), &msg)
	if err != nil {
		printError(err)
	}

	msgtype := ""
	color := ""
	if msg.Type == TYPE_BROADCAST {
		msgtype = TYPE_BROADCAST
		color = MAGENTA
	} else if msg.Type == TYPE_DIRECT_MSG {
		msgtype = TYPE_DIRECT_MSG
		color = YELLOW
	} else if msg.Type == TYPE_EXPLORER2 {
		msgtype = TYPE_EXPLORER2
		color = GREY
	}

	header := fmt.Sprintf("\n%s----RECEIVED %s----%s\n", color, msgtype, RESET)
	footer := fmt.Sprintf("%s--------------------------%s\n", color, RESET)

	m := header
	m += msgToString(msg)
	m += fmt.Sprintf("%s\n", footer)
	
	log.Print(m)
	printShell()
}


// Print all the opened streams of an host
func printOpenedStream(h host.Host) {
	header := fmt.Sprintf("\n%s####### NETWORK INFO #######%s\n", CYAN, RESET)
	footer := fmt.Sprintf("\n%s############################%s\n", CYAN, RESET)
	fmt.Print(header)
	connections := h.Network().Conns()
	for _, c := range connections {
		streams := c.GetStreams()
		for i, s := range streams {
			fmt.Printf("%d - stream: %s, id: %s\n", i, s, s.ID())
		}
		
	}
	
	fmt.Println(footer)
}


// Returns a list of all messages
// !! CAREFUL !! : the messages here are NOT IN CHRONOLOGICAL ORDER, because MessageContainer is a dictionnary, not a list!
func allMessages(messageContainer MessageContainer, mod string) string {
	var mod_string string = ""
	var color = ""
	var color_msg_top = ""
	var color_msg_bot = ""
	switch mod {
	case mod_show_del:
		mod_string = "+ DELIVERED MESSAGES +"
		color = BLUE_BG
	case mod_show_rcv:
		mod_string = "+ RECEIVED MESSAGES ++"
		color = CYAN_BG
	}


	header := fmt.Sprintf("\n%s++++++++++++++++++++++++++++++++++++++++++++%s+++++++++++++++++++++++++++++++++++++%s\n", color, mod_string, RESET)
	footer := fmt.Sprintf("%s+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++%s\n", color, RESET)
	ret := ""
	ret += header
	
	for k, messages := range messageContainer.messages {
		
		if messages[0].Type == TYPE_BROADCAST {
			color_msg_top = MAGENTA_BG
			color_msg_bot = MAGENTA
		} else if messages[0].Type == TYPE_DIRECT_MSG {
			color_msg_top = YELLOW_BG
			color_msg_bot = YELLOW
		} else if messages[0].Type == TYPE_DETECTOR {
			color_msg_top = GREY_BG
			color_msg_bot = WHITE
		} else if messages[0].Type == TYPE_EXPLORER || messages[0].Type == TYPE_EXPLORER2 {
			color_msg_top = GREY_BG
			color_msg_bot = GREY
		}


		msg := ""
		h := fmt.Sprintf("\n%s-ID: %s -%s", color_msg_top, k, RESET)
		f := fmt.Sprintf("%s-----------------------------------------------%s\n", color_msg_bot, RESET)
		
		msg += h
		for i, m := range messages {
			msg += fmt.Sprintf("\n%s- %d -------------------%s\n", CYAN, i, RESET)
			msg += msgToString(m)
			msg += "\n"
		}
		msg += f
		ret += msg
	}
	ret += footer
	return ret
}


// Returns a string with full node address with NODE_PRINTLAST characters. WHOLE_ADDR to print the whole address
func addressToPrint(address string, n_of_characters int) string {
	if n_of_characters == -1 {
		return address
	}
	if len(address) > n_of_characters {
		return address[len(address)-n_of_characters:]
	}
	return address
}


// Print cTop (Confirmed Topology) information
// cTop is a structure defined in topology.go
func (top CTop) toString() string {
	color_node := BLUE
	color_neigh := CYAN

	h := fmt.Sprintf("\n%s##### CONFIRMED TOPOLOGY #####%s\n", BLUE, RESET)
	f := fmt.Sprintf("%s#############################%s\n", BLUE, RESET)
	str := ""

	str += h

	for k, v := range top.tuples {
		node_to_print := addressToPrint(k, NODE_PRINTLAST)
		
		str += fmt.Sprintf("%sNode: %s%s\n",color_node, RESET, node_to_print)
		str += fmt.Sprintf("%s___Neighbourhood: %s\n", color_neigh, RESET)
		for i, n := range v {
			toprint := addressToPrint(n, NODE_PRINTLAST)
			str += fmt.Sprintf("%s	%d -%s %s\n", color_neigh, i, RESET, toprint)
		}
	}

	str += f

	return str	
}


// Print UTop
// uTop is a structure defined in topology.go
func (utop UTop) toString() string {
	color_node := BLUE
	color_neigh := CYAN
	color_visit	:= YELLOW
	
	h := fmt.Sprintf("\n%s#### UNCONFIRMED TOPOLOGY ####%s\n", MAGENTA, RESET)
	f := fmt.Sprintf("%s#############################%s\n", MAGENTA, RESET)
	str := ""

	str += h


	for k, v := range utop.tuples {
		node_to_print := addressToPrint(k, NODE_PRINTLAST)

		str += fmt.Sprintf("%sNode: %s%s\n", color_node, RESET, node_to_print)
		str += fmt.Sprintf("%s___Neighbourhood: %s\n", color_neigh, RESET)
		for i, n := range v[0] {
			toprint := addressToPrint(n, NODE_PRINTLAST)
			str += fmt.Sprintf("%s	%d -%s %s\n", color_neigh, i, RESET, toprint)
		}

		str += fmt.Sprintf("%s___Visited set: %s\n", color_visit, RESET)

		for i, n := range v[1] {
			toprint := addressToPrint(n, NODE_PRINTLAST)
			str += fmt.Sprintf("%s	%d -%s %s\n", color_visit, i, RESET, toprint)
		}
	}

	str += f
	return str	
}


// PrintGraph prints the adjacency list of the graph
func (g *Graph) PrintGraph() {
    fmt.Println("Graph:")
    for node, neighbors := range g.adjList {
        // Get the last 5 characters of the node
        nodeToPrint := addressToPrint(node, NODE_PRINTLAST)
        // Print the node and its neighbors
        fmt.Printf("%s -> [", nodeToPrint)
        for i, neighbor := range neighbors {
            // Get the last 5 characters of the neighbor
            neighborToPrint := addressToPrint(neighbor, NODE_PRINTLAST)
            if i > 0 {
                fmt.Printf(", ")
            }
            fmt.Printf("%s", neighborToPrint)
        }
        fmt.Println("]")
    }
    fmt.Println()
}

// Print DisjointPaths
func (dp *DisjointPaths) PrintDP() {
	fmt.Println("Disjoint Paths:")
	for node, paths := range dp.paths {
		nodetoprint := addressToPrint(node, NODE_PRINTLAST)
		fmt.Printf("Node %s :\n", nodetoprint)
		for i, path := range paths {
			fmt.Printf("\tPath %d : [", i)
			for _, p := range path {
				ptoprint := addressToPrint(p, NODE_PRINTLAST)
				fmt.Printf(" %s , ", ptoprint)
			}
			fmt.Printf(" ]\n")
		}
		fmt.Println()
	}
}

// Print Byzantine Detection Alert
func printByzantineAlert() {
	alert := fmt.Sprintf("%s!!!!----- BYZANTINE BEHAVIOUR DETECTED -----!!!!%s\n", RED, RESET)
	fmt.Println(alert)
	printShell()
}