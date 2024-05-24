# cfx-evm-sync-sdk

This tool retrieves block data from multiple RPC nodes concurrently and saves it as JSON files.

## Features

- Concurrently retrieves block data from multiple RPC nodes.
- Converts block data to JSON format and saves it as individual files.
- Supports configuration of RPC nodes using Viper.

## Dependencies

- Golang
- Web3go (https://github.com/openweb3/web3go)
- Viper

## Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/your-username/your-repository.git
   ```
   
2. Navigate to the project directory:
   
  ```bash
  cd your-repository
  ```
   
3. Build the project:

   ```bash
   go build
   ```

## Usage

### Config

Edit the `config.yaml` file in the `config` directory to specify the RPC nodes you want to connect to:

```yaml
rpc_nodes:
  - "https://node1.example.com"
  - "https://node2.example.com"
```

Edit the `config.yaml` file in the `config` directory to specify the starting and ending nodes to: 

```yaml
block:
  start: x
  end: y
```

### Running the Tool

Run the executable file with the following command:

```bash
./cfx-evm-sync-sdk
```

The tool will concurrently retrieve block data from the specified RPC nodes and save it as JSON files.
