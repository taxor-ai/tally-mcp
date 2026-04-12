# Template-Driven Tool Registry Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace hardcoded Go tool definitions and parsers with a file-based registry where each tool is a folder containing `request.xml` (Tally XML template), `tool.yaml` (MCP tool metadata + input schema), and `parser.yaml` (XPath-based response mapping) — so new tools can be added by dropping files, with zero Go code changes.

**Architecture:** A `Registry` loaded at startup scans the embedded `templates/` directory for `tool.yaml` files, reads sibling `request.xml` and `parser.yaml` files, and builds a map of `ToolDefinition`. The MCP handler does a single generic dispatch: render template → POST to Tally → apply XPath parser → return JSON. The `antchfx/xmlquery` library handles XPath evaluation; `gopkg.in/yaml.v3` parses the YAML metadata files.

**Tech Stack:** Go 1.25, `gopkg.in/yaml.v3`, `github.com/antchfx/xmlquery`, Go `embed.FS`, existing `go.uber.org/zap` logger

---

## File Map

### New files
| File | Responsibility |
|------|---------------|
| `pkg/tally/registry.go` | `ToolDefinition`, `Registry`, `LoadRegistry()` — scans embed.FS for tool.yaml files |
| `pkg/tally/parser.go` | `ParseResponse()` — generic XPath-based parser for list/object/import_result/raw |
| `pkg/tally/templates/company/get_companies/tool.yaml` | Tool metadata for get_companies |
| `pkg/tally/templates/company/get_companies/request.xml` | Tally XML request (moved) |
| `pkg/tally/templates/company/get_companies/parser.yaml` | XPath field mapping |
| `pkg/tally/templates/ledger/get_ledgers/tool.yaml` | Tool metadata |
| `pkg/tally/templates/ledger/get_ledgers/request.xml` | Tally XML request (moved) |
| `pkg/tally/templates/ledger/get_ledgers/parser.yaml` | XPath field mapping |
| `pkg/tally/templates/ledger/get_ledger_details/tool.yaml` | Tool metadata |
| `pkg/tally/templates/ledger/get_ledger_details/request.xml` | Tally XML request (moved) |
| `pkg/tally/templates/ledger/get_ledger_details/parser.yaml` | XPath field mapping |
| `pkg/tally/templates/ledger/create_ledger/tool.yaml` | Tool metadata |
| `pkg/tally/templates/ledger/create_ledger/request.xml` | Tally XML request (moved) |
| `pkg/tally/templates/ledger/create_ledger/parser.yaml` | import_result parser |
| `pkg/tally/templates/debtor_creditor/get_debtors/tool.yaml` | Tool metadata |
| `pkg/tally/templates/debtor_creditor/get_debtors/request.xml` | Tally XML request (moved) |
| `pkg/tally/templates/debtor_creditor/get_debtors/parser.yaml` | XPath field mapping |
| `pkg/tally/templates/debtor_creditor/get_creditors/tool.yaml` | Tool metadata |
| `pkg/tally/templates/debtor_creditor/get_creditors/request.xml` | Tally XML request (moved) |
| `pkg/tally/templates/debtor_creditor/get_creditors/parser.yaml` | XPath field mapping |
| `pkg/tally/templates/voucher/get_vouchers/tool.yaml` | Tool metadata |
| `pkg/tally/templates/voucher/get_vouchers/request.xml` | Tally XML request (moved) |
| `pkg/tally/templates/voucher/get_vouchers/parser.yaml` | XPath field mapping |
| `pkg/tally/templates/voucher/create_voucher/tool.yaml` | Tool metadata |
| `pkg/tally/templates/voucher/create_voucher/request.xml` | Tally XML request (moved) |
| `pkg/tally/templates/voucher/create_voucher/parser.yaml` | import_result parser |

### Modified files
| File | Change |
|------|--------|
| `pkg/tally/templates.go` | Change embed directive to `//go:embed templates`; export `EmbeddedFS`; keep `RenderTemplate()` as standalone func; remove `LoadTemplate()` |
| `pkg/tally/client.go` | Add `ExecuteXML(xmlContent string) ([]byte, error)`; remove all `Parse*Response` functions |
| `pkg/tally/models.go` | Remove response types (Company, Ledger, Debtor, Creditor, Voucher); keep only request types (CreateLedgerRequest, CreateVoucherRequest, LineItem) — these are removed in Task 8 |
| `pkg/mcp/handler.go` | Replace switch+typed handlers with single generic `HandleToolCall` using registry |
| `pkg/mcp/tools.go` | `AllTools(registry)` reads from registry instead of hardcoded list; remove all hardcoded tool funcs |
| `main.go` | Create registry via `tally.LoadRegistry(tally.EmbeddedFS)`, pass to `mcp.NewHandler` |
| `tests/integration/main_integration_test.go` | Update type assertions from `[]tally.Company` → `[]map[string]interface{}`; create registry in test setup |

### Deleted files (old flat templates — after new ones are created)
- `pkg/tally/templates/company/get_companies.xml`
- `pkg/tally/templates/ledger/get_ledgers.xml`
- `pkg/tally/templates/ledger/get_ledger_details.xml`
- `pkg/tally/templates/ledger/create_ledger.xml`
- `pkg/tally/templates/debtor_creditor/get_debtors.xml`
- `pkg/tally/templates/debtor_creditor/get_creditors.xml`
- `pkg/tally/templates/voucher/get_vouchers.xml`
- `pkg/tally/templates/voucher/create_voucher.xml`

---

## Task 1: Add Dependencies

**Files:**
- Modify: `go.mod`, `go.sum`

- [ ] **Step 1: Add yaml and xmlquery dependencies**

```bash
cd /Users/sree/Projects/Branding/tally-mcp
go get gopkg.in/yaml.v3
go get github.com/antchfx/xmlquery
go mod tidy
```

Expected output: `go.mod` updated with two new `require` entries, `go.sum` updated.

- [ ] **Step 2: Verify build still compiles**

```bash
go build ./...
```

Expected: no errors (existing code unchanged).

- [ ] **Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "deps: add yaml.v3 and antchfx/xmlquery for template-driven registry"
```

---

## Task 2: Create `registry.go` — Tool Definition Loading

**Files:**
- Create: `pkg/tally/registry.go`

- [ ] **Step 1: Write a failing unit test for registry loading**

Create `pkg/tally/registry_test.go`:

```go
package tally_test

import (
    "testing"
    "github.com/taxor-ai/tally-mcp/pkg/tally"
)

func TestRegistryLoadsBuiltinTools(t *testing.T) {
    registry, err := tally.LoadRegistry(tally.EmbeddedFS)
    if err != nil {
        t.Fatalf("LoadRegistry failed: %v", err)
    }
    tools := registry.All()
    if len(tools) == 0 {
        t.Fatal("expected built-in tools to be loaded")
    }
}

func TestRegistryGetUnknownTool(t *testing.T) {
    registry, err := tally.LoadRegistry(tally.EmbeddedFS)
    if err != nil {
        t.Fatalf("LoadRegistry failed: %v", err)
    }
    if registry.Get("nonexistent_tool") != nil {
        t.Fatal("expected nil for unknown tool")
    }
}
```

- [ ] **Step 2: Run test to confirm it fails**

```bash
cd /Users/sree/Projects/Branding/tally-mcp
go test ./pkg/tally/... -run TestRegistry -v
```

Expected: compile error — `tally.LoadRegistry` and `tally.EmbeddedFS` undefined.

- [ ] **Step 3: Create `pkg/tally/registry.go`**

```go
package tally

import (
    "embed"
    "fmt"
    "io/fs"
    "strings"

    "gopkg.in/yaml.v3"
)

// FieldSpec describes how to extract one field from XML via XPath.
type FieldSpec struct {
    XPath     string `yaml:"xpath"`
    Transform string `yaml:"transform,omitempty"` // number, integer, boolean, tally_date
}

// ParserSpec describes how to parse a Tally XML response.
// type values: "list", "object", "import_result", "raw"
type ParserSpec struct {
    Type       string               `yaml:"type"`
    ItemsXPath string               `yaml:"items_xpath,omitempty"` // for list: XPath to each item node
    RootXPath  string               `yaml:"root_xpath,omitempty"`  // for object: XPath to root node
    ResultKey  string               `yaml:"result_key,omitempty"`  // key used in response map
    Fields     map[string]FieldSpec `yaml:"fields,omitempty"`
}

// InputProperty is one parameter in a tool's input schema.
type InputProperty struct {
    Type        string   `yaml:"type"`
    Description string   `yaml:"description"`
    Enum        []string `yaml:"enum,omitempty"`
}

// InputSchema mirrors the JSON Schema object used in MCP's tools/list response.
type InputSchema struct {
    Properties map[string]InputProperty `yaml:"properties"`
    Required   []string                 `yaml:"required,omitempty"`
}

// ToMap converts InputSchema to map[string]interface{} for MCP protocol serialisation.
func (s InputSchema) ToMap() map[string]interface{} {
    props := make(map[string]interface{})
    for name, prop := range s.Properties {
        p := map[string]interface{}{
            "type":        prop.Type,
            "description": prop.Description,
        }
        if len(prop.Enum) > 0 {
            p["enum"] = prop.Enum
        }
        props[name] = p
    }
    result := map[string]interface{}{
        "type":       "object",
        "properties": props,
    }
    if len(s.Required) > 0 {
        result["required"] = s.Required
    }
    return result
}

// ToolDefinition is a fully loaded tool: metadata + XML template + parser spec.
type ToolDefinition struct {
    Name           string      `yaml:"name"`
    Description    string      `yaml:"description"`
    ImplicitParams []string    `yaml:"implicit_params,omitempty"`
    InputSchema    InputSchema `yaml:"input_schema"`
    RequestXML     string      `yaml:"-"` // loaded from request.xml
    Parser         ParserSpec  `yaml:"-"` // loaded from parser.yaml
}

// Registry holds all registered ToolDefinitions, keyed by tool name.
type Registry struct {
    tools map[string]*ToolDefinition
    order []string // preserves registration order for tools/list
}

func newRegistry() *Registry {
    return &Registry{tools: make(map[string]*ToolDefinition)}
}

// Get returns the ToolDefinition for the given name, or nil if not found.
func (r *Registry) Get(name string) *ToolDefinition {
    return r.tools[name]
}

// All returns all ToolDefinitions in registration order.
func (r *Registry) All() []*ToolDefinition {
    out := make([]*ToolDefinition, 0, len(r.order))
    for _, name := range r.order {
        if def, ok := r.tools[name]; ok {
            out = append(out, def)
        }
    }
    return out
}

// LoadRegistry scans the embedded FS for tool.yaml files and builds a Registry.
// Each tool lives in its own subdirectory alongside request.xml and parser.yaml.
func LoadRegistry(embeddedFS embed.FS) (*Registry, error) {
    reg := newRegistry()
    err := fs.WalkDir(embeddedFS, "templates", func(path string, d fs.DirEntry, walkErr error) error {
        if walkErr != nil {
            return walkErr
        }
        if d.IsDir() || !strings.HasSuffix(path, "/tool.yaml") {
            return nil
        }

        // dir is e.g. "templates/company/get_companies"
        lastSlash := strings.LastIndex(path, "/")
        dir := path[:lastSlash]

        // Parse tool.yaml
        toolData, err := embeddedFS.ReadFile(path)
        if err != nil {
            return fmt.Errorf("reading %s: %w", path, err)
        }
        var def ToolDefinition
        if err := yaml.Unmarshal(toolData, &def); err != nil {
            return fmt.Errorf("parsing %s: %w", path, err)
        }

        // Load request.xml
        requestPath := dir + "/request.xml"
        requestData, err := embeddedFS.ReadFile(requestPath)
        if err != nil {
            return fmt.Errorf("reading %s: %w", requestPath, err)
        }
        def.RequestXML = string(requestData)

        // Load parser.yaml (optional; default to raw)
        parserPath := dir + "/parser.yaml"
        if parserData, err := embeddedFS.ReadFile(parserPath); err == nil {
            if err := yaml.Unmarshal(parserData, &def.Parser); err != nil {
                return fmt.Errorf("parsing %s: %w", parserPath, err)
            }
        } else {
            def.Parser = ParserSpec{Type: "raw"}
        }

        reg.tools[def.Name] = &def
        reg.order = append(reg.order, def.Name)
        return nil
    })
    if err != nil {
        return nil, err
    }
    return reg, nil
}
```

- [ ] **Step 4: Update `pkg/tally/templates.go` to export `EmbeddedFS`**

Replace the contents of `pkg/tally/templates.go`:

```go
package tally

import (
    "embed"
    "fmt"
    "strings"
)

//go:embed templates
var EmbeddedFS embed.FS

// RenderTemplate fills {{ key | e }} and {{key}} placeholders in a template string.
func RenderTemplate(templateStr string, params map[string]string) string {
    result := templateStr
    for key, value := range params {
        result = strings.ReplaceAll(result, fmt.Sprintf("{{ %s | e }}", key), value)
        result = strings.ReplaceAll(result, fmt.Sprintf("{{%s}}", key), value)
    }
    return result
}
```

- [ ] **Step 5: Run tests to confirm they pass**

```bash
go test ./pkg/tally/... -run TestRegistry -v
```

Expected:
```
--- PASS: TestRegistryLoadsBuiltinTools (0.00s)
    registry_test.go:10: expected built-in tools to be loaded
```

Wait — the test will fail because the template subdirectories don't exist yet. The test *should* pass once templates are created in Task 3. For now, verify only that the code **compiles** without error:

```bash
go build ./pkg/tally/...
```

Expected: compiles without error (registry loads 0 tools from empty new structure, which is OK until Task 3).

- [ ] **Step 6: Commit**

```bash
git add pkg/tally/registry.go pkg/tally/registry_test.go pkg/tally/templates.go
git commit -m "feat: add Registry and ToolDefinition types with embed.FS loader"
```

---

## Task 3: Create `parser.go` — Generic XPath Parser

**Files:**
- Create: `pkg/tally/parser.go`
- Create: `pkg/tally/parser_test.go`

- [ ] **Step 1: Write failing unit tests**

Create `pkg/tally/parser_test.go`:

```go
package tally_test

import (
    "testing"
    "github.com/taxor-ai/tally-mcp/pkg/tally"
)

var sampleCompaniesXML = []byte(`<ENVELOPE>
  <BODY><DATA><COLLECTION>
    <COMPANY NAME="Alpha Corp"><GUID>guid-001</GUID></COMPANY>
    <COMPANY NAME="Beta Ltd"><GUID>guid-002</GUID></COMPANY>
  </COLLECTION></DATA></BODY>
</ENVELOPE>`)

func TestParseList(t *testing.T) {
    spec := tally.ParserSpec{
        Type:       "list",
        ItemsXPath: "//COLLECTION/COMPANY",
        ResultKey:  "companies",
        Fields: map[string]tally.FieldSpec{
            "name": {XPath: "@NAME"},
            "guid": {XPath: "GUID"},
        },
    }
    result, err := tally.ParseResponse(sampleCompaniesXML, spec)
    if err != nil {
        t.Fatalf("ParseResponse error: %v", err)
    }
    items, ok := result["companies"].([]map[string]interface{})
    if !ok {
        t.Fatalf("expected []map[string]interface{}, got %T", result["companies"])
    }
    if len(items) != 2 {
        t.Fatalf("expected 2 items, got %d", len(items))
    }
    if items[0]["name"] != "Alpha Corp" {
        t.Errorf("expected name=Alpha Corp, got %v", items[0]["name"])
    }
    if items[0]["guid"] != "guid-001" {
        t.Errorf("expected guid=guid-001, got %v", items[0]["guid"])
    }
}

var sampleLedgerXML = []byte(`<ENVELOPE>
  <BODY><DATA><COLLECTION>
    <LEDGER NAME="Cash">
      <PARENT>Current Assets</PARENT>
      <CLOSINGBALANCE>5000.00</CLOSINGBALANCE>
    </LEDGER>
  </COLLECTION></DATA></BODY>
</ENVELOPE>`)

func TestParseObject(t *testing.T) {
    spec := tally.ParserSpec{
        Type:      "object",
        RootXPath: "//COLLECTION/LEDGER[1]",
        ResultKey: "ledger",
        Fields: map[string]tally.FieldSpec{
            "name":    {XPath: "@NAME"},
            "parent":  {XPath: "PARENT"},
            "balance": {XPath: "CLOSINGBALANCE", Transform: "number"},
        },
    }
    result, err := tally.ParseResponse(sampleLedgerXML, spec)
    if err != nil {
        t.Fatalf("ParseResponse error: %v", err)
    }
    ledger, ok := result["ledger"].(map[string]interface{})
    if !ok {
        t.Fatalf("expected map, got %T", result["ledger"])
    }
    if ledger["name"] != "Cash" {
        t.Errorf("expected name=Cash, got %v", ledger["name"])
    }
    if ledger["balance"] != 5000.0 {
        t.Errorf("expected balance=5000.0, got %v", ledger["balance"])
    }
}

var sampleImportResultXML = []byte(`<ENVELOPE>
  <BODY><DATA><IMPORTRESULT>
    <CREATED>1</CREATED><ALTERED>0</ALTERED><DELETED>0</DELETED>
  </IMPORTRESULT></DATA></BODY>
</ENVELOPE>`)

func TestParseImportResult(t *testing.T) {
    spec := tally.ParserSpec{Type: "import_result"}
    result, err := tally.ParseResponse(sampleImportResultXML, spec)
    if err != nil {
        t.Fatalf("ParseResponse error: %v", err)
    }
    if result["success"] != true {
        t.Errorf("expected success=true, got %v", result["success"])
    }
    if result["created"] != 1 {
        t.Errorf("expected created=1, got %v", result["created"])
    }
}

func TestTransformTallyDate(t *testing.T) {
    spec := tally.ParserSpec{
        Type:       "list",
        ItemsXPath: "//VOUCHER",
        Fields: map[string]tally.FieldSpec{
            "date": {XPath: "DATE", Transform: "tally_date"},
        },
    }
    xml := []byte(`<ENVELOPE><BODY><DATA><COLLECTION>
        <VOUCHER><DATE>20240401</DATE></VOUCHER>
    </COLLECTION></DATA></BODY></ENVELOPE>`)
    result, err := tally.ParseResponse(xml, spec)
    if err != nil {
        t.Fatalf("ParseResponse error: %v", err)
    }
    items := result["items"].([]map[string]interface{})
    if items[0]["date"] != "2024-04-01" {
        t.Errorf("expected 2024-04-01, got %v", items[0]["date"])
    }
}
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
go test ./pkg/tally/... -run "TestParseList|TestParseObject|TestParseImportResult|TestTransformTallyDate" -v
```

Expected: compile error — `tally.ParseResponse` undefined.

- [ ] **Step 3: Create `pkg/tally/parser.go`**

```go
package tally

import (
    "fmt"
    "strconv"
    "strings"

    "github.com/antchfx/xmlquery"
)

// ParseResponse parses raw Tally XML using the given ParserSpec and returns
// a map[string]interface{} ready to serialise as the MCP tool result.
func ParseResponse(xmlData []byte, spec ParserSpec) (map[string]interface{}, error) {
    switch spec.Type {
    case "list":
        return parseList(xmlData, spec)
    case "object":
        return parseObject(xmlData, spec)
    case "import_result":
        return parseImportResult(xmlData)
    default: // "raw" or unrecognised
        return map[string]interface{}{
            "success": true,
            "data":    string(xmlData),
        }, nil
    }
}

func parseList(xmlData []byte, spec ParserSpec) (map[string]interface{}, error) {
    doc, err := xmlquery.Parse(strings.NewReader(string(xmlData)))
    if err != nil {
        return nil, fmt.Errorf("parse XML: %w", err)
    }
    nodes, err := xmlquery.QueryAll(doc, spec.ItemsXPath)
    if err != nil {
        return nil, fmt.Errorf("xpath %q: %w", spec.ItemsXPath, err)
    }
    items := make([]map[string]interface{}, 0, len(nodes))
    for _, node := range nodes {
        items = append(items, extractFields(node, spec.Fields))
    }
    key := spec.ResultKey
    if key == "" {
        key = "items"
    }
    return map[string]interface{}{
        "success": true,
        key:       items,
        "count":   len(items),
    }, nil
}

func parseObject(xmlData []byte, spec ParserSpec) (map[string]interface{}, error) {
    doc, err := xmlquery.Parse(strings.NewReader(string(xmlData)))
    if err != nil {
        return nil, fmt.Errorf("parse XML: %w", err)
    }
    xpath := spec.RootXPath
    if xpath == "" {
        xpath = "/*"
    }
    node, err := xmlquery.Query(doc, xpath)
    if err != nil || node == nil {
        return nil, fmt.Errorf("root node not found at %q", xpath)
    }
    key := spec.ResultKey
    if key == "" {
        key = "data"
    }
    return map[string]interface{}{
        "success": true,
        key:       extractFields(node, spec.Fields),
    }, nil
}

func parseImportResult(xmlData []byte) (map[string]interface{}, error) {
    doc, err := xmlquery.Parse(strings.NewReader(string(xmlData)))
    if err != nil {
        return nil, fmt.Errorf("parse XML: %w", err)
    }
    created := nodeInt(doc, "//IMPORTRESULT/CREATED")
    altered := nodeInt(doc, "//IMPORTRESULT/ALTERED")
    errMsg := nodeText(doc, "//IMPORTRESULT/LINEERROR")
    success := created > 0 || altered > 0
    result := map[string]interface{}{
        "success": success,
        "created": created,
        "altered": altered,
    }
    if errMsg != "" {
        result["error"] = errMsg
        result["success"] = false
    }
    return result, nil
}

func extractFields(node *xmlquery.Node, fields map[string]FieldSpec) map[string]interface{} {
    item := make(map[string]interface{}, len(fields))
    for fieldName, spec := range fields {
        var raw string
        if strings.HasPrefix(spec.XPath, "@") {
            raw = node.SelectAttr(spec.XPath[1:])
        } else {
            if child, _ := xmlquery.Query(node, spec.XPath); child != nil {
                raw = strings.TrimSpace(child.InnerText())
            }
        }
        item[fieldName] = applyTransform(raw, spec.Transform)
    }
    return item
}

func applyTransform(val, transform string) interface{} {
    val = strings.TrimSpace(val)
    switch transform {
    case "number":
        if f, err := strconv.ParseFloat(val, 64); err == nil {
            return f
        }
        return 0.0
    case "integer":
        if i, err := strconv.Atoi(val); err == nil {
            return i
        }
        return 0
    case "boolean":
        return strings.EqualFold(val, "yes") || val == "true" || val == "1"
    case "tally_date":
        if len(val) == 8 {
            return val[:4] + "-" + val[4:6] + "-" + val[6:]
        }
        return val
    default:
        return val
    }
}

// helpers for parseImportResult
func nodeText(doc *xmlquery.Node, xpath string) string {
    if n, _ := xmlquery.Query(doc, xpath); n != nil {
        return strings.TrimSpace(n.InnerText())
    }
    return ""
}

func nodeInt(doc *xmlquery.Node, xpath string) int {
    if i, err := strconv.Atoi(nodeText(doc, xpath)); err == nil {
        return i
    }
    return 0
}
```

- [ ] **Step 4: Run tests to confirm they pass**

```bash
go test ./pkg/tally/... -run "TestParseList|TestParseObject|TestParseImportResult|TestTransformTallyDate" -v
```

Expected: all 4 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add pkg/tally/parser.go pkg/tally/parser_test.go
git commit -m "feat: generic XPath parser — list, object, import_result, tally_date transform"
```

---

## Task 4: Create Template Directories for All 8 Tools

**Files:**
- Create all `request.xml`, `tool.yaml`, `parser.yaml` files under new subdirectory layout

> Note: Old flat `.xml` files are NOT deleted yet — they are removed in Task 5 after the build is confirmed.

- [ ] **Step 1: Create `company/get_companies/` files**

`pkg/tally/templates/company/get_companies/request.xml` — copy content from existing `company/get_companies.xml`:
```xml
<ENVELOPE>
    <HEADER>
        <VERSION>1</VERSION>
        <TALLYREQUEST>Export</TALLYREQUEST>
        <TYPE>Collection</TYPE>
        <ID>AllCompanies</ID>
    </HEADER>
    <BODY>
        <DESC>
            <STATICVARIABLES>
                <SVEXPORTFORMAT>$$SysName:XML</SVEXPORTFORMAT>
            </STATICVARIABLES>
            <TDL>
                <TDLMESSAGE>
                    <COLLECTION NAME="AllCompanies">
                        <TYPE>Company</TYPE>
                        <FETCH>NAME, GUID</FETCH>
                    </COLLECTION>
                </TDLMESSAGE>
            </TDL>
        </DESC>
    </BODY>
</ENVELOPE>
```

`pkg/tally/templates/company/get_companies/tool.yaml`:
```yaml
name: get_companies
description: List all companies available in Tally
implicit_params: []
input_schema:
  properties: {}
  required: []
```

`pkg/tally/templates/company/get_companies/parser.yaml`:
```yaml
type: list
items_xpath: //COLLECTION/COMPANY
result_key: companies
fields:
  name:
    xpath: "@NAME"
  guid:
    xpath: GUID
```

- [ ] **Step 2: Create `ledger/get_ledgers/` files**

`pkg/tally/templates/ledger/get_ledgers/request.xml` — copy from `ledger/get_ledgers.xml`:
```xml
<ENVELOPE>
    <HEADER>
        <VERSION>1</VERSION>
        <TALLYREQUEST>Export</TALLYREQUEST>
        <TYPE>Collection</TYPE>
        <ID>AllAccounts</ID>
    </HEADER>
    <BODY>
        <DESC>
            <STATICVARIABLES>
                <SVEXPORTFORMAT>$$SysName:XML</SVEXPORTFORMAT>
                <SVCURRENTCOMPANY>{{ company_name | e }}</SVCURRENTCOMPANY>
            </STATICVARIABLES>
            <TDL>
                <TDLMESSAGE>
                    <COLLECTION NAME="AllAccounts">
                        <TYPE>Ledger</TYPE>
                        <FETCH>NAME, PARENT, $ISAGGREGATE, $RESERVEDNAME</FETCH>
                    </COLLECTION>
                </TDLMESSAGE>
            </TDL>
        </DESC>
    </BODY>
</ENVELOPE>
```

`pkg/tally/templates/ledger/get_ledgers/tool.yaml`:
```yaml
name: get_ledgers
description: List all ledgers in Tally, optionally filtered by type
implicit_params:
  - company_name
input_schema:
  properties:
    filter_type:
      type: string
      description: "Filter ledgers by type: asset, liability, income, expense, debtor, creditor, or all"
      enum: [asset, liability, income, expense, debtor, creditor, all]
  required: []
```

`pkg/tally/templates/ledger/get_ledgers/parser.yaml`:
```yaml
type: list
items_xpath: //COLLECTION/LEDGER
result_key: ledgers
fields:
  name:
    xpath: "@NAME"
  parent:
    xpath: PARENT
```

- [ ] **Step 3: Create `ledger/get_ledger_details/` files**

`pkg/tally/templates/ledger/get_ledger_details/request.xml` — copy from `ledger/get_ledger_details.xml`:
```xml
<ENVELOPE>
    <HEADER>
        <VERSION>1</VERSION>
        <TALLYREQUEST>Export</TALLYREQUEST>
        <TYPE>Collection</TYPE>
        <ID>TargetAccount</ID>
    </HEADER>
    <BODY>
        <DESC>
            <STATICVARIABLES>
                <SVEXPORTFORMAT>$$SysName:XML</SVEXPORTFORMAT>
                <SVCURRENTCOMPANY>{{ company_name | e }}</SVCURRENTCOMPANY>
            </STATICVARIABLES>
            <TDL>
                <TDLMESSAGE>
                    <COLLECTION NAME="TargetAccount">
                        <TYPE>Ledger</TYPE>
                        <FETCH>NAME, PARENT, $ISAGGREGATE, $RESERVEDNAME, PARTYGSTIN, GSTRREGISTRATIONTYPE, STATENAME, INCOMEPAN, EMAIL, LEDGERPHONE, LEDGERMOBILE, ADDRESS, PINCODE, MAILINGNAME.LIST, ADDRESS.LIST</FETCH>
                        <FILTER>ByName</FILTER>
                    </COLLECTION>
                    <SYSTEM TYPE="Formula" NAME="ByName">($NAME = "{{ ledger_name | e }}")</SYSTEM>
                </TDLMESSAGE>
            </TDL>
        </DESC>
    </BODY>
</ENVELOPE>
```

Note: Template param is `ledger_name` (fixing bug in old template which used `name`).

`pkg/tally/templates/ledger/get_ledger_details/tool.yaml`:
```yaml
name: get_ledger_details
description: Get detailed information for a specific ledger including GST and contact details
implicit_params:
  - company_name
input_schema:
  properties:
    ledger_name:
      type: string
      description: Name of the ledger to query
  required: [ledger_name]
```

`pkg/tally/templates/ledger/get_ledger_details/parser.yaml`:
```yaml
type: object
root_xpath: //COLLECTION/LEDGER[1]
result_key: ledger
fields:
  name:
    xpath: "@NAME"
  parent:
    xpath: PARENT
  gstin:
    xpath: PARTYGSTIN
  gst_registration_type:
    xpath: GSTRREGISTRATIONTYPE
  state:
    xpath: STATENAME
  pan:
    xpath: INCOMEPAN
  email:
    xpath: EMAIL
  phone:
    xpath: LEDGERPHONE
  mobile:
    xpath: LEDGERMOBILE
  pincode:
    xpath: PINCODE
```

- [ ] **Step 4: Create `ledger/create_ledger/` files**

`pkg/tally/templates/ledger/create_ledger/request.xml` — copy from `ledger/create_ledger.xml`:
```xml
<ENVELOPE>
    <HEADER>
        <TALLYREQUEST>Import Data</TALLYREQUEST>
    </HEADER>
    <BODY>
        <IMPORTDATA>
            <REQUESTDESC>
                <REPORTNAME>All Masters</REPORTNAME>
                <STATICVARIABLES>
                    <SVCURRENTCOMPANY>{{ company_name | e }}</SVCURRENTCOMPANY>
                </STATICVARIABLES>
            </REQUESTDESC>
            <REQUESTDATA>
                <TALLYMESSAGE xmlns:UDF="TallyUDF">
                    <LEDGER NAME="{{ ledger_name | e }}" ACTION="Create">
                        <NAME.LIST>
                            <NAME>{{ ledger_name | e }}</NAME>
                        </NAME.LIST>
                        <PARENT>{{ parent | e }}</PARENT>
                    </LEDGER>
                </TALLYMESSAGE>
            </REQUESTDATA>
        </IMPORTDATA>
    </BODY>
</ENVELOPE>
```

`pkg/tally/templates/ledger/create_ledger/tool.yaml`:
```yaml
name: create_ledger
description: Create a new ledger account in Tally
implicit_params:
  - company_name
input_schema:
  properties:
    ledger_name:
      type: string
      description: Ledger name
    parent:
      type: string
      description: Parent group name (e.g. Sundry Debtors, Direct Expenses)
  required: [ledger_name, parent]
```

`pkg/tally/templates/ledger/create_ledger/parser.yaml`:
```yaml
type: import_result
```

- [ ] **Step 5: Create `debtor_creditor/get_debtors/` files**

`pkg/tally/templates/debtor_creditor/get_debtors/request.xml` — copy from `debtor_creditor/get_debtors.xml`:
```xml
<ENVELOPE>
    <HEADER>
        <VERSION>1</VERSION>
        <TALLYREQUEST>Export</TALLYREQUEST>
        <TYPE>Collection</TYPE>
        <ID>DebtorCollection</ID>
    </HEADER>
    <BODY>
        <DESC>
            <STATICVARIABLES>
                <SVEXPORTFORMAT>$$SysName:XML</SVEXPORTFORMAT>
                <SVCURRENTCOMPANY>{{ company_name | e }}</SVCURRENTCOMPANY>
            </STATICVARIABLES>
            <TDL>
                <TDLMESSAGE>
                    <COLLECTION NAME="DebtorCollection">
                        <TYPE>Ledger</TYPE>
                        <FETCH>NAME, PARENT, OPENINGBALANCE, CLOSINGBALANCE, CREDITLIMIT, STATENAME, EMAIL</FETCH>
                        <FILTER>IsSundryDebtor</FILTER>
                    </COLLECTION>
                    <SYSTEM TYPE="Formula" NAME="IsSundryDebtor">($Parent = "Sundry Debtors")</SYSTEM>
                </TDLMESSAGE>
            </TDL>
        </DESC>
    </BODY>
</ENVELOPE>
```

`pkg/tally/templates/debtor_creditor/get_debtors/tool.yaml`:
```yaml
name: get_debtors
description: List all debtor ledgers (Sundry Debtors) with balances
implicit_params:
  - company_name
input_schema:
  properties: {}
  required: []
```

`pkg/tally/templates/debtor_creditor/get_debtors/parser.yaml`:
```yaml
type: list
items_xpath: //COLLECTION/LEDGER
result_key: debtors
fields:
  name:
    xpath: "@NAME"
  parent:
    xpath: PARENT
  closing_balance:
    xpath: CLOSINGBALANCE
    transform: number
  credit_limit:
    xpath: CREDITLIMIT
    transform: number
  state:
    xpath: STATENAME
  email:
    xpath: EMAIL
```

- [ ] **Step 6: Create `debtor_creditor/get_creditors/` files**

`pkg/tally/templates/debtor_creditor/get_creditors/request.xml` — copy from `debtor_creditor/get_creditors.xml`:
```xml
<ENVELOPE>
    <HEADER>
        <VERSION>1</VERSION>
        <TALLYREQUEST>Export</TALLYREQUEST>
        <TYPE>Collection</TYPE>
        <ID>CreditorCollection</ID>
    </HEADER>
    <BODY>
        <DESC>
            <STATICVARIABLES>
                <SVEXPORTFORMAT>$$SysName:XML</SVEXPORTFORMAT>
                <SVCURRENTCOMPANY>{{ company_name | e }}</SVCURRENTCOMPANY>
            </STATICVARIABLES>
            <TDL>
                <TDLMESSAGE>
                    <COLLECTION NAME="CreditorCollection">
                        <TYPE>Ledger</TYPE>
                        <FETCH>NAME, PARENT, OPENINGBALANCE, CLOSINGBALANCE, CREDITLIMIT, STATENAME, EMAIL</FETCH>
                        <FILTER>IsSundryCreditor</FILTER>
                    </COLLECTION>
                    <SYSTEM TYPE="Formula" NAME="IsSundryCreditor">($Parent = "Sundry Creditors")</SYSTEM>
                </TDLMESSAGE>
            </TDL>
        </DESC>
    </BODY>
</ENVELOPE>
```

`pkg/tally/templates/debtor_creditor/get_creditors/tool.yaml`:
```yaml
name: get_creditors
description: List all creditor ledgers (Sundry Creditors) with balances
implicit_params:
  - company_name
input_schema:
  properties: {}
  required: []
```

`pkg/tally/templates/debtor_creditor/get_creditors/parser.yaml`:
```yaml
type: list
items_xpath: //COLLECTION/LEDGER
result_key: creditors
fields:
  name:
    xpath: "@NAME"
  parent:
    xpath: PARENT
  closing_balance:
    xpath: CLOSINGBALANCE
    transform: number
  credit_limit:
    xpath: CREDITLIMIT
    transform: number
  state:
    xpath: STATENAME
  email:
    xpath: EMAIL
```

- [ ] **Step 7: Create `voucher/get_vouchers/` files**

`pkg/tally/templates/voucher/get_vouchers/request.xml` — copy from `voucher/get_vouchers.xml`:
```xml
<ENVELOPE>
    <HEADER>
        <VERSION>1</VERSION>
        <TALLYREQUEST>Export</TALLYREQUEST>
        <TYPE>Collection</TYPE>
        <ID>AllVouchers</ID>
    </HEADER>
    <BODY>
        <DESC>
            <STATICVARIABLES>
                <SVEXPORTFORMAT>$$SysName:XML</SVEXPORTFORMAT>
                <SVCURRENTCOMPANY>{{ company_name | e }}</SVCURRENTCOMPANY>
                <SVFROMDATE>{{ from_date }}</SVFROMDATE>
                <SVTODATE>{{ to_date }}</SVTODATE>
            </STATICVARIABLES>
            <TDL>
                <TDLMESSAGE>
                    <COLLECTION NAME="AllVouchers">
                        <TYPE>Voucher</TYPE>
                        <FETCH>DATE, VOUCHERNUMBER, REFERENCE, NARRATION, VOUCHERTYPENAME</FETCH>
                    </COLLECTION>
                </TDLMESSAGE>
            </TDL>
        </DESC>
    </BODY>
</ENVELOPE>
```

`pkg/tally/templates/voucher/get_vouchers/tool.yaml`:
```yaml
name: get_vouchers
description: List vouchers in Tally for a date range
implicit_params:
  - company_name
input_schema:
  properties:
    from_date:
      type: string
      description: Start date in YYYYMMDD format (e.g. 20240101)
    to_date:
      type: string
      description: End date in YYYYMMDD format (e.g. 20241231)
  required: []
```

`pkg/tally/templates/voucher/get_vouchers/parser.yaml`:
```yaml
type: list
items_xpath: //COLLECTION/VOUCHER
result_key: vouchers
fields:
  date:
    xpath: DATE
    transform: tally_date
  voucher_number:
    xpath: VOUCHERNUMBER
  reference:
    xpath: REFERENCE
  narration:
    xpath: NARRATION
  voucher_type:
    xpath: VOUCHERTYPENAME
```

- [ ] **Step 8: Create `voucher/create_voucher/` files**

`pkg/tally/templates/voucher/create_voucher/request.xml`:
```xml
<ENVELOPE>
    <HEADER>
        <VERSION>1</VERSION>
        <TALLYREQUEST>Import</TALLYREQUEST>
        <TYPE>Data</TYPE>
        <ID>Vouchers</ID>
    </HEADER>
    <BODY>
        <DESC>
            <STATICVARIABLES>
                <SVCURRENTCOMPANY>{{ company_name | e }}</SVCURRENTCOMPANY>
            </STATICVARIABLES>
        </DESC>
        <DATA>
            <TALLYMESSAGE xmlns:UDF="TallyUDF">
                <VOUCHER VCHTYPE="{{ voucher_type | e }}" ACTION="Create" OBJVIEW="Accounting Voucher View">
                    <DATE>{{ date }}</DATE>
                    <VOUCHERTYPENAME>{{ voucher_type | e }}</VOUCHERTYPENAME>
                    <REFERENCE>{{ reference_number | e }}</REFERENCE>
                    <NARRATION>{{ notes | e }}</NARRATION>
                </VOUCHER>
            </TALLYMESSAGE>
        </DATA>
    </BODY>
</ENVELOPE>
```

`pkg/tally/templates/voucher/create_voucher/tool.yaml`:
```yaml
name: create_voucher
description: Create a new voucher in Tally (Journal, Payment, Receipt, etc.)
implicit_params:
  - company_name
input_schema:
  properties:
    voucher_type:
      type: string
      description: "Type of voucher: Journal, Payment, Receipt, Sales, Purchase"
      enum: [Journal, Payment, Receipt, Sales, Purchase]
    date:
      type: string
      description: Voucher date in YYYYMMDD format (e.g. 20240401)
    reference_number:
      type: string
      description: Reference number (optional)
    notes:
      type: string
      description: Narration / notes (optional)
  required: [voucher_type, date]
```

`pkg/tally/templates/voucher/create_voucher/parser.yaml`:
```yaml
type: import_result
```

- [ ] **Step 9: Verify registry test now passes**

```bash
go test ./pkg/tally/... -run TestRegistryLoadsBuiltinTools -v
```

Expected:
```
--- PASS: TestRegistryLoadsBuiltinTools
```

If it fails, check that each tool directory has a `tool.yaml` AND `request.xml`.

- [ ] **Step 10: Commit all template files**

```bash
git add pkg/tally/templates/company/get_companies/ \
        pkg/tally/templates/ledger/ \
        pkg/tally/templates/debtor_creditor/ \
        pkg/tally/templates/voucher/
git commit -m "feat: add tool.yaml and parser.yaml for all 8 built-in tools"
```

---

## Task 5: Update `client.go` — Add `ExecuteXML`, Remove `Parse*` Functions

**Files:**
- Modify: `pkg/tally/client.go`

- [ ] **Step 1: Add `ExecuteXML` method and remove all `Parse*` functions**

Replace `pkg/tally/client.go` content. Keep `NewClient`, `SetCompany`, `buildRPCURL`, `Ping`, `sanitizeXML`. Remove `ExecuteTemplate`, `ParseCompaniesResponse`, `ParseLedgersResponse`, `ParseLedgerDetailsResponse`, `ParseDebtorsResponse`, `ParseCreditorsResponse`, `ParseVouchersResponse`, `ParseCreateResponse`. Add `ExecuteXML`:

```go
package tally

import (
    "bytes"
    "fmt"
    "io"
    "net/http"
    "regexp"
    "time"
)

// Client wraps Tally XML-RPC communication
type Client struct {
    Host    string
    Port    int
    Company string
    Timeout int
    http    *http.Client
}

func NewClient(host string, port int, timeoutSeconds int) *Client {
    return &Client{
        Host:    host,
        Port:    port,
        Timeout: timeoutSeconds,
        http:    &http.Client{Timeout: time.Duration(timeoutSeconds) * time.Second},
    }
}

func (c *Client) SetCompany(companyName string) {
    c.Company = companyName
}

func (c *Client) buildRPCURL() string {
    return fmt.Sprintf("http://%s:%d/", c.Host, c.Port)
}

func (c *Client) Ping() error {
    url := c.buildRPCURL()
    resp, err := c.http.Get(url)
    if err != nil {
        return NewConnectionError(fmt.Sprintf("%s:%d", c.Host, c.Port), err.Error())
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
        return NewConnectionError(fmt.Sprintf("%s:%d", c.Host, c.Port), fmt.Sprintf("HTTP %d", resp.StatusCode))
    }
    return nil
}

// ExecuteXML posts rendered XML to Tally and returns the sanitised XML response.
func (c *Client) ExecuteXML(xmlContent string) ([]byte, error) {
    url := c.buildRPCURL()
    req, err := http.NewRequest("POST", url, bytes.NewReader([]byte(xmlContent)))
    if err != nil {
        return nil, NewConnectionError(fmt.Sprintf("%s:%d", c.Host, c.Port), fmt.Sprintf("request creation failed: %v", err))
    }
    req.Header.Set("Content-Type", "text/xml; charset=utf-8")

    resp, err := c.http.Do(req)
    if err != nil {
        return nil, NewConnectionError(fmt.Sprintf("%s:%d", c.Host, c.Port), fmt.Sprintf("request failed: %v", err))
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, NewConnectionError(fmt.Sprintf("%s:%d", c.Host, c.Port), fmt.Sprintf("failed to read response: %v", err))
    }
    if resp.StatusCode != http.StatusOK {
        return nil, NewConnectionError(fmt.Sprintf("%s:%d", c.Host, c.Port), fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)))
    }
    return sanitizeXML(body), nil
}

func sanitizeXML(data []byte) []byte {
    data = bytes.Map(func(r rune) rune {
        if (r >= 0x20 && r <= 0xD7FF) ||
            r == 0x09 || r == 0x0A || r == 0x0D ||
            (r >= 0xE000 && r <= 0xFFFD) ||
            (r >= 0x10000 && r <= 0x10FFFF) {
            return r
        }
        return -1
    }, data)
    controlEntityRegex := regexp.MustCompile(`&#(x?[0-9a-fA-F]+);`)
    return controlEntityRegex.ReplaceAllFunc(data, func(match []byte) []byte {
        entity := string(match)
        var val int
        var err error
        if len(entity) > 3 && entity[2] == 'x' {
            _, err = fmt.Sscanf(entity, "&#x%x;", &val)
        } else {
            _, err = fmt.Sscanf(entity, "&#%d;", &val)
        }
        if err == nil && val < 32 && val != 9 && val != 10 && val != 13 {
            return []byte("")
        }
        return match
    })
}
```

- [ ] **Step 2: Verify build**

```bash
go build ./pkg/tally/...
```

Expected: compiles. Some errors in `pkg/mcp/` are expected (handler still references old API) — that's fine.

- [ ] **Step 3: Commit**

```bash
git add pkg/tally/client.go
git commit -m "refactor: replace ExecuteTemplate+Parse* with generic ExecuteXML in client"
```

---

## Task 6: Update `pkg/mcp/handler.go` and `tools.go` — Generic Dispatch

**Files:**
- Modify: `pkg/mcp/handler.go`
- Modify: `pkg/mcp/tools.go`

- [ ] **Step 1: Rewrite `pkg/mcp/handler.go`**

```go
package mcp

import (
    "fmt"

    "github.com/taxor-ai/tally-mcp/pkg/logger"
    "github.com/taxor-ai/tally-mcp/pkg/tally"
)

// Handler processes MCP tool calls using the tool registry.
type Handler struct {
    client   *tally.Client
    registry *tally.Registry
    log      *logger.Logger
}

// NewHandler creates a new MCP handler backed by the given registry.
func NewHandler(client *tally.Client, registry *tally.Registry, log *logger.Logger) *Handler {
    return &Handler{client: client, registry: registry, log: log}
}

// ListTools returns all tools registered in the registry as MCP Tool structs.
func (h *Handler) ListTools() []Tool {
    defs := h.registry.All()
    tools := make([]Tool, 0, len(defs))
    for _, def := range defs {
        tools = append(tools, Tool{
            Name:        def.Name,
            Description: def.Description,
            InputSchema: def.InputSchema.ToMap(),
        })
    }
    return tools
}

// HandleToolCall dispatches a tool call generically:
//  1. Look up tool in registry
//  2. Build template params (implicit from client config + explicit from call)
//  3. Render request.xml template
//  4. POST to Tally
//  5. Apply parser.yaml spec and return result
func (h *Handler) HandleToolCall(toolName string, params map[string]interface{}) (interface{}, error) {
    def := h.registry.Get(toolName)
    if def == nil {
        return nil, fmt.Errorf("unknown tool: %s", toolName)
    }

    if h.log != nil {
        h.log.Infof("%s called", toolName)
    }

    // Build string params map
    templateParams := make(map[string]string, len(params)+2)

    // Inject implicit params from client config
    for _, implicit := range def.ImplicitParams {
        switch implicit {
        case "company_name":
            templateParams["company_name"] = h.client.Company
        }
    }

    // Merge caller-supplied params (string values only)
    for k, v := range params {
        if s, ok := v.(string); ok {
            templateParams[k] = s
        } else if v != nil {
            templateParams[k] = fmt.Sprintf("%v", v)
        }
    }

    // Render template
    rendered := tally.RenderTemplate(def.RequestXML, templateParams)

    // POST to Tally
    xmlResp, err := h.client.ExecuteXML(rendered)
    if err != nil {
        if h.log != nil {
            h.log.Warnf("%s failed: %v", toolName, err)
        }
        return nil, fmt.Errorf("tally request failed: %w", err)
    }

    // Parse response
    result, err := tally.ParseResponse(xmlResp, def.Parser)
    if err != nil {
        if h.log != nil {
            h.log.Warnf("%s parse error: %v", toolName, err)
        }
        return nil, fmt.Errorf("parse response failed: %w", err)
    }

    if h.log != nil {
        h.log.Infof("%s completed", toolName)
    }
    return result, nil
}
```

- [ ] **Step 2: Rewrite `pkg/mcp/tools.go`**

```go
package mcp

// Tool represents an MCP tool returned in tools/list.
type Tool struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    InputSchema map[string]interface{} `json:"inputSchema"`
}
```

Note: `AllTools()` is removed — callers now use `handler.ListTools()`.

- [ ] **Step 3: Update `main.go` to create registry and pass to handler**

In `main.go`, change:

```go
// OLD
handler := mcp.NewHandler(client, log)

// NEW — add registry creation after client creation:
registry, err := tally.LoadRegistry(tally.EmbeddedFS)
if err != nil {
    fmt.Fprintf(os.Stderr, "Error loading tool registry: %v\n", err)
    os.Exit(1)
}
handler := mcp.NewHandler(client, registry, log)
```

Also update `handleToolsList` in `main.go` (line ~196):

```go
// OLD
tools := mcp.AllTools()

// NEW
tools := handler.ListTools()
```

Update the `handleToolsList` signature to accept `handler`:

```go
func handleToolsList(req MCPRequest, handler *mcp.Handler, log *logger.Logger) {
    log.Debugf("Listing available tools")
    tools := handler.ListTools()
    response := MCPResponse{
        JSONRPC: "2.0",
        ID:      req.ID,
        Result: map[string]interface{}{
            "tools": tools,
        },
    }
    writeResponse(response)
}
```

- [ ] **Step 4: Build the entire project**

```bash
go build ./...
```

Expected: compiles without errors.

- [ ] **Step 5: Commit**

```bash
git add pkg/mcp/handler.go pkg/mcp/tools.go main.go
git commit -m "refactor: generic registry-based handler replaces hardcoded switch"
```

---

## Task 7: Delete Old Flat Template Files

**Files:**
- Delete: 8 old flat `.xml` files

- [ ] **Step 1: Remove old flat templates**

```bash
cd /Users/sree/Projects/Branding/tally-mcp
rm pkg/tally/templates/company/get_companies.xml
rm pkg/tally/templates/ledger/get_ledgers.xml
rm pkg/tally/templates/ledger/get_ledger_details.xml
rm pkg/tally/templates/ledger/create_ledger.xml
rm pkg/tally/templates/debtor_creditor/get_debtors.xml
rm pkg/tally/templates/debtor_creditor/get_creditors.xml
rm pkg/tally/templates/voucher/get_vouchers.xml
rm pkg/tally/templates/voucher/create_voucher.xml
```

- [ ] **Step 2: Build to confirm nothing broke**

```bash
go build ./...
```

Expected: clean build.

- [ ] **Step 3: Run unit tests**

```bash
go test ./pkg/tally/... -v
```

Expected: all parser and registry tests PASS.

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "chore: remove old flat XML templates superseded by subdirectory layout"
```

---

## Task 8: Update `models.go` — Remove Unused Response Types

**Files:**
- Modify: `pkg/tally/models.go`

- [ ] **Step 1: Remove response structs (Company, Ledger, Debtor, Creditor, Voucher, TallyResponse)**

The generic parser returns `[]map[string]interface{}` so these typed structs are no longer used. Replace `pkg/tally/models.go` with only the request types still needed by nothing after this refactor — so delete the file entirely:

```bash
rm /Users/sree/Projects/Branding/tally-mcp/pkg/tally/models.go
```

- [ ] **Step 2: Build to find any remaining references**

```bash
go build ./...
```

If any file references `tally.Company`, `tally.Ledger`, etc., it will error here. Fix by removing those references.

- [ ] **Step 3: Commit**

```bash
git add -A
git commit -m "chore: remove typed response models superseded by generic map parser"
```

---

## Task 9: Update Integration Tests

**Files:**
- Modify: `tests/integration/main_integration_test.go`

- [ ] **Step 1: Rewrite integration tests**

Replace the entire `tests/integration/main_integration_test.go` with:

```go
//go:build integration

package main

import (
    "os"
    "strconv"
    "testing"

    "github.com/taxor-ai/tally-mcp/pkg/logger"
    "github.com/taxor-ai/tally-mcp/pkg/mcp"
    "github.com/taxor-ai/tally-mcp/pkg/tally"
)

// setupHandler creates a registry-backed MCP handler from environment variables.
// Tests skip if TALLY_HOST is not set.
func setupHandler(t *testing.T) *mcp.Handler {
    t.Helper()
    host := os.Getenv("TALLY_HOST")
    if host == "" {
        t.Skip("TALLY_HOST not set — skipping integration test")
    }
    port := 9900
    if v := os.Getenv("TALLY_PORT"); v != "" {
        if p, err := strconv.Atoi(v); err == nil {
            port = p
        }
    }
    company := os.Getenv("TALLY_COMPANY")

    log, _ := logger.New("warn", "")
    client := tally.NewClient(host, port, 30)
    client.SetCompany(company)

    if err := client.Ping(); err != nil {
        t.Fatalf("cannot connect to Tally at %s:%d: %v", host, port, err)
    }

    registry, err := tally.LoadRegistry(tally.EmbeddedFS)
    if err != nil {
        t.Fatalf("LoadRegistry failed: %v", err)
    }
    return mcp.NewHandler(client, registry, log)
}

func TestGetCompaniesIntegration(t *testing.T) {
    handler := setupHandler(t)
    result, err := handler.HandleToolCall("get_companies", map[string]interface{}{})
    if err != nil {
        t.Fatalf("get_companies failed: %v", err)
    }
    m := result.(map[string]interface{})
    if m["success"] != true {
        t.Fatal("expected success=true")
    }
    companies := m["companies"].([]map[string]interface{})
    if len(companies) == 0 {
        t.Fatal("expected at least one company")
    }
    for i, c := range companies {
        if c["name"] == "" || c["name"] == nil {
            t.Errorf("company %d has empty name", i)
        }
        t.Logf("  Company %d: name=%v guid=%v", i+1, c["name"], c["guid"])
    }
    t.Logf("✓ get_companies: %d companies", len(companies))
}

func TestGetLedgersIntegration(t *testing.T) {
    handler := setupHandler(t)
    result, err := handler.HandleToolCall("get_ledgers", map[string]interface{}{})
    if err != nil {
        t.Fatalf("get_ledgers failed: %v", err)
    }
    m := result.(map[string]interface{})
    ledgers := m["ledgers"].([]map[string]interface{})
    t.Logf("✓ get_ledgers: %d ledgers", len(ledgers))
    for i, l := range ledgers {
        if i >= 3 {
            break
        }
        t.Logf("  Ledger %d: name=%v parent=%v", i+1, l["name"], l["parent"])
    }
}

func TestGetDebtorsIntegration(t *testing.T) {
    handler := setupHandler(t)
    result, err := handler.HandleToolCall("get_debtors", map[string]interface{}{})
    if err != nil {
        t.Fatalf("get_debtors failed: %v", err)
    }
    m := result.(map[string]interface{})
    debtors := m["debtors"].([]map[string]interface{})
    t.Logf("✓ get_debtors: %d debtors", len(debtors))
    for i, d := range debtors {
        if i >= 3 {
            break
        }
        t.Logf("  Debtor %d: name=%v balance=%v", i+1, d["name"], d["closing_balance"])
    }
}

func TestGetCreditorsIntegration(t *testing.T) {
    handler := setupHandler(t)
    result, err := handler.HandleToolCall("get_creditors", map[string]interface{}{})
    if err != nil {
        t.Fatalf("get_creditors failed: %v", err)
    }
    m := result.(map[string]interface{})
    creditors := m["creditors"].([]map[string]interface{})
    t.Logf("✓ get_creditors: %d creditors", len(creditors))
    for i, c := range creditors {
        if i >= 3 {
            break
        }
        t.Logf("  Creditor %d: name=%v balance=%v", i+1, c["name"], c["closing_balance"])
    }
}

func TestGetVouchersIntegration(t *testing.T) {
    handler := setupHandler(t)
    result, err := handler.HandleToolCall("get_vouchers", map[string]interface{}{
        "from_date": "20240101",
        "to_date":   "20241231",
    })
    if err != nil {
        t.Fatalf("get_vouchers failed: %v", err)
    }
    m := result.(map[string]interface{})
    vouchers := m["vouchers"].([]map[string]interface{})
    t.Logf("✓ get_vouchers: %d vouchers", len(vouchers))
    for i, v := range vouchers {
        if i >= 3 {
            break
        }
        t.Logf("  Voucher %d: number=%v date=%v type=%v", i+1, v["voucher_number"], v["date"], v["voucher_type"])
    }
}

func TestRegistryHasAllExpectedTools(t *testing.T) {
    registry, err := tally.LoadRegistry(tally.EmbeddedFS)
    if err != nil {
        t.Fatalf("LoadRegistry failed: %v", err)
    }
    expected := []string{
        "get_companies", "get_ledgers", "get_ledger_details",
        "create_ledger", "get_debtors", "get_creditors",
        "get_vouchers", "create_voucher",
    }
    for _, name := range expected {
        if registry.Get(name) == nil {
            t.Errorf("tool %q not found in registry", name)
        }
    }
    t.Logf("✓ registry has %d tools", len(registry.All()))
}

func TestAllGetToolsSequenceIntegration(t *testing.T) {
    handler := setupHandler(t)

    tools := []struct {
        name   string
        params map[string]interface{}
        key    string
    }{
        {"get_companies", nil, "companies"},
        {"get_ledgers", nil, "ledgers"},
        {"get_debtors", nil, "debtors"},
        {"get_creditors", nil, "creditors"},
        {"get_vouchers", map[string]interface{}{"from_date": "20240101", "to_date": "20241231"}, "vouchers"},
    }

    for _, tc := range tools {
        result, err := handler.HandleToolCall(tc.name, tc.params)
        if err != nil {
            t.Fatalf("%s failed: %v", tc.name, err)
        }
        m := result.(map[string]interface{})
        items := m[tc.key].([]map[string]interface{})
        t.Logf("✓ %s: %d items", tc.name, len(items))
    }
}
```

- [ ] **Step 2: Run unit (non-integration) tests to ensure no compile errors**

```bash
go test ./... -v 2>&1 | head -40
```

Expected: compile succeeds. Integration tests are skipped (no `TALLY_HOST`).

- [ ] **Step 3: Commit**

```bash
git add tests/integration/main_integration_test.go
git commit -m "test: update integration tests to use registry and map assertions"
```

---

## Task 10: End-to-End Verification

- [ ] **Step 1: Build the binary**

```bash
cd /Users/sree/Projects/Branding/tally-mcp
go build -o tally-mcp .
```

Expected: binary produced, no errors.

- [ ] **Step 2: Run all unit tests**

```bash
go test ./pkg/... -v
```

Expected: all tests PASS.

- [ ] **Step 3: Run integration tests against real Tally (if available)**

```bash
TALLY_HOST=<host> TALLY_PORT=9900 TALLY_COMPANY="<company>" \
  go test -tags=integration ./tests/integration/... -v -timeout 60s
```

Expected: all 7 integration tests PASS — verify output shows real company/ledger/debtor names.

- [ ] **Step 4: Verify tools/list works via stdin**

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | ./tally-mcp
```

Expected: JSON response with `tools` array containing all 8 tools.

- [ ] **Step 5: Final commit**

```bash
git add -A
git commit -m "feat: template-driven tool registry — tools defined by files, zero Go per tool"
```

---

## Self-Review

**Spec coverage:**
- ✅ Tools loaded from file system (tool.yaml + request.xml + parser.yaml)
- ✅ All 8 existing tools migrated
- ✅ Old flat XML deleted
- ✅ Generic handler replaces hardcoded switch
- ✅ Generic XPath parser replaces hardcoded Parse* functions
- ✅ Integration tests updated and passing
- ✅ `EmbeddedFS` exported for test access
- ✅ User extension model: drop files → tool appears (foundation in place)

**Notes:**
- `create_voucher` line items (for loop) are not implemented in this plan — template covers simple header-only vouchers. This is a known limitation to address in a follow-on plan.
- The old `models_test.go`, `client_test.go`, `templates_test.go` unit tests may reference removed types — they will fail to compile. Task 8 handles the build check; fix any stale test files found at that point by removing references to deleted types.
