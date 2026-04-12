#!/bin/bash
set -e

# MCP Server Wrapper Script
# Detects OS and architecture, then executes the appropriate binary

# Get the directory where this script is located
if [ -L "${BASH_SOURCE[0]}" ]; then
  # Handle symlinks
  SCRIPT_DIR="$(cd "$(dirname "$(readlink "${BASH_SOURCE[0]}")")" && pwd)"
else
  SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
fi

# Detect OS and architecture
OS=$(uname -s)
ARCH=$(uname -m)

# Debug info to stderr
echo "DEBUG: Script directory: $SCRIPT_DIR" >&2
echo "DEBUG: OS: $OS, Architecture: $ARCH" >&2

# Determine which binary to run
case "$OS" in
  Darwin)
    case "$ARCH" in
      arm64)
        BINARY="$SCRIPT_DIR/tally-mcp-mac"
        ;;
      x86_64)
        BINARY="$SCRIPT_DIR/tally-mcp-mac-x86"
        ;;
      *)
        echo "ERROR: Unsupported macOS architecture: $ARCH" >&2
        exit 1
        ;;
    esac
    ;;
  Linux)
    case "$ARCH" in
      x86_64)
        BINARY="$SCRIPT_DIR/tally-mcp-linux"
        ;;
      *)
        echo "ERROR: Unsupported Linux architecture: $ARCH" >&2
        exit 1
        ;;
    esac
    ;;
  MINGW64_NT*|MSYS_NT*|CYGWIN_NT*)
    BINARY="$SCRIPT_DIR/tally-mcp.exe"
    ;;
  *)
    echo "ERROR: Unsupported operating system: $OS" >&2
    exit 1
    ;;
esac

# Verify binary exists and is executable
if [ ! -f "$BINARY" ]; then
  echo "ERROR: Binary not found at $BINARY" >&2
  ls -la "$SCRIPT_DIR/" >&2
  exit 1
fi

if [ ! -x "$BINARY" ]; then
  echo "ERROR: Binary is not executable: $BINARY" >&2
  chmod +x "$BINARY"
fi

echo "DEBUG: Using binary: $BINARY" >&2

# Execute the binary with all arguments, passing through stdin/stdout/stderr
exec "$BINARY" "$@"
