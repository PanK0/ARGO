<p align="center">
  <img src="https://github.com/PanK0/ARGO/blob/main/pictures/ARGO.png?raw=true" alt="ARGO_logo"
    width="20%">
</p>

## ARGO - NODE-BY-NODE Network Setup

ARGO - Adversarial Robust Graph Operator is a software for testing reliable communication techniques in unknown networks in presence of Byzantine faults. 

### Example Case

In this example, nodes are opened one by one on the same machine.

Once nodes are running, they gain their own unique address.

The example consists in:

- **connecting** nodes
- **topology acquisition**
- **send a direct message** between two nodes

### Start the nodes

Locate in the `ARGO/src` folder.

Open nodes manually, by starting node by node.

To simply run a node, open a terminal and run

```
> ./argo
```

It is possible to run multiple nodes.

In this phase, nodes are not connected each other and they have no information about any given topology.

![node-by-node](https://github.com/PanK0/ARGO/blob/main/pictures/ex_nodebynode.png?raw=true)

### Connect the nodes

Since no topology information is given, nodes must be connected manually.

To do so, run the `-connect` command on the shell.

If the address of node A is `ADDRESS_A` and the address of node B is `ADDRESS_B`, to connect to node B from node A type on node A's shell the command:

```
> -connect ADDRESS_B
```

In this example, we are going to build a completely connected network, i.e. a node is connected with all other nodes in the network.

So, connect each node with all the others. The result in every single shell should be similar to the following one, with the proper differences in the addresses:

![node-by-node connect](https://github.com/PanK0/ARGO/blob/main/pictures/ex_nodebynode_connect.png?raw=true)

### Topology acquisition

The logic that manages protocols and topology is divided by the logic that manages actual connections for simulation purposes.

So, even if the nodes are practically connected, they need to **acquire** the topology, to be sure that all protocols will be correctly run.

Acquiring the topology means that each node reads the nodes it is connected to and saves their addresses in an internarl data structure, namely `cTop` (Confirmed Topology). Acquire the topology on each node with the following command:

```
> -topology ACQUIRE
```

The output on a single should be similar to the one below (only the last 5 chars of each address are printed, this specification can be changed in the code):

![node-by-node topacquire](https://github.com/PanK0/ARGO/blob/main/pictures/ex_nodebynode_topacquire.png?raw=true)

To **show** the topology, give the command: 

```
> -topology SHOW
```

Keep in mind that the shown topology of a node is the one that is known to that specific node, at least at startup. 

Topology can be discovered by running the proper protocols.

### Send a direct message

To send a message MESSAGE from node A to node B, on node A's shell give the command:

```
> -send ADDRESS_B -msg "MESSAGE"
```
The nodes **must be connected**, otherwise an error is raised.

The picture shows the process of sending (left) and receiving (right) node of a direct message:

![node-by-node send](https://github.com/PanK0/ARGO/blob/main/pictures/ex_nodebynode_send.png?raw=true)

