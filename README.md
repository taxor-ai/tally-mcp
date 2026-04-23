# Tally MCP - Claude Desktop Extension

Connect Claude to your Tally accounting software. Query companies, ledgers, debtors, creditors, and create new financial records directly from Claude conversations.

**Extension runs locally. No proprietary cloud service. No additional hosting required.**

## What is Tally MCP?

**Tally MCP** is a [Model Context Protocol](https://modelcontextprotocol.io/) (MCP) server packaged as a Claude Desktop Extension. It allows Claude to:

- Query your Tally company data
- List ledgers and accounts
- Look up debtor and creditor information
- Create new ledgers and vouchers
- All without leaving the Claude conversation interface

**The extension runs locally.** No separate cloud service to manage. Direct integration between Claude Desktop and your Tally instance.

Built with Go for cross-platform support (macOS, Linux, Windows).

## Features

✅ **Read Operations** (Fully Tested & Working)
- `get_companies` - List all companies in Tally
- `get_ledgers` - List all ledgers from the configured company
- `get_ledger_details` - Get detailed information for a specific ledger
- `get_debtors` - List all debtor ledgers with outstanding amounts
- `get_creditors` - List all creditor ledgers with outstanding amounts
- `get_sales_vouchers` - Fetch past sales vouchers for a specific customer/debtor ledger
- `get_journal_vouchers` - Fetch past journal vouchers for a specific vendor/creditor ledger

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

✅ **Local Extension, No Proprietary Cloud**
- Extension runs on your machine
- Direct connection between Claude and Tally
- No third-party cloud service to manage
- Open source and auditable

✅ **Well-Tested**
- Integration tests against real Tally instances
- Comprehensive test suite included
- All core features verified working

## How It Works

```
Your Machine:
├── Claude Desktop (installed)
├── Tally MCP Extension (.mcpb file)
└── Tally Server (local or network)

Data Flow:
Claude Desktop → Local MCP Extension → Tally

Extension runs locally. No separate cloud infrastructure to manage.
```

## Prerequisites

- **Claude Desktop** (native application, not claude.ai)
- **Tally Server** running locally or accessible at network address
- Tally company credentials

## Tally Setup

Before installing the extension, configure your Tally instance:

### 1. Enable XML API Access

Tally communicates via XML-RPC protocol. Ensure it's enabled:

1. Open **Tally** application
2. Go to **F1 (Help) → F1 (Help)** or **Gateway → Settings**
3. Navigate to **Network/Internet** or **TCP-IP Port Settings**
4. Confirm **TCP Port** is enabled (default: 9900)
5. Note the **port number** - you'll need this during installation

> Tally typically has XML-RPC enabled by default on port 9900.

### 2. Configure Voucher Numbering for Automatic Creation

The extension creates Sales and Journal vouchers with automatic numbering. Configure Tally's numbering method:

#### For Sales Vouchers:
1. In Tally, go to **Gateway → Masters → Voucher Type**
2. Select or open the **Sales** voucher type
3. Look for **Numbering Method** or **Number Series**
4. Set to: **Automatic** (not "Automatic with Manual Override")
   - This allows the extension to create vouchers without specifying manual voucher numbers
5. Save changes

#### For Journal Vouchers:
1. In Tally, go to **Gateway → Masters → Voucher Type**
2. Select or open the **Journal** voucher type
3. Set **Numbering Method** to: **Automatic**
4. Save changes

> **Important:** If numbering is set to "Automatic with Manual Override" or manual modes, voucher creation may fail with validation errors.

### 3. Verify Company Exists

1. In Tally, go to **Gateway → Masters → Company**
2. Note the **exact name** of your company (case-sensitive)
3. You'll need this exact name during extension configuration

## Installation

### For Users (5 Minutes, Simple Setup)

1. **Download** the `.mcpb` file from [Releases](https://github.com/taxor-ai/tally-mcp/releases)

2. **Install in Claude Desktop:**
   - Open Claude Desktop
   - Go to Settings > Extensions > Advanced settings
   - Click "Install extension"
   - Select the `.mcpb` file

3. **Configure:**
   - Enter your Tally Server host (localhost or internal IP)
   - Enter port (usually 9900)
   - Enter company name as it appears in Tally
   - Click Install

4. **Restart** Claude Desktop completely

5. **Verify** - Ask Claude: "What tools do I have access to?"

**That's it.** Extension runs locally. No separate service to host or manage.

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
bash build/claude_extension.sh
```

This creates a ready-to-install `.mcpb` file at `dist/tally-mcp-0.1.0.mcpb`.

Then install following the user instructions above.

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

## Using as Standalone MCP Server

If you want to use tally-mcp with other MCP-compatible clients, custom integrations, or run it independently from Claude Desktop:

### 1. Download the Binary

Download the appropriate pre-built binary for your OS from [Releases](https://github.com/taxor-ai/tally-mcp/releases):

- **macOS ARM64 (Apple Silicon):** `tally-mcp-mac`
- **macOS Intel:** `tally-mcp-mac-x86`
- **Linux:** `tally-mcp-linux`
- **Windows:** `tally-mcp.exe`

### 2. Make it Executable (macOS/Linux only)

```bash
chmod +x tally-mcp-linux    # For Linux
chmod +x tally-mcp-mac      # For macOS ARM64
chmod +x tally-mcp-mac-x86  # For macOS Intel
```

Windows `.exe` files are already executable.

### 3. Configure Environment Variables

Set the required configuration before running:

```bash
export TALLY_HOST=localhost          # or your Tally server IP/hostname
export TALLY_PORT=9900               # Tally server port (default: 9900)
export TALLY_COMPANY="Your Company"  # Company name exactly as in Tally
export TALLY_LOG_LEVEL=info          # Optional: debug, info, warn, error
```

**Windows (PowerShell):**
```powershell
$env:TALLY_HOST = "localhost"
$env:TALLY_PORT = "9900"
$env:TALLY_COMPANY = "Your Company"
```

### 4. Run the Server

```bash
./tally-mcp-linux      # Linux
./tally-mcp-mac        # macOS ARM64
./tally-mcp-mac-x86    # macOS Intel
./tally-mcp.exe        # Windows
```

The MCP server will start and listen on stdin/stdout for JSON-RPC requests from your MCP client.

### Using with Other MCP Clients

You can now integrate tally-mcp with any MCP-compatible client. The server implements the [Model Context Protocol](https://modelcontextprotocol.io/) and communicates via JSON-RPC 2.0 over stdin/stdout.

**Example integration in your MCP client config:**

```json
{
  "tools": {
    "tally": {
      "command": "/path/to/tally-mcp-linux",
      "env": {
        "TALLY_HOST": "your-server.local",
        "TALLY_PORT": "9900",
        "TALLY_COMPANY": "Your Company"
      }
    }
  }
}
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
  - *From Tally Setup:* Found in Tally's Network/TCP-IP settings
- **Tally Port** - Port where Tally is listening (default: 9900)
  - *From Tally Setup:* Confirmed in Tally's Network settings
- **Company Name** - Exact name of the company in Tally (case-sensitive)
  - *From Tally Setup:* Verified in Gateway → Masters → Company
- **Log Level** - Verbosity for debugging (debug, info, warn, error)

> **Before configuring here:** Complete the [Tally Setup](#tally-setup) section above

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
├── pkg/tally/
│   ├── client.go             # Tally XML-RPC client
│   ├── models.go             # Data models
│   ├── parser.go             # Response parsing
│   └── [other files]         # Supporting modules
├── tools/                     # Tool definitions (XML + YAML)
│   ├── company/              # Company tools
│   ├── ledger/               # Ledger tools
│   └── voucher/              # Voucher tools
├── tests/
│   ├── unit/                 # Unit tests
│   └── integration/          # Integration tests (requires live Tally)
├── dist/
│   └── tally-mcp-bundle/
│       ├── manifest.json     # Extension manifest
│       ├── icon.png          # Extension icon
│       └── server/           # Platform binaries
└── README.md
```

## Architecture

### Extension-Based Design

- Claude Desktop starts the MCP server locally
- Direct communication between Claude and your Tally instance
- No separate cloud service or hosted backend required

### MCP Protocol

Tally MCP implements the **Model Context Protocol**, a JSON-RPC 2.0 protocol over stdin/stdout:

1. Claude Desktop starts the server process locally
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

# Or use make target
make test-unit
```

**Integration Tests** (requires live Tally instance)

> **Note:** Integration tests use configuration from `.env.test` file—no command-line arguments needed!

#### 1. Setup Configuration (One-time)

```bash
# Copy the example configuration file
cp .env.test.example .env.test

# Edit .env.test with your Tally server connection details
nano .env.test
```

Example `.env.test` content:
```env
TALLY_HOST=your-tally-server.com
TALLY_PORT=9000
TALLY_COMPANY=Your Company Name
TALLY_LOG_LEVEL=info
```

#### 2. Run Integration Tests

```bash
# Run all integration tests (recommended)
make test-integration

# Or run directly with go test
go test -tags=integration -v ./tests/integration/...

# Run specific test
go test -tags=integration -run TestListAllCreditors -v ./tests/integration/...

# Run multiple specific tests
go test -tags=integration -run "TestGet.*Integration" -v ./tests/integration/...
```

#### Available Integration Tests

**Read Operations (Query Tally Data):**
- `TestGetCompaniesIntegration` - List all companies in Tally
- `TestGetLedgersIntegration` - Retrieve all ledgers
- `TestGetDebtorsIntegration` - Get debtor (customer) list
- `TestGetCreditorsIntegration` - Get creditor (vendor) list
- `TestListAllCreditors` - Detailed creditor list with full details
- `TestGetCreditorVouchersIntegration` - Fetch vendor transactions
- `TestGetDebtorVouchersIntegration` - Fetch customer transactions
- `TestRegistryHasAllExpectedTools` - Verify all tools are available
- `TestAllGetToolsSequenceIntegration` - Run all GET tools in sequence

**Write Operations (Create Records):**
- `TestCreateJournalVoucherIntegration` - Create a journal entry
- `TestCreateSalesVoucherIntegration` - Create a sales invoice

#### Configuration Priority

The config loader checks in this order (first found wins):
1. **Environment variables** (highest priority—useful for CI/CD)
2. `.env.local` (local development overrides)
3. `.env.test` (integration test configuration)
4. `.env` (default configuration)

**Example CI/CD usage:**
```bash
# Override .env.test values with environment variables in CI
TALLY_HOST=${{ secrets.TALLY_HOST }} \
TALLY_PORT=${{ secrets.TALLY_PORT }} \
TALLY_COMPANY=${{ secrets.TALLY_COMPANY }} \
go test -tags=integration -v ./tests/integration/...
```

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

### Voucher creation fails with exceptions

**Cause:** Tally voucher numbering is not set to "Automatic"

**Solution:**
1. Open Tally and navigate to **Gateway → Masters → Voucher Type**
2. For both **Sales** and **Journal** voucher types:
   - Select each voucher type
   - Change **Numbering Method** to **Automatic** (not "Automatic with Manual Override")
   - Save changes
3. Retry voucher creation in Claude

**Error messages that indicate this issue:**
- `EXCEPTIONS=1` in Tally response
- "Duplicate voucher number" 
- Voucher creation returns `created=0`

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
