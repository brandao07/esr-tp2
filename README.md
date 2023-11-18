# Over-The-Top (OTT) Streaming Application

This Go program enables Over-The-Top (OTT) streaming using UDP.
It allows for various modes such as Bootstrap, Node, Server, and Client to facilitate different functionalities
within the streaming network.

## Instructions

### Prerequisites

- Go installed on your system.
- Access to the terminal or command prompt.

### Build and Execution

This program provides a Makefile to simplify the build and execution process.

- **Build the Executable:**

  ```bash
  make build
  ```

  This command compiles the program and generates the executable file named `main` in the `./cmd` directory.

- **Run the Application:**

  - Execute the program with different modes and arguments:

    ```bash
    make run ARGS="[mode] [additional_arguments]"
    ```

    Replace `[mode]` with one of the following:

    - `Bootstrap`: Initialize the OTT streaming service.
    - `Node`: Participate in the streaming network.
    - `Server`: Stream content.
    - `Client`: Receive and display streamed content.

    Provide additional arguments based on the selected mode. Examples:

    - For `Node` mode: `make run ARGS="Node bootstrap_address myNodeId"`
    - For `Server` mode: `make run ARGS="Server bootstrap_address myNodeId video_to_stream"`
    - For `Client` mode: `make run ARGS="Client node_address video_to_stream"`

- **Clean Build Artifacts:**

  ```bash
  make clean
  ```

  This command removes the compiled program file and any build artifacts.
