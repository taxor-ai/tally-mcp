# Tally MCP - Claude Desktop Extension

Connect Claude to your Tally accounting software. Query companies, ledgers, debtors, creditors, and create new financial records directly from Claude conversations.

## What is Tally MCP?

**Tally MCP** is a [Model Context Protocol](https://modelcontextprotocol.io/) (MCP) server packaged as a Claude Desktop Extension. It allows Claude to:

- Query your Tally company data
- List ledgers and accounts
- Look up debtor and creditor information
- Create new ledgers and vouchers
- All without leaving the Claude conversation interface

Built with Go for cross-platform support (macOS, Linux, Windows).

## Features

✅ **Read Operations** (Fully Tested & Working)
- `get_companies` - List all companies in Tally
- `get_ledgers` - List all ledgers from the configured company
- `get_ledger_details` - Get detailed information for a specific ledger
- `get_debtors` - List all debtor ledgers with outstanding amounts
- `get_creditors` - List all creditor ledgers with outstanding amounts
- `get_creditor_vouchers` - Fetch past vouchers (transactions) for a specific creditor
- `get_debtor_vouchers` - Fetch past sales vouchers for a specific debtor/customer

✅ **Write Operations**
- `create_ledger` - Create a new ledger account
- `create_journal_voucher` - Create journal vouchers with multiple ledger entries (expense tracking)
- `create_sales_voucher` - Create sales vouchers (customer invoices)

✅ **Cross-Platform**
- macOS (Apple Silicon & Intel)
- Linux
- Windows

✅ **Easy Installation**
- Single `.mcpb` file
- Automated build script (`build/claude_extension.sh`)
- Configuration via Claude Desktop UI

✅ **Well-Tested**
- Integration tests against real Tally instances
- Comprehensive test suite included
- All core features verified working

## Prerequisites

- **Claude Desktop** (native application, not claude.ai)
- **Tally Server** running locally or accessible at network address
- Tally company credentials

## Installation

### For Users

1. **Download** the `.mcpb` file: `dist/tally-mcp-0.1.0.mcpb`

2. **Install in Claude Desktop:**
   - Open Claude Desktop
   - Settings → Customization → Connectors
   - Click "Add Custom Connector"
   - Select the `.mcpb` file

3. **Configure:**
   - Enter your Tally Server host
   - Enter port (usually 9900)
   - Enter company name as it appears in Tally
   - Click Install

4. **Restart** Claude Desktop completely

5. **Verify** - Ask Claude: "What tools do I have access to?"

### For Developers

See **Building from Source** section below.

## Building from Source

### Prerequisites

- Go 1.20 or higher
- Node.js 16+ (for mcpb package builder)
- Git

### Quick Build

The simplest way to build:

```bash
# Clone the repository
git clone https://github.com/taxor-ai/tally-mcp.git
cd tally-mcp

# Run the build script
./build/claude_extension.sh
```

This creates a ready-to-install `.mcpb` file at `dist/tally-mcp-0.1.0.mcpb`.

### Build Script Details

The automated build script at `build/claude_extension.sh` handles:

- Building binaries for all platforms (macOS ARM64/Intel, Linux, Windows)
- Copying extension metadata (manifest, icon)
- Validating the configuration
- Packaging as `.mcpb` file
- Displaying installation instructions

**Usage:**
```bash
./build/claude_extension.sh
```

**Output:**
- Extension file: `dist/tally-mcp-0.1.0.mcpb`
- Ready to install in Claude Desktop
- Includes SHA256 hash for verification

### Manual Build (Advanced)

If you need more control, build step by step:

```bash
# Create directory structure
mkdir -p dist/tally-mcp-bundle/server

# Build binaries
GOOS=darwin GOARCH=arm64 go build -o dist/tally-mcp-bundle/server/tally-mcp-mac .
GOOS=darwin GOARCH=amd64 go build -o dist/tally-mcp-bundle/server/tally-mcp-mac-x86 .
GOOS=linux GOARCH=amd64 go build -o dist/tally-mcp-bundle/server/tally-mcp-linux .
GOOS=windows GOARCH=amd64 go build -o dist/tally-mcp-bundle/server/tally-mcp.exe .

# Copy metadata
cp integrations/claude/manifest.json dist/tally-mcp-bundle/
cp integrations/claude/icon.png dist/tally-mcp-bundle/
cp tally-mcp dist/tally-mcp-bundle/server/
chmod +x dist/tally-mcp-bundle/server/tally-mcp

# Package
npx @anthropic-ai/mcpb pack dist/tally-mcp-bundle dist/tally-mcp-0.1.0.mcpb
```

## Configuration

### Environment Variables

The extension reads these environment variables:

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `TALLY_HOST` | Tally server hostname or IP | - | Yes |
| `TALLY_PORT` | Tally server port | 9900 | No |
| `TALLY_COMPANY` | Company name in Tally | - | Yes |
| `TALLY_LOG_LEVEL` | Log level: debug, info, warn, error | info | No |
| `TALLY_LOG_FILE` | Path to log file (optional) | - | No |

### From Claude Desktop

When you install the extension, Claude Desktop prompts for configuration:

- **Tally Server Host** - Your Tally server address (localhost, 192.168.x.x, etc.)
- **Tally Port** - Port where Tally is listening (default: 9900)
- **Company Name** - Exact name of the company in Tally
- **Log Level** - Verbosity for debugging

## Usage

Once installed, ask Claude naturally:

```
You: "List all companies in Tally"
Claude: Uses get_companies tool → Shows results

You: "What ledgers does [Company Name] have?"
Claude: Uses get_ledgers tool → Shows ledger list

You: "Get details for the Sales ledger"
Claude: Uses get_ledger_details tool → Shows full details

You: "Create a new ledger called 'Marketing Expenses'"
Claude: Uses create_ledger tool → Creates and confirms
```

Claude automatically uses the right tool for your request. No special syntax needed.

## Project Structure

```
tally-mcp/
├── main.go                    # MCP server entry point
├── config/
│   └── config.go             # Configuration management
├── logger/
│   └── logger.go             # Structured logging (zap)
├── mcp/
│   ├── handler.go            # Tool call routing
│   ├── tools.go              # Tool definitions
│   └── response.go           # Response formatting
├── tally/
│   ├── client.go             # Tally XML-RPC client
│   ├── models.go             # Data models
│   ├── templates/            # XML-RPC templates
│   └── [parser files]        # Response parsing
├── dist/
│   └── tally-mcp-bundle/
│       ├── manifest.json     # Extension manifest (v0.2)
│       ├── icon.png          # Extension icon
│       └── server/           # Platform binaries
└── README.md
```

## Architecture

### MCP Protocol

Tally MCP implements the **Model Context Protocol**, a JSON-RPC 2.0 protocol over stdin/stdout:

1. Claude Desktop starts the server process
2. Server reads JSON-RPC requests from stdin
3. Server processes and responds via stdout
4. All logs go to stderr (preserving stdout for protocol)

### Tool Definitions

Each tool defines:
- **Name** - Tool identifier (e.g., `get_companies`)
- **Description** - What the tool does
- **InputSchema** - JSON Schema for parameters
- **IsWrite** - Whether it modifies data

### Tally Integration

Uses **XML-RPC protocol** to communicate with Tally:
- Sends XML requests to `http://[host]:[port]/`
- Parses XML responses
- Handles Tally-specific data formats

## Development

### Running Locally

```bash
# Build the binary
go build -o tally-mcp .

# Set environment variables
export TALLY_HOST=localhost
export TALLY_PORT=9900
export TALLY_COMPANY="Your Company"
export TALLY_LOG_LEVEL=debug

# Run the server
./tally-mcp
```

### Testing

**Unit Tests**
```bash
# Run all unit tests
go test ./...

# With coverage
go test -cover ./...
```

**Integration Tests** (requires live Tally instance)
```bash
# Test against real Tally server
TALLY_HOST=localhost TALLY_PORT=9000 TALLY_COMPANY="Your Company" \
go test -tags=integration -v ./tests/integration/...

# Run specific test
TALLY_HOST=localhost TALLY_PORT=9000 TALLY_COMPANY="Your Company" \
go test -tags=integration -run TestGetLedgersIntegration -v ./tests/integration/...

# Test all GET commands in sequence
TALLY_HOST=localhost TALLY_PORT=9000 TALLY_COMPANY="Your Company" \
go test -tags=integration -run TestAllGetToolsSequenceIntegration -v ./tests/integration/...
```

**Available Integration Tests:**
- `TestGetCompaniesRealTally` - Verify company retrieval
- `TestGetLedgersIntegration` - Verify ledger retrieval
- `TestGetDebtorsIntegration` - Verify debtor retrieval
- `TestGetCreditorsIntegration` - Verify creditor retrieval
- `TestAllGetToolsSequenceIntegration` - Comprehensive test of all GET tools

### Adding New Tools

1. **Define the tool** in `mcp/tools.go`:
   ```go
   func GetYourToolTool() Tool {
       return Tool{
           Name: "your_tool",
           Description: "...",
           InputSchema: map[string]interface{}{...},
           IsWrite: false,
       }
   }
   ```

2. **Add to AllTools()** in `mcp/tools.go`

3. **Implement handler** in `mcp/handler.go`:
   ```go
   case "your_tool":
       return h.handleYourTool(params)
   ```

4. **Create handler method** with Tally API calls

5. **Test** before rebuilding package

## Troubleshooting

### Tools not showing up

**Cause:** JSON field names incorrect

**Solution:** Ensure struct tags are lowercase:
```go
type Tool struct {
    Name        string `json:"name"`
    Description string `json:"description"`
    InputSchema map... `json:"inputSchema"`
}
```

### "Server transport closed unexpectedly"

**Cause:** Missing configuration or logging to stdout

**Solution:**
- Verify all required env vars are set
- Check logs at `~/Library/Logs/Claude/mcp.log`
- Ensure logger outputs to stderr, not stdout

### Empty results from tools

**Cause:** Multiple possible reasons:
1. Tally server not responding
2. Company name is incorrect or empty
3. Tally instance doesn't have data for that company
4. Network connectivity issues

**Solution:**
- Verify Tally server is running at the configured host/port
- **Double-check company name** - Must match exactly as shown in Tally
- Verify network connectivity: `ping [TALLY_HOST]`
- Check environment variables are set correctly:
  ```bash
  echo $TALLY_HOST
  echo $TALLY_PORT
  echo $TALLY_COMPANY
  ```
- Enable debug logging to see XML requests/responses:
  ```bash
  export TALLY_LOG_LEVEL=debug
  ```

**Testing connectivity:**
```bash
# Run integration tests against your Tally instance
TALLY_HOST=your.tally.host TALLY_PORT=9000 TALLY_COMPANY="Company Name" \
go test -tags=integration -v ./tests/integration/...
```

See [blog post](../blogs/creating-claude-desktop-extension.md) for more detailed troubleshooting.

## Logging

Logs are written to:
- **stderr** - Real-time logging (to console)
- **File** - Optional, set `TALLY_LOG_FILE` env var

Format: JSON structured logs (zap library)

Example log output:
```json
{"level":"info","ts":1712973600.847,"logger":"","msg":"get_companies completed","count":1}
```

## Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

MIT License - See LICENSE file for details

## Resources

- [Anthropic Desktop Extensions Guide](https://www.anthropic.com/engineering/desktop-extensions)
- [Model Context Protocol Docs](https://modelcontextprotocol.io/)
- [Claude Desktop Debugging](https://modelcontextprotocol.io/docs/tools/debugging)
- [MCP Package Builder](https://github.com/anthropics/mcpb)

## Support

- Issues: [GitHub Issues](https://github.com/taxor-ai/tally-mcp/issues)
- Questions: Check [blog post](../blogs/creating-claude-desktop-extension.md) or open a discussion
- Tally Support: [Tally Documentation](https://tallysolutions.com/)

## Authors

Built by [Taxor AI](https://taxor.ai)

---

**Made with ❤️ for accounting professionals using Claude**
