package cmd

import (
	"context"
	"fmt"

	"github.com/raksul/go-clickup/clickup"
	"github.com/spf13/cobra"
	"github.com/timimsms/cu/internal/api"
	"github.com/timimsms/cu/internal/config"
	"github.com/timimsms/cu/internal/output"
)

// fieldRow is the rendered shape of a custom field, shared by every output
// format so table, json, yaml and csv agree on the columns.
type fieldRow struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	ID    string `json:"id"`
	Value string `json:"value"`
}

func newFieldRow(f *clickup.CustomField) fieldRow {
	return fieldRow{
		Name:  f.Name,
		Type:  f.Type,
		ID:    f.ID,
		Value: api.FormatCustomFieldValue(f),
	}
}

// fieldWriteResult reports what a set or clear actually wrote, so a scripted
// caller can confirm which field was resolved rather than re-reading the task.
type fieldWriteResult struct {
	TaskID  string `json:"task_id"`
	Field   string `json:"field"`
	FieldID string `json:"field_id"`
	Value   string `json:"value,omitempty"`
	Cleared bool   `json:"cleared"`
}

var fieldCmd = &cobra.Command{
	Use:   "field",
	Short: "Manage custom field values",
	Long: `Read and write ClickUp custom fields.

Fields are resolved by name (case-insensitive) or by id. Values are coerced to
the shape each field type expects — epoch milliseconds for dates, option ids
for dropdowns and labels — so names can be used on the command line.`,
}

var fieldListCmd = &cobra.Command{
	Use:   "list",
	Short: "List custom fields available on a list",
	RunE: func(cmd *cobra.Command, args []string) error {
		listID, _ := cmd.Flags().GetString("list")
		if listID == "" {
			listID = config.GetString("default_list")
		}
		if listID == "" {
			return fmt.Errorf("no list specified: pass --list or set a default with 'cu list default <list-id>'")
		}

		client, err := api.NewClient()
		if err != nil {
			return err
		}

		fields, err := client.GetCustomFields(context.Background(), listID)
		if err != nil {
			return err
		}

		rows := make([]fieldRow, 0, len(fields))
		for i := range fields {
			rows = append(rows, newFieldRow(&fields[i]))
		}
		return output.Format(outputFormat, rows)
	},
}

var fieldGetCmd = &cobra.Command{
	Use:   "get <task-id> [field]",
	Short: "Show custom field values on a task",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.NewClient()
		if err != nil {
			return err
		}

		task, err := client.GetTask(context.Background(), args[0])
		if err != nil {
			return err
		}

		fields := task.CustomFields
		if len(args) == 2 {
			f, err := api.FindCustomField(fields, args[1])
			if err != nil {
				return err
			}
			fields = []clickup.CustomField{*f}
		}

		rows := make([]fieldRow, 0, len(fields))
		for i := range fields {
			rows = append(rows, newFieldRow(&fields[i]))
		}
		return output.Format(outputFormat, rows)
	},
}

var fieldSetCmd = &cobra.Command{
	Use:   "set <task-id> <field> <value>",
	Short: "Set a custom field value on a task",
	Long: `Set a custom field value on a task.

The field is resolved against the fields accessible from the task's own list,
so space-level and folder-level fields work the same as list-level ones.

Examples:
  cu field set 86dxbeqyt Repo git@github.com:owner/repo.git
  cu field set 86dxbeqyt "Last synced" today
  cu field set 86dxbeqyt Machine "Chilastra,Sunrunner"`,
	Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID, name, raw := args[0], args[1], args[2]

		client, err := api.NewClient()
		if err != nil {
			return err
		}
		ctx := context.Background()

		field, err := resolveTaskField(ctx, client, taskID, name)
		if err != nil {
			return err
		}

		value, err := api.CoerceCustomFieldValue(field, raw)
		if err != nil {
			return err
		}

		if err := client.SetCustomFieldValue(ctx, taskID, field.ID, value); err != nil {
			return err
		}

		if outputFormat != "table" {
			return output.Format(outputFormat, fieldWriteResult{
				TaskID: taskID, Field: field.Name, FieldID: field.ID, Value: raw, Cleared: false,
			})
		}

		fmt.Printf("Set %s on task %s\n", field.Name, taskID)
		return nil
	},
}

var fieldClearCmd = &cobra.Command{
	Use:   "clear <task-id> <field>",
	Short: "Clear a custom field value on a task",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID, name := args[0], args[1]

		client, err := api.NewClient()
		if err != nil {
			return err
		}
		ctx := context.Background()

		field, err := resolveTaskField(ctx, client, taskID, name)
		if err != nil {
			return err
		}

		if err := client.RemoveCustomFieldValue(ctx, taskID, field.ID); err != nil {
			return err
		}

		if outputFormat != "table" {
			return output.Format(outputFormat, fieldWriteResult{
				TaskID: taskID, Field: field.Name, FieldID: field.ID, Cleared: true,
			})
		}

		fmt.Printf("Cleared %s on task %s\n", field.Name, taskID)
		return nil
	},
}

// resolveTaskField finds a field by name or id using the fields accessible
// from the task's own list, which is what makes space-level fields resolvable
// without the caller knowing where the field was defined.
func resolveTaskField(ctx context.Context, client *api.Client, taskID, nameOrID string) (*clickup.CustomField, error) {
	task, err := client.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if task.List.ID == "" {
		return nil, fmt.Errorf("could not determine the list for task %s", taskID)
	}

	fields, err := client.GetCustomFields(ctx, task.List.ID)
	if err != nil {
		return nil, err
	}
	return api.FindCustomField(fields, nameOrID)
}

func init() {
	fieldListCmd.Flags().StringP("list", "l", "", "List ID (defaults to the configured default list)")

	fieldCmd.AddCommand(fieldListCmd)
	fieldCmd.AddCommand(fieldGetCmd)
	fieldCmd.AddCommand(fieldSetCmd)
	fieldCmd.AddCommand(fieldClearCmd)
}
