package main

import (
	"bufio"
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
)

// Manages the input from the console to perform the wanted actions
func manageConsoleInput(ctx context.Context, h host.Host, 
	messageContainer *MessageContainer, deliveredMessages *MessageContainer, disjointPaths *DisjointPaths, 
	topology *Topology) (*bufio.ReadWriter, error) {
	stdReader := bufio.NewReader(os.Stdin)
	// endless loop
	for {
		printShell()
		inputData, err_readStr := stdReader.ReadString('\n')
		if err_readStr != nil {
			log.Println(err_readStr)
			return nil, err_readStr
		}
		
		// Parse the console input
		inputData_words := strings.Fields(inputData)

		// Print help panel
		command, idx := findElement(inputData_words, cmd_help) 
		if command == cmd_help {
			// Print full help panel
			if len(inputData_words) == 1 {
				printStartMessage(h, mod_help_def)
			// Print relative help panel
			} else if len(inputData_words) == 2 {
				printStartMessage(h, inputData_words[idx+1])
			}
		}

		// Print node information
		command, _ = findElement(inputData_words, cmd_info)
		if command == cmd_info {
			printNodeInfo(h)
		} 

		// Connect to a node
		command, idx = findElement(inputData_words, cmd_connect)
		if command == cmd_connect {
			dest := inputData_words[idx+1]
			connectNodes(ctx, h, dest, topology)
		}

		// Connect to all nodes
		command, _ = findElement(inputData_words, cmd_connect_all)
		if command == cmd_connect_all {
			connectAllNodes(ctx, h, topology)
		}

		// Send a message to a node
		command, idx = findElement(inputData_words, cmd_send)
		if command == cmd_send {
			targetNode_address := inputData_words[idx+1]
			command, _ = findElement(inputData_words, cmd_msg)
			if command == cmd_msg {
				message := extractMessage(inputData)

				// Generate an ID for the message
				// !! WARNING !!
				// If it happens that two nodes generate a message at the same time a collision may happen.
				// This is only for test purposes
				timestamp := time.Now().Unix()
    			hasher := sha1.New()
				hasher.Write([]byte(fmt.Sprintf("%d", timestamp)))
				msgid := fmt.Sprintf("%x", hasher.Sum(nil))

				data := Message{ID: msgid, Type: TYPE_DIRECT_MSG, Sender: getNodeAddress(h, ADDR_DEFAULT), Source: getNodeAddress(h, ADDR_DEFAULT), Target: targetNode_address, Content: message}
				dataBytes, err := json.Marshal(data)
				if err != nil {
					fmt.Println("Error marshalling data while sending a direct message:", err)
				}

				sendMessage(ctx, h, string(dataBytes))
			}
		}

		// Send broadcast
		command, idx = findElement(inputData_words, cmd_broadcast)
		if command == cmd_broadcast {
			targetNode_address := inputData_words[idx+1]
			command, _ = findElement(inputData_words, cmd_msg)
			if command == cmd_msg {
				message := extractMessage(inputData)

				// Generate an ID for the message
				// !! WARNING !!
				// If it happens that two nodes generate a message at the same time a collision may happen.
				// This is only for test purposes
				timestamp := time.Now().Unix()
    			hasher := sha1.New()
				hasher.Write([]byte(fmt.Sprintf("%d", timestamp)))
				msgid := fmt.Sprintf("%x", hasher.Sum(nil))
				
				var path []string
				// When sending a broadcast from a node, that node is both the sender and the source of the message
				data := Message{ID: msgid, Type: TYPE_BROADCAST, Sender: getNodeAddress(h, ADDR_DEFAULT), Source: getNodeAddress(h, ADDR_DEFAULT), Target: targetNode_address, Content: message, Path: path}
				dataBytes, err := json.Marshal(data)
				if err != nil {
					fmt.Println("Error marshalling data while sending a broadcast:", err)
				}
				sendBroadcast(ctx, h, string(dataBytes))
				
			}
			
		}

		// Force the delivery of a message
		command, idx = findElement(inputData_words, cmd_deliver)
		if command == cmd_deliver {
			if len(inputData_words) != 2 {
				fmt.Println("Please provide a message ID to deliver")
			}
			// If inputData_words[idx+1] is equal to "ALL" then deliver all the messages
			if inputData_words[idx+1] == mod_deliver_all {
				for k := range messageContainer.messages {
					dolevR_deliver(messageContainer, k, deliveredMessages)
				}
			} else {
			message_id := inputData_words[idx+1]
			dolevR_deliver(messageContainer, message_id, deliveredMessages)
			}
		}

		// Show all messages
		command, idx = findElement(inputData_words, cmd_show)
		if command == cmd_show {
			if len(inputData_words) != 1 {
				if inputData_words[idx+1] == mod_show_del {
				// print deliveredMessages instead
				fmt.Println(allMessages(*deliveredMessages, mod_show_del))
				} else if inputData_words[idx+1] == mod_show_rcv {
				// print messageContainer instead	
				fmt.Println(allMessages(*messageContainer, mod_show_rcv))
				}
			} else {
				// print error message
				fmt.Println("Please provide show mod: DEL for delivered messages or RCV for received messages")
			}
		}
		
		// Print network information
		command, _ = findElement(inputData_words, cmd_network)
		if command == cmd_network {
			printOpenedStream(h)
		}

		// Print, acquire, load or force topology
		command, idx = findElement(inputData_words, cmd_topology)
		if command == cmd_topology {
			if len(inputData_words) == 1 {
				fmt.Println("Please provide a topology mod: SHOW to show the current topology, LOAD to load the topology from file, FORCE <NODE> to force a change in the topology")
			} else if len(inputData_words) == 2 {
				// -topology LOAD (from topology.csv file, replacing the current cTop)
				if inputData_words[idx+1] == mod_top_load {
					topology_graph := LoadGraphFromCSV(topology_path)
					topology.ctop.loadNeigh(topology_graph, getNodeAddress(h, ADDR_DEFAULT))
					// topology.ctop = *loadCTop(topology_graph) // Uncomment this to laod the whole topology
					fmt.Println(topology.ctop.toString())					
				// -topology SHOW and WHOLE
				} else if inputData_words[idx+1] == mod_top_show {
					fmt.Println(topology.ctop.toString())
				} else if inputData_words[idx+1] == mod_top_whole {
					fmt.Println(topology.ctop.toString())
					fmt.Println(topology.utop.toString())
				} else if inputData_words[idx+1] == mod_top_acquire {
					// Acquire the topology from the network
					acquireTopology(h, topology)
					fmt.Println("Topology acquired")
					fmt.Println(topology.ctop.toString())
				} else {
					fmt.Println("Please provide a topology mod: SHOW to show the current topology, LOAD to load the topology from file, FORCE <NODE> to force a change in the topology")
				}
			} else if len(inputData_words) == 3 {
				// -topology FORCE <NODE> (force this node in topology file, by changing this node's address to the provided one)
				if inputData_words[idx+1] == mod_top_force {
					if inputData_words[idx+2] != "" {
						ReplaceInCSV(topology_path, getNodeAddress(h, ADDR_DEFAULT), inputData_words[idx+2])
						topology_graph := LoadGraphFromCSV(topology_path)
						topology.ctop.loadNeigh(topology_graph, getNodeAddress(h, ADDR_DEFAULT))
						//topology.ctop = *loadCTop(topology_graph) // Uncomment this to laod the whole topology
						fmt.Println(topology.ctop.toString())
					} else {
						fmt.Println("Please provide a topology mod: SHOW to show the current topology, LOAD to load the topology from file, FORCE <NODE> to force a change in the topology")
					}
				} else {
					fmt.Println("Please provide a topology mod: SHOW to show the current topology, LOAD to load the topology from file, FORCE <NODE> to force a change in the topology")
				}
			}
		}

		// Build graph
		command, _ = findElement(inputData_words, cmd_graph)
		if command == cmd_graph {
			if command == cmd_graph {
				g := generateGraph(topology, mod_graph_byz)
				event := g.GraphToString()
				logEvent(h.ID().String(), false, event)
				g.PrintGraph()
			}
		}

		// Show disjoint Paths
		command, _ = findElement(inputData_words, cmd_djp)
		if command == cmd_djp {
			event := disjointPaths.toEvent()
			logEvent(h.ID().String(), PRINTOPTION, event)
			fmt.Println(disjointPaths.toString())
		}

		// Send detector message
		command, _ = findElement(inputData_words, cmd_detector)
		if command == cmd_detector {
			// Generate an ID for the message
			timestamp := time.Now().Unix()
			hasher := sha1.New()
			hasher.Write([]byte(fmt.Sprintf("%d", timestamp)))
			msgid := fmt.Sprintf("%x", hasher.Sum(nil))
			neighbourhood := topology.ctop.GetNeighbourhood(getNodeAddress(h, ADDR_DEFAULT))
			var detector_message Message = 
			Message{
					ID: msgid, 
					Type: TYPE_DETECTOR, 
					Sender: "", 
					Source: getNodeAddress(h, ADDR_DEFAULT), 
					Target: "",
					Content: "",
					Neighbourhood: neighbourhood,
					Path: []string{},
				}
			
			sendDetector(ctx, h, detector_message)
		}

		// CombinedRC messages
		command, idx = findElement(inputData_words, cmd_crc)
		if command == cmd_crc {
			if len(inputData_words) == 1 {
				fmt.Println("Provide correct input for CombinedRC")
			} else if len(inputData_words) > 1 {
				// Generate an ID for the message
				timestamp := time.Now().Unix()
				hasher := sha1.New()
				hasher.Write([]byte(fmt.Sprintf("%d", timestamp)))
				msgid := fmt.Sprintf("%x", hasher.Sum(nil))
				neighbourhood := topology.ctop.GetNeighbourhood(getNodeAddress(h, ADDR_DEFAULT))
				var visitedSet []string
				var crc_message Message = 
				Message{
					ID: msgid,
					Type: "", 
					Sender: "", 
					Source: getNodeAddress(h, ADDR_DEFAULT), 
					Target: "",
					Content: "",
					Neighbourhood: neighbourhood,
					Path: visitedSet,
				}

				if len(inputData_words) == 2 && inputData_words[idx+1] == mod_crc_exp {
					crc_message.Type = TYPE_CRC_EXP
				} else if len(inputData_words) == 3 && 
							inputData_words[idx+1] == mod_crc_rou &&
							inputData_words[idx+2] != "" {
					crc_message.Type = TYPE_CRC_ROU
					crc_message.Target = inputData_words[idx+2]
				} else if len(inputData_words) > 3 &&
							inputData_words[idx+1] == mod_crc_cnt &&
							inputData_words[idx+2] == cmd_msg {
					crc_message.Type = TYPE_CRC_CNT
					crc_message.Target = inputData_words[idx+3]
					crc_message.Content = extractMessage(inputData)
				}
				// Send crc_exp2 message
				sendCombinedRC(ctx, h, crc_message, topology, disjointPaths)
			}
		}

		command, idx = findElement(inputData_words, cmd_master)
		if command == cmd_master {
			// Generate an ID for the message
			timestamp := time.Now().Unix()
			hasher := sha1.New()
			hasher.Write([]byte(fmt.Sprintf("%d", timestamp)))
			msgid := fmt.Sprintf("%x", hasher.Sum(nil))
			var neighbourhood []string
			var visitedSet []string
			var master_message Message = 
			Message{
				ID: msgid, 
				Type: TYPE_MASTER, 
				Sender: getNodeAddress(h, ADDR_DEFAULT), 
				Source: getNodeAddress(h, ADDR_DEFAULT), 
				Target: "",
				Content: "",
				Neighbourhood: neighbourhood,
				Path: visitedSet,
			}
			if len(inputData_words) == 2 {
				if inputData_words[idx+1] == mst_connect {
					connectNodes(ctx, h, master_address, topology)
				} else if inputData_words[idx+1] == mst_top {
					master_message.Type = mst_top
					sendTopology(ctx, h, master_message)
				} else {
					master_message.Content = inputData_words[idx+1]
					sendMaster(ctx, h, master_message)
				}
			}

			
		}

		// Transform this node into a byzantine
		command, idx = findElement(inputData_words, cmd_byzantine)
		if len(inputData_words) == 1 {
			if command == cmd_byzantine {
				if !byzantine_status {		
					var err error
					bz, err = LoadByzantineConfig(BYZANTINE_CONFIG)
					if err != nil {
						printError(err)
					}		
					color_info = RED
					event := fmt.Sprintf("byzantine - Node %s is now a byzantine", addressToPrint(h.ID().String(), NODE_PRINTLAST))
					logEvent(h.ID().String(), PRINTOPTION, event)
				} else {
					event := fmt.Sprintf("byzantine - Node %s is no more a byzantine", addressToPrint(h.ID().String(), NODE_PRINTLAST))
					logEvent(h.ID().String(), PRINTOPTION, event)
					color_info = GREEN
				}
				byzantine_status = !byzantine_status
			}
		} else if len(inputData_words) == 2  && inputData_words[idx+1] == BYZ_GENERATE {
			// Generate a fake explorer2 message
			timestamp := time.Now().Unix()
			hasher := sha1.New()
			hasher.Write([]byte(fmt.Sprintf("%d", timestamp)))
			msgid := fmt.Sprintf("%x", hasher.Sum(nil))
			var neighbourhood []string
			var visitedSet []string
			var fake_message Message = 
			Message{
				ID: msgid,
				Type: TYPE_CRC_EXP, 
				Sender: getNodeAddress(h, ADDR_DEFAULT), 
				Source: topology.GetRandomNeighbour(), 
				Target: "",
				Content: "",
				Neighbourhood: neighbourhood,
				Path: visitedSet,
			}
			event := fmt.Sprintf("byzantine - Propagating fake message with source %s. . .", addressToPrint(fake_message.Source, NODE_PRINTLAST))
			logEvent(h.ID().String(), PRINTOPTION, event)
			sendCombinedRC(ctx, h, fake_message, topology, disjointPaths)
		}
			
		

		// invoke test function
		command, _ = findElement(inputData_words, "-test")
		if command == "-test" {
			test()
		}
		
	}
}