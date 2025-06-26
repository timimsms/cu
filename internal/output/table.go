package output

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"text/tabwriter"
	"time"
)

// TableFormatter formats output as a table
type TableFormatter struct {
	Writer       io.Writer
	NoHeader     bool
	Columns      []string
	ShowEmpty    bool
	ColorEnabled bool
}

func (f *TableFormatter) Format(data interface{}) error {
	if f.Writer == nil {
		f.Writer = os.Stdout
	}

	// Create tab writer for aligned columns
	w := tabwriter.NewWriter(f.Writer, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// Handle different data types
	rv := reflect.ValueOf(data)
	switch rv.Kind() {
	case reflect.Slice:
		return f.formatSlice(w, rv)
	case reflect.Map:
		return f.formatMap(w, rv)
	case reflect.Struct:
		return f.formatStruct(w, rv)
	default:
		// For simple types, just print the value
		fmt.Fprintln(w, data)
		return nil
	}
}

func (f *TableFormatter) formatSlice(w io.Writer, rv reflect.Value) error {
	if rv.Len() == 0 {
		if f.ShowEmpty {
			fmt.Fprintln(w, "No items found")
		}
		return nil
	}

	// Get headers from the first item
	first := rv.Index(0)
	headers, err := f.getHeaders(first.Interface())
	if err != nil {
		return err
	}

	// Print headers
	if !f.NoHeader && len(headers) > 0 {
		fmt.Fprintln(w, strings.Join(headers, "\t"))
		// Print separator
		var sep []string
		for range headers {
			sep = append(sep, strings.Repeat("-", 10))
		}
		fmt.Fprintln(w, strings.Join(sep, "\t"))
	}

	// Print rows
	for i := 0; i < rv.Len(); i++ {
		row, err := f.getRow(rv.Index(i).Interface(), headers)
		if err != nil {
			return err
		}
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}

	return nil
}

func (f *TableFormatter) formatMap(w io.Writer, rv reflect.Value) error {
	if !f.NoHeader {
		fmt.Fprintln(w, "KEY\tVALUE")
		fmt.Fprintln(w, "---\t-----")
	}

	for _, key := range rv.MapKeys() {
		value := rv.MapIndex(key)
		fmt.Fprintf(w, "%v\t%v\n", key.Interface(), f.formatValue(value.Interface()))
	}

	return nil
}

func (f *TableFormatter) formatStruct(w io.Writer, rv reflect.Value) error {
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	rt := rv.Type()
	if !f.NoHeader {
		fmt.Fprintln(w, "FIELD\tVALUE")
		fmt.Fprintln(w, "-----\t-----")
	}

	for i := 0; i < rv.NumField(); i++ {
		field := rt.Field(i)
		value := rv.Field(i)
		
		// Skip unexported fields
		if field.PkgPath != "" {
			continue
		}

		// Use json tag if available
		name := field.Name
		if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
			parts := strings.Split(tag, ",")
			if parts[0] != "" {
				name = parts[0]
			}
		}

		fmt.Fprintf(w, "%s\t%v\n", name, f.formatValue(value.Interface()))
	}

	return nil
}

func (f *TableFormatter) getHeaders(item interface{}) ([]string, error) {
	if f.Columns != nil && len(f.Columns) > 0 {
		return f.Columns, nil
	}

	var headers []string
	rv := reflect.ValueOf(item)
	
	switch rv.Kind() {
	case reflect.Map:
		// For maps, use specified columns or all keys
		for _, key := range rv.MapKeys() {
			headers = append(headers, fmt.Sprint(key.Interface()))
		}
	case reflect.Struct, reflect.Ptr:
		if rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
		}
		rt := rv.Type()
		for i := 0; i < rt.NumField(); i++ {
			field := rt.Field(i)
			// Skip unexported fields
			if field.PkgPath != "" {
				continue
			}
			// Use json tag if available
			name := field.Name
			if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
				parts := strings.Split(tag, ",")
				if parts[0] != "" {
					name = parts[0]
				}
			}
			headers = append(headers, name)
		}
	default:
		return nil, fmt.Errorf("unsupported type for table headers: %T", item)
	}

	return headers, nil
}

func (f *TableFormatter) getRow(item interface{}, headers []string) ([]string, error) {
	var row []string
	rv := reflect.ValueOf(item)

	switch rv.Kind() {
	case reflect.Map:
		for _, header := range headers {
			key := reflect.ValueOf(header)
			value := rv.MapIndex(key)
			if value.IsValid() {
				row = append(row, f.formatValue(value.Interface()))
			} else {
				row = append(row, "")
			}
		}
	case reflect.Struct, reflect.Ptr:
		if rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
		}
		rt := rv.Type()
		for i := 0; i < rt.NumField(); i++ {
			field := rt.Field(i)
			// Skip unexported fields
			if field.PkgPath != "" {
				continue
			}
			value := rv.Field(i)
			row = append(row, f.formatValue(value.Interface()))
		}
	default:
		return nil, fmt.Errorf("unsupported type for table row: %T", item)
	}

	return row, nil
}

func (f *TableFormatter) formatValue(v interface{}) string {
	if v == nil {
		return ""
	}

	switch val := v.(type) {
	case time.Time:
		if val.IsZero() {
			return ""
		}
		// Format as relative time if recent
		now := time.Now()
		diff := now.Sub(val)
		if diff < 24*time.Hour && diff > -24*time.Hour {
			return formatRelativeTime(val)
		}
		return val.Format("2006-01-02")
	case *time.Time:
		if val == nil || val.IsZero() {
			return ""
		}
		return f.formatValue(*val)
	case []string:
		return strings.Join(val, ", ")
	case []interface{}:
		var items []string
		for _, item := range val {
			items = append(items, fmt.Sprint(item))
		}
		return strings.Join(items, ", ")
	default:
		s := fmt.Sprint(v)
		// Truncate long strings
		if len(s) > 50 {
			s = s[:47] + "..."
		}
		return s
	}
}

func formatRelativeTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < -time.Hour:
		return fmt.Sprintf("in %d hours", int(-diff.Hours()))
	case diff < -time.Minute:
		return fmt.Sprintf("in %d minutes", int(-diff.Minutes()))
	case diff < 0:
		return "in a moment"
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		return fmt.Sprintf("%d minutes ago", int(diff.Minutes()))
	case diff < 24*time.Hour:
		return fmt.Sprintf("%d hours ago", int(diff.Hours()))
	default:
		return t.Format("2006-01-02")
	}
}