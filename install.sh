#!/usr/bin/env sh
set -e

REPO="dicedb/dicedb-cli"
LATEST_RELEASE=$(curl -s https://api.github.com/repos/$REPO/releases/latest | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
VERSION=$(echo $LATEST_RELEASE | sed 's/^v//')

# Detect the operating system and architecture
OS=$(uname -s)
ARCH=$(uname -m)

# Convert OS/ARCH to the naming convention used in releases
case $OS in
  Linux) OS="linux" ;;
  Darwin) OS="darwin" ;;
  CYGWIN*|MINGW32*|MSYS*|MINGW*) OS="windows" ;;
  *) echo "OS not supported"; exit 1 ;;
esac

case $ARCH in
  x86_64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) echo "Architecture not supported"; exit 1 ;;
esac

BINARY="dicedb-cli_${VERSION}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/$LATEST_RELEASE/$BINARY"

echo "Downloading $BINARY..."
curl -L $URL -o /tmp/$BINARY

# Extract and move to /usr/local/bin
tar -xzf /tmp/$BINARY -C /tmp
chmod 777 /tmp/dicedb-cli

DICEDB_DIR=/usr/local/dicedb
DICEDB_BIN_DIR=$DICEDB_DIR/bin

if [ ! -d "$DICEDB_DIR" ]; then
  sudo mkdir -p $DICEDB_DIR
fi

if [ ! -d "$DICEDB_BIN_DIR" ]; then
  sudo mkdir -p $DICEDB_BIN_DIR
  sudo chmod 777 $DICEDB_BIN_DIR
fi

mv /tmp/dicedb-cli $DICEDB_BIN_DIR
sudo ln -sf $DICEDB_BIN_DIR/dicedb-cli /usr/local/bin/dicedb-cli

echo "\n
██████╗ ██╗ ██████╗███████╗██████╗ ██████╗ 
██╔══██╗██║██╔════╝██╔════╝██╔══██╗██╔══██╗
██║  ██║██║██║     █████╗  ██║  ██║██████╔╝
██║  ██║██║██║     ██╔══╝  ██║  ██║██╔══██╗
██████╔╝██║╚██████╗███████╗██████╔╝██████╔╝
╚═════╝ ╚═╝ ╚═════╝╚══════╝╚═════╝ ╚═════╝
"
echo "> if you get 'command not found' error, add '/usr/local/bin' to your 'PATH' variable."
echo "\nDiceDB CLI installation complete ✓"
