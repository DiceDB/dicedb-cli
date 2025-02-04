module github.com/dicedb/dicedb-cli

go 1.23.5

require (
	github.com/chzyer/readline v1.5.1
	github.com/dicedb/dicedb-go v1.0.1
	github.com/fatih/color v1.18.0
	github.com/spf13/cobra v1.8.1
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/sys v0.25.0 // indirect
	google.golang.org/protobuf v1.36.4 // indirect
)

replace github.com/dicedb/dicedb-go v1.0.1 => ../dicedb-go
