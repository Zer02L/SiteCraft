package pages

import (
	"encoding/json"
	"sitecraft/internal/models"
)

// ParseFeaturesPayload converts raw block payload map into typed FeaturesPayload
func ParseFeaturesPayload(payload map[string]interface{}) *models.FeaturesPayload {
	if payload == nil {
		return nil
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return nil
	}
	var res models.FeaturesPayload
	if err := json.Unmarshal(b, &res); err != nil {
		return nil
	}
	return &res
}

// ParseFAQPayload converts raw block payload map into typed FAQPayload
func ParseFAQPayload(payload map[string]interface{}) *models.FAQPayload {
	if payload == nil {
		return nil
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return nil
	}
	var res models.FAQPayload
	if err := json.Unmarshal(b, &res); err != nil {
		return nil
	}
	return &res
}

// ParseStepsPayload converts raw block payload map into typed StepsPayload
func ParseStepsPayload(payload map[string]interface{}) *models.StepsPayload {
	if payload == nil {
		return nil
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return nil
	}
	var res models.StepsPayload
	if err := json.Unmarshal(b, &res); err != nil {
		return nil
	}
	return &res
}
