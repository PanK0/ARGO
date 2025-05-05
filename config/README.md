 
<p align="center">
  <img src="https://github.com/PanK0/ARGO/blob/main/pictures/ARGO.png?raw=true" alt="ARGO_logo"
    width="30%">
</p>

## ARGO - Configuration files

ARGO - Adversarial Robust Graph Operator is a software for testing reliable communication techniques in unknown networks in presence of Byzantine faults.

### `config/`
- `byzantine.config`  : configuration file to simulate byzantine processes.
- `topology.csv`      : topology of a 4 nodes graph, given into a .csv file.
- `topology2.csv`     : topology of a 8 nodes graph, given into a .csv file.

### File topology.csv
In this file it is possible to give a naive representation of a graph.

This file can be used when dealing with a **known topology** a priori, I.E. when it is required that each node knows every other node in the network within its neighbourhood, or either to easily set up a netwrok that is known for the programmer, but can stay unknown for the nodes.

Nodes can force their address into the *topology.csv* file: this operation will change a letter in the file with the node's address (**FORCE TOPOLOGY**, see later).

The location of the *topology.csv* file is saved in a dedicated constant in *constants.go* file.

![topology.csv](https://github.com/PanK0/ARGO/blob/main/pictures/topology.png?raw=true)


### File byzantine.config Byzantine configuration file

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
- Alterations: accepts a string, that may be `neighbourhood` or `path`. This randomly alterates the content of the specified field of the message by deleting an element.