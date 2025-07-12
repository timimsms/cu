package output

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTableFormatter(t *testing.T) {
	t.Run("TableFormatter formats data", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := &TableFormatter{Writer: &buf}
		
		data := []map[string]string{
			{"id": "123", "name": "test"},
		}
		
		err := formatter.Format(data)
		assert.NoError(t, err)
		assert.NotEmpty(t, buf.String())
	})
}

func TestJSONFormatter(t *testing.T) {
	t.Run("JSONFormatter formats data", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := &JSONFormatter{Writer: &buf}
		data := map[string]string{"id": "123", "name": "test"}
		
		err := formatter.Format(data)
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "123")
		assert.Contains(t, buf.String(), "test")
	})
}

func TestYAMLFormatter(t *testing.T) {
	t.Run("YAMLFormatter formats data", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := &YAMLFormatter{Writer: &buf}
		data := map[string]string{"id": "123", "name": "test"}
		
		err := formatter.Format(data)
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "123")
		assert.Contains(t, buf.String(), "test")
	})
}

func TestFormat(t *testing.T) {
	tests := []struct {
		name       string
		format     string
		shouldFail bool
	}{
		{"json formatter", "json", false},
		{"yaml formatter", "yaml", false},
		{"table formatter", "table", false},
		{"csv formatter", "csv", false},
		{"invalid formatter", "invalid", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			old := os.Stdout
			_, w, _ := os.Pipe()
			os.Stdout = w
			
			data := map[string]string{"id": "123", "name": "test"}
			err := Format(tt.format, data)
			
			// Restore stdout
			_ = w.Close()
			os.Stdout = old
			
			if tt.shouldFail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCSVFormatter(t *testing.T) {
	t.Run("CSVFormatter formats slice data", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := &CSVFormatter{Writer: &buf}
		
		data := []map[string]string{
			{"id": "123", "name": "test"},
			{"id": "456", "name": "test2"},
		}
		
		err := formatter.Format(data)
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "123")
		assert.Contains(t, buf.String(), "test")
	})
}