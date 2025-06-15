package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

type ServiceInfo struct {
	Name       string `json:"name"`
	URL        string `json:"url"`
	SchemaPath string `json:"schemaPath"`
}

type TypeClash struct {
	TypeName string   `json:"typeName"`
	Services []string `json:"services"`
}

type ServiceConfig struct {
	Services    map[string]ServiceDetails `json:"services"`
	BuildTime   string                    `json:"buildTime"`
	Version     string                    `json:"version"`
	TypeClashes []TypeClash               `json:"typeClashes"`
}

type ServiceDetails struct {
	URL       string   `json:"url"`
	Queries   []string `json:"queries"`
	Mutations []string `json:"mutations"`
	Types     []string `json:"types"`
}

type SchemaMerger struct {
	services    []ServiceInfo
	schemas     map[string]*ast.Schema
	typeClashes []TypeClash
}

func NewSchemaMerger() *SchemaMerger {
	return &SchemaMerger{
		services: []ServiceInfo{
			{
				Name:       "afl",
				URL:        getEnvOrDefault("AFL_SERVICE_URL", "http://localhost:8081/query"),
				SchemaPath: "../../services/afl/api/graphql/schema.graphqls",
			},
			{
				Name:       "ffl",
				URL:        getEnvOrDefault("FFL_SERVICE_URL", "http://localhost:8080/query"),
				SchemaPath: "../../services/ffl/api/graphql/schema.graphqls",
			},
		},
		schemas: make(map[string]*ast.Schema),
	}
}

func (sm *SchemaMerger) LoadSchemas() error {
	fmt.Println("🔄 Loading schemas from services...")

	for _, service := range sm.services {
		// Read schema file
		content, err := os.ReadFile(service.SchemaPath)
		if err != nil {
			return fmt.Errorf("failed to read %s schema at %s: %w", service.Name, service.SchemaPath, err)
		}

		// Parse schema
		schema, gqlErr := gqlparser.LoadSchema(&ast.Source{
			Name:    service.Name + ".graphql",
			Input:   string(content),
			BuiltIn: false,
		})
		if gqlErr != nil {
			return fmt.Errorf("failed to parse %s schema: %w", service.Name, gqlErr)
		}

		sm.schemas[service.Name] = schema
		fmt.Printf("✅ Loaded %s schema (%d types)\n", service.Name, len(schema.Types))
	}

	return nil
}

func (sm *SchemaMerger) DetectTypeClashes() error {
	fmt.Println("⚔️  Detecting type name clashes...")

	typeMap := make(map[string][]string)

	// Collect all type names from all services
	for serviceName, schema := range sm.schemas {
		for typeName := range schema.Types {
			// Skip built-in GraphQL types and root operation types
			if isBuiltinType(typeName) || isRootOperationType(typeName) {
				continue
			}
			typeMap[typeName] = append(typeMap[typeName], serviceName)
		}
	}

	// Find clashes
	var clashes []TypeClash
	for typeName, services := range typeMap {
		if len(services) > 1 {
			sort.Strings(services) // For consistent output
			clashes = append(clashes, TypeClash{
				TypeName: typeName,
				Services: services,
			})
		}
	}

	if len(clashes) > 0 {
		fmt.Printf("❌ %d type clash(es) detected:\n", len(clashes))
		for _, clash := range clashes {
			fmt.Printf("  - %s: %s\n", clash.TypeName, strings.Join(clash.Services, ", "))
		}

		fmt.Println("\n🔧 Recommendations:")
		for _, clash := range clashes {
			for _, service := range clash.Services {
				prefix := strings.ToUpper(service)
				fmt.Printf("  - Rename '%s' to '%s%s' in %s service\n",
					clash.TypeName, prefix, clash.TypeName, service)
			}
		}

		sm.typeClashes = clashes
		return fmt.Errorf("schema build failed due to %d type name clash(es)", len(clashes))
	}

	fmt.Println("✅ No type clashes detected")
	return nil
}

func (sm *SchemaMerger) MergeSchemas() (*ast.Schema, error) {
	fmt.Println("🔨 Merging schemas...")

	// Start with an empty schema definition
	mergedSource := &ast.Source{
		Name:  "merged-schema.graphql",
		Input: "",
	}

	var allTypeDefs strings.Builder

	// Build merged Query type
	allTypeDefs.WriteString("type Query {\n")
	allTypeDefs.WriteString("  _gateway: GatewayInfo!\n")
	
	// Add queries from all services
	for _, service := range sm.services {
		schema := sm.schemas[service.Name]
		if queryType, exists := schema.Types["Query"]; exists {
			fmt.Printf("  📋 Merging %s queries...\n", service.Name)
			for _, field := range queryType.Fields {
				// Skip introspection fields
				if strings.HasPrefix(field.Name, "__") {
					continue
				}
				allTypeDefs.WriteString("  " + field.Name)
				if len(field.Arguments) > 0 {
					allTypeDefs.WriteString("(")
					for i, arg := range field.Arguments {
						if i > 0 {
							allTypeDefs.WriteString(", ")
						}
						allTypeDefs.WriteString(arg.Name + ": " + arg.Type.String())
					}
					allTypeDefs.WriteString(")")
				}
				allTypeDefs.WriteString(": " + field.Type.String() + "\n")
			}
		}
	}
	allTypeDefs.WriteString("}\n\n")

	// Build merged Mutation type if any service has mutations
	hasMutations := false
	for _, service := range sm.services {
		if _, exists := sm.schemas[service.Name].Types["Mutation"]; exists {
			hasMutations = true
			break
		}
	}
	
	if hasMutations {
		allTypeDefs.WriteString("type Mutation {\n")
		for _, service := range sm.services {
			schema := sm.schemas[service.Name]
			if mutationType, exists := schema.Types["Mutation"]; exists {
				fmt.Printf("  📋 Merging %s mutations...\n", service.Name)
				for _, field := range mutationType.Fields {
					allTypeDefs.WriteString("  " + field.Name)
					if len(field.Arguments) > 0 {
						allTypeDefs.WriteString("(")
						for i, arg := range field.Arguments {
							if i > 0 {
								allTypeDefs.WriteString(", ")
							}
							allTypeDefs.WriteString(arg.Name + ": " + arg.Type.String())
						}
						allTypeDefs.WriteString(")")
					}
					allTypeDefs.WriteString(": " + field.Type.String() + "\n")
				}
			}
		}
		allTypeDefs.WriteString("}\n\n")
	}

	// Add gateway metadata type
	allTypeDefs.WriteString(`type GatewayInfo {
  version: String!
  services: [String!]!
  lastBuild: String!
  uptime: Float!
}

`)

	// Merge all non-root types from service schemas
	for _, service := range sm.services {
		schema := sm.schemas[service.Name]
		
		fmt.Printf("  📋 Merging %s types...\n", service.Name)
		
		// Convert schema back to SDL, excluding root types
		schemaSDL := formatSchemaToSDLExcludingRootTypes(schema)
		if schemaSDL != "" {
			allTypeDefs.WriteString(fmt.Sprintf("# %s Service Types\n", strings.ToUpper(service.Name)))
			allTypeDefs.WriteString(schemaSDL)
			allTypeDefs.WriteString("\n")
		}
	}

	// Parse the merged schema
	mergedSource.Input = allTypeDefs.String()
	mergedSchema, gqlErr := gqlparser.LoadSchema(mergedSource)
	if gqlErr != nil {
		return nil, fmt.Errorf("failed to parse merged schema: %w", gqlErr)
	}

	fmt.Printf("✅ Schema merged successfully (%d total types)\n", len(mergedSchema.Types))
	return mergedSchema, nil
}

func (sm *SchemaMerger) GenerateServiceConfig() ServiceConfig {
	fmt.Println("⚙️  Generating service configuration...")

	config := ServiceConfig{
		Services:    make(map[string]ServiceDetails),
		BuildTime:   time.Now().Format(time.RFC3339),
		Version:     getEnvOrDefault("BUILD_VERSION", "dev"),
		TypeClashes: sm.typeClashes,
	}

	for _, service := range sm.services {
		schema := sm.schemas[service.Name]
		details := ServiceDetails{
			URL:       service.URL,
			Queries:   []string{},
			Mutations: []string{},
			Types:     []string{},
		}

		// Extract queries, mutations, and types
		for typeName, typeDef := range schema.Types {
			if isBuiltinType(typeName) {
				continue
			}

			switch typeName {
			case "Query":
				for _, field := range typeDef.Fields {
					details.Queries = append(details.Queries, field.Name)
				}
			case "Mutation":
				for _, field := range typeDef.Fields {
					details.Mutations = append(details.Mutations, field.Name)
				}
			default:
				if typeDef.Kind == ast.Object || typeDef.Kind == ast.InputObject {
					details.Types = append(details.Types, typeName)
				}
			}
		}

		// Sort for consistent output
		sort.Strings(details.Queries)
		sort.Strings(details.Mutations)
		sort.Strings(details.Types)

		config.Services[service.Name] = details
		fmt.Printf("  ✅ %s: %d queries, %d mutations, %d types\n",
			service.Name, len(details.Queries), len(details.Mutations), len(details.Types))
	}

	return config
}

func (sm *SchemaMerger) WriteOutputs(schema *ast.Schema, config ServiceConfig) error {
	fmt.Println("📝 Writing output files...")

	// Ensure gateway directory exists
	gatewayDir := "../../gateway"
	if err := os.MkdirAll(filepath.Join(gatewayDir, "generated"), 0755); err != nil {
		return fmt.Errorf("failed to create gateway/generated directory: %w", err)
	}

	// Write unified schema
	schemaPath := filepath.Join(gatewayDir, "generated", "unified-schema.graphql")
	schemaContent := formatSchemaToSDL(schema)
	if err := os.WriteFile(schemaPath, []byte(schemaContent), 0644); err != nil {
		return fmt.Errorf("failed to write unified schema: %w", err)
	}
	fmt.Printf("  ✅ %s\n", schemaPath)

	// Write service configuration
	configPath := filepath.Join(gatewayDir, "generated", "service-config.json")
	configBytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal service config: %w", err)
	}
	if err := os.WriteFile(configPath, configBytes, 0644); err != nil {
		return fmt.Errorf("failed to write service config: %w", err)
	}
	fmt.Printf("  ✅ %s\n", configPath)

	// Write build info
	buildInfo := map[string]interface{}{
		"timestamp":   time.Now().Format(time.RFC3339),
		"services":    []string{"afl", "ffl"},
		"typeCount":   len(schema.Types),
		"clashCount":  len(sm.typeClashes),
		"toolVersion": "1.0.0",
	}
	buildInfoPath := filepath.Join(gatewayDir, "generated", "build-info.json")
	buildInfoBytes, err := json.MarshalIndent(buildInfo, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal build info: %w", err)
	}
	if err := os.WriteFile(buildInfoPath, buildInfoBytes, 0644); err != nil {
		return fmt.Errorf("failed to write build info: %w", err)
	}
	fmt.Printf("  ✅ %s\n", buildInfoPath)

	return nil
}

func formatSchemaToSDL(schema *ast.Schema) string {
	var result strings.Builder

	// Format each type definition by manually constructing SDL
	for _, typeName := range getSortedTypeNames(schema.Types) {
		if isBuiltinType(typeName) {
			continue
		}

		typeDef := schema.Types[typeName]
		formatTypeDefinition(&result, typeDef)
		result.WriteString("\n")
	}

	return result.String()
}

func formatSchemaToSDLExcludingRootTypes(schema *ast.Schema) string {
	var result strings.Builder

	// Format each type definition by manually constructing SDL, excluding root types
	for _, typeName := range getSortedTypeNames(schema.Types) {
		if isBuiltinType(typeName) || isRootOperationType(typeName) {
			continue
		}

		typeDef := schema.Types[typeName]
		formatTypeDefinition(&result, typeDef)
		result.WriteString("\n")
	}

	return result.String()
}

func formatTypeDefinition(w *strings.Builder, def *ast.Definition) {
	switch def.Kind {
	case ast.Object:
		w.WriteString("type " + def.Name)
		if len(def.Interfaces) > 0 {
			w.WriteString(" implements ")
			for i, iface := range def.Interfaces {
				if i > 0 {
					w.WriteString(" & ")
				}
				w.WriteString(iface)
			}
		}
		w.WriteString(" {\n")
		for _, field := range def.Fields {
			w.WriteString("  " + field.Name)
			if len(field.Arguments) > 0 {
				w.WriteString("(")
				for i, arg := range field.Arguments {
					if i > 0 {
						w.WriteString(", ")
					}
					w.WriteString(arg.Name + ": " + arg.Type.String())
				}
				w.WriteString(")")
			}
			w.WriteString(": " + field.Type.String() + "\n")
		}
		w.WriteString("}")
	case ast.InputObject:
		w.WriteString("input " + def.Name + " {\n")
		for _, field := range def.Fields {
			w.WriteString("  " + field.Name + ": " + field.Type.String() + "\n")
		}
		w.WriteString("}")
	case ast.Enum:
		w.WriteString("enum " + def.Name + " {\n")
		for _, value := range def.EnumValues {
			w.WriteString("  " + value.Name + "\n")
		}
		w.WriteString("}")
	case ast.Scalar:
		w.WriteString("scalar " + def.Name)
	case ast.Interface:
		w.WriteString("interface " + def.Name + " {\n")
		for _, field := range def.Fields {
			w.WriteString("  " + field.Name + ": " + field.Type.String() + "\n")
		}
		w.WriteString("}")
	case ast.Union:
		w.WriteString("union " + def.Name + " = ")
		for i, member := range def.Types {
			if i > 0 {
				w.WriteString(" | ")
			}
			w.WriteString(member)
		}
	}
}

func getSortedTypeNames(types map[string]*ast.Definition) []string {
	var names []string
	for name := range types {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func isBuiltinType(typeName string) bool {
	// Skip GraphQL built-in scalar types
	builtins := map[string]bool{
		"String": true, "Int": true, "Float": true, "Boolean": true, "ID": true,
	}
	
	// Skip all introspection types (anything starting with __)
	if strings.HasPrefix(typeName, "__") {
		return true
	}
	
	return builtins[typeName]
}

func isRootOperationType(typeName string) bool {
	return typeName == "Query" || typeName == "Mutation" || typeName == "Subscription"
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	fmt.Println("🚀 XFFL Schema Merger v1.0.0")
	fmt.Println("=" + strings.Repeat("=", 40))

	merger := NewSchemaMerger()

	// Load schemas
	if err := merger.LoadSchemas(); err != nil {
		fmt.Printf("💥 Failed to load schemas: %v\n", err)
		os.Exit(1)
	}

	// Detect type clashes
	if err := merger.DetectTypeClashes(); err != nil {
		fmt.Printf("💥 %v\n", err)
		os.Exit(1)
	}

	// Merge schemas
	mergedSchema, err := merger.MergeSchemas()
	if err != nil {
		fmt.Printf("💥 Failed to merge schemas: %v\n", err)
		os.Exit(1)
	}

	// Generate service configuration
	config := merger.GenerateServiceConfig()

	// Write outputs
	if err := merger.WriteOutputs(mergedSchema, config); err != nil {
		fmt.Printf("💥 Failed to write outputs: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("🎉 Schema merge completed successfully!")
	fmt.Printf("   Build time: %s\n", config.BuildTime)
	fmt.Printf("   Services: %s\n", strings.Join([]string{"afl", "ffl"}, ", "))
	fmt.Printf("   Gateway files written to: ../../gateway/generated/\n")
}
