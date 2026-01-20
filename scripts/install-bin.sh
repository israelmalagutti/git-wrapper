#!/bin/bash
# Quick local install script - builds and installs gw
# For development use. For production, use: make install

set -e
cd "$(dirname "$0")/.."
make install
