<p align="center">
  <img src="https://github.com/PanK0/ARGO/blob/main/pictures/ARGO.png?raw=true" alt="ARGO_logo"
    width="20%">
</p>

## ARGO - AUTO-START a network

ARGO - Adversarial Robust Graph Operator is a software for testing reliable communication techniques in unknown networks in presence of Byzantine faults.  

### Example Case

In this example, nodes are opened by a dedicated python script on the same machine, with automatic force of the topology, so that the user must only **load** the topology and **connect all** the peers.

Once nodes are running, they gain their own unique address.

The example consists in:

- **opening** 4 nodes with the python script that automates topology forcing on the *topology.csv* file
- **Loading the complete topology** from the *topology.csv* file
- **Connect all** the nodes, automatically

In this example, we will use the *topology.csv* file in `ARGO/config`.

The file appears to be shaped like this, representing a graph as an adjaciency list:

![topology.csv](https://github.com/PanK0/ARGO/blob/main/pictures/topology.png?raw=true)

### Start the nodes

First of all, be sure that the *open_nodes.py* script and the *topology.csv* file are aligned.

To do so, inside the script is present an array that **MUST CONTAIN ALL THE EXACT LETTERS** that represent nodes in the topology.csv file: that is, if the topology.csv file changes, the array in the script MUST be modified as well to ensure a proper functioning of the network.

Locate in the `ARGO/src` folder and run the command

```
> python3 open_nodes.py 4 ./argo auto
```

This command will open 4 different nodes, not yet connected each other, and forces their addresses on the *topology.csv* file.

### Topology load

Once all the nodes are up and running, the *topology.csv* is complete with all the addresses of the network.

There is **no need of forcing** the topology, since it has been done automatically at every node start.

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

Now the network is ready to be used.