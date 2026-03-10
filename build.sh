#!/bin/bash

# Exit on error
set -e

echo "Building gobuild..."
go build -o gobuild ./cmd/gobuild

echo "Installing gobuild to /usr/local/bin..."
sudo mv gobuild /usr/local/bin/gobuild

echo "Installation complete! You can now run 'gobuild' from anywhere."
