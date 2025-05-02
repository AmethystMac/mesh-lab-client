#!/bin/bash

# Path to directories
BUILD_PATH="$PWD/build"
CMD_PATH="$PWD/cmd"

# Paths to your executables 
GETH_PATH="$CMD_PATH/geth"
CLEF_PATH="$CMD_PATH/clef"

# Building binaries
echo 'Building Binaries...';
go build -o $BUILD_PATH $GETH_PATH
go build -o $BUILD_PATH $CLEF_PATH

echo 'Build successful!'

# Command to start Geth in a new terminal window
# gnome-terminal -- bash -c "
# echo 'Starting Geth...';
# $BUILD_PATH/geth
# read -p 'Press Enter to exit Geth';"

# Command to start Clef in a new terminal window, using Clef as the signer
# gnome-terminal -- bash -c "
# echo 'Starting Clef...';
# $BUILD_PATH/clef
# # read -p 'Press Enter to exit Clef';"

# Optional: You can add more setup steps here if necessary
