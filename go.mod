module github.com/DiceDB/dicedb-cli

go 1.23.5

require (
	github.com/c-bata/go-prompt v0.2.6
	github.com/chzyer/readline v1.5.1
	github.com/dicedb/dicedb-go v1.0.1
	github.com/fatih/color v1.18.0
	github.com/spf13/cobra v1.8.1
	google.golang.org/protobuf v1.36.4
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/mattn/go-tty v0.0.3 // indirect
	github.com/pkg/term v1.2.0-beta.2 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/sys v0.25.0 // indirect
)

replace github.com/DiceDB/dicedb-go => ../dicedb-go
