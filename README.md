# PortPilot Port Forwarding Tool

<img src="PortPilot.png" alt="drawing" width="200"/>
This is a command-line tool written in Go that enables port forwarding to multiple Kubernetes services. It reads service configuration from a YAML file and establishes port forwarding connections to the specified services.

## Prerequisites

Before contributing to this tool, ensure that you have the following prerequisites installed:

- Go programming language (https://golang.org/dl/)
- Kubernetes client Go library (`go get k8s.io/client-go@v0.24.0`)

## Usage

1. Clone the repository and install the tool using `go install`.
2. Create a YAML file named `.portpilot.yaml`. The YAML file should contain the configuration for the services you want to port forward. Below is an example of the YAML file structure:

   ```yaml
   services:
   - name: example-service1
     remotePort: 8080
     localPort: 8081
   - name: example-service2
     remotePort: 8888
     localPort: 8889
   ```

   Customize the `name`, `remotePort`, and `localPort` fields for each service as needed.

3. Open a terminal and navigate to the directory containing the `services.yaml` file.
4. Build and run the tool using the following command:

   ```shell
   portpilot
   ```

   The tool will read the `services.yaml` file, establish port forwarding connections to the specified services, and print the local URLs for accessing each service.

5. To stop the port forwarding, press `Ctrl+C` in the terminal.

## Configuration

- The Kubernetes configuration is automatically read from the default `kubeconfig` file location (`$HOME/.kube/config`). Make sure you have a valid `kubeconfig` file configured for your cluster.
- The tool uses the Kubernetes client Go library to interact with the cluster and perform the port forwarding.

## Limitations

- The tool assumes the Kubernetes cluster is reachable using the default `kubeconfig` configuration.
- Only the first matching pod is used for each service. If multiple pods match the service selector, only the first pod will be used for port forwarding.

## License

This tool is licensed under the [MIT License](LICENSE).
