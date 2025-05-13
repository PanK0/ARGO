<p align="center">
  <img src="https://github.com/PanK0/ARGO/blob/main/pictures/ARGO.png?raw=true" alt="ARGO_logo"
    width="20%">
</p>

## ARGO - MASTER-SLAVE - Set up a network from a remote peer

ARGO - Adversarial Robust Graph Operator is a software for testing reliable communication techniques in unknown networks in presence of Byzantine faults. 

### Example Case

In this example, nodes are opened by a dedicated python script on the same machine, with automatic force of the topology. Nodes are also connected at start time to a master node, from which the user can give the master commands to rapidly start the network.

Once nodes are running, they gain their own unique address and they are also connected to a master node.

The example consists in:

- **Opening** 4 nodes with the python script that automates topology forcing on the *topology.csv* file
- **Use the master node** to **load** the topology and **connect** all the nodes

In this example, we will use the *topology.csv* file in `ARGO/config`.

The file appears to be shaped like this, representing a graph as an adjaciency list:

![topology.csv](https://github.com/PanK0/ARGO/blob/main/pictures/topology.png?raw=true)

### Start the nodes

Open the master node: locate in the `ARGO/src` folder and run the command

```
> ./argo
```

**The master node doesn't need to be specified anywhere in the *topology.csv* file, nor in the python script letters array.**

In this example, the master node appears to have this address:

![master node](https://github.com/PanK0/ARGO/blob/main/pictures/ex_masterslave_master.png?raw=true)


Then proceed by opening the other nodes.

Be sure that the *open_nodes.py* script and the *topology.csv* file are aligned.

To do so, inside the script is present an array that **MUST CONTAIN ALL THE EXACT LETTERS** that represent nodes in the topology.csv file: that is, if the topology.csv file changes, the array in the script MUST be modified as well to ensure a proper functioning of the network.

Locate in the `ARGO/src` folder and run the command

```
> python3 open_nodes.py 4 ./argo auto MASTER_ADDRESS
```

That, in our case, is:

```
> python3 open_nodes.py 4 ./argo auto /ip4/192.168.1.7/tcp/41111/p2p/12D3KooWRs3ynee96N2CRz9rh9H2orX66wYWtUMqwcRDDmBq74Pk
```

This command will open 4 different nodes, not yet connected each other, and forces their addresses on the *topology.csv* file and also connects every single node to a master node, previously opened. 

As it is possible to see, nodes start already connected to the master node.

### Use the master node to set up the network

From the master, give the commands, in order:

```
> -master TOP_LOAD
```

and

```
> -master CONNECTALL
```

Then, close the master node by giving the command:

```
> -master DISCONNECT
```

The other nodes will receive the instructions from the master and will print their actions on their own shell: after closing the master, the network is up and ready to be used.

![subordinate nodes](https://github.com/PanK0/ARGO/blob/main/pictures/ex_masterslave_subordinates.png?raw=true)
