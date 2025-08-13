# CARES: Phase 1 Development Briefing

## Objective
To build and test the two core, local components of a CARES node: the Function Runner and the Resource Reporter. This phase involves no networking and focuses on creating stable, independent building blocks.

---

## Task 1: The Function Runner

### What is this?
The primary operational component of the system is a Go module, the designated function of which is to facilitate the execution of a specified Docker image within a containerized environment. The module's sole responsibility encompasses the invocation of the function contained within the image and the subsequent capture of any output generated therefrom.

### Why are we building it first?
We need to ensure we can reliably interact with the Docker daemon from our Go code. By building this component in isolation, we can perfect the logic of starting a container and capturing its output without worrying about network requests or other complexities. This component directly fulfills requirement **F-304**.

### Implementation Plan:
1. **Create a Package**: Create a new Go package, for example, `internal/runner`.
2. **Define the Function**: Inside this package, create a function with the signature:
   ```go
   func RunContainer(imageName string) (output string, err error)
   ```
3. **Execute the Command**: Use Go's standard `os/exec` package.
4. **Construct the command**:  
   ```go
   cmd := exec.Command("docker", "run", "--rm", imageName)
   ```  
   The `--rm` flag is crucial as it automatically cleans up the container after it exits.
5. **Capture the output**: Use `cmd.CombinedOutput()`. This is the simplest way to get both the standard output and standard error in one byte slice, which is perfect for our needs.
6. **Return the Result**: Convert the byte slice output to a string and return it along with any error that occurred.

### How to Test:
- Create a temporary `main.go` file that imports the `runner` package.
- Call `runner.RunContainer("hello-world")` and print the output and error to the console.
- If you see the "Hello from Docker!" message, this task is complete.

---

## Task 2: The Resource Reporter

### What is this?
This is the sensory organ of our node. It's a Go module that inspects the host system and reports on its current CPU and Memory load.

### Why are we building it now?
This component provides the raw data for the "Cost-Aware" part of the scheduler (**F-302**, **F-401**). The Orchestrator will eventually need this data from every node to make intelligent decisions. We build it now to confirm we can gather these metrics accurately.

### Implementation Plan:
1. **Add Dependency**: Use the industry-standard `gopsutil` library.  
   Run:
   ```bash
   go get github.com/shirou/gopsutil/v3/...
   ```
2. **Create a Package**: Create a new Go package, for example, `internal/metrics`.
3. **Implement Metric Functions**:
   - **CPU Usage**:  
     ```go
     func GetCPUUsage() (float64, error) {
         percent, err := cpu.Percent(time.Second, false)
         if err != nil {
             return 0, err
         }
         return percent[0], nil
     }
     ```
   - **Memory Usage**:  
     ```go
     func GetMemoryUsage() (float64, error) {
         vmStat, err := mem.VirtualMemory()
         if err != nil {
             return 0, err
         }
         return vmStat.UsedPercent, nil
     }
     ```

### How to Test:
- Create another temporary `main.go` file that imports the `metrics` package.
- Call both `GetCPUUsage` and `GetMemoryUsage` in a loop every few seconds and print the results.
- You should see the percentages change as you use your computer.

---

## Phase 1 Outcome
At the end of this phase, you will have zero networked code. Instead, you will have two highly reliable, independently tested Go packages that form the foundation of the Node's capabilities. This allows us to move into Phase 2 with confidence that the core logic is solid.
