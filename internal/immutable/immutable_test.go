package immutable

import (
	"strings"
	"testing"

	"pgregory.net/rapid"
)

// --------------------------------------------------------------------------
// Property 5: Immutable field enforcement
// Validates: Requirements 4.5, 7.6, 7.7, 13.6
// --------------------------------------------------------------------------

func TestProperty5_TeamLabelTeamIdChangeRejected(t *testing.T) {
	t.Log("Feature: crossplane-provider-linear, Property 5: Immutable field enforcement")

	rapid.Check(t, func(t *rapid.T) {
		original := rapid.StringMatching(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`).Draw(t, "originalTeamId")
		updated := rapid.StringMatching(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`).Draw(t, "updatedTeamId")

		if original == updated {
			return
		}

		observed := map[string]any{"team_id": original}
		desired := map[string]any{"team_id": updated}

		err := CheckImmutableFields(observed, desired, TeamLabelConfig)
		if err == nil {
			t.Fatalf("expected rejection when changing teamId from %q to %q", original, updated)
		}
		assertContains(t, err.Error(), "teamId")
		assertContains(t, err.Error(), "immutable")
	})
}

func TestProperty5_WorkflowStateTypeChangeRejected(t *testing.T) {
	t.Log("Feature: crossplane-provider-linear, Property 5: Immutable field enforcement")

	validTypes := []string{"triage", "backlog", "unstarted", "started", "completed", "canceled"}

	rapid.Check(t, func(t *rapid.T) {
		origIdx := rapid.IntRange(0, len(validTypes)-1).Draw(t, "origIdx")
		updIdx := rapid.IntRange(0, len(validTypes)-1).Draw(t, "updIdx")

		if origIdx == updIdx {
			return
		}

		observed := map[string]any{
			"type":    validTypes[origIdx],
			"team_id": "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
		}
		desired := map[string]any{
			"type":    validTypes[updIdx],
			"team_id": "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
		}

		err := CheckImmutableFields(observed, desired, WorkflowStateConfig)
		if err == nil {
			t.Fatalf("expected rejection when changing type from %q to %q",
				validTypes[origIdx], validTypes[updIdx])
		}
		assertContains(t, err.Error(), "type")
		assertContains(t, err.Error(), "immutable")
	})
}

func TestProperty5_WorkflowStateTeamIdChangeRejected(t *testing.T) {
	t.Log("Feature: crossplane-provider-linear, Property 5: Immutable field enforcement")

	rapid.Check(t, func(t *rapid.T) {
		original := rapid.StringMatching(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`).Draw(t, "originalTeamId")
		updated := rapid.StringMatching(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`).Draw(t, "updatedTeamId")

		if original == updated {
			return
		}

		observed := map[string]any{
			"type":    "started",
			"team_id": original,
		}
		desired := map[string]any{
			"type":    "started",
			"team_id": updated,
		}

		err := CheckImmutableFields(observed, desired, WorkflowStateConfig)
		if err == nil {
			t.Fatalf("expected rejection when changing teamId from %q to %q", original, updated)
		}
		assertContains(t, err.Error(), "teamId")
		assertContains(t, err.Error(), "immutable")
	})
}

func TestProperty5_UnchangedImmutableFieldsAccepted(t *testing.T) {
	t.Log("Feature: crossplane-provider-linear, Property 5: Immutable field enforcement")

	rapid.Check(t, func(t *rapid.T) {
		teamID := rapid.StringMatching(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`).Draw(t, "teamId")

		observed := map[string]any{"team_id": teamID}
		desired := map[string]any{"team_id": teamID}

		err := CheckImmutableFields(observed, desired, TeamLabelConfig)
		if err != nil {
			t.Fatalf("unexpected rejection when teamId unchanged (%q): %v", teamID, err)
		}

		wsType := rapid.SampledFrom([]string{"triage", "backlog", "unstarted", "started", "completed", "canceled"}).Draw(t, "type")

		observed2 := map[string]any{"type": wsType, "team_id": teamID}
		desired2 := map[string]any{"type": wsType, "team_id": teamID}

		err = CheckImmutableFields(observed2, desired2, WorkflowStateConfig)
		if err != nil {
			t.Fatalf("unexpected rejection when type and teamId unchanged: %v", err)
		}
	})
}

func TestProperty5_ErrorMessagesContainFieldName(t *testing.T) {
	t.Log("Feature: crossplane-provider-linear, Property 5: Immutable field enforcement")

	rapid.Check(t, func(t *rapid.T) {
		fieldName := rapid.StringMatching(`[a-z]{3,10}`).Draw(t, "fieldName")
		displayName := rapid.StringMatching(`[a-zA-Z]{3,15}`).Draw(t, "displayName")
		original := rapid.StringMatching(`[a-z]{5,10}`).Draw(t, "original")
		updated := rapid.StringMatching(`[a-z]{5,10}`).Draw(t, "updated")

		if original == updated {
			return
		}

		config := ResourceConfig{
			ResourceKind: "TestResource",
			Fields: []FieldConfig{
				{FieldName: fieldName, DisplayName: displayName},
			},
		}

		observed := map[string]any{fieldName: original}
		desired := map[string]any{fieldName: updated}

		err := CheckImmutableFields(observed, desired, config)
		if err == nil {
			t.Fatalf("expected error when changing %s from %q to %q", fieldName, original, updated)
		}
		assertContains(t, err.Error(), displayName)
	})
}

// assertContains is a helper that fails the test if s does not contain substr.
func assertContains(t *rapid.T, s, substr string) {
	if !strings.Contains(s, substr) {
		t.Fatalf("string %q does not contain %q", s, substr)
	}
}
