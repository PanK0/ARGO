import shutil
import string
import pandas as pd
import re
import glob
import sys
import os
import matplotlib.pyplot as plt
import matplotlib.dates as mdates
import glob

csv_files = glob.glob("../config/*.csv")
if not csv_files:
    print("No CSV file found in ../config/")
    sys.exit()
config_csv_path = csv_files[0]  # Take the first CSV file found

# Function to parse log files and extract timestamps, node IDs, and events
def parse_logs(log_files):
    log_entries = []

    for log_file in log_files:
        with open(log_file, "r") as file:
            for line in file:
                match = re.match(r"\[(.*?)\] \[(.*?)\] (.*)", line.strip())
                if match:
                    timestamp, node_id, event = match.groups()
                    log_entries.append({"Timestamp": timestamp, "Node": node_id, "Event": event})

    return log_entries

# Function to create a structured table with nodes as rows and events as columns
def create_event_table(log_entries):
    df = pd.DataFrame(log_entries)
    
    # Convert timestamps to datetime for sorting
    df["Timestamp"] = pd.to_datetime(df["Timestamp"], format='mixed')
    df = df.sort_values(by="Timestamp")

    # Pivot table: Nodes as rows, Events as columns
    event_table = df.pivot(index="Node", columns="Timestamp", values="Event")

    return event_table, df  # Return both full log and event table

# Function to save logs to Excel and CSV
def save_to_files(event_table, full_log, excel_output="log_output.xlsx", csv_output="log_output.csv"):
    # Save event_table only if it has a manageable number of columns
    max_excel_columns = 16384
    with pd.ExcelWriter(excel_output) as writer:
        '''
        if event_table.shape[1] <= max_excel_columns:
            event_table.to_excel(writer, sheet_name="Event Table")
        else:
            print(f"Event Table NOT saved to Excel: too many columns ({event_table.shape[1]})")
        '''
        full_log.to_excel(writer, sheet_name="Full Log", index=False)
    # Always save both as CSV (no column limit)
    event_table.to_csv("event_table.csv")
    full_log.to_csv(csv_output, index=False)
    print(f"Logs saved to {excel_output} and {csv_output}")

# Function to filter logs by event type
def filter_logs_by_event(log_entries, event_type):
    return [entry for entry in log_entries if event_type in entry["Event"]]

# Function to plot event timeline
def plot_event_timeline(df):
    fig, ax = plt.subplots(figsize=(12, 6))

    # Assign unique colors for each node
    node_colors = {node: plt.cm.tab10(i) for i, node in enumerate(df["Node"].unique())}

    for _, row in df.iterrows():
        timestamp = row["Timestamp"]
        node = row["Node"]
        event = row["Event"]

        # Plot each event as a scatter point
        ax.scatter(timestamp, node, color=node_colors[node], label=node if node not in ax.get_legend_handles_labels()[1] else "")

        # Annotate event names
        ax.text(timestamp, node, event, fontsize=9, verticalalignment="bottom", horizontalalignment="right", rotation=30)

    # Formatting the plot
    ax.set_xlabel("Timestamp")
    ax.set_ylabel("Node")
    ax.set_title("Event Timeline Across Nodes")
    ax.xaxis.set_major_formatter(mdates.DateFormatter("%H:%M:%S"))
    plt.xticks(rotation=45)
    plt.grid(True, linestyle="--", alpha=0.7)
    
    # Add legend for nodes
    plt.legend(title="Nodes", loc="upper left", bbox_to_anchor=(1, 1))

    # Show the plot
    plt.tight_layout()
    plt.savefig("event_timeline.png", dpi=300)


# Function to delete all generated files
def clear_logs():
    file_patterns = ["*.log", "*.csv", "*.xlsx", "*.png"]
    deleted_files = 0

    for pattern in file_patterns:
        for file in glob.glob(pattern):
            try:
                os.remove(file)
                print(f"Deleted: {file}")
                deleted_files += 1
            except Exception as e:
                print(f"Error deleting {file}: {e}")

    if deleted_files == 0:
        print("No files found to delete.")
    else:
        print(f"Deleted {deleted_files} files.")

# Function to replace all the topology files from the main folder ../ to ../config/
def replace_topology_files():
    main_folder = "../"
    config_folder = "../config/"
    for file in glob.glob(os.path.join(main_folder, "topology*.csv")):
        shutil.copy(file, config_folder)
        print(f"Replaced: {file} -> {config_folder}")

def update_log_output_with_node_letters():
    nodes_file = "nodes.txt"
    log_output_file = "log_output.xlsx"

    # Check if nodes.txt exists
    if not os.path.isfile(nodes_file):
        print("nodes.txt not found.")
        return

    # Parse nodes.txt to build mapping {node_id: letter}
    mapping = {}
    with open(nodes_file, "r") as f:
        for line in f:
            match = re.search(r"Topology updated: node (\w) -> (\w{5})", line)
            if match:
                letter, node_id = match.group(1), match.group(2)
                mapping[node_id] = letter

    if not mapping:
        print("No node mappings found in nodes.txt.")
        return

    # Read all sheets from log_output.csv (if multi-sheet, e.g. Excel)
    try:
        # Try Excel first (multi-sheet)
        xls = pd.ExcelFile(log_output_file)
        sheets = {sheet: xls.parse(sheet) for sheet in xls.sheet_names}
        is_excel = True
    except Exception:
        # Fallback to single CSV
        df = pd.read_csv(log_output_file, dtype=str)
        sheets = {"Sheet1": df}
        is_excel = False

    # Replace node_id with node_id + " " + letter in all sheets
    for sheet_name, df in sheets.items():
        for node_id, letter in mapping.items():
            df.replace(node_id, f"{node_id} {letter}", inplace=True)
        sheets[sheet_name] = df

    # Write back the updated data
    if is_excel:
        with pd.ExcelWriter(log_output_file, engine='openpyxl', mode='w') as writer:
            for sheet_name, df in sheets.items():
                df.to_excel(writer, sheet_name=sheet_name, index=False)
    else:
        # Single CSV
        sheets["Sheet1"].to_csv(log_output_file, index=False)

    print("log_output.csv updated with node letters.")

def convert_topology_and_mark_byzantines(config_csv_path=config_csv_path, log_xlsx_path="log_output.xlsx"):
    # 1. Read the topology CSV (adjacency list)
    topo_df = pd.read_csv(config_csv_path, header=0)
    # Flatten all node addresses in the first column
    node_addresses = topo_df.iloc[:, 0].dropna().unique().tolist()

    # 2. Read the "Full Log" sheet and find byzantine nodes
    log_df = pd.read_excel(log_xlsx_path, sheet_name="Full Log", engine="openpyxl")
    byz_nodes = set()
    for event, node in zip(log_df["Event"], log_df["Node"]):
        match = re.match(r"byzantine - Node (\w{5}) is now a byzantine", event)
        if match:
            byz_nodes.add(match.group(1))

    # 3. Map node addresses to letters, mark byzantines as lowercase
    letter_map = {}
    letters = list(string.ascii_uppercase)
    for i, addr in enumerate(node_addresses):
        node_id = addr[-5:]
        if node_id in byz_nodes:
            letter_map[addr] = letters[i].lower()
        else:
            letter_map[addr] = letters[i]

    # 4. Rewrite topology with letters
    new_topo = []
    for _, row in topo_df.iterrows():
        node_addr = row.iloc[0]
        neighbours = [n for n in row.iloc[1:] if pd.notnull(n)]
        node_letter = letter_map.get(node_addr, node_addr)
        neighbour_letters = [letter_map.get(n, n) for n in neighbours]
        new_topo.append([node_letter] + neighbour_letters)

    # 5. Prepare original topology with only last 5 chars for each node
    original_topo = []
    for _, row in topo_df.iterrows():
        node_addr = row.iloc[0]
        neighbours = [n for n in row.iloc[1:] if pd.notnull(n)]
        node_id = node_addr[-5:] if isinstance(node_addr, str) else node_addr
        neighbour_ids = [n[-5:] if isinstance(n, str) else n for n in neighbours]
        original_topo.append([node_id] + neighbour_ids)

    # 6. Combine: modified topology, empty row, original topology (last 5 chars)
    combined_topo = new_topo + [[""] * topo_df.shape[1]] + original_topo

    # Save to the same .xlsx file in a new sheet "topology"
    combined_topo_df = pd.DataFrame(combined_topo)
    with pd.ExcelWriter(log_xlsx_path, engine="openpyxl", mode="a", if_sheet_exists="replace") as writer:
        combined_topo_df.to_excel(writer, sheet_name="topology", header=False, index=False)
    print(f"Topology with letters/byzantines and original (last 5 chars) saved to '{log_xlsx_path}' in sheet")

# Main execution
if __name__ == "__main__":

    # Check if the script was run with "clear"
    if len(sys.argv) > 1 and sys.argv[1] == "clear":
        clear_logs()
        #replace_topology_files()
        sys.exit()

    # Find all log files in the directory
    log_files = glob.glob("*.log")  # Matches all .log files in the current directory
    
    if not log_files:
        print("No log files found.")
        exit()

    logs = parse_logs(log_files)
    
    if logs:
        # Generate event table and full log data
        event_table, full_log = create_event_table(logs)

        # Save to Excel and CSV
        save_to_files(event_table, full_log)

        # Update log output with node letters
        update_log_output_with_node_letters()

        # Write topology
        convert_topology_and_mark_byzantines()

        # Plot event timeline
        # plot_event_timeline(full_log)

        # Example: Filter logs for "handleExplorer2()"
        """
        filtered_logs = filter_logs_by_event(logs, "handleExplorer2()")
        if filtered_logs:
            print("\nFiltered Logs (handleExplorer2):")
            for log in filtered_logs:
                print(log)
        """

    else:
        print("No valid log entries found.")
