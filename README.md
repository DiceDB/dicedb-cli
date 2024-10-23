# DiceDB CLI

[DiceDB](https://dicedb.io) is a redis-compliant, in-memory, real-time, and reactive database optimized for modern hardware and for building and scaling truly real-time applications. This is a command line interface for it.

## Get Started

```sh
sudo su
curl -sL https://raw.githubusercontent.com/DiceDB/dicedb-cli/refs/heads/master/install.sh | sh
```

### Setting up DiceDB from source for development and contributions

To run DiceDB CLI for local development or running from source, you will need

1. [Golang](https://go.dev/)
2. Any of the below supported platform environments:
    1. [Linux based environment](https://en.wikipedia.org/wiki/Comparison_of_Linux_distributions)
    2. [OSX (Darwin) based environment](https://en.wikipedia.org/wiki/MacOS)
    3. WSL under Windows

```bash
git clone https://github.com/dicedb/dicedb-cli
cd dicedb-cli
go run main.go
```

## Usage

Run the executable to start the interactive prompt

```bash
$ dicedb-cli
```

You should see

```sh
Connected to DiceDB. Type 'exit' or press Ctrl+D to exit.
dicedb>
```

### Basic Commands

You can execute any DiceDB command directly:

```bash
dicedb> SET k1 v1
OK
dicedb> GET k1
v1
dicedb> DEL k1
1
```

### Watch Commands

> To use `.WATCH` commands, make sure your DiceDB server is running with the flag `--enable-watch --enable-multithreading`.

Receive updated results of supported commands using their `.WATCH` variants. These commands keep the prompt in a persistent state, displaying updates when the monitored data changes. Start watching a key:

```bash
dicedb> GET.WATCH k1
```

The prompt changes to indicate watch mode:

```
dicedb(get.watch)>
```

In other terminal, connect the CLI to the same database and fire

```
dicedb> SET k1 v2
```

As the value of key `k1` changes, the new value is emitted to the client connected on the first terminal in real-time.

```
Command: GET
Fingerprint: 2402418009
Data: "v2"
```

To exit the watch mode, use the corresponding `.UNWATCH` command: 

```bash
dicedb(get.watch)> GET.UNWATCH 2402418009
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

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
