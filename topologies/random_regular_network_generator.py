import sys
import csv
import networkx as nx


# Main execution
if __name__ == "__main__":
    if len(sys.argv) < 3 :
        print("Usage: python3 script.py <k=connectivity> <n=nodes>")
        sys.exit(1)
    
    k = int(sys.argv[1])
    n = int(sys.argv[2])

    print(f"Generating random regular network with k={k} and n={n}")
    G = nx.random_regular_graph(k, n)

    if n <= 26 :
        mapping = {i : chr(65+i) for i in G.nodes}
        G = nx.relabel_nodes(G, mapping)

    for node, neighbors in G.adjacency() :
        print(f"{node} : {list(neighbors)}")

    filename = f"{n}nodes_{k}connected.csv"
    with open(filename, mode="w", newline="") as file :
        writer = csv.writer(file)

        # Write header
        #writer.writerow(["NODE", "NEIGHBORS", ""])
        writer.writerow(["NODE", "NEIGHBORS"] + [""] * (k - 1))

        # Write nodes and neighbors
        for node, neighbors in G.adjacency() :
            writer.writerow([node] + list(neighbors))

    print("DONE\n")
    print(G.nodes)