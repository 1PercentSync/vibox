#!/bin/bash
# Install and start code-server

set -e

echo "=== Installing code-server ==="

# Install dependencies
apt-get update
apt-get install -y curl wget

# Download and install code-server
curl -fsSL https://code-server.dev/install.sh | sh

echo "=== code-server installed successfully ==="

# Create config directory
mkdir -p ~/.config/code-server

# Configure code-server (no password, bind to all interfaces)
echo "bind-addr: 0.0.0.0:8080" > ~/.config/code-server/config.yaml
echo "auth: none" >> ~/.config/code-server/config.yaml
echo "cert: false" >> ~/.config/code-server/config.yaml

echo "=== Configuration created ==="
echo "Config: ~/.config/code-server/config.yaml"
echo ""
echo "=== Starting code-server in background ==="
echo "Access at: http://localhost:8080"
echo "Or use port forwarding to access from outside"
echo ""

# Start code-server in background
nohup code-server > /var/log/vibox/code-server.log 2>&1 &

# Wait a moment for startup
sleep 2

# Check if it's running
if pgrep -f code-server > /dev/null; then
    echo "✓ code-server is running on port 8080"
    echo "Logs: /var/log/vibox/code-server.log"
else
    echo "✗ Failed to start code-server"
    exit 1
fi
