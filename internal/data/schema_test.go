package data

import (
	"slices"
	"testing"

	"github.com/pocketbase/pocketbase/core"
)

func TestSiteSettingsSelectFields(t *testing.T) {
	app := setupTestApp(t)
	defer app.ResetBootstrapState()

	col, err := app.FindCollectionByNameOrId("site_settings")
	if err != nil {
		t.Fatalf("site_settings collection not found: %v", err)
	}

	// Verify header_variant is a SelectField
	headerField := col.Fields.GetByName("header_variant")
	if headerField == nil {
		t.Fatal("header_variant field missing")
	}
	headerSelect, ok := headerField.(*core.SelectField)
	if !ok {
		t.Fatalf("expected header_variant to be *core.SelectField, got %T", headerField)
	}
	if headerSelect.MaxSelect != 1 {
		t.Errorf("expected header_variant MaxSelect=1, got %d", headerSelect.MaxSelect)
	}
	expectedHeaderVals := []string{"classic", "minimal", "centered"}
	for _, val := range expectedHeaderVals {
		if !slices.Contains(headerSelect.Values, val) {
			t.Errorf("expected header_variant to contain %q, got %v", val, headerSelect.Values)
		}
	}

	// Verify footer_variant is a SelectField
	footerField := col.Fields.GetByName("footer_variant")
	if footerField == nil {
		t.Fatal("footer_variant field missing")
	}
	footerSelect, ok := footerField.(*core.SelectField)
	if !ok {
		t.Fatalf("expected footer_variant to be *core.SelectField, got %T", footerField)
	}
	if footerSelect.MaxSelect != 1 {
		t.Errorf("expected footer_variant MaxSelect=1, got %d", footerSelect.MaxSelect)
	}
	expectedFooterVals := []string{"classic", "minimal", "columns"}
	for _, val := range expectedFooterVals {
		if !slices.Contains(footerSelect.Values, val) {
			t.Errorf("expected footer_variant to contain %q, got %v", val, footerSelect.Values)
		}
	}

	// Verify validation: saving with invalid header_variant should fail
	record := core.NewRecord(col)
	record.Set("site_name", "Test Site")
	record.Set("header_variant", "invalid_variant")
	if err := app.Save(record); err == nil {
		t.Error("expected error when saving invalid header_variant, got nil")
	}

	// Verify valid value succeeds
	record.Set("header_variant", "minimal")
	record.Set("footer_variant", "columns")
	if err := app.Save(record); err != nil {
		t.Errorf("expected success when saving valid variants, got %v", err)
	}

	// Verify migration from TextField: simulate existing collection with TextField
	col.Fields.RemoveByName("header_variant")
	col.Fields.Add(&core.TextField{Name: "header_variant"})
	if err := app.Save(col); err != nil {
		t.Fatalf("failed to downgrade header_variant to TextField: %v", err)
	}

	// Re-run initSiteSettingsCollection to test upgrade logic
	if err := initSiteSettingsCollection(app); err != nil {
		t.Fatalf("initSiteSettingsCollection failed on existing collection: %v", err)
	}

	reloadedCol, err := app.FindCollectionByNameOrId("site_settings")
	if err != nil {
		t.Fatalf("failed to reload site_settings: %v", err)
	}
	upgradedField := reloadedCol.Fields.GetByName("header_variant")
	if _, ok := upgradedField.(*core.SelectField); !ok {
		t.Fatalf("expected upgraded header_variant to be *core.SelectField, got %T", upgradedField)
	}
}
