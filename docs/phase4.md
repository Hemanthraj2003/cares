CARES Project: Phase 04 - Remote Execution and Live TUI Logging
Objective: The goal of this phase is to implement the complete, end-to-end function execution workflow. This involves creating the "brain" of the Orchestrator—the scheduler—which selects the optimal Worker node for a given task. A user will be able to invoke a registered function via a REST API call, and the system will automatically handle the scheduling, execution, and result retrieval, all while providing rich, real-time feedback in the TUI log panels.

Prerequisites:

A completed Phase 03 application: An Orchestrator that can accept function registrations via its TUI and a Worker that can connect to it.

A pre-existing Docker executor module capable of running a container from a given image name and capturing its output.

Architectural Design for Phase 04
This phase introduces a new critical component and connects several existing ones:

Scheduler Component (Orchestrator): A new, stateless logic component will be created within the Orchestrator. Its sole responsibility is to analyze the current state of the NodeRegistry and apply a cost model to select the best candidate for execution.

REST API invoke Endpoint (Orchestrator): The previously defined POST /invoke/{function_name} endpoint will be made fully functional. It will act as the trigger for the entire execution workflow, coordinating between the FunctionRegistry, the Scheduler, and the gRPC client.

Execution RPC (Worker): A new gRPC method will be added to the ClusterService. The Worker will implement the handler for this method, which will directly call the Docker executor module.

TUI Logging: Both the Orchestrator and Worker TUIs will be enhanced to provide a detailed, step-by-step narrative of the scheduling and execution process as it happens.

Detailed Task Breakdown
Task ID

Task Description

Detailed Implementation Steps & Requirements

P4-T1

Implement the Execution RPC

1. Extend .proto Contract: Add a new RPC to the ClusterService in your cluster.proto file: rpc ExecuteFunction(FunctionRequest) returns (FunctionResult);. <br>     a. FunctionRequest should contain the docker_image (string). <br>     b. FunctionResult should contain the output (string) and a boolean success flag. <br> 2. Regenerate gRPC Code: Run the protoc compiler again to generate the updated client and server stubs. <br> 3. Implement Worker Handler: In the Worker's ClusterService implementation, create the handler for ExecuteFunction. <br>     a. This handler will receive the FunctionRequest. <br>     b. It must log a message to its TUI: [INFO] Received execution request for image '...'. <br>     c. It will then call your existing Docker executor module, passing the docker_image. <br>     d. It must wait for the execution to complete, capture the container's standard output, and return it in the FunctionResult.

P4-T2

Implement the Scheduler Core Logic

1. Create Scheduler Component: In the Orchestrator's codebase, create a new file/module for the Scheduler. <br> 2. Implement SelectNodeForExecution Method: This function will take the NodeRegistry as input. <br>     a. It will first filter the registry to get a list of all nodes with a status of Active. <br>     b. It will then iterate through this active list and calculate a "cost score" for each node. Initial Cost Model: score = (cpu_usage _ 0.5) + (memory_usage _ 0.5). <br>     c. The function will return the Node object that has the lowest calculated score. <br>     d. It must handle the case where no active nodes are available, returning an error.

P4-T3

Implement the Invocation API Endpoint

1. Activate Endpoint: In the Orchestrator's HTTP server, implement the handler for the POST /invoke/{function_name} route. <br> 2. Workflow Implementation: The handler will perform the following sequence: <br>     a. Lookup: Read the function_name from the URL path. Access the FunctionRegistry to get the corresponding Docker image. If not found, return a 404 Not Found error. <br>     b. Schedule: Call the Scheduler.SelectNodeForExecution() method to get the optimal worker node. If it returns an error (e.g., no nodes available), return a 503 Service Unavailable error. <br>     c. Dispatch: Use the Orchestrator's gRPC client to call the ExecuteFunction RPC on the selected worker's address. <br>     d. Return: Wait for the FunctionResult from the worker and return its output to the original HTTP caller with a 200 OK status.

P4-T4

Provide Live Feedback to TUIs

1. Orchestrator TUI Logging: Instrument the invoke endpoint handler with detailed logging that is sent to the Orchestrator's TUI log pane. The logs should narrate the process: <br>     a. [INFO] Received invocation for 'my-func'. <br>     b. [INFO] Selecting optimal node from N active candidates... <br>     c. [INFO] Node 'xyz-123' selected with lowest cost. <br>     d. [INFO] Dispatching job to 'xyz-123'... <br>     e. [INFO] Job complete. Received result from worker. <br> 2. Worker TUI Logging: The Worker's ExecuteFunction handler should also log its progress: <br>     a. [INFO] Starting container for image 'hello-world'... <br>     b. [INFO] Container finished. Sending result back to Orchestrator.

Outcome & Verification Criteria
To confirm the successful completion of this phase, the following conditions must be met:

End-to-End Execution: With the cluster running and a function registered, you can use a command-line tool like curl to make a POST request to the invoke endpoint (e.g., curl -X POST http://localhost:8080/invoke/my-test-func).

Correct Output: The curl command receives the correct standard output from the executed Docker container (e.g., the "Hello from Docker!" message).

Verifiable Load Balancing: If you run two workers and manually stress the CPU of one (e.g., with a while true; do true; done loop), subsequent function invocations should be visibly scheduled on the non-stressed worker, as evidenced by the TUI logs.

Real-Time TUI Narrative: The entire process can be observed in real-time across the TUIs. The Orchestrator log shows the scheduling decision, and moments later, the chosen Worker log shows the execution events.
