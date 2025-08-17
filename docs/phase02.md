CARES Project: Phase 02 - Cluster Formation and TUI Visualization
Objective: The goal of this phase is to evolve the standalone application from Phase 01 into a networked system capable of forming a multi-node cluster. This involves introducing two distinct operational modes: "Orchestrator" and "Worker". The Orchestrator's TUI will be significantly enhanced to act as a central dashboard, displaying a real-time list of all connected Worker nodes and their respective resource metrics.

Prerequisites:

A completed Phase 01 application: A runnable binary that displays a TUI with live local CPU and Memory usage.

A gRPC library and Protobuf compiler for your chosen language.

Architectural Design for Phase 02
The single application will now have a branching logic at startup to determine its role.

Orchestrator Mode: The application will act as a gRPC server. It will listen for incoming connections from Workers. It will maintain a new, central data structure—the NodeRegistry—to keep track of every worker's state. Its TUI will be rendered based on the data in this registry.

Worker Mode: The application will act as a gRPC client. After starting, it will connect to the Orchestrator and continuously stream its local metrics (using the logic from Phase 01) to the server. Its TUI will continue to show its own local stats, just as it did in Phase 01.

Detailed Task Breakdown
Task ID

Task Description

Detailed Implementation Steps & Requirements

P2-T1

Implement Startup Mode Selection

1. Modify TUI Model: Add a new state to your main TUI model to track the current application view (e.g., view_mode: 'selection'). <br> 2. Create Selection View: When the application starts, render a new initial view that presents two selectable options: [1] Create a new cluster (Orchestrator) and [2] Join an existing cluster (Worker). <br> 3. Handle User Input: Implement the logic to handle keyboard input (arrow keys and Enter) to select a mode. <br> 4. Input Screen for Worker: If "Join" is selected, transition the TUI to a new view with a text input field, prompting the user to enter the Orchestrator's address (e.g., 127.0.0.1:50051).

P2-T2

Define and Implement gRPC Communication

1. Create .proto File: Create a file named cluster.proto. Define a ClusterService with two RPCs: <br>     a. rpc JoinCluster(NodeInfo) returns (Acknowledgement); <br>     b. rpc Heartbeat(stream NodeMetrics) returns (stream OrchestratorCommand); <br> 2. Generate Code: Use the Protobuf compiler (protoc) to generate the server and client stub code in your language. <br> 3. Integrate gRPC Server (Orchestrator): If Orchestrator mode is selected, start the gRPC server in a background thread/goroutine so it doesn't block the TUI. Implement placeholder handlers for the RPCs that log messages to the TUI. <br> 4. Integrate gRPC Client (Worker): If Worker mode is selected, use the address from the input screen to connect to the Orchestrator. Call JoinCluster and then begin the Heartbeat stream, sending the metrics gathered by your existing module.

P2-T3

Implement Orchestrator State Management

1. Define Node Struct: Create a struct/class to represent a Worker node. It should contain fields for NodeID, Address, Status (e.g., "Active", "Disconnected"), LastSeen (timestamp), CPUUsage, and MemoryUsage. <br> 2. Create NodeRegistry: This will be a thread-safe (mutex-protected) map or dictionary that stores Node structs, keyed by NodeID. This registry must be part of the Orchestrator's main state. <br> 3. Update gRPC Handlers: <br>     a. The JoinCluster handler must add the new node to the NodeRegistry. <br>     b. The Heartbeat handler must find the corresponding node in the registry and update its CPUUsage, MemoryUsage, and LastSeen timestamp with every message it receives.

P2-T4

Enhance the Orchestrator TUI

1. Conditional Rendering: The TUI's main View() function must now render different layouts based on the application's mode. <br> 2. New Orchestrator Layout: The central pane of the Orchestrator TUI should be replaced with a dynamic list or table. <br> 3. Render from State: The TUI's rendering loop will iterate over the NodeRegistry. For each Node in the registry, it will render a row in the table containing: <br>     a. The Node ID. <br>    ˆb. The Node's Status. <br>     c. A progress bar for CPU usage, driven by the node.CPUUsage value. <br>     d. A progress bar for Memory usage, driven by the node.MemoryUsage value. <br> 4. Update Log Pane: The Orchestrator's log pane should now display cluster-level events like [INFO] Node 'xyz-123' has joined the cluster.

Outcome & Verification Criteria
To confirm the successful completion of this phase, the following conditions must be met:

Dual Mode Functionality: The same ./cares binary can be launched in two different modes based on user selection in the TUI.

Orchestrator Launch: Starting in "Orchestrator" mode displays a TUI with an empty list of nodes, ready to accept connections.

Worker Connection: Starting a second instance in "Worker" mode allows you to input the Orchestrator's address. Upon connecting:

The Worker's TUI continues to function as it did in Phase 01, showing its local stats.

A new entry for the Worker appears in the Orchestrator's TUI node list.

Live Dashboard: The CPU and Memory progress bars for the Worker inside the Orchestrator's TUI update in real-time, accurately reflecting the metrics being sent from the Worker.

Multi-Node Display: Connecting a third and fourth instance as Workers results in them also appearing in the Orchestrator's list, with all entries updating independently.
