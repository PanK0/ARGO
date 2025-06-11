# [UNDER CONSTRUCTION]

<p align="center">
  <img src="https://github.com/PanK0/ARGO/blob/main/pictures/ARGO.png?raw=true" alt="ARGO_logo"
    width="30%">
</p>


# ARGO

ARGO - Adversarial Robust Graph Operator is a software for testing reliable communication techniques in unknown networks in presence of Byzantine faults.

This software is mainly based on the following studies:

- `Discovering Network Topology in the Presence of Byzantine Faults - Nesterenko, Tixeuil`
- `Boosting the Efficiency of Byzantine-Tolerant Reliable Communication - Bonomi, Farina, Tixeuil`
- `Tractable Reliable Communication in Compromised Networks - Giovanni Farina`
 


## Files description

### `config/`
- `byzantine.config`  : configuration file to simulate byzantine processes.
- `topology.csv`      : topology of a 4 nodes graph, given into a .csv file.
- `topology2.csv`     : topology of a 8 nodes graph, given into a .csv file.

### `docker/`
- `dockerfile` : dockerfile to create an image of a node.

### `examples/`
Directory in which are presented some examples on how to start a network

### `logs/`
- `log_parser.py` : script to create a table of events of a node.
- `requirements.txt` : requirements for the `log_parser.py` script.

### `src/`
- `byzantine.go` : operations to set up and configure a byzantine node.
- `constants.go` : constants used in the program.
- `graph.go` : graph management in order to help the reconstruction of the network topology. Implements Ford-Fulkerson algorithm for max flow, that is useful to determine the number of disjoint paths between two endpoints, and other basic graph operations.
- `disjoint_paths.go` : data structure to trace the Disjoint Paths Solution.
. `graph.go` : graph representation of the topology.
- `list.go` : operations on lists.
- `main.go` : main file, where the message lists are stored and the nodes are run.
- `manage_console_input.go` : takes the input given from the user through the console and translates it into operations for the nodes to perform.
- `master.go` : defining a master protocol in order to manage other nodes via a remote one. Useful when working with big networks.
- `message_container.go` : data struct and operations that stores messages and groups them by their ID.
- `message.go` : message type data struct definition.
- `node_operations.go` : creation and connection of nodes, plus some other features.
- `output_print_functions.go` : all the functions used to print the output on the console.
- `protocol_*.go` : filse that describe the protocols.
- `protocols_operations.go` : where the magic happens. Here are implemented the functions that take the messages given in input and send them as direct messages or broadcasts. It also contains the stream handlers, that are supposed to react when a message arrives on the stream.
- `topology.go` : contains topology information, like uTop and cTop and some operations.
- `utils.go` : utility functions.

- `open_nodes.py` : ptyhon script to automatically open *n* nodes.



### File config/topology.csv
In this file it is possible to give a naive representation of a graph.

This file can be used when dealing with a **known topology** a priori, I.E. when it is required that each node knows every other node in the network within its neighbourhood, or either to easily set up a netwrok that is known for the programmer, but can stay unknown for the nodes.

Nodes can force their address into the *topology.csv* file: this operation will change a letter in the file with the node's address (**FORCE TOPOLOGY**, see later).

The location of the *topology.csv* file is saved in a dedicated constant in *constants.go* file.

![topology.csv](https://github.com/PanK0/ARGO/blob/main/pictures/topology.png?raw=true)


### Script src/open_nodes.py
With this script it is possible to open nodes in groups, instead of opening them one by one. (See later)

The script opens the specified number of nodes.

It also offers the possibility to simoultaneusly open multiple nodes AND make the nodes substitute their address with a letter in the file *topology.csv* (This action is called **FORCE TOPOLOGY**), so that, once the *topology.csv* file will be filled with addresses instead of letters, nodes can load the complete topology (or only their neighbourhood) inside their internal structure (**LOAD TOPOLOGY** action).

FORCE action also include LOAD action, while LOAD action doesn't affect the *topology.csv* file.

To do so, inside the script is present an array that **MUST CONTAIN ALL THE EXACT LETTERS** that represent nodes in the *topology.csv* file: that is, if the *topology.csv* file changes, the array in the script **MUST** be modified as well to ensure a proper functioning of the network.

# BUILDING and STARTING the system

## Build
Go in the `ARGO/src` folder.

To generate the **argo** executable, run the following:

```
> go build
```

## Start
The system can be started in various ways:

- [NODE-BY-NODE](https://github.com/PanK0/ARGO/blob/main/examples/01_NODE-BY-NODE.md):     manually, by starting node by node 
- [GROUP-START](https://github.com/PanK0/ARGO/blob/main/examples/02_GROUP-START.md):    	via script *open_nodes.py* by opening the wanted number of nodes
- [AUTO-START](https://github.com/PanK0/ARGO/blob/main/examples/03_AUTO-START.md):   	via script *open_nodes.py* by opening the wanted number of nodes AND automatically force the topology from the *topology.csv* file
- [MASTER-SLAVE](https://github.com/PanK0/ARGO/blob/main/examples/04_MASTER-SLAVE.md):     manually or via script *open_nodes.py*, by passing ```-d "MASTER_ADDRESS"``` as argument

When a node is opened, an multiaddress with a random ID is assigned for the node.

For simplicity, the shown address is the Local Network address, useful for testing purposes to let the nodes communicate inside the same LAN. 

If for some reason (for example, all nodes run on the same machine and there is no internet connection) it is desired to show loopback addresses, constant ```ADDR_DEFAULT``` in file *constants.go* must be modified with the value ```LOOPBACK``` and the software must be built again.


Once a node starts, some instructions are suggested and node's information is printed.

To call a more complete view with all the available commands type `-help` on the shell.

This view can be called again with the ```-help PROTOCOLS``` command.

![commands](https://github.com/PanK0/ARGO/blob/main/pictures/commands.png?raw=true)


## CONNECT 
Once up and running, to communicate nodes must be connected. 

Connections are bidirectional, so if node A is connected to node B, then node B is connected to node A.

Nodes can be connected manually in pairs or, if the neighbourhood is already known (I.E. nodes has LOADED their topology from *topology.csv* file), the action can be automated.

**KEEP IN MIND** that **connections** and **topology information** are strictly related, but also independant: this may cause redundancy in the managed information, but it is necessary to obtain a clear and simple access and focus on what is needed when it's needed.

**!!! WARNING**: connecting node A with node B from node B causes the addition of node A in node B's confirmed topology internal information. **However** node A needs to **ACQUIRE** the topology to add node B to its confirmed topology internal structure.

### Manually connect pair of nodes
If the address of node A is `ADDRESS_A` and the address of node B is `ADDRESS_B`, to connect to node B from node A type on node A's shell the command:

```
> -connect ADDRESS_B
```

### Automatically connect a node with its neighbourhood
It is possible to connect the current node to all neighbours present in the topology. This can be done only after LOADING the topology in the node's topology internal structure, or at least its neighbourhood: it is no more than cycling through the node's neighbours and connecting nodes one by one, but it's done automatically.

```
> -connectall
```


## TOPOLOGY MANAGEMENT
Each node preserves **its own view** of the network topology in its internal structure.

When the topology is known a priori (**KNOWN TOPOLOGY assumption**), all nodes share the same knowledge of the whole topology. This may happen, for example, if the network is started involving *topology.csv* file.

When the topology is not known a priori (**UNKNOWN TOPOLOGY assumption**), nodes must be connected one by one. When two nodes are connected, ```-topology ACQUIRE``` command should be launched to ensure that the nodes are aware of their neighbourhood and registered it into their confirmed topology internal structure (**KNOWN NEIGHBOURHOOD assumption**).


### Topology related commands - SHOW

Show confirmed topology:

```
> -topology SHOW
```

Show both Unconfirmed and Confirmed topology (useful when dealing with EXPLORER protocol):

```
> -topology WHOLE
```

### Topology related commands - unknown topology but known neighbourhood assumption

Acquire topology from the node's network information - basically adds all connected nodes to the current node's confirmed topology internal information:

```
> -topology ACQUIRE
```

### Topology related commands - known topology assumption

Load a new confirmed neighbourhood from the file whose path is saved in *constants.go* at ```topology_path``` - default *topology.csv* file:

```
> -topology LOAD
```

Change node <NODE> in topology.csv with this node's address:

```
> -topology FORCE <NODE>
```


## SENDING MESSAGES

Messages are identified by their type. Types are:
- DIRECTMSG = Direct message
- BROADCAST = Broadcast message
- DETECTOR  = Detector message
- EXPLORER  = Explorer message
- EXPLORER2 = Explorer2 message
- COMBINEDRC= CombinedRC message, divided in EXPLORER2, ROUTE, CONTENT

Once received, messages are placed in a dedicated message container, that is an internal structure of a node. They can also be **DELIVERED** and so moved in another message container for delivered messages. 

Delivery can be performed by invoking the dedicated ```-deliver <FLAG>``` command (More information by running the *-help* command).

![Message Container](https://github.com/PanK0/ARGO/blob/main/pictures/messagecontainer.jpeg?raw=true)

### Send a direct message 
To send a message MESSAGE from node A to node B:

```
> -send ADDRESS_B -msg "MESSAGE"
```


### Send a broadcast message
To send a broadcast with message MESSAGE from any node to node A:

```
> -broadcast address_A -msg "MESSAGE"
```

Run this command into any node that is different from A and it will send the message MESSAGE to all its peers through the network. The broadcast message will circulate in the network until it reaches its destination, if it can, or it will stops naturally under certain conditions.

Once a node X receives a broadcast message, it forwards it to any node that is not in the path.

![Broadcast example](https://github.com/PanK0/ARGO/blob/main/pictures/naive_broadcast_example.png?raw=true)


### Run DETECTOR and EXPLORER protocols
These two protocols are described in *Discovering Network Topology in the Presence of Byzantine Faults - 2009 - Nesterenko, Tixeuil*. 

Their purpose is to reconstruct the network topology under certain conditions (described in the paper).

However, **EXPLORER** has been proven to be wrong (see more [here](http://antares.cs.kent.edu/~mikhail/Research/topology.errata.html)), and so it has to be replaced with **EXPLORER2**, described @ `Tractable Reliable Communication in Compromised Networks, Giovanni Farina - cpt. 9.3, 9.4`

These protocols imply the sending of some messages in the network. These messages **DO NOT USE** the described above broadcast communication primitive, but they implement their own broadcast.

Detector and Explorer can be invoked by running the ```-detector``` and/or ```-explorer``` commands.


## MASTER
By connecting the nodes to a master node M, M can remotely send instructions.

By now, from master it is possible to command all nodes at once:

```
> -master TOPACQUIRE : nodes acquire the topology from their connected peers. Useful when a network is started without a *topology.csv* file
> -master TOPLOAD : nodes load their knowable topology (that is their neighbourhood or the full topology, this can be changed in the code) from the *topology.csv* file
> -master CONNECTALL : nodes establish a connection with all other nodes in their neighbourhood
> -master EXP : nodes send their CombinedRC Exploration Message one by one, with a time interval of one second
> -master GRAPH : nodes produce their graph of the topology
> -master DJP : nodes print their Disjoint Paths Solution computed in respect of other nodes
> -master LOG : master requires the *.log* file from other nodes, that respond with the file. Then master saves the file at *ARGO/logs/r_NODEADDRESS.log*
> -master TOP : master sends the updated topology to all the nodes. Nodes will replace their *topology.csv* file with the one sent by the master
> -master DISCONNECT : master disconnects from the nodes
```
**TO DO** : implement a very well functioning version

# BYZANTINES
Byzantines are processes that may deviate from the normal expected behavior.

Byzantines can be of 3 types:

- Type 1: a process that introduces a delay inside the system
- Type 2: a process that drops the messages with a certain rate and doesn't relay
- Type 3: a process that alters information

It is possible to make a node byzantine by giving the proper command on the node. The interface will turn red to better identify the byzantine:

```
> -byzantine
```

The byzantine process will be loaded coherently with the configuration file in *config/byzantine.config*.

## Byzantine configuration file

The byzantine configuration file has this kind of structure:

```
Type1=true
Type2=false
Type3=false
Delay=500
DropRate=0.3
Alterations=neighbourhood
```

- Type1, Type2 and Type3 entries are trivial: they accept a boolean value true/false
- Delay: accepts an int that indicates the number of milliseconds of delay to introduce in a Type1 byzantine
- DropRate: accepts a float r, with 0 < r < 1, that indicates the probability to drop a message in a Type2 byzantine
- Alterations: accepts a string, that may be `neighbourhood` or `path`. This randomly alterates the content of the specified field of the message by deleting an element. It can also be `swap`, that swaps two nodes in the path of a message.

A Byzantine can also generate a spurious message, I.E. a message with a random source that is different from the actual byzantine node and a void path. To do so, **after activating a byzantine**, give the command:

```
> -byzantine FAKE
```


# LOGS
In `/logs/` are saved logs created by using `logEvent()` function in `utils.go`. You can basically write whatever you want in the logs. 

By inserting this function in the code, it is possible to create a *NODE_ADDRESS.log* file so that the wanted events are saved in the file.

By calling the dedicated python script `log_parser.py` all the saved logs in the directory are put into an excel file that represents a timeline of the recorded events: **after formatting the file**, with low effort it can be a nice resource to have a clear view of what happens in the whole system, and it is very useful especially when testing the network on a local machine.

Install the requirements with 

```
> pip3 install -r requirements.txt
```

Then, when the directory will be full of *.log* files, run the script to create the wanted files:

```
> python3 log_parser.py
```

To clear all the generated files, run:

```
> python3 log_parser.py clear
```

This is an example of the generated timeline after a bit of formatting and coloring:

![Timeline](https://github.com/PanK0/ARGO/blob/main/pictures/timeline.png?raw=true)

**TO DO** : this mechanism could be refined, but for now it works

# DOCKER SETTINGS
**!!! WARNING** : when using docker, the whole network must be run manually, i.e. no automatic mode is available: this is because each docker image is the image of a single node. Thus, nodes don't share the *topology.csv* file.

## Build the image
Build the docker image. From the main folder:

```
docker build -t argo -f docker/Dockerfile .
```

The command builds a docker image named *argo* by using the dockerfile in folder *docker/*.

## Run a container
Run the docker container (one for each node):

```
docker run -dit --rm --name nodeA argo 
```

Where: 
- *-d* runs the container in detached mode (background). Remove *-d* to directly start *argo* bash.
- *-i* keeps standard input open for interactive commands
- *-t* allocates a pseudo-TTY for the container
- *--rm* removes the container after it stops
- *--name nodeA* names the running container as 'nodeA'
- *argo* is the image for which to create a container

## Access the container's shell
By accessing the container's shell it is possible to run the node.

```
docker exec -it nodeA bash
```

Then navigate into the proper folder and run the executable:

```
dockerbash> cd src
dockerbash> ./argo
```

To exit, type:

```
dockerbash> exit
```


# WARNING
Since this is still a test:

- For each message is created a new stream. In a long run, this causes a saturation. To cope with this, a function openStream() is called, that closes currently opened streams.
- To deliver a broadcast it is used a disjoint paths method. To deliver explorer2 messages, intersection is used.

## WARNING - ADDRESSES AND CONNECTIONS
- Nodes multiaddresses can be viewed by uncommenting function `pintNodeInfo()` in *output_print_functions.go*.
- For simplicity, nodes communicate through the Local Area Network (LAN). This option is set in constant `ADDR_DEFAULT` in *constants.go*. In case of **missing internet connection** it would be wise to change the value of `ADDR_DEFAULT` from `LAN` to `LOOPBACK` in order to let the nodes communicate on the local machine.
- When acquiring topology the function `acquireTopology()` in *node_operations.go* is called. This function calls the `addNeighbour()`, defined in *topology.go* that checks whether it already exists another node with the same address in the topology. By now, this check is done on the full addresses of the nodes, not only on their IDs: this means that it is virtually possible to add to the topology the same node with both its Loopback and LAN addresses.