package api

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/raksul/go-clickup/clickup"
)

// GetCustomFields returns the custom fields accessible from a list. Fields
// created at space or folder level are included, which is why the same field
// id comes back for every list beneath that location.
func (c *Client) GetCustomFields(ctx context.Context, listID string) ([]clickup.CustomField, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	fields, _, err := c.client.CustomFields.GetAccessibleCustomFields(ctx, listID)
	if err != nil {
		return nil, c.handleError(err)
	}
	return fields, nil
}

// SetCustomFieldValue writes a single custom field value on a task. The value
// map is the request body, e.g. {"value": "https://github.com/owner/repo"} —
// build it with CoerceCustomFieldValue rather than by hand.
func (c *Client) SetCustomFieldValue(ctx context.Context, taskID string, fieldID string, value map[string]interface{}) error {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return err
	}

	if _, err := c.client.CustomFields.SetCustomFieldValue(ctx, taskID, fieldID, value, &clickup.CustomFieldOptions{}); err != nil {
		return c.handleError(err)
	}
	return nil
}

// RemoveCustomFieldValue clears a custom field value on a task.
func (c *Client) RemoveCustomFieldValue(ctx context.Context, taskID string, fieldID string) error {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return err
	}

	if _, err := c.client.CustomFields.RemoveCustomFieldValue(ctx, taskID, fieldID, &clickup.CustomFieldOptions{}); err != nil {
		return c.handleError(err)
	}
	return nil
}

// FindCustomField resolves a field by id, or by name case-insensitively.
// An ambiguous name is an error rather than a silent first-match.
func FindCustomField(fields []clickup.CustomField, nameOrID string) (*clickup.CustomField, error) {
	var matches []clickup.CustomField
	for _, f := range fields {
		if f.ID == nameOrID {
			return &f, nil
		}
		if strings.EqualFold(f.Name, nameOrID) {
			matches = append(matches, f)
		}
	}

	switch len(matches) {
	case 0:
		names := make([]string, 0, len(fields))
		for _, f := range fields {
			names = append(names, f.Name)
		}
		return nil, fmt.Errorf("no custom field %q (available: %s)", nameOrID, strings.Join(names, ", "))
	case 1:
		return &matches[0], nil
	default:
		return nil, fmt.Errorf("custom field name %q is ambiguous; use the field id", nameOrID)
	}
}

// fieldOptions extracts the {id, name|label} option list from a field's
// type_config, which the SDK models as an untyped interface{}.
func fieldOptions(f *clickup.CustomField) []map[string]interface{} {
	cfg, ok := f.TypeConfig.(map[string]interface{})
	if !ok {
		return nil
	}
	raw, ok := cfg["options"].([]interface{})
	if !ok {
		return nil
	}

	out := make([]map[string]interface{}, 0, len(raw))
	for _, o := range raw {
		if m, ok := o.(map[string]interface{}); ok {
			out = append(out, m)
		}
	}
	return out
}

func optionLabel(o map[string]interface{}) string {
	// Dropdowns use "name", labels use "label".
	for _, k := range []string{"name", "label"} {
		if s, ok := o[k].(string); ok && s != "" {
			return s
		}
	}
	return ""
}

// resolveOption maps an option name (or id) to its id.
func resolveOption(f *clickup.CustomField, input string) (string, error) {
	opts := fieldOptions(f)
	labels := make([]string, 0, len(opts))

	for _, o := range opts {
		id, _ := o["id"].(string)
		label := optionLabel(o)
		labels = append(labels, label)
		if id == input || strings.EqualFold(label, input) {
			return id, nil
		}
	}
	return "", fmt.Errorf("no option %q on field %q (available: %s)", input, f.Name, strings.Join(labels, ", "))
}

// CoerceCustomFieldValue turns a command-line string into the request body
// ClickUp expects for that field's type. Types differ enough — epoch
// milliseconds for dates, option uuids for labels — that passing a raw string
// through would silently write the wrong thing.
func CoerceCustomFieldValue(f *clickup.CustomField, raw string) (map[string]interface{}, error) {
	switch f.Type {
	case "url", "text", "short_text", "email", "phone", "location":
		return map[string]interface{}{"value": raw}, nil

	case "number":
		n, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return nil, fmt.Errorf("field %q expects a number, got %q", f.Name, raw)
		}
		return map[string]interface{}{"value": n}, nil

	case "checkbox":
		b, err := strconv.ParseBool(raw)
		if err != nil {
			return nil, fmt.Errorf("field %q expects true or false, got %q", f.Name, raw)
		}
		return map[string]interface{}{"value": b}, nil

	case "date":
		ms, err := parseDateMillis(raw)
		if err != nil {
			return nil, fmt.Errorf("field %q: %w", f.Name, err)
		}
		return map[string]interface{}{"value": ms}, nil

	case "drop_down":
		id, err := resolveOption(f, raw)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"value": id}, nil

	case "labels":
		// Labels are multi-select: accept a comma-separated list of names.
		parts := strings.Split(raw, ",")
		ids := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			id, err := resolveOption(f, p)
			if err != nil {
				return nil, err
			}
			ids = append(ids, id)
		}
		return map[string]interface{}{"value": ids}, nil

	default:
		// Unknown or unsupported types (formula, rollup, progress…) are not
		// guessed at — a wrong write is worse than a clear refusal.
		return nil, fmt.Errorf("setting fields of type %q is not supported", f.Type)
	}
}

// parseDateMillis accepts epoch milliseconds, an ISO date, an RFC3339
// timestamp, or "today"/"now".
func parseDateMillis(raw string) (int64, error) {
	raw = strings.TrimSpace(raw)

	switch strings.ToLower(raw) {
	case "now":
		return time.Now().UnixMilli(), nil
	case "today":
		y, m, d := time.Now().Date()
		return time.Date(y, m, d, 0, 0, 0, 0, time.Local).UnixMilli(), nil
	}

	if ms, err := strconv.ParseInt(raw, 10, 64); err == nil {
		return ms, nil
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02 15:04", "2006-01-02"} {
		if t, err := time.ParseInLocation(layout, raw, time.Local); err == nil {
			return t.UnixMilli(), nil
		}
	}
	return 0, fmt.Errorf("cannot parse %q as a date (try 2006-01-02, an RFC3339 timestamp, epoch milliseconds, today, or now)", raw)
}

// FormatCustomFieldValue renders a field's current value for display.
func FormatCustomFieldValue(f *clickup.CustomField) string {
	if f.Value == nil {
		return ""
	}

	switch f.Type {
	case "labels":
		vals, ok := f.Value.([]interface{})
		if !ok {
			return fmt.Sprintf("%v", f.Value)
		}
		out := make([]string, 0, len(vals))
		for _, v := range vals {
			id, _ := v.(string)
			label := id
			for _, o := range fieldOptions(f) {
				if oid, _ := o["id"].(string); oid == id {
					if l := optionLabel(o); l != "" {
						label = l
					}
					break
				}
			}
			out = append(out, label)
		}
		return strings.Join(out, ", ")

	case "drop_down":
		id := fmt.Sprintf("%v", f.Value)
		for _, o := range fieldOptions(f) {
			if oid, _ := o["id"].(string); oid == id {
				if l := optionLabel(o); l != "" {
					return l
				}
			}
		}
		return id

	case "date":
		switch v := f.Value.(type) {
		case string:
			if ms, err := strconv.ParseInt(v, 10, 64); err == nil {
				return time.UnixMilli(ms).Format("2006-01-02 15:04")
			}
			return v
		case float64:
			return time.UnixMilli(int64(v)).Format("2006-01-02 15:04")
		}
	}

	return fmt.Sprintf("%v", f.Value)
}
