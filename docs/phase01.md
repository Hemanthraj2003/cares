CARES Project: Phase 01 - The Standalone, Self-Aware Node
Objective: The primary goal of this phase is to construct a single, self-contained command-line application that serves as the foundational building block for the entire CARES system. This application will launch a Text-based User Interface (TUI) that provides a real-time visual representation of the local system's resource utilization (CPU and Memory).

Modules Required:

Metrics Reporter Module: A pre-existing module capable of fetching the host system's current CPU and Memory usage percentages.

TUI Library: A suitable library for the chosen programming language (Bubble Tea ).

Architectural Design for Phase 01
The application will operate on two primary concurrent processes:

The Main/UI Thread: This thread is responsible for initializing the TUI, handling user input (though none is required in this phase), and rendering all UI components. It will act as the receiver of data from the metrics thread.

The Background Metrics Thread: This is a separate, non-blocking thread (or goroutine) dedicated to periodically polling the Metrics Reporter Module for system vitals. Its sole responsibility is to fetch data and pass it to the UI thread.

Communication between these threads must be thread-safe, typically achieved using channels or a producer-consumer queue.

Detailed Task Breakdown
Task ID

Task Description

Detailed Implementation Steps & Requirements

P1-T1

Initialize TUI Framework

1. Dependency Integration: Add the chosen TUI library as a dependency to the project. <br> 2. Application Entry Point: Create the main function. This function's primary responsibility is to instantiate the main TUI model/component and start the library's event loop (e.g., program.Run()). <br> 3. Model/State Struct: Define the core struct or class for the TUI application. This will hold the state, such as the current CPU percentage, memory percentage, and a list of log messages.

P1-T2

Design and Build the UI Layout

1. Component Structure: The main view should be composed of three distinct rectangular areas (panes). <br> 2. Title Pane (Top): A static text component at the top of the screen. It must display the text: CARES - Standalone Mode. <br> 3. Vitals Pane (Middle): This pane will contain two progress bar components. Each progress bar must have a text label. The required labels are CPU Usage: and Memory Usage:. The progress bars should be initialized to 0%. <br> 4. Log Pane (Bottom): This pane will be a scrollable text area. It should be initialized with a log message like [INFO] CARES application initialized in standalone mode..

P1-T3

Integrate the Metrics Module

1. Module Import: Import the existing metrics reporter module into the project. <br> 2. Background Process: Create a function that will run as a separate thread/goroutine. This function will contain an infinite loop. <br> 3. Polling Mechanism: Inside the loop, the thread will first call the metrics reporter to get CPU and Memory data. Then, it will sleep for a fixed duration (recommended: 2 seconds). <br> 4. Data Transmission: The data fetched from the metrics module must be sent back to the main UI thread using a thread-safe channel.

P1-T4

Connect Backend Metrics to Frontend TUI

1. Channel Setup: The main application struct will hold the receiving end of the channel. The background metrics thread will be given the sending end. <br> 2. TUI Update Logic: The TUI's main update function (part of the event loop) must handle incoming messages from the metrics channel. <br> 3. State Mutation: When a new metrics message is received, the TUI must update the corresponding values (e.g., cpu_usage, memory_usage) in its state struct. <br> 4. Re-rendering: This state update must trigger a re-render of the UI, which will cause the progress bars to visually reflect the new values. <br> 5. Logging: Upon receiving a metrics update, a new message (e.g., [INFO] Metrics received: CPU at X%, Memory at Y%) should be appended to the list of logs, causing the log pane to update.

Outcome & Verification Criteria
To confirm the successful completion of this phase, the following conditions must be met:

Successful Compilation: The project must compile into a single executable binary without errors.

Application Execution: Running the binary from a terminal (./cares) must immediately launch the full-screen TUI without requiring any command-line arguments.

Visual Correctness:

The TUI must display the three distinct panes (Title, Vitals, Logs) as specified.

The title must be correct.

The progress bars for CPU and Memory must be visible and correctly labeled.

Dynamic Updates:

The progress bars must show updated values approximately every 2 seconds.

The values shown must plausibly reflect the system's actual resource usage.

The log panel must show a new entry corresponding to each metrics update, causing the content to scroll or grow.

bubble tea is instaled and ready
