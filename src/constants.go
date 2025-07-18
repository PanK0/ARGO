package main

var master_address string = ""
var thisnode_address string = ""
var color_info = GREEN
var color_desc = CYAN
var byzantine_status = false
var MAX_BYZANTINES		= 0

const (
	// Byzantine related constants
	BYZANTINE_CONFIG	= "../config/byzantine.config"
	BYZ_NEIGHBOURHOOD	= "neighbourhood"
	BYZ_PATH			= "path"
	BYZ_SWAP_PATH		= "swap"
	BYZ_ALTER_ID		= "msgid"
	BYZ_GENERATE		= "FAKE"

	// Address related constants
	ADDR_DEFAULT	= "LAN"
	ADDR_LOOPBACK	= "LOOPBACK"
	ADDR_LAN		= "LAN"
	ADDR_LB_POS		= 0
	ADDR_LAN_POS	= 3
	WHOLE_ADDR		= -1				// For a node, print the whole address
	NODE_PRINTLAST	= 5					// For a node, print only the last n characters of its address

	// Protocol related constants
	PROTOCOL_CHAT	= "/chat/"
	PROTOCOL_NAB	= "/nab/"			// This is actually a Byzantine Reliable Broadcast, granted by DolevU protocol
	PROTOCOL_EXP	= "/exp/"			// Protocol fo Explorer algorithm (WARNING: explorer has been proved wrong)
	PROTOCOL_DET	= "/det/"			// Protocol for Detector algorithm
	PROTOCOL_EXP2	= "/exp2/"			// Protocol for Explorer2 algorithm
	PROTOCOL_MST	= "/mst/"			// Protocol to manage master-slave operations
	PROTOCOL_CRC	= "/crc/"			// Protocol for CombinedRC algorithm

	TYPE_BROADCAST	= "BROADCAST"
	TYPE_DIRECT_MSG	= "DIRECTMSG" 
	TYPE_DETECTOR	= "DETECTOR"
	TYPE_EXPLORER	= "EXPLORER"
	TYPE_EXPLORER2	= "EXPLORER2"
	TYPE_MASTER		= "MASTER"
	TYPE_CRC		= "COMBINED RC"
	TYPE_CRC_CNT	= "COMBINEDRC_CNT"	// Content type for combinedRC message exchange
	TYPE_CRC_ROU	= "COMBINEDRC_ROU"	// Route type for combinedRC message exchange
	TYPE_CRC_EXP	= "COMBINEDRC_EXP"	// Exploration type for combinedRC message exchange

	// Commands
	cmd_help 		= "-help"
	cmd_info 		= "-info"
	cmd_connect 	= "-connect"
	cmd_connect_all	= "-connectall"
	cmd_send 		= "-send"
	cmd_broadcast	= "-broadcast"
	cmd_detector	= "-detector"
	cmd_explorer	= "-explorer"
	cmd_explorer2	= "-exp2"
	cmd_deliver		= "-deliver"
	cmd_msg 		= "-msg"
	cmd_show		= "-show"	
	cmd_network		= "-network"
	cmd_topology	= "-topology"
	cmd_graph		= "-graph"
	cmd_djp			= "-djp"
	cmd_master		= "-master"
	cmd_byzantine	= "-byzantine"
	cmd_crc			= "-crc"

	// Master commands
	mst_top_acquire	= "TOPACQUIRE"
	mst_top_load	= "TOPLOAD"
	mst_connectall	= "CONNECTALL"
	mst_disconnect	= "DISCONNECT"
	mst_connect 	= "CONNECT"
	mst_crc_exp		= "EXP"
	mst_graph		= "GRAPH"
	mst_djp			= "DJP"
	mst_log			= "LOG"
	mst_top			= "SENDTOP"
	mst_printprot	= "PROTOCOLS"
	mst_reset		= "RESET"
	
	// Mode in which some commands are called
	mod_help_def	= "DEFAULT"
	mod_help_help	= "HELP"
	mod_help_node	= "NODE"
	mod_help_prot	= "PROTOCOLS"
	mod_help_msg	= "MSG"
	mod_help_net	= "NETWORK"
	mod_help_mst	= "MASTER"
	mod_show_del	= "DEL"
	mod_show_rcv	= "RCV"
	mod_deliver_all	= "ALL"
	mod_top_show	= "SHOW"
	mod_top_whole	= "WHOLE"
	mod_top_acquire	= "ACQUIRE"
	mod_top_load	= "LOAD"
	mod_top_neigh	= "NEIGH"
	mod_top_force	= "FORCE"
	mod_crc_exp		= "EXP"
	mod_crc_rou		= "ROU"
	mod_crc_cnt		= "SEND"
	mod_graph_byz	= true
	
	
	start_automatic	= "auto"			// Start the network automatically
	topology_path 	= "../config/6nodes_3connected.csv"	// Path to the topology file

	// Colors for fancy terminal printing
	RED		= "\x1b[31m"
	GREEN 	= "\x1b[32m"
	YELLOW 	= "\x1b[33m"
	BLUE 	= "\x1b[34m"
	CYAN 	= "\x1b[36m"
	WHITE	= "\x1b[37m"
	GREY	= "\x1b[90m"
	MAGENTA	= "\x1b[95m"
	RESET	= "\x1b[0m"
	
	RED_BG		= "\x1b[41m"
	GREEN_BG 	= "\x1b[42m"
	YELLOW_BG 	= "\x1b[43m"
	BLUE_BG 	= "\x1b[44m"
	CYAN_BG 	= "\x1b[46m"
	WHITE_BG	= "\x1b[47m"
	GREY_BG		= "\x1b[100m"
	MAGENTA_BG 	= "\x1b[45m"

	// Log constants
	LOGDIR		= "../logs"
	PRINTOPTION	= true
	
)