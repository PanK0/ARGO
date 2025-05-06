<p align="center">
  <img src="https://github.com/PanK0/ARGO/blob/main/pictures/ARGO.png?raw=true" alt="ARGO_logo"
    width="20%">
</p>

## ARGO - GROUP-START Network Setup

ARGO - Adversarial Robust Graph Operator is a software for testing reliable communication techniques in unknown networks in presence of Byzantine faults.  

### Example Case

In this example, nodes are opened by a dedicated python script on the same machine.

Once nodes are running, they gain their own unique address.

The example consists in:

- **opening** 4 nodes with the python script
- **Forcing topology** on the *topology.csv* file
- **Loading topology** from the *topology.csv* file
- **Connect all** the nodes, automatically
- **Send a broadcast message** from a *source* node with the purpose of reaching a *target* node

In this example, we will use the *topology.csv* file in `ARGO/config`.

The file appears to be shaped like this, representing a graph as an adjaciency list:

![topology.csv](https://github.com/PanK0/ARGO/blob/main/pictures/topology.png?raw=true)



### Start the nodes

First of all, be sure that the *open_nodes.py* script and the *topology.csv* file are aligned.

To do so, inside the script is present an array that **MUST CONTAIN ALL THE EXACT LETTERS** that represent nodes in the topology.csv file: that is, if the topology.csv file changes, the array in the script MUST be modified as well to ensure a proper functioning of the network.

Locate in the `ARGO/src` folder and run the command

```
> python3 open_nodes.py 4 ./argo
```

This command will open 4 different nodes, not yet connected each other.

For simplicity's sake, let's identify the nodes with the correspondant letter of the *topology.csv* file:

![groupstart.csv](https://github.com/PanK0/ARGO/blob/main/pictures/ex_groupstart.png?raw=true)

### Topology force

Once the nodes are up and running, **force the topology**.

> What does **FORCING** the topology mean?
> Forcing the topology means that each node replaces the correspondant letter inside the *topology.csv* file
> with its address.
> This can be done to avoid connecting manually node per node, giving instead only two commands per node.

On every node, run the command

```
> -topology FORCE <NODE>
```

Taking care to replace the <NODE> argument with the letter you want to change the node's address with.

This operation must be done on each node, changing the proper letter.

Forcing topology on a node will permanently affect the *topology.csv* file, replacing the letters with the node's address. 

What can be seen is that, while going on forcing the topology on every node, the previous node's address is shown in the next one: this means that the node's address has been correctly replaced.

![groupstart.csv](https://github.com/PanK0/ARGO/blob/main/pictures/ex_groupstart_topacquire.png?raw=true)

### Topology load

At the end, after forcing the topology on every single node, the *topology.csv* is complete with all the addresses of the network.

Now it is time to **load** the topology from the file, so that each node will have a complete view of its neighbourhood. Run this command on every node:

```
> -topology LOAD
```

### Connect all the nodes

To automatically connect a node with all its neighbours in the topology, run the command 

```
> -connectall
```

Run this command on each node's shell.

### Send broadcast

To send a broadcast with message MESSAGE from a node X to a node W, run this command on node X terminal:

```
> -broadcast <ADDRESS_W> -msg "MESSAGE"
```

Run this command into a node that is different from W and it will send the message MESSAGE to all its peers through the network. The broadcast message will circulate in the network until it reaches its destination, if it can, or it will stops naturally under certain conditions.

Once a node receives a broadcast message, it forwards it to any node that is not in the path.

After sending a broadcast message, each node will decide to forward it or not, until it is sure the message reached the destination (if possible).

So, after giving this command from node A:

```
> -broadcast /ip4/192.168.1.7/tcp/33149/p2p/12D3KooWGTqsipVdEqPqFynwrzVLd5dM1TLoEMgNXsTxr6d7U1Cy -msg "Hello there!"
```

that is a broadcast message which target is node W, what appears in node W is this:

![groupstart broadcast.csv](https://github.com/PanK0/ARGO/blob/main/pictures/ex_groupstart_broadcast.png?raw=true)

It is possible to notice that node W received several copies of the same message (see the field MSG ID), with the same content, but each copy traveled from a different path.