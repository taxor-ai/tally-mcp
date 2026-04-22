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
	if err != nil {
		return nil, fmt.Errorf("root node not found at %q", xpath)
	}
	key := spec.ResultKey
	if key == "" {
		key = "data"
	}
	if node == nil {
		return map[string]interface{}{
			"success": false,
			"error":   "not found",
			key:       nil,
		}, nil
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
	exceptions := nodeInt(doc, "//IMPORTRESULT/EXCEPTIONS")
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
	if exceptions > 0 {
		result["error"] = "Tally returned exception(s) during import"
		result["success"] = false
	}
	return result, nil
}

func extractFields(node *xmlquery.Node, fields map[string]FieldSpec) map[string]interface{} {
	item := make(map[string]interface{}, len(fields))
	for fieldName, spec := range fields {
		// Check if this is a nested list
		if spec.ItemsXPath != "" {
			nodes, err := xmlquery.QueryAll(node, spec.ItemsXPath)
			if err == nil {
				nestedItems := make([]map[string]interface{}, 0, len(nodes))
				for _, nestedNode := range nodes {
					nestedItems = append(nestedItems, extractFields(nestedNode, spec.Fields))
				}
				item[fieldName] = nestedItems
			}
			continue
		}

		// Handle simple fields
		var raw string
		if spec.XPath == "" {
			// No xpath specified, skip
			continue
		}
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
