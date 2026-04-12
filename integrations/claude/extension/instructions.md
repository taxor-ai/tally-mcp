# Tally MCP Tools: Adding and Updating

This guide explains how to extend and modify the Tally MCP server with new tools and update existing ones without modifying any Go code.

## Quick Start

1. Create a new template directory
2. Add three configuration files
3. Restart the MCP server
4. The new tool is automatically available

## Directory Structure

Templates are organized by category. Each tool is its own subdirectory:

```
templates/
├── company/
│   └── get_companies/
│       ├── tool.yaml        (metadata & input schema)
│       ├── request.xml      (Tally XML request template)
│       └── parser.yaml      (XPath response parser)
├── ledger/
│   ├── get_ledgers/
│   ├── get_ledger_details/
│   └── create_ledger/
├── debtor_creditor/
│   ├── get_debtors/
│   └── get_creditors/
└── your_category/
    └── your_new_tool/       ← Add here
        ├── tool.yaml
        ├── request.xml
        └── parser.yaml
```

## File Specifications

### 1. tool.yaml - Tool Metadata

Defines the tool name, description, and input parameters.

```yaml
name: your_tool_name
description: Brief description of what this tool does
implicit_params:
  - company_name          # Parameters auto-injected from config
input_schema:
  properties:
    param1:
      type: string
      description: Description of param1
    param2:
      type: string
      description: Description of param2
      enum:
        - option1
        - option2
  required:
    - param1              # Required parameters
```

**Field Reference:**
- `name`: Tool identifier (snake_case, must be unique)
- `description`: User-facing description
- `implicit_params`: Parameters injected automatically (e.g., company_name from server config)
- `input_schema.properties`: Input parameters and their types
- `input_schema.required`: Which parameters are mandatory

**Supported Types:**
- `string` - Text input
- `number` - Floating point numbers
- `integer` - Whole numbers
- `boolean` - True/false

### 2. request.xml - Tally Request Template

XML request sent to Tally's XML-RPC API. Use Tally's TDL syntax.

```xml
<ENVELOPE>
    <HEADER>
        <VERSION>1</VERSION>
        <TALLYREQUEST>Export</TALLYREQUEST>
        <TYPE>Collection</TYPE>
        <ID>YourCollectionName</ID>
    </HEADER>
    <BODY>
        <DESC>
            <STATICVARIABLES>
                <SVEXPORTFORMAT>$$SysName:XML</SVEXPORTFORMAT>
                <SVCURRENTCOMPANY>{{ company_name | e }}</SVCURRENTCOMPANY>
                <SVYOURPARAM>{{ param1 }}</SVYOURPARAM>
            </STATICVARIABLES>
            <TDL>
                <TDLMESSAGE>
                    <COLLECTION NAME="YourCollectionName">
                        <TYPE>Ledger</TYPE>
                        <FETCH>NAME, PARENT, CLOSINGBALANCE</FETCH>
                    </COLLECTION>
                </TDLMESSAGE>
            </TDL>
        </DESC>
    </BODY>
</ENVELOPE>
```

**Template Variables:**
- `{{ param_name }}` - Insert parameter value as-is
- `{{ param_name | e }}` - Insert parameter value XML-escaped (safer)

Always use `| e` for user input to prevent XML injection.

### 3. parser.yaml - Response Parser

Maps Tally's XML response to structured data using XPath.

```yaml
type: list                          # Parser type: list, object, import_result, raw
items_xpath: //COLLECTION/LEDGER    # XPath to each item (for list type)
result_key: ledgers                 # Key in response JSON
fields:
  name:
    xpath: NAME
  parent:
    xpath: PARENT
  balance:
    xpath: CLOSINGBALANCE
    transform: number               # Field transformation (optional)
```

**Parser Types:**
- `list` - Returns array of items. Use `items_xpath` to locate each item.
- `object` - Returns single object. Use `root_xpath` to locate the root.
- `import_result` - Parse Tally creation response (IMPORTDATA format).
- `raw` - Return raw XML, no parsing.

**Field Transformations:**
- `number` - Convert to decimal (e.g., "100.50" → 100.50)
- `integer` - Convert to whole number (e.g., "100.50" → 100)
- `boolean` - Convert to boolean (e.g., "Yes" → true)
- `tally_date` - Convert Tally date format YYYYMMDD → YYYY-MM-DD
- *(omit for string fields)*

## Example: Add a "Get Company Info" Tool

**Step 1: Create directory**
```bash
mkdir -p templates/company/get_company_info
cd templates/company/get_company_info
```

**Step 2: Create tool.yaml**
```yaml
name: get_company_info
description: Get detailed information about a company by name
implicit_params:
  - company_name
input_schema:
  properties: {}
  required: []
```

**Step 3: Create request.xml**
```xml
<ENVELOPE>
    <HEADER>
        <VERSION>1</VERSION>
        <TALLYREQUEST>Export</TALLYREQUEST>
        <TYPE>Collection</TYPE>
        <ID>CompanyInfo</ID>
    </HEADER>
    <BODY>
        <DESC>
            <STATICVARIABLES>
                <SVEXPORTFORMAT>$$SysName:XML</SVEXPORTFORMAT>
                <SVCURRENTCOMPANY>{{ company_name | e }}</SVCURRENTCOMPANY>
            </STATICVARIABLES>
            <TDL>
                <TDLMESSAGE>
                    <COLLECTION NAME="CompanyInfo">
                        <TYPE>Company</TYPE>
                        <FETCH>NAME, MAILING.LIST, PHONE, EMAIL</FETCH>
                    </COLLECTION>
                </TDLMESSAGE>
            </TDL>
        </DESC>
    </BODY>
</ENVELOPE>
```

**Step 4: Create parser.yaml**
```yaml
type: object
root_xpath: //COLLECTION/COMPANY
result_key: company_info
fields:
  name:
    xpath: NAME
  phone:
    xpath: PHONE
  email:
    xpath: EMAIL
```

**Step 5: Restart the MCP server**

The tool is now available! Test it:
```bash
curl -X POST http://localhost:9000/ \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_company_info","arguments":{}}}'
```

## Testing Your Tool

### With the MCP Server
1. Place templates in the `templates/` directory
2. Set `TALLY_TEMPLATES_DIR` to the templates directory (or use default)
3. Restart the server
4. Call the tool via MCP protocol

### Manual Testing
1. Open `request.xml` and verify the TDL syntax matches your Tally version
2. Test the XPath expressions in `parser.yaml` against actual Tally responses
3. Ensure all field transformations are correct

## Troubleshooting

### Tool not appearing in tools/list
- Check that `tool.yaml` exists and is valid YAML
- Verify `name` field matches your directory name (recommended)
- Restart the server to reload templates

### Request fails with "Cannot connect to Tally"
- Verify Tally is running and accessible at the configured host:port
- Check that `{{ company_name | e }}` is escaped in request.xml
- Ensure company name in request matches a company in Tally

### Parser returns empty results
- Check `items_xpath` / `root_xpath` matches your Tally response structure
- Use a tool like `xmllint` to test XPath expressions
- Verify Tally is returning data for the requested parameters

### "Connection reset by peer"
- The Tally server may have crashed (too large a result set)
- Try limiting the query with filters or date ranges
- Check Tally logs for errors

## Common Tally TDL Patterns

### Export a collection with filters
```xml
<COLLECTION NAME="FilteredLedgers">
    <TYPE>Ledger</TYPE>
    <FILTER>OPENING.BALANCE > 0</FILTER>
    <FETCH>NAME, PARENT, OPENINGBALANCE</FETCH>
</COLLECTION>
```

### Group by category
```xml
<COLLECTION NAME="LedgersByGroup">
    <TYPE>Ledger</TYPE>
    <GROUPBY>PARENT</GROUPBY>
    <FETCH>PARENT, COUNT(*) as Count</FETCH>
</COLLECTION>
```

### Date range filter
```xml
<STATICVARIABLES>
    <SVFROMDATE>{{ from_date }}</SVFROMDATE>
    <SVTODATE>{{ to_date }}</SVTODATE>
</STATICVARIABLES>
<COLLECTION NAME="Transactions">
    <TYPE>Journal</TYPE>
    <FILTER>DATE >= $SVFROMDATE AND DATE &lt;= $SVTODATE</FILTER>
    <FETCH>DATE, REFKEY, NARRATION, AMOUNT</FETCH>
</COLLECTION>
```

## Best Practices

1. **Use XML escaping** - Always use `{{ param | e }}` for user input
2. **Minimize data** - Use FETCH to select only needed fields
3. **Add filters** - Reduce result set to avoid server crashes
4. **Document parameters** - Write clear descriptions in tool.yaml
5. **Test incrementally** - Start with simple queries, add complexity
6. **Version your tools** - Include version info in tool description if relevant

## Adding Tools: User Instructions

### Via File Explorer (Manual)

1. **Navigate to the templates folder:**
   - **macOS/Linux:** `~/.config/Claude/extensions/local.mcpb.taxor-ai.tally-mcp/server/templates/`
   - **Windows:** `%APPDATA%\Claude\extensions\local.mcpb.taxor-ai.tally-mcp\server\templates\`

2. **Create a new directory:**
   ```
   templates/your_category/your_tool_name/
   ```

3. **Create three files** in that directory:
   - `tool.yaml` - Tool metadata (see "File Specifications" above)
   - `request.xml` - Tally XML template (see example above)
   - `parser.yaml` - Response parser config (see example above)

4. **Restart Claude Desktop** to load the new tool

**Note:** Changes to templates take effect on restart without rebuilding.

## Adding Tools: Agent Instructions (Claude Code / Claude Copilot)

Agents can programmatically add tools using file tools:

### Step 1: Determine Template Directory

The templates directory is located at:
- **macOS/Linux:** `~/.config/Claude/extensions/local.mcpb.taxor-ai.tally-mcp/server/templates/`
- **Windows:** `%APPDATA%\Claude\extensions\local.mcpb.taxor-ai.tally-mcp\server\templates\`

### Step 2: Create Directory Structure

Use file operations to create:
```
{TEMPLATES_DIR}/your_category/your_tool_name/
```

### Step 3: Create tool.yaml

Create file at `{TEMPLATES_DIR}/your_category/your_tool_name/tool.yaml`:
```yaml
name: your_tool_name
description: Description of what this tool does
implicit_params:
  - company_name
input_schema:
  properties:
    param1:
      type: string
      description: Parameter description
  required:
    - param1
```

### Step 4: Create request.xml

Create file at `{TEMPLATES_DIR}/your_category/your_tool_name/request.xml`:
```xml
<ENVELOPE>
    <HEADER>
        <VERSION>1</VERSION>
        <TALLYREQUEST>Export</TALLYREQUEST>
        <TYPE>Collection</TYPE>
        <ID>YourToolName</ID>
    </HEADER>
    <BODY>
        <DESC>
            <STATICVARIABLES>
                <SVEXPORTFORMAT>$$SysName:XML</SVEXPORTFORMAT>
                <SVCURRENTCOMPANY>{{ company_name | e }}</SVCURRENTCOMPANY>
                <SVPARAM>{{ param1 }}</SVPARAM>
            </STATICVARIABLES>
            <TDL>
                <TDLMESSAGE>
                    <COLLECTION NAME="YourToolName">
                        <TYPE>Ledger</TYPE>
                        <FETCH>NAME, PARENT</FETCH>
                    </COLLECTION>
                </TDLMESSAGE>
            </TDL>
        </DESC>
    </BODY>
</ENVELOPE>
```

### Step 5: Create parser.yaml

Create file at `{TEMPLATES_DIR}/your_category/your_tool_name/parser.yaml`:
```yaml
type: list
items_xpath: //COLLECTION/LEDGER
result_key: results
fields:
  name:
    xpath: NAME
  parent:
    xpath: PARENT
```

### Step 6: Restart Claude Desktop

1. Close Claude Desktop completely
2. Reopen Claude Desktop
3. The new tool will be available in the tools list

**Note:** Use the `Write` tool to create files. Always create all three required files in a single operation.

## Updating Existing Tools

To modify an existing tool's behavior, parameters, or response parsing:

### Step 1: Locate the Tool

Find the tool directory:
```
{TEMPLATES_DIR}/category/tool_name/
```

For example: `~/.config/Claude/extensions/local.mcpb.taxor-ai.tally-mcp/server/templates/ledger/get_ledgers/`

### Step 2: Update tool.yaml (if changing metadata or parameters)

Edit the tool's `tool.yaml` to:
- Change the `description`
- Add or remove parameters from `input_schema.properties`
- Adjust `required` parameters
- Modify `implicit_params`

**Example:** Adding a new optional parameter to `get_ledgers`:
```yaml
name: get_ledgers
description: Get ledgers with optional filtering
implicit_params:
  - company_name
input_schema:
  properties:
    parent_ledger:                    # New parameter
      type: string
      description: Filter by parent ledger (optional)
  required: []
```

### Step 3: Update request.xml (if changing the Tally request)

Edit the template to:
- Use new template variables: `{{ parent_ledger }}`
- Add filters: `<FILTER>PARENT = {{ parent_ledger | e }}</FILTER>`
- Change TDL syntax or FETCH fields
- Adjust COLLECTION structure

**Example:** Adding a parent ledger filter:
```xml
<STATICVARIABLES>
    <SVEXPORTFORMAT>$$SysName:XML</SVEXPORTFORMAT>
    <SVCURRENTCOMPANY>{{ company_name | e }}</SVCURRENTCOMPANY>
    <SVPARENT>{{ parent_ledger | e }}</SVPARENT>
</STATICVARIABLES>
<TDL>
    <TDLMESSAGE>
        <COLLECTION NAME="Ledgers">
            <TYPE>Ledger</TYPE>
            <FILTER>PARENT = $SVPARENT</FILTER>
            <FETCH>NAME, PARENT, CLOSINGBALANCE</FETCH>
        </COLLECTION>
    </TDLMESSAGE>
</TDL>
```

### Step 4: Update parser.yaml (if changing response structure)

Modify the parser to:
- Add new fields to extract from response
- Change field transformations
- Adjust XPath expressions
- Rename `result_key` if needed

**Example:** Adding balance extraction with number transformation:
```yaml
type: list
items_xpath: //COLLECTION/LEDGER
result_key: ledgers
fields:
  name:
    xpath: NAME
  parent:
    xpath: PARENT
  closing_balance:               # New field
    xpath: CLOSINGBALANCE
    transform: number            # Convert string to number
  opening_balance:               # New field
    xpath: OPENINGBALANCE
    transform: number
```

### Step 5: Restart Claude Desktop

1. Close Claude Desktop completely
2. Reopen Claude Desktop
3. The updated tool will be available with changes applied

**Note:** All template files must remain valid YAML/XML. Invalid syntax will cause the tool to fail on restart.

### User Instructions for Updating

If using file explorer manually:
1. Navigate to the tool directory (see Step 1)
2. Edit the three files in a text editor
3. Save changes
4. Restart Claude Desktop

If using Claude Code/Copilot:
1. Open the templates folder
2. Edit each file using the editor
3. Save files
4. Restart Claude Desktop

## Need Help?

Refer to:
- Tally's TDL documentation
- XPath specification (for parser.yaml)
- Existing tool examples in `templates/`

For server issues, check logs at `~/.config/tally-mcp/logs/` (if configured).
