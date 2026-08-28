package api

import (
	"testing"
	"time"

	"github.com/raksul/go-clickup/clickup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// labelsField mirrors the shape ClickUp returns: type_config is untyped, with
// options carrying "label" for labels fields and "name" for dropdowns.
func labelsField() *clickup.CustomField {
	return &clickup.CustomField{
		ID:   "a258ab77",
		Name: "Machine",
		Type: "labels",
		TypeConfig: map[string]interface{}{
			"options": []interface{}{
				map[string]interface{}{"id": "opt-chi", "label": "Chilastra"},
				map[string]interface{}{"id": "opt-sun", "label": "Sunrunner"},
				map[string]interface{}{"id": "opt-abe", "label": "Aberama Gold"},
			},
		},
	}
}

func dropdownField() *clickup.CustomField {
	return &clickup.CustomField{
		ID:   "dd-1",
		Name: "Stage",
		Type: "drop_down",
		TypeConfig: map[string]interface{}{
			"options": []interface{}{
				map[string]interface{}{"id": "o1", "name": "Scoping"},
				map[string]interface{}{"id": "o2", "name": "Shipped"},
			},
		},
	}
}

func TestFindCustomField(t *testing.T) {
	fields := []clickup.CustomField{
		{ID: "f1", Name: "Repo", Type: "url"},
		{ID: "f2", Name: "Last synced", Type: "date"},
	}

	t.Run("by id", func(t *testing.T) {
		f, err := FindCustomField(fields, "f2")
		require.NoError(t, err)
		assert.Equal(t, "Last synced", f.Name)
	})

	t.Run("by name, case-insensitive", func(t *testing.T) {
		f, err := FindCustomField(fields, "rEpO")
		require.NoError(t, err)
		assert.Equal(t, "f1", f.ID)
	})

	t.Run("unknown name lists what is available", func(t *testing.T) {
		_, err := FindCustomField(fields, "Nope")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Repo")
		assert.Contains(t, err.Error(), "Last synced")
	})

	t.Run("ambiguous name is an error, not a first match", func(t *testing.T) {
		dupes := []clickup.CustomField{
			{ID: "a", Name: "Owner", Type: "text"},
			{ID: "b", Name: "owner", Type: "text"},
		}
		_, err := FindCustomField(dupes, "Owner")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ambiguous")
	})
}

func TestCoerceCustomFieldValue(t *testing.T) {
	t.Run("string types pass through", func(t *testing.T) {
		for _, typ := range []string{"url", "text", "short_text", "email", "phone"} {
			v, err := CoerceCustomFieldValue(&clickup.CustomField{Name: "F", Type: typ}, "hello")
			require.NoError(t, err, typ)
			assert.Equal(t, map[string]interface{}{"value": "hello"}, v)
		}
	})

	t.Run("number", func(t *testing.T) {
		v, err := CoerceCustomFieldValue(&clickup.CustomField{Name: "N", Type: "number"}, "42.5")
		require.NoError(t, err)
		assert.Equal(t, 42.5, v["value"])

		_, err = CoerceCustomFieldValue(&clickup.CustomField{Name: "N", Type: "number"}, "abc")
		assert.Error(t, err)
	})

	t.Run("checkbox", func(t *testing.T) {
		v, err := CoerceCustomFieldValue(&clickup.CustomField{Name: "C", Type: "checkbox"}, "true")
		require.NoError(t, err)
		assert.Equal(t, true, v["value"])

		_, err = CoerceCustomFieldValue(&clickup.CustomField{Name: "C", Type: "checkbox"}, "maybe")
		assert.Error(t, err)
	})

	t.Run("date accepts several forms", func(t *testing.T) {
		f := &clickup.CustomField{Name: "D", Type: "date"}

		v, err := CoerceCustomFieldValue(f, "1755835200000")
		require.NoError(t, err)
		assert.Equal(t, int64(1755835200000), v["value"])

		v, err = CoerceCustomFieldValue(f, "2026-08-22")
		require.NoError(t, err)
		want := time.Date(2026, 8, 22, 0, 0, 0, 0, time.Local).UnixMilli()
		assert.Equal(t, want, v["value"])

		v, err = CoerceCustomFieldValue(f, "today")
		require.NoError(t, err)
		y, m, d := time.Now().Date()
		assert.Equal(t, time.Date(y, m, d, 0, 0, 0, 0, time.Local).UnixMilli(), v["value"])

		_, err = CoerceCustomFieldValue(f, "not-a-date")
		assert.Error(t, err)
	})

	t.Run("dropdown resolves option name to id", func(t *testing.T) {
		v, err := CoerceCustomFieldValue(dropdownField(), "shipped")
		require.NoError(t, err)
		assert.Equal(t, "o2", v["value"])

		v, err = CoerceCustomFieldValue(dropdownField(), "o1")
		require.NoError(t, err)
		assert.Equal(t, "o1", v["value"])
	})

	t.Run("labels resolve a comma-separated list", func(t *testing.T) {
		v, err := CoerceCustomFieldValue(labelsField(), "Chilastra, Aberama Gold")
		require.NoError(t, err)
		assert.Equal(t, []string{"opt-chi", "opt-abe"}, v["value"])
	})

	t.Run("unknown label option is rejected with the valid set", func(t *testing.T) {
		_, err := CoerceCustomFieldValue(labelsField(), "Chilastra,Nope")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Sunrunner")
	})

	t.Run("unsupported types refuse rather than guess", func(t *testing.T) {
		_, err := CoerceCustomFieldValue(&clickup.CustomField{Name: "P", Type: "manual_progress"}, "50")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not supported")
	})
}

func TestFormatCustomFieldValue(t *testing.T) {
	t.Run("nil value renders empty", func(t *testing.T) {
		assert.Equal(t, "", FormatCustomFieldValue(&clickup.CustomField{Type: "url"}))
	})

	t.Run("labels render as names", func(t *testing.T) {
		f := labelsField()
		f.Value = []interface{}{"opt-chi", "opt-sun"}
		assert.Equal(t, "Chilastra, Sunrunner", FormatCustomFieldValue(f))
	})

	t.Run("unknown label id falls back to the id", func(t *testing.T) {
		f := labelsField()
		f.Value = []interface{}{"opt-gone"}
		assert.Equal(t, "opt-gone", FormatCustomFieldValue(f))
	})

	t.Run("dropdown renders as name", func(t *testing.T) {
		f := dropdownField()
		f.Value = "o2"
		assert.Equal(t, "Shipped", FormatCustomFieldValue(f))
	})

	t.Run("date renders from epoch milliseconds", func(t *testing.T) {
		ms := time.Date(2026, 8, 22, 9, 30, 0, 0, time.Local).UnixMilli()
		f := &clickup.CustomField{Type: "date", Value: float64(ms)}
		assert.Equal(t, "2026-08-22 09:30", FormatCustomFieldValue(f))

		f = &clickup.CustomField{Type: "date", Value: "1755835200000"}
		assert.Equal(t, time.UnixMilli(1755835200000).Format("2006-01-02 15:04"), FormatCustomFieldValue(f))
	})
}
