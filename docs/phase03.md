CARES Project: Phase 03 - Function Registration and TUI Display
Objective: The goal of this phase is to introduce the concept of "functions" into the CARES ecosystem. This will be achieved by creating a user-facing REST API on the Orchestrator and building an interactive form within the TUI, allowing users to register a function (defined as a Docker image). The Orchestrator's TUI will be updated to display a live list of all registered functions.

Prerequisites:

A completed Phase 02 application: A binary that can run as an Orchestrator or Worker, with a fully functional cluster dashboard on the Orchestrator TUI.

An HTTP server library for your chosen language.

Architectural Design for Phase 03
The Orchestrator's architecture will be expanded to run two concurrent network servers:

gRPC Server (Internal): The existing server from Phase 02, which continues to handle all communication with Worker nodes (heartbeats, etc.).

REST API Server (External): A new HTTP server that will listen on a different port (e.g., 8080). This server is designed for user-facing interactions, including those initiated from the TUI itself.

A new stateful component will be added to the Orchestrator:

FunctionRegistry: A thread-safe (mutex-protected) map or dictionary. It will store the mapping between a user-defined function name (string) and its corresponding Docker image URI (string).

Detailed Task Breakdown
Task ID

Task Description

Detailed Implementation Steps & Requirements

P3-T1

Implement Orchestrator REST API Server

1. Dependency Integration: Add a lightweight HTTP/REST framework library to your project's dependencies. <br> 2. Concurrent Server Launch: In the Orchestrator's startup logic, create and run the new HTTP server in its own background thread/goroutine. This ensures it runs alongside the TUI and the gRPC server without blocking. <br> 3. Configuration: The listening port for the REST API (e.g., 8080) should be configurable, perhaps via a command-line flag.

P3-T2

Implement the Function Registry

1. Define Function Struct: Create a simple struct/class to represent a function. It should contain fields for Name (string) and Image (string). <br> 2. Create FunctionRegistry: Instantiate a thread-safe map (e.g., map[string]Function protected by a sync.Mutex in Go). This registry will be a part of the Orchestrator's main state, alongside the NodeRegistry.

P3-T3

Create the Function Upload Endpoint

1. Define Endpoint: On the HTTP server, register a new route handler for POST /functions. <br> 2. Request Handling: The handler must be able to parse an incoming JSON request body. It should expect a JSON object with two fields: name and image. <br> 3. State Mutation: Upon successfully parsing the request, the handler must acquire a lock on the FunctionRegistry and add the new Function data to it. <br> 4. Response: The handler should return a success response (e.g., HTTP status 201 Created) that includes the path the user will use to invoke the function later. For example: {"status": "success", "invoke_path": "/invoke/my-test-func"}. This makes the next step clear.

P3-T4

Enhance Orchestrator TUI Layout

1. Add New Pane: Modify the Orchestrator's TUI layout to include a new pane, typically next to or below the "Nodes" list. This pane should be clearly labeled "Registered Functions". <br> 2. Render from State: The TUI's main rendering loop must now also read from the FunctionRegistry. It will iterate over the functions in the registry and display them in the new pane, showing at least the function name and its associated Docker image. This list must update automatically as new functions are added.

P3-T5

Implement TUI-based Function Registration

1. Add Keybinding: Implement a keybinding in the Orchestrator TUI (e.g., pressing 'a' for "add"). <br> 2. Create Registration Form: When the key is pressed, the TUI should display a new view or a modal form on top of the existing layout. This form must contain two text input fields: "Function Name" and "Docker Image URL". <br> 3. HTTP Client Logic: Upon form submission (e.g., pressing Enter), the TUI will act as an HTTP client. It will construct a JSON payload from the form data and send a POST request to its own backend API endpoint (http://localhost:8080/functions). <br> 4. Handle Response: The TUI should handle the response from the API. On success, it should close the form and log a confirmation message (e.g., "Successfully registered 'my-test-func'"). On failure, it should display an error message.

Outcome & Verification Criteria
To confirm the successful completion of this phase, the following conditions must be met:

Concurrent Server Operation: The Orchestrator application successfully runs the gRPC server and the REST API server simultaneously without issues.

TUI Update: The Orchestrator's TUI now displays the new "Registered Functions" pane, which is initially empty.

Interactive Function Registration: While the Orchestrator TUI is running, a user can press a specific key (e.g., 'a') to bring up a form for registering a new function.

Form Submission and API Call: After filling out the function name and Docker image in the TUI form and submitting it, the TUI successfully makes a POST request to its own backend /functions endpoint.

Real-Time Visual Confirmation: Immediately after the form is submitted successfully, the new function (e.g., "my-test-func") appears in the "Registered Functions" list within the live Orchestrator TUI, and a success message appears in the log.

Note: The /invoke/my-test-func endpoint will be made functional in Phase 04. This phase only deals with registering the function and confirming its API path. The load balancing and execution logic are the focus of the next phase.
