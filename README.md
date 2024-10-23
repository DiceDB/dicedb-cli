# DiceDB CLI Client

A command-line interface client for [DiceDB](https://dicedb.io).
## Table of Contents

- [Features](#features)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Usage](#usage)
    - [Basic Commands](#basic-commands)
    - [Subscription Commands](#subscription-commands)
    - [Custom Watch Commands](#custom-watch-commands)
- [Command Autocompletion](#command-autocompletion)
- [For Developers](#for-developers)
    - [Code Structure](#code-structure)
    - [Contributing](#contributing)
- [License](#license)

## Features

- **Interactive Prompt**: Provides an interactive shell to execute DiceDB commands.
- **Command Autocompletion**: Offers autocompletion for DiceDB commands when typing.
- **Watch Commands**: Supports custom `.WATCH` and `.UNWATCH` commands to reactively receive updated results of specific commands.
- **Authentication**: Supports `AUTH` command to authenticate with the DiceDB server.

## Prerequisites

- **Go version 1.23 or later recommended
- **DiceDB Server**: A running instance of DiceDB. You can find setup instructions on the [DiceDB GitHub page](https://github.com/dicedb/dice).

## Installation

1. **Clone the Repository**:

   ```bash
   git clone https://github.com/DiceDB/dicedb-cli.git
   cd dicedb-cli
   ```

2. **Download Dependencies**:

   ```bash
   go mod download
   ```

3. **Build the Application**:

   ```bash
   go build -o dicedb-cli
   ```

   This will generate an executable named `dicedb-cli` in the current directory.

## Usage

### Starting the CLI

Run the executable to start the interactive prompt:

```bash
./dicedb-cli
```

You should see:

```
Connected to DiceDB. Type 'exit' or press Ctrl+D to exit.
dicedb>
```

### Basic Commands

You can execute any DiceDB command directly:

```bash
dicedb> SET mykey "Hello, World!"
OK
dicedb> GET mykey
"Hello, World!"
dicedb> DEL mykey
(integer) 1
```

### Watch Commands

Receive updated results of supported commands using their `.WATCH` variants. These commands keep the prompt in a persistent state, displaying updates when the monitored data changes.

**Start Watching a Key**:

```bash
dicedb> GET.WATCH mykey
```

The prompt changes to indicate watch mode:

```
dicedb(get.watch)>
```

**Receive Updates**:

When the value of `mykey` changes, the new value is displayed:

```
Command: GET
Fingerprint: 1234567890
Data: "New Value"
```

**Stop Watching a Key**:

To exit the watch mode, use the corresponding `.UNWATCH` command: 
```bash
dicedb(get.watch)> GET.UNWATCH 1234567890 # Use the fingerprint from the watch output
```
Output
```
OK
dicedb>
```

### Exiting the CLI

Type `exit` or press `Ctrl+D` to exit the CLI:

```bash
dicedb> exit
```

## Command Autocompletion

The CLI provides autocompletion for DiceDB commands when typing the first word. Start typing a command and press `Tab` to see suggestions:

```bash
dicedb> S    # Press Tab
```

Suggestions:

- `SET`
- `SUBSCRIBE`
- `SADD`
- `SCAN`
- ...

### Contributing

Contributions are welcome! Here's how you can help:

1. **Fork the Repository**:

   Click on the `Fork` button at the top right of the repository page.

2. **Clone Your Fork**:

   ```bash
   git clone https://github.com/yourusername/dicedb-cli-client.git
   cd dicedb-cli-client
   ```

3. **Create a New Branch**:

   ```bash
   git checkout -b feature/my-new-feature
   ```

4. **Make Your Changes**:

    - Add new features or fix bugs.
    - Ensure your code follows the project's coding standards.

5. **Commit and Push**:

   ```bash
   git commit -m 'Add new feature'
   git push origin feature/my-new-feature
   ```

6. **Submit a Pull Request**:

   Go to the repository on GitHub and open a pull request.

### Development Setup

Ensure you have Go installed and set up on your machine.

**Download Dependencies**:

```bash
go mod download
```

**Run the Application**:

```bash
go run .
```

**Testing**:

Add tests as needed to ensure code reliability.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---
