package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

// AsyncAPISpec represents the structure we need from AsyncAPI files
type AsyncAPISpec struct {
	AsyncAPI   string            `yaml:"asyncapi"`
	Info       map[string]interface{} `yaml:"info"`
	Channels   map[string]Channel     `yaml:"channels"`
	Components Components             `yaml:"components"`
}

type Channel struct {
	Description string    `yaml:"description"`
	Publish     Operation `yaml:"publish,omitempty"`
	Subscribe   Operation `yaml:"subscribe,omitempty"`
}

type Operation struct {
	Summary     string  `yaml:"summary,omitempty"`
	Description string  `yaml:"description,omitempty"`
	Message     Message `yaml:"message"`
}

type Message struct {
	Ref     string                 `yaml:"$ref,omitempty"`
	Name    string                 `yaml:"name,omitempty"`
	Payload map[string]interface{} `yaml:"payload,omitempty"`
}

type Components struct {
	Messages map[string]MessageComponent `yaml:"messages"`
	Schemas  map[string]interface{}      `yaml:"schemas"`
}

type MessageComponent struct {
	Name        string                 `yaml:"name"`
	Title       string                 `yaml:"title,omitempty"`
	Summary     string                 `yaml:"summary,omitempty"`
	Description string                 `yaml:"description,omitempty"`
	Payload     map[string]interface{} `yaml:"payload"`
	Examples    []interface{}          `yaml:"examples,omitempty"`
}

// GoStruct represents a Go struct to be generated
type GoStruct struct {
	Name        string
	Description string
	Fields      []GoField
}

type GoField struct {
	Name         string
	Type         string
	JSONTag      string
	ValidateTag  string
	Description  string
}

const goStructTemplate = `// Package generated contains auto-generated Go structs from AsyncAPI specifications
// DO NOT EDIT - This file is generated automatically
package generated

{{range .}}
// {{.Name}} {{.Description}}
type {{.Name}} struct {
{{range .Fields}}	{{.Name}} {{.Type}} ` + "`json:\"{{.JSONTag}}\"{{if .ValidateTag}} validate:\"{{.ValidateTag}}\"{{end}}`" + ` // {{.Description}}
{{end}}}

{{end}}`

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Usage: go run generate-structs.go <asyncapi-dir> <output-file>")
	}

	asyncapiDir := os.Args[1]
	outputFile := os.Args[2]

	structs, err := generateStructsFromAsyncAPI(asyncapiDir)
	if err != nil {
		log.Fatalf("Failed to generate structs: %v", err)
	}

	if err := writeGoFile(outputFile, structs); err != nil {
		log.Fatalf("Failed to write Go file: %v", err)
	}

	fmt.Printf("Generated %d structs in %s\n", len(structs), outputFile)
}

func generateStructsFromAsyncAPI(dir string) ([]GoStruct, error) {
	var structs []GoStruct

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-YAML files
		if d.IsDir() || (!strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml")) {
			return nil
		}

		// Skip the unified spec (it references other files)
		if strings.Contains(path, "xffl-events") {
			return nil
		}

		fileStructs, err := parseAsyncAPIFile(path)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		structs = append(structs, fileStructs...)
		return nil
	})

	return structs, err
}

func parseAsyncAPIFile(filePath string) ([]GoStruct, error) {
	yamlData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var spec AsyncAPISpec
	if err := yaml.Unmarshal(yamlData, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	var structs []GoStruct

	// Generate structs from message components
	for messageName, message := range spec.Components.Messages {
		goStruct := GoStruct{
			Name:        messageName + "Payload",
			Description: message.Description,
		}

		if message.Payload != nil {
			fields, err := convertSchemaToFields(message.Payload, spec.Components.Schemas)
			if err != nil {
				return nil, fmt.Errorf("failed to convert schema for %s: %w", messageName, err)
			}
			goStruct.Fields = fields
			structs = append(structs, goStruct)
		}
	}

	// Generate structs from component schemas
	for schemaName, schema := range spec.Components.Schemas {
		if schemaMap, ok := schema.(map[string]interface{}); ok {
			goStruct := GoStruct{
				Name:        schemaName,
				Description: fmt.Sprintf("represents %s data structure", schemaName),
			}

			fields, err := convertSchemaToFields(schemaMap, spec.Components.Schemas)
			if err != nil {
				return nil, fmt.Errorf("failed to convert schema for %s: %w", schemaName, err)
			}
			goStruct.Fields = fields
			structs = append(structs, goStruct)
		}
	}

	return structs, nil
}

func convertSchemaToFields(schema map[string]interface{}, components map[string]interface{}) ([]GoField, error) {
	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		return nil, nil
	}

	required := make(map[string]bool)
	if req, ok := schema["required"].([]interface{}); ok {
		for _, r := range req {
			if str, ok := r.(string); ok {
				required[str] = true
			}
		}
	}

	var fields []GoField
	for propName, propSchema := range properties {
		field := GoField{
			Name:    toPascalCase(propName),
			JSONTag: propName,
		}

		if propMap, ok := propSchema.(map[string]interface{}); ok {
			// Handle $ref
			if ref, ok := propMap["$ref"].(string); ok {
				field.Type = extractTypeFromRef(ref)
			} else {
				field.Type = convertJSONTypeToGo(propMap)
			}

			// Add validation tags
			field.ValidateTag = buildValidationTag(propMap, required[propName])

			// Add description
			if desc, ok := propMap["description"].(string); ok {
				field.Description = desc
			}
		}

		fields = append(fields, field)
	}

	return fields, nil
}

func toPascalCase(s string) string {
	if s == "" {
		return s
	}
	
	// Handle camelCase to PascalCase
	if len(s) > 1 && s[0] >= 'a' && s[0] <= 'z' {
		return strings.ToUpper(string(s[0])) + s[1:]
	}
	
	return s
}

func extractTypeFromRef(ref string) string {
	parts := strings.Split(ref, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return "interface{}"
}

func convertJSONTypeToGo(schema map[string]interface{}) string {
	schemaType, ok := schema["type"].(string)
	if !ok {
		return "interface{}"
	}

	switch schemaType {
	case "string":
		return "string"
	case "integer":
		return "int"
	case "number":
		return "float64"
	case "boolean":
		return "bool"
	case "array":
		if items, ok := schema["items"].(map[string]interface{}); ok {
			itemType := convertJSONTypeToGo(items)
			return "[]" + itemType
		}
		return "[]interface{}"
	case "object":
		return "map[string]interface{}"
	default:
		return "interface{}"
	}
}

func buildValidationTag(schema map[string]interface{}, isRequired bool) string {
	var tags []string

	if isRequired {
		tags = append(tags, "required")
	}

	if min, ok := schema["minimum"].(float64); ok {
		tags = append(tags, fmt.Sprintf("min=%v", min))
	}

	if max, ok := schema["maximum"].(float64); ok {
		tags = append(tags, fmt.Sprintf("max=%v", max))
	}

	if enum, ok := schema["enum"].([]interface{}); ok {
		var enumValues []string
		for _, v := range enum {
			if str, ok := v.(string); ok {
				enumValues = append(enumValues, str)
			}
		}
		if len(enumValues) > 0 {
			tags = append(tags, fmt.Sprintf("oneof=%s", strings.Join(enumValues, " ")))
		}
	}

	return strings.Join(tags, ",")
}

func writeGoFile(filename string, structs []GoStruct) error {
	tmpl, err := template.New("structs").Parse(goStructTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	return tmpl.Execute(file, structs)
}