package validation

import (
	"testing"

	"xffl/pkg/events/generated"
)

func TestStructValidator(t *testing.T) {
	validator, err := NewStructValidator()
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	t.Run("ValidPlayerMatchUpdatedPayload", func(t *testing.T) {
		validPayload := generated.PlayerMatchUpdatedPayload{
			PlayerSeasonId: 123,
			ClubMatchId:    456,
			OldStats: generated.PlayerStats{
				Kicks:     10,
				Handballs: 5,
				Marks:     3,
				Hitouts:   0,  // Zero is valid for hitouts
				Tackles:   2,
				Goals:     1,
				Behinds:   0,  // Zero is valid for behinds
			},
			NewStats: generated.PlayerStats{
				Kicks:     12,
				Handballs: 6,
				Marks:     4,
				Hitouts:   0,  // Zero is valid for hitouts
				Tackles:   3,
				Goals:     2,
				Behinds:   1,
			},
		}

		err := validator.ValidateStruct(&validPayload)
		if err != nil {
			t.Errorf("Expected valid payload to pass validation, got: %v", err)
		}
	})

	t.Run("InvalidPlayerMatchUpdatedPayload_MissingRequired", func(t *testing.T) {
		invalidPayload := generated.PlayerMatchUpdatedPayload{
			// Missing PlayerSeasonId and ClubMatchId
			OldStats: generated.PlayerStats{
				Kicks: 10,
			},
			NewStats: generated.PlayerStats{
				Kicks: 12,
			},
		}

		err := validator.ValidateStruct(&invalidPayload)
		if err == nil {
			t.Error("Expected invalid payload to fail validation")
		}
	})

	t.Run("InvalidPlayerMatchUpdatedPayload_NegativeValues", func(t *testing.T) {
		invalidPayload := generated.PlayerMatchUpdatedPayload{
			PlayerSeasonId: -1, // Invalid: less than minimum
			ClubMatchId:    456,
			OldStats: generated.PlayerStats{
				Kicks:     -5, // Invalid: negative value
				Handballs: 5,
				Marks:     3,
				Hitouts:   0,
				Tackles:   2,
				Goals:     1,
				Behinds:   0,
			},
			NewStats: generated.PlayerStats{
				Kicks:     12,
				Handballs: 6,
				Marks:     4,
				Hitouts:   0,
				Tackles:   3,
				Goals:     2,
				Behinds:   1,
			},
		}

		err := validator.ValidateStruct(&invalidPayload)
		if err == nil {
			t.Error("Expected payload with negative values to fail validation")
		}
	})

	t.Run("ValidFantasyScoreCalculatedPayload", func(t *testing.T) {
		validPayload := generated.FantasyScoreCalculatedPayload{
			PlayerSeasonId: 123,
			ClubMatchId:    456,
			AflScore:       85,
			FantasyScore:   92,
			Source:         "AFL.PlayerMatchUpdated",
		}

		err := validator.ValidateStruct(&validPayload)
		if err != nil {
			t.Errorf("Expected valid payload to pass validation, got: %v", err)
		}
	})

	t.Run("InvalidFantasyScoreCalculatedPayload_InvalidSource", func(t *testing.T) {
		invalidPayload := generated.FantasyScoreCalculatedPayload{
			PlayerSeasonId: 123,
			ClubMatchId:    456,
			AflScore:       85,
			FantasyScore:   92,
			Source:         "InvalidSource", // Not in oneof list
		}

		err := validator.ValidateStruct(&invalidPayload)
		if err == nil {
			t.Error("Expected payload with invalid source to fail validation")
		}
	})

	t.Run("GetSupportedEventTypes", func(t *testing.T) {
		types := validator.GetSupportedEventTypes()
		expectedTypes := []string{"AFL.PlayerMatchUpdated", "FFL.FantasyScoreCalculated"}
		
		if len(types) != len(expectedTypes) {
			t.Errorf("Expected %d event types, got %d", len(expectedTypes), len(types))
		}

		typeMap := make(map[string]bool)
		for _, eventType := range types {
			typeMap[eventType] = true
		}

		for _, expectedType := range expectedTypes {
			if !typeMap[expectedType] {
				t.Errorf("Expected event type %s not found in supported types", expectedType)
			}
		}
	})
}