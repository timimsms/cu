package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tim/cu/internal/auth"
	"github.com/tim/cu/internal/output"
)

var (
	apiMethod string
	apiData   string
	apiHeaders []string
)

var apiCmd = &cobra.Command{
	Use:   "api <endpoint>",
	Short: "Make direct API requests to ClickUp",
	Long: `Make direct API requests to the ClickUp API.

This command provides direct access to any ClickUp API endpoint,
useful for operations not yet implemented in the CLI or for
advanced use cases.

Examples:
  # Get all workspaces
  cu api /team

  # Get a specific task
  cu api /task/abc123

  # Create a new task (with data)
  cu api /list/def456/task -X POST -d '{"name": "New Task"}'

  # Update task with PATCH
  cu api /task/abc123 -X PATCH -d '{"name": "Updated Name"}'

  # Delete a task
  cu api /task/abc123 -X DELETE

  # Get tasks with query parameters
  cu api "/list/def456/task?archived=false&page=0"

  # Get custom fields for a list
  cu api /list/def456/field

  # Pass custom headers
  cu api /team -H "X-Custom-Header: value"

The endpoint should be the path after https://api.clickup.com/api/v2/
For example, use "/team" for https://api.clickup.com/api/v2/team`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		endpoint := args[0]
		
		// Ensure endpoint starts with /
		if !strings.HasPrefix(endpoint, "/") {
			endpoint = "/" + endpoint
		}

		// Get authentication token
		authMgr := auth.NewManager()
		token, err := authMgr.GetCurrentToken()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Not authenticated. Run 'cu auth login' first.\n")
			os.Exit(1)
		}

		// Build full URL
		baseURL := "https://api.clickup.com/api/v2"
		fullURL := baseURL + endpoint

		// Prepare request body if data provided
		var body io.Reader
		if apiData != "" {
			// Validate JSON if content-type suggests it
			if strings.Contains(apiData, "{") || strings.Contains(apiData, "[") {
				var js json.RawMessage
				if err := json.Unmarshal([]byte(apiData), &js); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: data does not appear to be valid JSON: %v\n", err)
				}
			}
			body = bytes.NewBufferString(apiData)
		}

		// Create request
		req, err := http.NewRequest(apiMethod, fullURL, body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create request: %v\n", err)
			os.Exit(1)
		}

		// Set headers
		req.Header.Set("Authorization", token.Value)
		req.Header.Set("Content-Type", "application/json")
		
		// Add custom headers
		for _, header := range apiHeaders {
			parts := strings.SplitN(header, ":", 2)
			if len(parts) == 2 {
				req.Header.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
			}
		}

		// Make request
		client := &http.Client{
			Timeout: 30 * time.Second,
		}
		
		resp, err := client.Do(req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Request failed: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		// Read response
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read response: %v\n", err)
			os.Exit(1)
		}

		// Check for non-2xx status codes
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			fmt.Fprintf(os.Stderr, "API request failed with status %d: %s\n", resp.StatusCode, resp.Status)
			
			// Try to parse error response
			var errResp map[string]interface{}
			if err := json.Unmarshal(respBody, &errResp); err == nil {
				if msg, ok := errResp["err"].(string); ok {
					fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
				} else if msg, ok := errResp["error"].(string); ok {
					fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
				} else {
					// Output the full error response
					output.Format("json", errResp)
				}
			} else {
				fmt.Fprintf(os.Stderr, "Response: %s\n", string(respBody))
			}
			os.Exit(1)
		}

		// Parse response as JSON
		var result interface{}
		if err := json.Unmarshal(respBody, &result); err != nil {
			// If not JSON, output as string
			fmt.Println(string(respBody))
			return
		}

		// Format output
		format := cmd.Flag("output").Value.String()
		if err := output.Format(format, result); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	apiCmd.Flags().StringVarP(&apiMethod, "method", "X", "GET", "HTTP method (GET, POST, PUT, PATCH, DELETE)")
	apiCmd.Flags().StringVarP(&apiData, "data", "d", "", "Request body data (JSON)")
	apiCmd.Flags().StringArrayVarP(&apiHeaders, "header", "H", []string{}, "Custom headers (format: 'Header: value')")
}