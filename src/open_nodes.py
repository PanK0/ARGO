import os
import sys
import time
import subprocess

# Array of node identifiers (cycling through these in auto mode)
nodes4 = ["A", "B", "C", "W"]
nodes5 = ["A", "B", "C", "D", "W"]
nodes8 = ["A", "B", "C", "D", "E", "F", "G", "W"]
nodes9 = ["A", "B", "C", "D", "E", "F", "X", "Y", "Z"]

def open_terminals(n_shells, command, automod=None, address=None):
    """
    Open multiple terminal windows and execute the given command.
    
    :param n_shells: Number of terminals to open
    :param command: Base command to execute
    :param automod: If "auto", assigns node names
    :param address: Optional address to append (-d ADDRESS)
    """
    if n_shells == 4:
        nodes_l = nodes4
    elif n_shells == 5:
        nodes_l = nodes5
    elif n_shells == 8:
        nodes_l = nodes8
    elif n_shells == 9:
        nodes_l = nodes9
    else:
        print("Error: n_shells must be 4 or 8 when using automod.")
        sys.exit(1)

    for i in range(n_shells):
        full_command = command  # Start with the base command
        
        # If automod is enabled, add -m auto -n <node>
        if automod == "auto":
            node = nodes_l[i % len(nodes_l)]  # Cycle through node names
            full_command += f" -m auto -n {node}"

        # Append -d ADDRESS if provided
        if address:
            full_command += f" -d {address}"

        # Open the terminal and execute the command
        subprocess.Popen(["konsole", "--workdir", os.getcwd(), "-e", "bash", "-i", "-c", f"{full_command}; exec bash"])
        
        print(f"Terminal {i + 1}: {full_command}")
        time.sleep(1)  # Short delay to avoid race conditions

    print(f"{n_shells} terminals opened in directory {os.getcwd()}")

if __name__ == "__main__":
    # Ensure correct number of arguments
    if len(sys.argv) < 3:
        print("Usage: python3 open_nodes.py <n_shells> <command> [automod] [ADDRESS]")
        sys.exit(1)

    # Parse arguments
    n_shells = int(sys.argv[1])
    command = sys.argv[2]
    automod = sys.argv[3] if len(sys.argv) > 3 else None
    address = sys.argv[4] if len(sys.argv) > 4 else None

    # Run the function
    open_terminals(n_shells, command, automod, address)
 
