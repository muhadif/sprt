#!/bin/bash

# sprt Installation Script

# Set colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Print banner
echo -e "${GREEN}"
echo "  _____                  _   _      _ _ "
echo " / ____|                | | (_)    | (_)"
echo "| (___  _ __   ___  _ __| |_ _  ___| |_ "
echo " \___ \| '_ \ / _ \| '__| __| |/ __| | |"
echo " ____) | |_) | (_) | |  | |_| | (__| | |"
echo "|_____/| .__/ \___/|_|   \__|_|\___|_|_|"
echo "       | |                              "
echo "       |_|                              "
echo -e "${NC}"
echo "Spotify CLI Client - Installation Script"
echo "========================================"
echo

# Check if running as root
if [ "$EUID" -ne 0 ]; then
  echo -e "${YELLOW}Warning: Not running as root. Installation may fail if you don't have sufficient permissions.${NC}"
  echo "You may need to run this script with sudo."
  echo
  read -p "Continue anyway? (y/n) " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${RED}Installation aborted.${NC}"
    exit 1
  fi
fi

# Detect OS
OS="unknown"
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
  OS="linux"
elif [[ "$OSTYPE" == "darwin"* ]]; then
  OS="darwin"
else
  echo -e "${RED}Unsupported operating system: $OSTYPE${NC}"
  echo "This script supports Linux and macOS only."
  exit 1
fi

echo -e "${GREEN}Detected OS: ${OS}${NC}"

# Set installation paths
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="$HOME/.sprt"

# Check if binary exists in current directory
if [ ! -f "./sprt" ]; then
  echo -e "${RED}Error: sprt binary not found in current directory.${NC}"
  echo "Make sure you're running this script from the directory containing the sprt binary."
  exit 1
fi

# Create config directory if it doesn't exist
echo "Creating configuration directory..."
mkdir -p "$CONFIG_DIR"
if [ $? -ne 0 ]; then
  echo -e "${RED}Error: Failed to create configuration directory.${NC}"
  exit 1
fi

# Install binary
echo "Installing sprt binary to $INSTALL_DIR..."
cp ./sprt "$INSTALL_DIR/"
if [ $? -ne 0 ]; then
  echo -e "${RED}Error: Failed to copy binary to $INSTALL_DIR.${NC}"
  echo "You may need to run this script with sudo."
  exit 1
fi

# Make binary executable
chmod +x "$INSTALL_DIR/sprt"
if [ $? -ne 0 ]; then
  echo -e "${RED}Error: Failed to make binary executable.${NC}"
  exit 1
fi

echo -e "${GREEN}Installation completed successfully!${NC}"
echo
echo "You can now run sprt by typing 'sprt' in your terminal."
echo
echo "To get started, run:"
echo "  sprt auth init"
echo
echo "For more information, see the README.md file or run:"
echo "  sprt --help"
echo
echo "Enjoy using sprt!"