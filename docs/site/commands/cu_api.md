## cu api

Make direct API requests to ClickUp

### Synopsis

Make direct API requests to the ClickUp API.

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
For example, use "/team" for https://api.clickup.com/api/v2/team

```
cu api <endpoint> [flags]
```

### Options

```
  -d, --data string          Request body data (JSON)
  -H, --header stringArray   Custom headers (format: 'Header: value')
  -h, --help                 help for api
  -X, --method string        HTTP method (GET, POST, PUT, PATCH, DELETE) (default "GET")
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.config/cu/config.yml)
      --debug           enable debug mode
  -o, --output string   output format (table|json|yaml|csv) (default "table")
```

### SEE ALSO

* [cu](cu.md)	 - A GitHub CLI-inspired command-line interface for ClickUp

###### Auto generated by spf13/cobra on 29-Jun-2025
