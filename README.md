# SyncEra Indexer

SyncEra Indexer is a Web3 indexer for zkSync Era, written in Go and aimed at being easy to get started with and a good learning project.

It connects to zkSync Era RPC nodes to sync block, transaction, and log data in real time, parses common DeFi and DEX contract events (for example, SyncSwap), and writes the normalized results into MySQL and Redis. The codebase is kept clear and modular so you can read, debug, and extend it while understanding how a blockchain indexer is structured and implemented.

If you want to:
- Get started with Web3 backend or blockchain data indexing;
- Learn how to use Go to connect to zkSync Era and consume blocks and logs;
- Build a simple, extensible on-chain data service for yourself or your team,

this repository can serve as a hands-on scaffold to extend with more protocols, more metrics, and external APIs.


## Quick Start

### 1. Start the databases

**Mac / Linux:**
```bash
make up
```

**Windows:**
```batch
start-windows.bat
```

### 2. Connect to the databases

Use your desktop tools to connect:

**MySQL (Navicat):**
- Host: `localhost`
- Port: `3307`
- User: `scanner`
- Password: `scannerpass`
- Database: `syncswap`

**Redis (Another Redis Desktop Manager):**
- Host: `localhost`
- Port: `6380`

### 3. Configure the project

```bash
cp config/config.yaml.example config/config.yaml
# Edit config.yaml and fill in your RPC endpoint
```

### 4. Run the program

```bash
make deps  # Install dependencies
make run   # Start running
```

## Common Commands

```bash
make up      # Start services
make down    # Stop services
make logs    # View logs
make db      # Enter MySQL
make redis   # Enter Redis
make clean   # Remove all data (dangerous)
```

## Ports

> Non-default ports are used to avoid conflicts with other projects:

- MySQL: `3307` (default 3306)
- Redis: `6380` (default 6379)

## Project Structure

```
scan-chain/
├── config/              # Configuration files
├── docker/              # Docker configuration
├── main.go              # Main program
├── Makefile             # Mac/Linux commands
└── start-windows.bat    # Windows startup script
```

## Tech Stack

- Go 1.24+
- MySQL 8.0
- Redis 7
- Docker

## License

MIT

## Contact

For work or project inquiries, reach out via email: `austin.rate@foxmail.com`
