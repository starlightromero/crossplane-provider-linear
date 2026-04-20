// Package validation provides reusable field validation functions for Linear
// Crossplane provider resources. All validators return nil on valid input and a
// descriptive error containing the field name and violated constraint on invalid
// input.
package validation

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// Pre-compiled regex patterns (compiled once at package level).
var (
	uuidPattern     = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	hexColorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
	teamKeyPattern  = regexp.MustCompile(`^[A-Z0-9]{1,5}$`)
)

// Valid enum value sets.
var (
	validAutoArchivePeriods = map[int]bool{1: true, 3: true, 6: true, 9: true, 12: true}
	validAutoClosePeriods   = map[int]bool{0: true, 1: true, 3: true, 6: true, 9: true, 12: true}
	validWorkflowStateTypes = map[string]bool{
		"triage": true, "backlog": true, "unstarted": true,
		"started": true, "completed": true, "canceled": true,
	}
	validTemplateTypes = map[string]bool{
		"issue": true, "project": true, "document": true,
	}
)

// ValidateUUID validates that value matches the UUID format.
func ValidateUUID(field, value string) error {
	if !uuidPattern.MatchString(value) {
		return fmt.Errorf("%s: must be a valid UUID matching %s, got %q", field, uuidPattern.String(), value)
	}
	return nil
}

// ValidateHexColor validates that value matches the #rrggbb hex color format.
func ValidateHexColor(field, value string) error {
	if !hexColorPattern.MatchString(value) {
		return fmt.Errorf("%s: must be a valid hex color matching %s, got %q", field, hexColorPattern.String(), value)
	}
	return nil
}

// ValidateTeamKey validates that value is 1-5 uppercase alphanumeric characters.
func ValidateTeamKey(field, value string) error {
	if !teamKeyPattern.MatchString(value) {
		return fmt.Errorf("%s: must be 1-5 uppercase alphanumeric characters matching %s, got %q", field, teamKeyPattern.String(), value)
	}
	return nil
}

// ValidateNameMinLength validates that value meets the minimum length for the
// given resource type. Team names require at least 2 characters; all other
// resource names require at least 1 character.
func ValidateNameMinLength(field, value string, minLen int) error {
	if len(value) < minLen {
		return fmt.Errorf("%s: must be at least %d character(s), got %d", field, minLen, len(value))
	}
	return nil
}

// ValidateTeamName validates that a Team name has at least 2 characters.
func ValidateTeamName(field, value string) error {
	return ValidateNameMinLength(field, value, 2)
}

// ValidateName validates that a resource name has at least 1 character.
// Use for TeamLabel, Template, WorkflowState, WorkspaceLabel, and other
// non-Team resources.
func ValidateName(field, value string) error {
	return ValidateNameMinLength(field, value, 1)
}

// ValidateAutoArchivePeriod validates that value is one of the allowed auto
// archive period values: 1, 3, 6, 9, 12.
func ValidateAutoArchivePeriod(field string, value int) error {
	if !validAutoArchivePeriods[value] {
		return fmt.Errorf("%s: must be one of [1, 3, 6, 9, 12], got %d", field, value)
	}
	return nil
}

// ValidateAutoClosePeriod validates that value is one of the allowed auto close
// period values: 0, 1, 3, 6, 9, 12.
func ValidateAutoClosePeriod(field string, value int) error {
	if !validAutoClosePeriods[value] {
		return fmt.Errorf("%s: must be one of [0, 1, 3, 6, 9, 12], got %d", field, value)
	}
	return nil
}

// ValidateWorkflowStateType validates that value is one of the allowed workflow
// state types: triage, backlog, unstarted, started, completed, canceled.
func ValidateWorkflowStateType(field, value string) error {
	if !validWorkflowStateTypes[value] {
		allowed := []string{"triage", "backlog", "unstarted", "started", "completed", "canceled"}
		return fmt.Errorf("%s: must be one of [%s], got %q", field, strings.Join(allowed, ", "), value)
	}
	return nil
}

// ValidateTemplateType validates that value is one of the allowed template
// types: issue, project, document.
func ValidateTemplateType(field, value string) error {
	if !validTemplateTypes[value] {
		allowed := []string{"issue", "project", "document"}
		return fmt.Errorf("%s: must be one of [%s], got %q", field, strings.Join(allowed, ", "), value)
	}
	return nil
}

// ValidateJSON validates that value is valid JSON.
func ValidateJSON(field, value string) error {
	if !json.Valid([]byte(value)) {
		return fmt.Errorf("%s: must be valid JSON", field)
	}
	return nil
}

// ValidateFiscalMonth validates that value is an integer in [0, 11].
func ValidateFiscalMonth(field string, value int) error {
	if value < 0 || value > 11 {
		return fmt.Errorf("%s: must be an integer between 0 and 11 inclusive, got %d", field, value)
	}
	return nil
}
