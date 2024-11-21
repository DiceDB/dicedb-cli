# DiceDB CLI

[DiceDB](https://dicedb.io) is a redis-compliant, reactive, scalable, highly-available, unified cache optimized for modern hardware. This is a command line interface for it.

## Get Started

```sh
sudo su
curl -sL https://raw.githubusercontent.com/DiceDB/dicedb-cli/refs/heads/master/install.sh | sh
```

## Usage

Run the executable to start the interactive prompt (REPL)

```bash
$ dicedb-cli
```

You should see

```sh
dicedb (localhost:7379)>
```

To connect to some other host or port, you can pass the flags `--host` and `--port` with apt parameters.
You can also get all available parameters by firing

```sh
$ dicedb-cli --help
```

## Firing commands

You can execute any DiceDB command directly:

```bash
dicedb (localhost:7379)> SET k1 v1
OK
dicedb (localhost:7379)> GET k1
"v1"
dicedb (localhost:7379)> DEL k1
1
```

You can find all available commands at [dicedb.io/docs](https://dicedb.io/docs).

## Setting up DiceDB from source for development and contributions

To run DiceDB CLI for local development or running from source, you will need

1. [Golang](https://go.dev/)
2. Any of the below supported platform environments:
    1. [Linux based environment](https://en.wikipedia.org/wiki/Comparison_of_Linux_distributions)
    2. [OSX (Darwin) based environment](https://en.wikipedia.org/wiki/MacOS)
    3. WSL under Windows

```bash
$ git clone https://github.com/dicedb/dicedb-cli
$ cd dicedb-cli
$ go run main.go
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
