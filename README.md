# DiceDB CLI

A command-line interface for DiceDB.

## Features

- Interactive command prompt
- Command auto-completion
- Support for all standard DiceDB commands

## Installation

### Prerequisites

- Go 1.23 or higher
- DiceDB server running locally or accessible network connection

### Installing from source

```bash
# Clone the repository
git clone https://github.com/DiceDB/dicedb-cli
cd dicedb-cli

# Build the CLI
go build -o dicedb-cli

# Optional: Install globally
sudo mv dicedb-cli /usr/local/bin/
```

## Usage

### Starting the CLI

```bash
./dicedb-cli
```

By default, the CLI connects to `localhost:7379`. You'll see a prompt indicating successful connection:

```
Connected to DiceDB. Type 'exit' or press Ctrl+D to exit.
dicedb>
```

### Basic Commands

Here are some common commands you can use:

```
dicedb> SET mykey "Hello World"
OK
dicedb> GET mykey
"Hello World"
dicedb> DEL mykey
(integer) 1
```

### Authentication

If your DiceDB server requires authentication:

```
dicedb> AUTH your_password
OK
```

### Command Auto-completion

Press TAB to see available commands or complete partial commands:

```
dicedb> S[TAB]
SADD     SET      SMEMBERS SUBSCRIBE
```

## Supported Commands

The CLI supports all standard DiceDB commands, including:

- **Key-Value Operations**: GET, SET, DEL, INCR, DECR, EXISTS
- **Lists**: LPUSH, RPUSH, LPOP, RPOP
- **Sets**: SADD, SREM, SMEMBERS
- **Hashes**: HSET, HGET, HDEL
- **Server Management**: AUTH, PING, INFO
- **Key Space**: KEYS, EXPIRE, TTL

## Exiting the CLI

You can exit the CLI in two ways:
- Type `exit` and press Enter
- Press Ctrl+D

## Error Handling

The CLI provides clear error messages for common scenarios:

```
dicedb> GET
Error: wrong number of arguments for 'get' command
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.