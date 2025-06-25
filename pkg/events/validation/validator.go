package validation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
	"xffl/pkg/events"
)

// AsyncAPIValidator validates events against AsyncAPI schemas
type AsyncAPIValidator struct {
	schemas map[string]*gojsonschema.Schema
}

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

// NewAsyncAPIValidator creates a new validator that loads schemas from AsyncAPI specs
func NewAsyncAPIValidator(schemaDir string) (*AsyncAPIValidator, error) {
	validator := &AsyncAPIValidator{
		schemas: make(map[string]*gojsonschema.Schema),
	}

	if err := validator.loadSchemasFromAsyncAPI(schemaDir); err != nil {
		return nil, fmt.Errorf("failed to load AsyncAPI schemas: %w", err)
	}

	return validator, nil
}

// ValidateEvent validates a domain event against its AsyncAPI schema
func (v *AsyncAPIValidator) ValidateEvent(event events.DomainEvent) error {
	eventType := event.EventType()
	
	// Get schema for this event type
	schema, exists := v.schemas[eventType]
	if !exists {
		// If no schema exists, log warning but don't fail
		// This allows for graceful degradation
		return nil
	}

	// Convert event data to JSON for validation
	eventData := event.EventData()
	eventJSON, err := json.Marshal(eventData)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	// Create document loader from event JSON
	documentLoader := gojsonschema.NewBytesLoader(eventJSON)

	// Validate against schema
	result, err := schema.Validate(documentLoader)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	if !result.Valid() {
		// Collect all validation errors
		var errors []string
		for _, desc := range result.Errors() {
			errors = append(errors, desc.String())
		}
		return fmt.Errorf("event validation failed for %s: %s", eventType, strings.Join(errors, "; "))
	}

	return nil
}

// loadSchemasFromAsyncAPI loads JSON schemas by parsing AsyncAPI YAML files
func (v *AsyncAPIValidator) loadSchemasFromAsyncAPI(schemaDir string) error {
	// Walk through all YAML files in the schema directory
	err := filepath.WalkDir(schemaDir, func(path string, d os.DirEntry, err error) error {
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

		return v.loadAsyncAPIFile(path)
	})

	if err != nil {
		return fmt.Errorf("failed to walk schema directory: %w", err)
	}

	return nil
}

// loadAsyncAPIFile loads and parses a single AsyncAPI YAML file
func (v *AsyncAPIValidator) loadAsyncAPIFile(filePath string) error {
	// Read the YAML file
	yamlData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read AsyncAPI file %s: %w", filePath, err)
	}

	// Parse the AsyncAPI spec
	var spec AsyncAPISpec
	if err := yaml.Unmarshal(yamlData, &spec); err != nil {
		return fmt.Errorf("failed to parse AsyncAPI file %s: %w", filePath, err)
	}

	// Extract schemas from channels and components
	return v.extractSchemasFromSpec(spec, filePath)
}

// extractSchemasFromSpec extracts JSON schemas from AsyncAPI spec
func (v *AsyncAPIValidator) extractSchemasFromSpec(spec AsyncAPISpec, filePath string) error {
	// Process each channel to find event types and their schemas
	for channelName, channel := range spec.Channels {
		eventType := channelName // Channel name is the event type

		// Get the message schema from publish operation
		var messageRef string
		if channel.Publish.Message.Ref != "" {
			messageRef = channel.Publish.Message.Ref
		} else if channel.Subscribe.Message.Ref != "" {
			messageRef = channel.Subscribe.Message.Ref
		}

		if messageRef == "" {
			continue // Skip channels without message references
		}

		// Parse the reference (e.g., "#/components/messages/PlayerMatchUpdated")
		parts := strings.Split(messageRef, "/")
		if len(parts) != 4 || parts[0] != "#" || parts[1] != "components" || parts[2] != "messages" {
			continue // Skip invalid references
		}
		messageName := parts[3]

		// Get the message component
		messageComponent, exists := spec.Components.Messages[messageName]
		if !exists {
			continue // Skip if message component not found
		}

		// Extract the payload schema
		payloadSchema := messageComponent.Payload
		if payloadSchema == nil {
			continue // Skip if no payload schema
		}

		// Resolve schema references if needed
		resolvedSchema, err := v.resolveSchemaReferences(payloadSchema, spec.Components.Schemas)
		if err != nil {
			return fmt.Errorf("failed to resolve schema references for %s: %w", eventType, err)
		}

		// Convert to JSON and create gojsonschema.Schema
		schemaJSON, err := json.Marshal(resolvedSchema)
		if err != nil {
			return fmt.Errorf("failed to marshal schema for %s: %w", eventType, err)
		}

		schema, err := gojsonschema.NewSchema(gojsonschema.NewBytesLoader(schemaJSON))
		if err != nil {
			return fmt.Errorf("failed to create JSON schema for %s: %w", eventType, err)
		}

		v.schemas[eventType] = schema
	}

	return nil
}

// resolveSchemaReferences resolves $ref references in schemas
func (v *AsyncAPIValidator) resolveSchemaReferences(schema map[string]interface{}, components map[string]interface{}) (map[string]interface{}, error) {
	// For now, we'll implement basic reference resolution
	// In a full implementation, you'd handle nested references recursively
	
	resolved := make(map[string]interface{})
	for key, value := range schema {
		if key == "$ref" {
			// Handle reference resolution
			refPath, ok := value.(string)
			if !ok {
				return nil, fmt.Errorf("invalid $ref value: %v", value)
			}
			
			// Parse reference (e.g., "#/components/schemas/PlayerStats")
			parts := strings.Split(refPath, "/")
			if len(parts) == 4 && parts[0] == "#" && parts[1] == "components" && parts[2] == "schemas" {
				schemaName := parts[3]
				if refSchema, exists := components[schemaName]; exists {
					// Replace the $ref with the actual schema
					refMap, ok := refSchema.(map[string]interface{})
					if ok {
						for refKey, refValue := range refMap {
							resolved[refKey] = refValue
						}
					}
				}
			}
		} else {
			resolved[key] = value
		}
	}
	
	return resolved, nil
}

// GetSupportedEventTypes returns list of event types that have validation schemas
func (v *AsyncAPIValidator) GetSupportedEventTypes() []string {
	var types []string
	for eventType := range v.schemas {
		types = append(types, eventType)
	}
	return types
}