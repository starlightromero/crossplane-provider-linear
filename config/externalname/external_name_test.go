// Package externalname contains property-based tests for external name
// annotation round-trip correctness across all resource types.
//
// The compose/parse functions replicate the exact logic from each resource's
// config package (team, teamlabel, etc.) so we can verify the round-trip
// property without importing the ujconfig dependency.
package externalname

import (
	"fmt"
	"strings"
	"testing"

	"pgregory.net/rapid"
)

// --------------------------------------------------------------------------
// Property 8: External name annotation round-trip
// Validates: Requirements 11.5
//
// For each resource type, we replicate the exact SetIdentifierArgumentFn and
// GetExternalNameFn logic from the per-resource config packages and verify
// that composing an external name from identity fields and parsing it back
// yields the original fields.
// --------------------------------------------------------------------------

// Shared helpers for common external name patterns.

// colonSplitSetIdentifier splits an external name by ":" into two fields.
func colonSplitSetIdentifier(base map[string]any, externalName, field1, field2 string) {
	parts := strings.SplitN(externalName, ":", 2)
	if len(parts) == 2 {
		base[field1] = parts[0]
		base[field2] = parts[1]
	}
}

// colonJoinGetExternalName joins two fields with ":" to form an external name.
func colonJoinGetExternalName(tfstate map[string]any, field1, field2 string) (string, error) {
	v1, _ := tfstate[field1].(string)
	v2, _ := tfstate[field2].(string)
	if v1 == "" || v2 == "" {
		return "", fmt.Errorf("cannot determine external name: %s or %s is empty", field1, field2)
	}
	return v1 + ":" + v2, nil
}

// idGetExternalName reads the "id" field from tfstate as the external name.
func idGetExternalName(tfstate map[string]any) (string, error) {
	if id, ok := tfstate["id"].(string); ok && id != "" {
		return id, nil
	}
	return "", nil
}

// --- Team: external name = key (e.g., "ENG") ---

func teamSetIdentifier(base map[string]any, externalName string) {
	base["key"] = externalName
}

func teamGetExternalName(tfstate map[string]any) (string, error) {
	if key, ok := tfstate["key"].(string); ok && key != "" {
		return key, nil
	}
	return "", fmt.Errorf("cannot determine external name: key field is empty")
}

// --- TeamLabel: external name = label_name:team_key ---

func teamLabelSetIdentifier(base map[string]any, externalName string) {
	colonSplitSetIdentifier(base, externalName, "name", "team_key")
}

func teamLabelGetExternalName(tfstate map[string]any) (string, error) {
	return colonJoinGetExternalName(tfstate, "name", "team_key")
}

// --- TeamWorkflow: external name = key or key:branch_pattern:is_regex ---

func teamWorkflowSetIdentifier(base map[string]any, externalName string) {
	parts := strings.SplitN(externalName, ":", 3)
	if len(parts) >= 1 {
		base["key"] = parts[0]
	}
	if len(parts) >= 3 {
		base["branch_pattern"] = parts[1]
		base["is_regex"] = parts[2]
	}
}

func teamWorkflowGetExternalName(tfstate map[string]any) (string, error) {
	key, _ := tfstate["key"].(string)
	if key == "" {
		return "", fmt.Errorf("cannot determine external name: key field is empty")
	}
	pattern, _ := tfstate["branch_pattern"].(string)
	isRegex, _ := tfstate["is_regex"].(string)
	if pattern != "" {
		return key + ":" + pattern + ":" + isRegex, nil
	}
	return key, nil
}

// --- Template: external name = id (UUID) ---

func templateSetIdentifier(base map[string]any, externalName string) {
	base["id"] = externalName
}

func templateGetExternalName(tfstate map[string]any) (string, error) {
	return idGetExternalName(tfstate)
}

// --- WorkflowState: external name = workflow_state_name:team_key ---

func workflowStateSetIdentifier(base map[string]any, externalName string) {
	colonSplitSetIdentifier(base, externalName, "name", "team_key")
}

func workflowStateGetExternalName(tfstate map[string]any) (string, error) {
	return colonJoinGetExternalName(tfstate, "name", "team_key")
}

// --- WorkspaceLabel: external name = label_name ---

func workspaceLabelSetIdentifier(base map[string]any, externalName string) {
	base["name"] = externalName
}

func workspaceLabelGetExternalName(tfstate map[string]any) (string, error) {
	if name, ok := tfstate["name"].(string); ok && name != "" {
		return name, nil
	}
	return "", fmt.Errorf("cannot determine external name: name field is empty")
}

// --- WorkspaceSettings: external name = id (UUID) ---

func workspaceSettingsSetIdentifier(base map[string]any, externalName string) {
	base["id"] = externalName
}

func workspaceSettingsGetExternalName(tfstate map[string]any) (string, error) {
	return idGetExternalName(tfstate)
}

// --- Workspace: external name = id (UUID) ---

func workspaceSetIdentifier(base map[string]any, externalName string) {
	base["id"] = externalName
}

func workspaceGetExternalName(tfstate map[string]any) (string, error) {
	return idGetExternalName(tfstate)
}

// --------------------------------------------------------------------------
// Property tests — split per resource type to reduce cognitive complexity
// --------------------------------------------------------------------------

func TestProperty8_TeamKeyRoundTrip(t *testing.T) {
	t.Log("Feature: crossplane-provider-linear, Property 8: External name annotation round-trip")

	rapid.Check(t, func(t *rapid.T) {
		key := rapid.StringMatching(`[A-Z0-9]{1,5}`).Draw(t, "key")

		base := make(map[string]any)
		teamSetIdentifier(base, key)

		extName, err := teamGetExternalName(base)
		if err != nil {
			t.Fatalf("GetExternalNameFn failed for key %q: %v", key, err)
		}
		if extName != key {
			t.Fatalf("round-trip failed: input key=%q, got external name=%q", key, extName)
		}
	})
}

func TestProperty8_TeamLabelRoundTrip(t *testing.T) {
	t.Log("Feature: crossplane-provider-linear, Property 8: External name annotation round-trip")

	rapid.Check(t, func(t *rapid.T) {
		labelName := rapid.StringMatching(`[a-zA-Z0-9 ]{1,20}`).Draw(t, "labelName")
		teamKey := rapid.StringMatching(`[A-Z0-9]{1,5}`).Draw(t, "teamKey")
		composed := labelName + ":" + teamKey

		base := make(map[string]any)
		teamLabelSetIdentifier(base, composed)

		extName, err := teamLabelGetExternalName(base)
		if err != nil {
			t.Fatalf("GetExternalNameFn failed for %q: %v", composed, err)
		}
		if extName != composed {
			t.Fatalf("round-trip failed: input=%q, got=%q", composed, extName)
		}
		if base["name"] != labelName {
			t.Fatalf("name field mismatch: want %q, got %q", labelName, base["name"])
		}
		if base["team_key"] != teamKey {
			t.Fatalf("team_key field mismatch: want %q, got %q", teamKey, base["team_key"])
		}
	})
}

func TestProperty8_TeamWorkflowKeyOnlyRoundTrip(t *testing.T) {
	t.Log("Feature: crossplane-provider-linear, Property 8: External name annotation round-trip")

	rapid.Check(t, func(t *rapid.T) {
		key := rapid.StringMatching(`[A-Z0-9]{1,5}`).Draw(t, "key")

		base := make(map[string]any)
		teamWorkflowSetIdentifier(base, key)

		extName, err := teamWorkflowGetExternalName(base)
		if err != nil {
			t.Fatalf("GetExternalNameFn failed for key %q: %v", key, err)
		}
		if extName != key {
			t.Fatalf("round-trip failed: input key=%q, got=%q", key, extName)
		}
	})
}

func TestProperty8_TeamWorkflowBranchRoundTrip(t *testing.T) {
	t.Log("Feature: crossplane-provider-linear, Property 8: External name annotation round-trip")

	rapid.Check(t, func(t *rapid.T) {
		key := rapid.StringMatching(`[A-Z0-9]{1,5}`).Draw(t, "key")
		branchPattern := rapid.StringMatching(`[a-zA-Z0-9/\*\-]{1,20}`).Draw(t, "branchPattern")
		isRegex := rapid.SampledFrom([]string{"true", "false"}).Draw(t, "isRegex")
		composed := key + ":" + branchPattern + ":" + isRegex

		base := make(map[string]any)
		teamWorkflowSetIdentifier(base, composed)

		extName, err := teamWorkflowGetExternalName(base)
		if err != nil {
			t.Fatalf("GetExternalNameFn failed for %q: %v", composed, err)
		}
		if extName != composed {
			t.Fatalf("round-trip failed: input=%q, got=%q", composed, extName)
		}
		if base["key"] != key {
			t.Fatalf("key field mismatch: want %q, got %q", key, base["key"])
		}
		if base["branch_pattern"] != branchPattern {
			t.Fatalf("branch_pattern mismatch: want %q, got %q", branchPattern, base["branch_pattern"])
		}
		if base["is_regex"] != isRegex {
			t.Fatalf("is_regex mismatch: want %q, got %q", isRegex, base["is_regex"])
		}
	})
}

func TestProperty8_TemplateIdRoundTrip(t *testing.T) {
	t.Log("Feature: crossplane-provider-linear, Property 8: External name annotation round-trip")

	rapid.Check(t, func(t *rapid.T) {
		id := rapid.StringMatching(
			`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`,
		).Draw(t, "id")

		base := make(map[string]any)
		templateSetIdentifier(base, id)

		extName, err := templateGetExternalName(base)
		if err != nil {
			t.Fatalf("GetExternalNameFn failed for id %q: %v", id, err)
		}
		if extName != id {
			t.Fatalf("round-trip failed: input id=%q, got=%q", id, extName)
		}
	})
}

func TestProperty8_WorkflowStateRoundTrip(t *testing.T) {
	t.Log("Feature: crossplane-provider-linear, Property 8: External name annotation round-trip")

	rapid.Check(t, func(t *rapid.T) {
		stateName := rapid.StringMatching(`[a-zA-Z0-9 ]{1,20}`).Draw(t, "stateName")
		teamKey := rapid.StringMatching(`[A-Z0-9]{1,5}`).Draw(t, "teamKey")
		composed := stateName + ":" + teamKey

		base := make(map[string]any)
		workflowStateSetIdentifier(base, composed)

		extName, err := workflowStateGetExternalName(base)
		if err != nil {
			t.Fatalf("GetExternalNameFn failed for %q: %v", composed, err)
		}
		if extName != composed {
			t.Fatalf("round-trip failed: input=%q, got=%q", composed, extName)
		}
		if base["name"] != stateName {
			t.Fatalf("name field mismatch: want %q, got %q", stateName, base["name"])
		}
		if base["team_key"] != teamKey {
			t.Fatalf("team_key field mismatch: want %q, got %q", teamKey, base["team_key"])
		}
	})
}

func TestProperty8_WorkspaceLabelRoundTrip(t *testing.T) {
	t.Log("Feature: crossplane-provider-linear, Property 8: External name annotation round-trip")

	rapid.Check(t, func(t *rapid.T) {
		name := rapid.StringMatching(`[a-zA-Z0-9 ]{1,20}`).Draw(t, "name")

		base := make(map[string]any)
		workspaceLabelSetIdentifier(base, name)

		extName, err := workspaceLabelGetExternalName(base)
		if err != nil {
			t.Fatalf("GetExternalNameFn failed for name %q: %v", name, err)
		}
		if extName != name {
			t.Fatalf("round-trip failed: input name=%q, got=%q", name, extName)
		}
	})
}

func TestProperty8_WorkspaceSettingsIdRoundTrip(t *testing.T) {
	t.Log("Feature: crossplane-provider-linear, Property 8: External name annotation round-trip")

	rapid.Check(t, func(t *rapid.T) {
		id := rapid.StringMatching(
			`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`,
		).Draw(t, "id")

		base := make(map[string]any)
		workspaceSettingsSetIdentifier(base, id)

		extName, err := workspaceSettingsGetExternalName(base)
		if err != nil {
			t.Fatalf("GetExternalNameFn failed for id %q: %v", id, err)
		}
		if extName != id {
			t.Fatalf("round-trip failed: input id=%q, got=%q", id, extName)
		}
	})
}

func TestProperty8_WorkspaceIdRoundTrip(t *testing.T) {
	t.Log("Feature: crossplane-provider-linear, Property 8: External name annotation round-trip")

	rapid.Check(t, func(t *rapid.T) {
		id := rapid.StringMatching(
			`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`,
		).Draw(t, "id")

		base := make(map[string]any)
		workspaceSetIdentifier(base, id)

		extName, err := workspaceGetExternalName(base)
		if err != nil {
			t.Fatalf("GetExternalNameFn failed for id %q: %v", id, err)
		}
		if extName != id {
			t.Fatalf("round-trip failed: input id=%q, got=%q", id, extName)
		}
	})
}
