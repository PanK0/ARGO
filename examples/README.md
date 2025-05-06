<p align="center">
  <img src="https://github.com/PanK0/ARGO/blob/main/pictures/ARGO.png?raw=true" alt="ARGO_logo"
    width="20%">
</p>

## ARGO - Examples

ARGO - Adversarial Robust Graph Operator is a software for testing reliable communication techniques in unknown networks in presence of Byzantine faults.

## BUILD

Go in the `ARGO/src` folder.

To generate the **argo** executable, run the following:

```
> go build
```

## EXAMPLES

Here are provided some examples on how to set up and run the network.

For more specific information, visit the single examples.

The system can be started in various ways:

- NODE-BY-NODE:     manually, by starting node by node 
- NODE-BY-AUTO:     manually with automatic topology load: start every node one by one and replace its address with the correspondant letter in the *topology.csv* file
- GROUP-BY-NODE:    via script *open_nodes.py* by opening the wanted number of nodes
- GROUP-BY-AUTO:    via script *open_nodes.py* by opening the wanted number of nodes AND automatically force the topology from the *topology.csv* file
- MASTER-SLAVE:     manually or via script *open_nodes.py*, by passing ```-d "MASTER_ADDRESS"``` as argument

When a node is opened, an multiaddress with a random ID is assigned for the node.

For simplicity, the shown address is the Local Network address, useful for testing purposes to let the nodes communicate inside the same LAN. 

If for some reason (for example, all nodes run on the same machine and there is no internet connection) it is desired to show loopback addresses, constant ```ADDR_DEFAULT``` in file *constants.go* must be modified with the value ```LOOPBACK``` and the software must be built again.


Once a node starts, some instructions are suggested and node's information is printed.

To call a more complete view with all the available commands type `-help` on the shell.

This view can be called again with the ```-help PROTOCOLS``` command.

![commands](https://github.com/PanK0/ARGO/blob/main/pictures/commands.png?raw=true)

### NODE-BY-NODE mode
Open nodes manually, by starting node by node
To simply run a node, open a terminal and run

```
> ./argo
```

It is possible to run multiple nodes.

### NODE-BY-AUTO mode
Open node by node, one per time, and modify neighbourhood information contained in *topology.csv* file with the node's address.

This procedure is the same as running a node and then giving the ```-topology FORCE <NODE>``` command: basically the opened node substitutes its node adrress to the correspondant letter in the *topology.csv* file.

For example, to run a node in automatic mode and FORCE its address in the topology.csv file replacing it with all 'A' occurrencies:

```
> ./argo -m auto -n A
```

After applying the changes in the *topology.csv* file, neighbourhood is loaded on the node's internal structure. However the loaded neighbourhood may be composed by a mix of node addresses and single letters: this means that not all nodes have forced their topology yet, so it may be required to LOAD again the topology, maybe when the last node forced its address, by using the command ```-topology LOAD```.

**KEEP IN MIND** that running the whole network in this mode implies that only the last opened node will have a complete view on the correct topology with addresses (and not with letters), so it is necessary to **LOAD** the topology on every other node by running the ```-topology LOAD``` command on each node.

**!!! WARNING**: *topology.csv* file will be permanently modified after each FORCE action such that nodes addresses will replace the letters indicating generic nodes. This means that to perform again a FORCE operation, the *topology.csv* file must be restored to its original state, with letters instead of addresses. Take into account that node addresses are randomically generated when a node is started, so it is very unlikely that two nodes in a certain time of this universe's life will have the same address.

### GROUP-BY-NODE mode: run multiple nodes with the dedicated script
Open the console in the same directory of the script `open_nodes.py`. 

The script will accept as arguments <number_of_terminals> and <name_of_the_executable>, in this order.

For example, to open three nodes run:

```
> python3 open_nodes.py 3 ./argo
```

### GROUP-BY-AUTO mode: run multiple nodes with automatic topology FORCE

Open multiple nodes and make them modify topology information contained in *topology.csv* file with the node's address.

This procedure is the same as repeating NODE-BY-AUTO mode multiple times, each time FORCING a different node in the *topology.csv* file. Nodes (represented by letters) in the *topology.csv* file are automatically replaced with each node's address.

To do so, inside the script *open_nodes.py* is present an array that **MUST CONTAIN ALL THE EXACT LETTERS** that represent nodes in the *topology.csv* file: that is, if the *topology.csv* file changes, the array in the script **MUST** be modified as well to ensure a proper functioning of the network. So, be sure that the `nodes` array in `open_nodes.py` is consistent in quantity and value with the nodes in file `topology.csv`

For example, to run four nodes and make them automatically FORCE their address into the *topology.csv* file:

```
> python3 open_nodes.py 4 ./argo auto
```

After applying the changes in the *topology.csv* file, neighbourhood is loaded on the node's internal structure. However the loaded topology may be composed by a mix of node addresses and single letters: this means that not all nodes have forced their topology yet, so it may be required to LOAD again the topology, maybe when the last node forced its address, by using the command ```-topology LOAD```.

**KEEP IN MIND** that running the whole network in this mode implies that only the last opened node will have a complete view on the correct topology with addresses (and not with letters), so it is necessary to **LOAD** the topology on every other node by running the ```-topology LOAD``` command on each node.

**!!! WARNING**: *topology.csv* file will be permanently modified after each FORCE action such that nodes addresses will replace the letters indicating generic nodes. This means that to perform again a FORCE operation, the *topology.csv* file must be restored to its original state, with letters instead of addresses. Take into account that node addresses are randomically generated when a node is started, so it is very unlikely that two nodes in a certain time of this universe's life will have the same address.


### MASTER-SLAVE
Open a node (or multiple nodes) and connect them to the master address.

**!!! WARNING: MASTER NODE** is invisible to the topology. However it may be affected by protocol message exchanges if not closed before starting the experiments.

To open one node and connect to the master:

```
> ./argo -d MASTER_ADDRESS
```

To open three nodes and connect to the master:

```
> python3 open_nodes.py 3 ./argo -d MASTER_ADDRESS
```

It also work in automatic mode (omit the -d):

```
> python3 open_nodes.py 8 ./argo auto MASTER_ADDRESS
```

