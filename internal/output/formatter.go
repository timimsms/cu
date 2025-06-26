package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

// Formatter is an interface for output formatters
type Formatter interface {
	Format(data interface{}) error
}

// Format formats and prints data according to the specified format
func Format(format string, data interface{}) error {
	var formatter Formatter
	
	switch strings.ToLower(format) {
	case "json":
		formatter = &JSONFormatter{Writer: os.Stdout}
	case "yaml", "yml":
		formatter = &YAMLFormatter{Writer: os.Stdout}
	case "csv":
		formatter = &CSVFormatter{Writer: os.Stdout}
	case "table":
		formatter = &TableFormatter{Writer: os.Stdout}
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
	
	return formatter.Format(data)
}

// JSONFormatter formats output as JSON
type JSONFormatter struct {
	Writer io.Writer
}

func (f *JSONFormatter) Format(data interface{}) error {
	encoder := json.NewEncoder(f.Writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// YAMLFormatter formats output as YAML
type YAMLFormatter struct {
	Writer io.Writer
}

func (f *YAMLFormatter) Format(data interface{}) error {
	encoder := yaml.NewEncoder(f.Writer)
	defer encoder.Close()
	return encoder.Encode(data)
}

// CSVFormatter formats output as CSV
type CSVFormatter struct {
	Writer io.Writer
}

func (f *CSVFormatter) Format(data interface{}) error {
	writer := csv.NewWriter(f.Writer)
	defer writer.Flush()

	// Handle different data types
	switch v := data.(type) {
	case [][]string:
		return writer.WriteAll(v)
	case []map[string]interface{}:
		if len(v) == 0 {
			return nil
		}
		// Extract headers
		var headers []string
		for k := range v[0] {
			headers = append(headers, k)
		}
		if err := writer.Write(headers); err != nil {
			return err
		}
		// Write rows
		for _, row := range v {
			var values []string
			for _, h := range headers {
				values = append(values, fmt.Sprint(row[h]))
			}
			if err := writer.Write(values); err != nil {
				return err
			}
		}
		return nil
	default:
		// Try to convert to slice of maps using reflection
		rv := reflect.ValueOf(data)
		if rv.Kind() == reflect.Slice {
			var rows [][]string
			for i := 0; i < rv.Len(); i++ {
				item := rv.Index(i).Interface()
				row, err := structToSlice(item)
				if err != nil {
					return err
				}
				rows = append(rows, row)
			}
			return writer.WriteAll(rows)
		}
		return fmt.Errorf("unsupported data type for CSV output")
	}
}

func structToSlice(v interface{}) ([]string, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	
	var result []string
	switch rv.Kind() {
	case reflect.Struct:
		for i := 0; i < rv.NumField(); i++ {
			result = append(result, fmt.Sprint(rv.Field(i).Interface()))
		}
	case reflect.Map:
		for _, key := range rv.MapKeys() {
			result = append(result, fmt.Sprint(rv.MapIndex(key).Interface()))
		}
	default:
		result = append(result, fmt.Sprint(v))
	}
	return result, nil
}