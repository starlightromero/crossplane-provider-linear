package validation

import (
	"strings"
	"testing"
)

func TestValidateUUID(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid lowercase", "550e8400-e29b-41d4-a716-446655440000", false},
		{"valid uppercase", "550E8400-E29B-41D4-A716-446655440000", false},
		{"valid mixed case", "550e8400-E29B-41d4-a716-446655440000", false},
		{"empty string", "", true},
		{"missing dashes", "550e8400e29b41d4a716446655440000", true},
		{"too short", "550e8400-e29b-41d4-a716", true},
		{"invalid chars", "gggggggg-gggg-gggg-gggg-gggggggggggg", true},
		{"extra chars", "550e8400-e29b-41d4-a716-4466554400001", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUUID("teamId", tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUUID() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				if !strings.Contains(err.Error(), "teamId") {
					t.Errorf("error should contain field name 'teamId', got: %s", err.Error())
				}
			}
		})
	}
}

func TestValidateHexColor(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid lowercase", "#6b5ce7", false},
		{"valid uppercase", "#6B5CE7", false},
		{"valid mixed", "#aaBBcc", false},
		{"missing hash", "6b5ce7", true},
		{"too short", "#6b5ce", true},
		{"too long", "#6b5ce77", true},
		{"invalid chars", "#gggggg", true},
		{"empty", "", true},
		{"three digit", "#abc", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHexColor("color", tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateHexColor() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), "color") {
				t.Errorf("error should contain field name 'color', got: %s", err.Error())
			}
		})
	}
}

func TestValidateTeamKey(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid single char", "A", false},
		{"valid five chars", "ENG12", false},
		{"valid all digits", "12345", false},
		{"lowercase", "eng", true},
		{"too long", "ABCDEF", true},
		{"empty", "", true},
		{"special chars", "AB-CD", true},
		{"spaces", "AB CD", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTeamKey("key", tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTeamKey() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), "key") {
				t.Errorf("error should contain field name 'key', got: %s", err.Error())
			}
		})
	}
}

func TestValidateTeamName(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid two chars", "AB", false},
		{"valid long", "Engineering", false},
		{"one char", "A", true},
		{"empty", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTeamName("name", tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTeamName() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), "name") {
				t.Errorf("error should contain field name 'name', got: %s", err.Error())
			}
		})
	}
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid one char", "A", false},
		{"valid long", "Bug Report", false},
		{"empty", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateName("name", tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateName() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), "name") {
				t.Errorf("error should contain field name 'name', got: %s", err.Error())
			}
		})
	}
}

func TestValidateAutoArchivePeriod(t *testing.T) {
	valid := []int{1, 3, 6, 9, 12}
	for _, v := range valid {
		if err := ValidateAutoArchivePeriod("autoArchivePeriod", v); err != nil {
			t.Errorf("ValidateAutoArchivePeriod(%d) unexpected error: %v", v, err)
		}
	}
	invalid := []int{0, 2, 4, 5, 7, 8, 10, 11, 13, -1}
	for _, v := range invalid {
		err := ValidateAutoArchivePeriod("autoArchivePeriod", v)
		if err == nil {
			t.Errorf("ValidateAutoArchivePeriod(%d) expected error, got nil", v)
		}
		if err != nil && !strings.Contains(err.Error(), "autoArchivePeriod") {
			t.Errorf("error should contain field name, got: %s", err.Error())
		}
	}
}

func TestValidateAutoClosePeriod(t *testing.T) {
	valid := []int{0, 1, 3, 6, 9, 12}
	for _, v := range valid {
		if err := ValidateAutoClosePeriod("autoClosePeriod", v); err != nil {
			t.Errorf("ValidateAutoClosePeriod(%d) unexpected error: %v", v, err)
		}
	}
	invalid := []int{2, 4, 5, 7, 8, 10, 11, 13, -1}
	for _, v := range invalid {
		err := ValidateAutoClosePeriod("autoClosePeriod", v)
		if err == nil {
			t.Errorf("ValidateAutoClosePeriod(%d) expected error, got nil", v)
		}
		if err != nil && !strings.Contains(err.Error(), "autoClosePeriod") {
			t.Errorf("error should contain field name, got: %s", err.Error())
		}
	}
}

func TestValidateWorkflowStateType(t *testing.T) {
	valid := []string{"triage", "backlog", "unstarted", "started", "completed", "canceled"}
	for _, v := range valid {
		if err := ValidateWorkflowStateType("type", v); err != nil {
			t.Errorf("ValidateWorkflowStateType(%q) unexpected error: %v", v, err)
		}
	}
	invalid := []string{"", "cancelled", "done", "open", "closed"}
	for _, v := range invalid {
		err := ValidateWorkflowStateType("type", v)
		if err == nil {
			t.Errorf("ValidateWorkflowStateType(%q) expected error, got nil", v)
		}
		if err != nil && !strings.Contains(err.Error(), "type") {
			t.Errorf("error should contain field name, got: %s", err.Error())
		}
	}
}

func TestValidateTemplateType(t *testing.T) {
	valid := []string{"issue", "project", "document"}
	for _, v := range valid {
		if err := ValidateTemplateType("type", v); err != nil {
			t.Errorf("ValidateTemplateType(%q) unexpected error: %v", v, err)
		}
	}
	invalid := []string{"", "bug", "task", "Issue"}
	for _, v := range invalid {
		err := ValidateTemplateType("type", v)
		if err == nil {
			t.Errorf("ValidateTemplateType(%q) expected error, got nil", v)
		}
		if err != nil && !strings.Contains(err.Error(), "type") {
			t.Errorf("error should contain field name, got: %s", err.Error())
		}
	}
}

func TestValidateJSON(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid object", `{"key":"value"}`, false},
		{"valid array", `[1,2,3]`, false},
		{"valid string", `"hello"`, false},
		{"valid number", `42`, false},
		{"valid null", `null`, false},
		{"valid bool", `true`, false},
		{"empty string", "", true},
		{"invalid json", `{key: value}`, true},
		{"trailing comma", `{"a":1,}`, true},
		{"plain text", `not json`, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateJSON("data", tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), "data") {
				t.Errorf("error should contain field name 'data', got: %s", err.Error())
			}
		})
	}
}

func TestValidateFiscalMonth(t *testing.T) {
	for v := 0; v <= 11; v++ {
		if err := ValidateFiscalMonth("fiscalYearStartMonth", v); err != nil {
			t.Errorf("ValidateFiscalMonth(%d) unexpected error: %v", v, err)
		}
	}
	invalid := []int{-1, -100, 12, 13, 100}
	for _, v := range invalid {
		err := ValidateFiscalMonth("fiscalYearStartMonth", v)
		if err == nil {
			t.Errorf("ValidateFiscalMonth(%d) expected error, got nil", v)
		}
		if err != nil && !strings.Contains(err.Error(), "fiscalYearStartMonth") {
			t.Errorf("error should contain field name, got: %s", err.Error())
		}
	}
}
