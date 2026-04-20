package validation

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"pgregory.net/rapid"
)

// Reference patterns for biconditional checks (independent of production code).
var (
	refUUIDPattern     = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	refHexColorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
	refTeamKeyPattern  = regexp.MustCompile(`^[A-Z0-9]{1,5}$`)
)

// --------------------------------------------------------------------------
// Property 1: UUID field validation
// Validates: Requirements 5.5, 13.1
// --------------------------------------------------------------------------

func TestProperty1_UUIDFieldValidation_RandomStrings(t *testing.T) {
	t.Log("Feature: crossplane-provider-linear, Property 1: UUID field validation")

	rapid.Check(t, func(t *rapid.T) {
		value := rapid.StringMatching(`[0-9a-fA-F\-]{0,40}`).Draw(t, "value")
		err := ValidateUUID("teamId", value)
		matches := refUUIDPattern.MatchString(value)

		if matches && err != nil {
			t.Fatalf("valid UUID %q rejected: %v", value, err)
		}
		if !matches && err == nil {
			t.Fatalf("invalid UUID %q accepted", value)
		}
	})
}

func TestProperty1_UUIDFieldValidation_ValidUUIDs(t *testing.T) {
	t.Log("Feature: crossplane-provider-linear, Property 1: UUID field validation")

	hexChar := rapid.StringMatching(`[0-9a-fA-F]`)
	rapid.Check(t, func(t *rapid.T) {
		uuid := fmt.Sprintf("%s-%s-%s-%s-%s",
			drawHexN(t, hexChar, 8),
			drawHexN(t, hexChar, 4),
			drawHexN(t, hexChar, 4),
			drawHexN(t, hexChar, 4),
			drawHexN(t, hexChar, 12),
		)
		if err := ValidateUUID("teamId", uuid); err != nil {
			t.Fatalf("valid UUID %q rejected: %v", uuid, err)
		}
	})
}

func TestProperty1_UUIDFieldValidation_ArbitraryStrings(t *testing.T) {
	t.Log("Feature: crossplane-provider-linear, Property 1: UUID field validation")

	rapid.Check(t, func(t *rapid.T) {
		value := rapid.String().Draw(t, "value")
		err := ValidateUUID("teamId", value)
		matches := refUUIDPattern.MatchString(value)

		if matches && err != nil {
			t.Fatalf("valid UUID %q rejected: %v", value, err)
		}
		if !matches && err == nil {
			t.Fatalf("invalid UUID %q accepted", value)
		}
	})
}

// --------------------------------------------------------------------------
// Property 2: Hex color field validation
// Validates: Requirements 3.6, 4.6, 7.8, 8.5, 13.2
// --------------------------------------------------------------------------

func TestProperty2_HexColorFieldValidation_RandomStrings(t *testing.T) {
	t.Log("Feature: crossplane-provider-linear, Property 2: Hex color field validation")

	rapid.Check(t, func(t *rapid.T) {
		value := rapid.StringMatching(`[#0-9a-fA-F]{0,10}`).Draw(t, "value")
		err := ValidateHexColor("color", value)
		matches := refHexColorPattern.MatchString(value)

		if matches && err != nil {
			t.Fatalf("valid hex color %q rejected: %v", value, err)
		}
		if !matches && err == nil {
			t.Fatalf("invalid hex color %q accepted", value)
		}
	})
}

func TestProperty2_HexColorFieldValidation_ValidColors(t *testing.T) {
	t.Log("Feature: crossplane-provider-linear, Property 2: Hex color field validation")

	rapid.Check(t, func(t *rapid.T) {
		hex := rapid.StringMatching(`[0-9a-fA-F]{6}`).Draw(t, "hex")
		color := "#" + hex
		if err := ValidateHexColor("color", color); err != nil {
			t.Fatalf("valid hex color %q rejected: %v", color, err)
		}
	})
}

func TestProperty2_HexColorFieldValidation_ArbitraryStrings(t *testing.T) {
	t.Log("Feature: crossplane-provider-linear, Property 2: Hex color field validation")

	rapid.Check(t, func(t *rapid.T) {
		value := rapid.String().Draw(t, "value")
		err := ValidateHexColor("color", value)
		matches := refHexColorPattern.MatchString(value)

		if matches && err != nil {
			t.Fatalf("valid hex color %q rejected: %v", value, err)
		}
		if !matches && err == nil {
			t.Fatalf("invalid hex color %q accepted", value)
		}
	})
}

// --------------------------------------------------------------------------
// Property 3: Team key validation
// Validates: Requirements 3.4, 13.4
// --------------------------------------------------------------------------

func TestProperty3_TeamKeyValidation_RandomStrings(t *testing.T) {
	t.Log("Feature: crossplane-provider-linear, Property 3: Team key validation")

	rapid.Check(t, func(t *rapid.T) {
		value := rapid.StringMatching(`[A-Za-z0-9]{0,8}`).Draw(t, "value")
		err := ValidateTeamKey("key", value)
		matches := refTeamKeyPattern.MatchString(value)

		if matches && err != nil {
			t.Fatalf("valid team key %q rejected: %v", value, err)
		}
		if !matches && err == nil {
			t.Fatalf("invalid team key %q accepted", value)
		}
	})
}

func TestProperty3_TeamKeyValidation_ValidKeys(t *testing.T) {
	t.Log("Feature: crossplane-provider-linear, Property 3: Team key validation")

	rapid.Check(t, func(t *rapid.T) {
		length := rapid.IntRange(1, 5).Draw(t, "length")
		key := rapid.StringMatching(fmt.Sprintf(`[A-Z0-9]{%d}`, length)).Draw(t, "key")
		if err := ValidateTeamKey("key", key); err != nil {
			t.Fatalf("valid team key %q rejected: %v", key, err)
		}
	})
}

func TestProperty3_TeamKeyValidation_ArbitraryStrings(t *testing.T) {
	t.Log("Feature: crossplane-provider-linear, Property 3: Team key validation")

	rapid.Check(t, func(t *rapid.T) {
		value := rapid.String().Draw(t, "value")
		err := ValidateTeamKey("key", value)
		matches := refTeamKeyPattern.MatchString(value)

		if matches && err != nil {
			t.Fatalf("valid team key %q rejected: %v", value, err)
		}
		if !matches && err == nil {
			t.Fatalf("invalid team key %q accepted", value)
		}
	})
}

// --------------------------------------------------------------------------
// Property 4: Name minimum length validation
// Validates: Requirements 3.5, 4.4, 6.4, 7.4, 8.4, 13.3
// --------------------------------------------------------------------------

func TestProperty4_TeamNameAcceptedIffLengthGE2(t *testing.T) {
	t.Log("Feature: crossplane-provider-linear, Property 4: Name minimum length validation")

	rapid.Check(t, func(t *rapid.T) {
		value := rapid.String().Draw(t, "name")
		err := ValidateTeamName("name", value)
		valid := len(value) >= 2

		if valid && err != nil {
			t.Fatalf("valid team name %q (len=%d) rejected: %v", value, len(value), err)
		}
		if !valid && err == nil {
			t.Fatalf("invalid team name %q (len=%d) accepted", value, len(value))
		}
	})
}

func TestProperty4_OtherNameAcceptedIffLengthGE1(t *testing.T) {
	t.Log("Feature: crossplane-provider-linear, Property 4: Name minimum length validation")

	rapid.Check(t, func(t *rapid.T) {
		value := rapid.String().Draw(t, "name")
		err := ValidateName("name", value)
		valid := len(value) >= 1

		if valid && err != nil {
			t.Fatalf("valid name %q (len=%d) rejected: %v", value, len(value), err)
		}
		if !valid && err == nil {
			t.Fatalf("invalid name %q (len=%d) accepted", value, len(value))
		}
	})
}

func TestProperty4_NameMinLengthArbitrary(t *testing.T) {
	t.Log("Feature: crossplane-provider-linear, Property 4: Name minimum length validation")

	rapid.Check(t, func(t *rapid.T) {
		minLen := rapid.IntRange(0, 10).Draw(t, "minLen")
		value := rapid.String().Draw(t, "name")
		err := ValidateNameMinLength("name", value, minLen)
		valid := len(value) >= minLen

		if valid && err != nil {
			t.Fatalf("name %q (len=%d, min=%d) rejected: %v", value, len(value), minLen, err)
		}
		if !valid && err == nil {
			t.Fatalf("name %q (len=%d, min=%d) accepted", value, len(value), minLen)
		}
	})
}

// --------------------------------------------------------------------------
// Property 6: Template data JSON validation
// Validates: Requirements 6.5
// --------------------------------------------------------------------------

func TestProperty6_ValidJSONAlwaysAccepted(t *testing.T) {
	t.Log("Feature: crossplane-provider-linear, Property 6: Template data JSON validation")

	rapid.Check(t, func(t *rapid.T) {
		data := generateValidJSON(t)
		if err := ValidateJSON("data", string(data)); err != nil {
			t.Fatalf("valid JSON %q rejected: %v", string(data), err)
		}
	})
}

func TestProperty6_ArbitraryStringsAcceptedIffValidJSON(t *testing.T) {
	t.Log("Feature: crossplane-provider-linear, Property 6: Template data JSON validation")

	rapid.Check(t, func(t *rapid.T) {
		value := rapid.String().Draw(t, "value")
		err := ValidateJSON("data", value)
		isValid := json.Valid([]byte(value))

		if isValid && err != nil {
			t.Fatalf("valid JSON %q rejected: %v", value, err)
		}
		if !isValid && err == nil {
			t.Fatalf("invalid JSON %q accepted", value)
		}
	})
}

// generateValidJSON produces a valid JSON byte slice from random Go values.
func generateValidJSON(t *rapid.T) []byte {
	choice := rapid.IntRange(0, 4).Draw(t, "jsonType")
	var raw interface{}
	switch choice {
	case 0:
		raw = map[string]interface{}{
			rapid.StringMatching(`[a-z]{1,8}`).Draw(t, "key"): rapid.IntRange(-1000, 1000).Draw(t, "val"),
		}
	case 1:
		raw = []int{rapid.IntRange(-100, 100).Draw(t, "elem")}
	case 2:
		raw = rapid.String().Draw(t, "str")
	case 3:
		raw = rapid.IntRange(-10000, 10000).Draw(t, "num")
	default:
		raw = rapid.Bool().Draw(t, "bool")
	}
	data, _ := json.Marshal(raw)
	return data
}

// --------------------------------------------------------------------------
// Property 7: Fiscal year start month range validation
// Validates: Requirements 9.5
// --------------------------------------------------------------------------

func TestProperty7_FiscalMonthAcceptedIffInRange(t *testing.T) {
	t.Log("Feature: crossplane-provider-linear, Property 7: Fiscal year start month range validation")

	rapid.Check(t, func(t *rapid.T) {
		value := rapid.IntRange(-1000, 1000).Draw(t, "month")
		err := ValidateFiscalMonth("fiscalYearStartMonth", value)
		valid := value >= 0 && value <= 11

		if valid && err != nil {
			t.Fatalf("valid fiscal month %d rejected: %v", value, err)
		}
		if !valid && err == nil {
			t.Fatalf("invalid fiscal month %d accepted", value)
		}
	})
}

func TestProperty7_FiscalMonthBoundaryValues(t *testing.T) {
	t.Log("Feature: crossplane-provider-linear, Property 7: Fiscal year start month range validation")

	rapid.Check(t, func(t *rapid.T) {
		value := rapid.IntRange(-5, 16).Draw(t, "month")
		err := ValidateFiscalMonth("fiscalYearStartMonth", value)
		valid := value >= 0 && value <= 11

		if valid && err != nil {
			t.Fatalf("valid fiscal month %d rejected: %v", value, err)
		}
		if !valid && err == nil {
			t.Fatalf("invalid fiscal month %d accepted", value)
		}
	})
}

// --------------------------------------------------------------------------
// Property 9: Validation errors identify field and constraint
// Validates: Requirements 13.7
// --------------------------------------------------------------------------

func TestProperty9_UUIDErrorContainsFieldAndConstraint(t *testing.T) {
	t.Log("Feature: crossplane-provider-linear, Property 9: Validation errors identify field and constraint")

	rapid.Check(t, func(t *rapid.T) {
		field := rapid.StringMatching(`[a-zA-Z]{1,20}`).Draw(t, "field")
		value := rapid.StringMatching(`[^0-9a-fA-F\-]{1,10}`).Draw(t, "value")
		err := ValidateUUID(field, value)
		if err == nil {
			if refUUIDPattern.MatchString(value) {
				return
			}
			t.Fatalf("invalid UUID %q accepted for field %q", value, field)
		}
		assertErrorContains(t, err, field, "UUID")
	})
}

func TestProperty9_HexColorErrorContainsFieldAndConstraint(t *testing.T) {
	t.Log("Feature: crossplane-provider-linear, Property 9: Validation errors identify field and constraint")

	rapid.Check(t, func(t *rapid.T) {
		field := rapid.StringMatching(`[a-zA-Z]{1,20}`).Draw(t, "field")
		value := rapid.StringMatching(`[^#0-9a-fA-F]{1,10}`).Draw(t, "value")
		err := ValidateHexColor(field, value)
		if err == nil {
			if refHexColorPattern.MatchString(value) {
				return
			}
			t.Fatalf("invalid hex color %q accepted for field %q", value, field)
		}
		assertErrorContains(t, err, field, "hex color")
	})
}

func TestProperty9_TeamKeyErrorContainsFieldAndConstraint(t *testing.T) {
	t.Log("Feature: crossplane-provider-linear, Property 9: Validation errors identify field and constraint")

	rapid.Check(t, func(t *rapid.T) {
		field := rapid.StringMatching(`[a-zA-Z]{1,20}`).Draw(t, "field")
		value := rapid.StringMatching(`[a-z]{1,10}`).Draw(t, "value")
		err := ValidateTeamKey(field, value)
		if err == nil {
			if refTeamKeyPattern.MatchString(value) {
				return
			}
			t.Fatalf("invalid team key %q accepted for field %q", value, field)
		}
		assertErrorContains(t, err, field, "alphanumeric")
	})
}

func TestProperty9_NameLengthErrorContainsFieldAndConstraint(t *testing.T) {
	t.Log("Feature: crossplane-provider-linear, Property 9: Validation errors identify field and constraint")

	rapid.Check(t, func(t *rapid.T) {
		field := rapid.StringMatching(`[a-zA-Z]{1,20}`).Draw(t, "field")
		err := ValidateName(field, "")
		if err == nil {
			t.Fatalf("empty name accepted for field %q", field)
		}
		assertErrorContains(t, err, field, "character")
	})
}

func TestProperty9_JSONErrorContainsFieldAndConstraint(t *testing.T) {
	t.Log("Feature: crossplane-provider-linear, Property 9: Validation errors identify field and constraint")

	rapid.Check(t, func(t *rapid.T) {
		field := rapid.StringMatching(`[a-zA-Z]{1,20}`).Draw(t, "field")
		value := rapid.StringMatching(`[^"{\[0-9ntf]{1,10}`).Draw(t, "value")
		err := ValidateJSON(field, value)
		if err == nil {
			if json.Valid([]byte(value)) {
				return
			}
			t.Fatalf("invalid JSON %q accepted for field %q", value, field)
		}
		assertErrorContains(t, err, field, "JSON")
	})
}

func TestProperty9_FiscalMonthErrorContainsFieldAndConstraint(t *testing.T) {
	t.Log("Feature: crossplane-provider-linear, Property 9: Validation errors identify field and constraint")

	rapid.Check(t, func(t *rapid.T) {
		field := rapid.StringMatching(`[a-zA-Z]{1,20}`).Draw(t, "field")
		value := rapid.OneOf(
			rapid.IntRange(-1000, -1),
			rapid.IntRange(12, 1000),
		).Draw(t, "value")
		err := ValidateFiscalMonth(field, value)
		if err == nil {
			t.Fatalf("invalid fiscal month %d accepted for field %q", value, field)
		}
		assertErrorContains(t, err, field, "between")
	})
}

// --------------------------------------------------------------------------
// Helpers
// --------------------------------------------------------------------------

// drawHexN draws n hex characters and concatenates them.
func drawHexN(t *rapid.T, gen *rapid.Generator[string], n int) string {
	var sb strings.Builder
	for i := 0; i < n; i++ {
		sb.WriteString(gen.Draw(t, fmt.Sprintf("hex_%d", i)))
	}
	return sb.String()
}

// fatalIfMissing checks that err message contains both the field name and
// a constraint keyword. Works with both *testing.T and *rapid.T via the
// common interface.
type fatalfHelper interface {
	Fatalf(format string, args ...interface{})
}

func assertErrorContains(t fatalfHelper, err error, field, constraint string) {
	msg := err.Error()
	if !strings.Contains(msg, field) {
		t.Fatalf("error %q does not contain field name %q", msg, field)
	}
	if !strings.Contains(strings.ToLower(msg), strings.ToLower(constraint)) {
		t.Fatalf("error %q does not contain constraint description %q", msg, constraint)
	}
}
