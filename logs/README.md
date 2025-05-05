<p align="center">
  <img src="https://github.com/PanK0/ARGO/blob/main/pictures/ARGO.png?raw=true" alt="ARGO_logo"
    width="20%">
</p>

## ARGO - Logs

ARGO - Adversarial Robust Graph Operator is a software for testing reliable communication techniques in unknown networks in presence of Byzantine faults. 

## LOGS
In `ARGO/logs/` are saved logs created by using `logEvent()` function in `utils.go`. You can basically write whatever you want in the logs. 

By inserting this function in the code, it is possible to create a *NODE_ADDRESS.log* file so that the wanted events are saved in the file.

By calling the dedicated python script `log_parser.py` all the saved logs in the directory are put into an excel file that represents a timeline of the recorded events: **after formatting the file**, with low effort it can be a nice resource to have a clear view of what happens in the whole system, and it is very useful especially when testing the network on a local machine.

This is an example of the generated timeline after a bit of formatting and coloring:

![Timeline](https://github.com/PanK0/ARGO/blob/main/pictures/timeline.jpeg?raw=true)

### Make log_parser.py work

Install the requirements with 

```
> pip3 install -r requirements.txt
```

Then, when the directory will be full of *.log* files, run the script to create the wanted files:

```
> python3 log_parser.py
```

To clear all the generated files, run:

```
> python3 log_parser.py clear
```

### How to place logs in the code

Insert the function `logEvent()`, defined in *ARGO/src/utils.go*, wherever it is needed.

The function has this structure:
`func logEvent(nodeID string, printoption bool, event string)`

Where:
- `nodeID` is the id of the node that is registering the event
- `printoption` is a boolean that is put to *true* if you want to also print the log on the shell, *false* otherwise
- `event` is the event to register


**TO DO** : this mechanism could be refined, but for now it works