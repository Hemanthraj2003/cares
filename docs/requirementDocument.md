# CARES: Cost-Aware Resource Allocation and Execution Scheduler

## 1. Introduction

This document outlines the requirements and design for the **Cost-Aware Resource Allocation and Execution Scheduler (CARES)**, a command-line interface (CLI) application. The primary goal of CARES is to provide a decentralized and cost-aware platform for executing user-defined functions. The system operates as a cluster of interconnected nodes, intelligently distributing incoming function requests based on a defined cost model and real-time resource availability.

## 2. Project Goals

The overarching goal of the CARES project is to demonstrate a functional, distributed system that can:

- Create and manage a dynamic cluster of computing nodes.

- Receive and orchestrate the execution of user-submitted functions (packaged as Docker images).

- Implement a load-balancing mechanism that considers resource availability and a cost-based metric.

- Provide a simple CLI for user interaction, node management, and cluster status monitoring.

- Be easily deployable and testable using Docker containers to simulate a multi-node environment.

## 3. Functional Requirements

### 3.1 Cluster Management

- **F-101:** The application **must** support a "cluster creation" mode, which initializes an orchestrator node and generates a unique cluster identifier (e.g., an IP address and port or a unique key).

- **F-102:** The application **must** support a "join cluster" mode, allowing a new node to discover and connect to the orchestrator using the cluster identifier.

- **F-103:** The orchestrator **must** maintain a registry of all active nodes within the cluster.

- **F-104:** The orchestrator **must** detect and handle node disconnections gracefully (e.g., by marking them as inactive).

### 3.2 Function Management

- **F-201:** The CLI **must** provide a command for users to upload a function, specified as a Docker image reference.

- **F-202:** The orchestrator **must** store a registry of all uploaded functions and their associated Docker images.

- **F-203:** Upon successful upload, the orchestrator **must** return a unique API endpoint or handler URL that users can use to trigger the function.

### 3.3 Request Execution and Load Balancing

- **F-301:** The orchestrator **must** expose an API endpoint for users to invoke a registered function.

- **F-302:** The orchestrator **must** implement a **cost-aware scheduling algorithm** to select the optimal node for each incoming function request.

  - **Initial Cost Model:** The algorithm will prioritize nodes with the lowest current load and highest available resources (CPU, RAM).

- **F-303:** The orchestrator **must** forward the function request to the selected node.

- **F-304:** The node **must** be able to receive a function request and execute the corresponding Docker image.

- **F-305:** The node **must** return the result of the function execution back to the orchestrator.

- **F-306:** The orchestrator **must** return the final result to the original user.

### 3.4 Monitoring and Statistics

- **F-401:** The orchestrator **must** periodically collect real-time statistics from each node, including:

  - CPU and Memory usage.

  - Current number of executing functions.

  - Success/failure rate of function executions.

- **F-402:** The CLI **must** provide a command for users to view the collected statistics for all nodes in the cluster.

## 4. Non-Functional Requirements

- **NF-101 (Performance):** Function requests should be executed with minimal latency.

- **NF-102 (Scalability):** The system should be able to support a growing number of nodes and function requests without significant degradation in performance.

- **NF-103 (Reliability):** The system should be resilient to node failures and continue operating.

- **NF-104 (Usability):** The CLI should be simple and intuitive for a user to understand and operate.

## 5. System Architecture and Design

### 5.1 High-Level Architecture

The CARES system follows a decentralized master-worker architecture with a single **Orchestrator** (the master) and multiple **Nodes** (the workers).

### 5.2 Component Breakdown

#### **Orchestrator Component**

- **Cluster Discovery Service:** Responsible for listening for new nodes joining the cluster and maintaining the list of active nodes.

- **Function Registry:** A data structure that maps function identifiers (like API endpoints) to Docker image details.

- **Scheduler/Load Balancer:** The core logic that implements the cost-aware algorithm to select the best node for a given request. It uses the Node Monitor's data to make decisions.

- **API Gateway:** An HTTP server that handles incoming function upload and execution requests from users.

- **Node Monitor:** A service that periodically pings nodes and collects resource and performance metrics.

#### **Node Component**

- **Agent Service:** A lightweight service running on each node that communicates with the orchestrator.

- **Function Runner:** The component responsible for pulling the specified Docker image and executing the function inside it.

- **Resource Reporter:** Gathers real-time resource usage data (CPU, RAM) and success rates, sending it to the orchestrator.

#### **TUI Component**

- A user-facing, keyboard-navigable Text-based User Interface (TUI) that provides an interactive, real-time view of the application's state. The TUI will dynamically adjust its display based on the node's role (Orchestrator or Worker).

- **Initial Setup:** When the `cares` command is first run, the TUI will present an interactive menu asking the user to choose the node's role: `Create a new cluster` (Orchestrator) or `Join an existing cluster` (Worker). If joining, the TUI will prompt for the cluster ID.

- **For the Orchestrator:**

  - The TUI will display a list of all connected worker nodes.

  - For each node, it will show a real-time visualization of its health, including progress bars for CPU and memory usage.

  - A separate pane will display a consolidated log of all cluster activities and function requests.

  - The TUI will also have interactive elements (e.g., menu items) to trigger commands like `upload-function` and `monitor-status`.

- **For the Worker Node:**

  - The TUI will show a real-time display of its own resource usage (CPU, RAM) using progress bars.

  - A dedicated section will stream logs related to its specific activities, such as function executions and communications with the orchestrator.

### 5.3 Communication Protocol

- **Orchestrator-to-Node & Node-to-Orchestrator:** A robust, bi-directional communication protocol like **gRPC** or **WebSockets** will be used to ensure efficient and low-latency message passing for monitoring and job execution.

- **User-to-Orchestrator:** A standard **RESTful API** will be exposed to allow for simple function upload and execution calls.

## 6. Deployment and Execution Model

### 6.1 Docker Containerization

The entire CARES application will be bundled into a single Docker image. When the image is run, the user will execute the `cares` command within the container's terminal. This will launch a TUI that guides the user through the initial setup, eliminating the need for complex command-line arguments.

### 6.2 Simulation Environment

To demonstrate the functionality, a single host machine will be used to run multiple Docker containers.

- The orchestrator will be started in one container, which will output a cluster identifier after the TUI guides the user through the setup.

- The user will then manually start other containers, each representing a node.

- The user will then run the `cares` command in each of these new containers and, guided by the TUI, manually input the cluster identifier to connect them to the orchestrator.

- The CARES application is designed to be network-agnostic. While Docker's networking is used to facilitate communication in this simulated environment, the core functionality relies on a configurable network address, making it suitable for deployment across diverse network setups, including air-gapped nodes with manual configuration. This setup effectively simulates a distributed environment on a single machine.
