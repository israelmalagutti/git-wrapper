#!/bin/bash
set -e

# Build the binary and wait for completion
make build

# Copy to /usr/local/bin
sudo cp bin/gw /usr/local/bin/gw
echo "Installed Git-Wrapper (gw) CLI!"
