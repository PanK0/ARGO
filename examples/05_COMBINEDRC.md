p align="center">
  <img src="https://github.com/PanK0/ARGO/blob/main/pictures/ARGO.png?raw=true" alt="ARGO_logo"
    width="20%">
</p>

## ARGO - MASTER-SLAVE - Set up a network from a remote peer

ARGO - Adversarial Robust Graph Operator is a software for testing reliable communication techniques in unknown networks in presence of Byzantine faults. 

### Example Case

In this example, nodes are opened by a dedicated python script on the same machine, with automatic force of the topology. Nodes are also connected at start time to a master node, from which the user can give the master commands to rapidly start the network.

Once nodes are running, they gain their own unique address and they are also connected to a master node.

After the network is properly set up, send the commands from the Master to all the nodes to make them spread the CombinedRC Exploration message.

Once all the node will have the topology information, send a CombinedRC Route message from node B to node C, so that C can automatically save the Disjoint Paths leading to B in its Disjoint Paths Solution.

Then, send a CombinedRC Content message from C to B: this message will cross only the disjoint paths to arrive to B from C.

Finally, request the logs from the Master nodes.

The example consists in:

- **Opening** 4 nodes with the python script that automates topology forcing on the *topology.csv* file
- **Use the master node** to **load** the topology and **connect** all the nodes
- **Use the master node** to spread each node's topology with the CombinedRC Exploration messaage
- Send a CombinedRC Route message from node B to node C
- Send a CombinedRC Content message from node C to node B
- Collect all the *.log* files from the Master node

In this example, we will use the *topology.csv* file in `ARGO/config`.

The file appears to be shaped like this, representing a graph as an adjaciency list:

![topology.csv](https://github.com/PanK0/ARGO/blob/main/pictures/topology.png?raw=true)