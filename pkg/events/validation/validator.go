package validation

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/go-playground/validator/v10"
	"xffl/pkg/events"
	"xffl/pkg/events/generated"
)

// StructValidator validates events using go-playground/validator with generated structs
type StructValidator struct {
	validator    *validator.Validate
	eventStructs map[string]reflect.Type
}

// NewStructValidator creates a new validator that uses generated Go structs
func NewStructValidator() (*StructValidator, error) {
	v := validator.New()

	// Register custom validation functions if needed
	// v.RegisterValidation("custom", customValidationFunc)

	validator := &StructValidator{
		validator:    v,
		eventStructs: make(map[string]reflect.Type),
	}

	// Register event types with their corresponding struct types
	validator.registerEventTypes()

	return validator, nil
}

// registerEventTypes maps event type names to their Go struct types
func (v *StructValidator) registerEventTypes() {
	// Map event types to their payload struct types
	v.eventStructs["AFL.PlayerMatchUpdated"] = reflect.TypeOf(generated.PlayerMatchUpdatedPayload{})
	v.eventStructs["FFL.FantasyScoreCalculated"] = reflect.TypeOf(generated.FantasyScoreCalculatedPayload{})
}

// ValidateEvent validates a domain event against its generated Go struct
func (v *StructValidator) ValidateEvent(event events.DomainEvent) error {
	eventType := event.EventType()

	// Get the struct type for this event
	structType, exists := v.eventStructs[eventType]
	if !exists {
		// If no struct exists, log warning but don't fail
		// This allows for graceful degradation
		return nil
	}

	// Create a new instance of the struct
	structInstance := reflect.New(structType).Interface()

	// Convert event data to JSON and then unmarshal into the struct
	eventData := event.EventData()
	eventJSON, err := json.Marshal(eventData)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	if err := json.Unmarshal(eventJSON, structInstance); err != nil {
		return fmt.Errorf("failed to unmarshal event data into struct: %w", err)
	}

	// Validate the struct using go-playground/validator
	if err := v.validator.Struct(structInstance); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			return fmt.Errorf("event validation failed for %s: %s", eventType, formatValidationErrors(validationErrors))
		}
		return fmt.Errorf("validation error for %s: %w", eventType, err)
	}

	return nil
}

// ValidateStruct directly validates a struct (useful for testing)
func (v *StructValidator) ValidateStruct(s interface{}) error {
	return v.validator.Struct(s)
}

// GetSupportedEventTypes returns list of event types that have validation structs
func (v *StructValidator) GetSupportedEventTypes() []string {
	var types []string
	for eventType := range v.eventStructs {
		types = append(types, eventType)
	}
	return types
}

// formatValidationErrors converts validator.ValidationErrors to a readable string
func formatValidationErrors(errors validator.ValidationErrors) string {
	var errorMessages []string
	for _, err := range errors {
		var message string
		switch err.Tag() {
		case "required":
			message = fmt.Sprintf("field '%s' is required", err.Field())
		case "min":
			message = fmt.Sprintf("field '%s' must be at least %s", err.Field(), err.Param())
		case "max":
			message = fmt.Sprintf("field '%s' must be at most %s", err.Field(), err.Param())
		case "oneof":
			message = fmt.Sprintf("field '%s' must be one of: %s", err.Field(), err.Param())
		default:
			message = fmt.Sprintf("field '%s' failed validation '%s'", err.Field(), err.Tag())
		}
		errorMessages = append(errorMessages, message)
	}
	
	return fmt.Sprintf("[%s]", fmt.Sprintf("%v", errorMessages))
}