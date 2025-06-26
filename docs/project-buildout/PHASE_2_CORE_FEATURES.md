# Phase 2: Core Features

## Overview
Implement the essential ClickUp operations that provide value to developers, mirroring GitHub CLI functionality where applicable.

## Prerequisites
- [ ] Phase 1 completed and validated
- [ ] Authentication working with ClickUp API
- [ ] Output formatting system functional
- [ ] Test infrastructure in place

## Task Checklist

### 1. Task Management Commands
#### `cu task list`
- [ ] Implement basic task listing from default list
- [ ] Add filtering options:
  - [ ] `--space <name/id>`
  - [ ] `--folder <name/id>` 
  - [ ] `--list <name/id>`
  - [ ] `--assignee <username/id>`
  - [ ] `--status <status>`
  - [ ] `--tag <tag>`
  - [ ] `--priority <priority>`
  - [ ] `--due <date>` (today, tomorrow, week, overdue)
- [ ] Add sorting options:
  - [ ] `--sort <field>` (created, updated, due, priority)
  - [ ] `--order <asc/desc>`
- [ ] Implement pagination:
  - [ ] `--limit <n>` (default: 30)
  - [ ] `--page <n>`
- [ ] Add output customization:
  - [ ] `--columns <col1,col2>` for table output
  - [ ] Full JSON response with `--json`
- [ ] Cache workspace structure for name resolution

#### `cu task create`
- [ ] Interactive mode (default):
  - [ ] Prompt for title
  - [ ] Prompt for description (optional)
  - [ ] List selection (with fuzzy search)
  - [ ] Assignee selection
  - [ ] Priority selection
  - [ ] Due date input
- [ ] Non-interactive mode with flags:
  - [ ] `--title <title>` (required)
  - [ ] `--description <desc>`
  - [ ] `--list <list>`
  - [ ] `--assignee <user>`
  - [ ] `--priority <priority>`
  - [ ] `--due <date>`
  - [ ] `--tag <tag>` (multiple)
- [ ] Support markdown in description
- [ ] Return created task URL and ID
- [ ] Add `--open` flag to open in browser

#### `cu task view <task-id>`
- [ ] Display task details in readable format
- [ ] Show all fields:
  - [ ] Title, description, status
  - [ ] Assignees, watchers
  - [ ] Priority, tags, due date
  - [ ] Created/updated timestamps
  - [ ] Comments count
  - [ ] Subtasks
  - [ ] Custom fields
- [ ] Add `--comments` flag to include comments
- [ ] Add `--web` flag to open in browser
- [ ] Support both task ID and URL as input

#### `cu task update <task-id>`
- [ ] Update specific fields:
  - [ ] `--title <title>`
  - [ ] `--description <desc>`
  - [ ] `--status <status>`
  - [ ] `--assignee <user>` (add)
  - [ ] `--unassign <user>` (remove)
  - [ ] `--priority <priority>`
  - [ ] `--due <date>`
  - [ ] `--tag <tag>` (add)
  - [ ] `--remove-tag <tag>`
- [ ] Interactive mode for status change
- [ ] Validate status transitions
- [ ] Show before/after diff

#### `cu task close <task-id>`
- [ ] Change task status to closed/done
- [ ] Support custom "done" statuses
- [ ] Add `--comment <text>` for closing comment
- [ ] Bulk close with multiple IDs

#### `cu task reopen <task-id>`
- [ ] Change task status to open/todo
- [ ] Support custom "open" statuses
- [ ] Add reopening comment option

#### `cu task delete <task-id>`
- [ ] Soft delete task
- [ ] Require confirmation (unless `--force`)
- [ ] Support bulk delete
- [ ] Show deleted task summary

#### `cu task comment <task-id>`
- [ ] Add comment to task:
  - [ ] Interactive mode (editor)
  - [ ] `--message <text>` flag
  - [ ] Support markdown
  - [ ] File attachment support (future)
- [ ] List comments with `--list`
- [ ] Delete comment with `--delete <comment-id>`

### 2. List Management Commands
#### `cu list list`
- [ ] Show all lists in workspace/space
- [ ] Hierarchical display (space → folder → list)
- [ ] Filter by:
  - [ ] `--space <name/id>`
  - [ ] `--folder <name/id>`
  - [ ] `--archived` (include archived)
- [ ] Show task counts per list
- [ ] Highlight default list

#### `cu list create`
- [ ] Create new list:
  - [ ] `--name <name>` (required)
  - [ ] `--space <space>` 
  - [ ] `--folder <folder>`
  - [ ] `--status <status>` (multiple, custom statuses)
- [ ] Copy settings from template list
- [ ] Set as default after creation

#### `cu list view <list-id>`
- [ ] Show list details:
  - [ ] Name, description
  - [ ] Space/folder location  
  - [ ] Custom statuses
  - [ ] Task count by status
  - [ ] Members with access
- [ ] Include task preview

#### `cu list update <list-id>`
- [ ] Update list properties:
  - [ ] `--name <name>`
  - [ ] `--description <desc>`
  - [ ] `--add-status <status>`
  - [ ] `--remove-status <status>`
- [ ] Reorder statuses

#### `cu list archive <list-id>`
- [ ] Archive list (soft delete)
- [ ] Require confirmation
- [ ] Show archived task count

#### `cu list default <list-id>`
- [ ] Set default list for current project
- [ ] Save to `.cu.yml`
- [ ] Validate list exists and is accessible

### 3. Space & Folder Commands
#### `cu space list`
- [ ] List all spaces in workspace
- [ ] Show space details:
  - [ ] Member count
  - [ ] List count
  - [ ] Privacy settings
- [ ] Filter private/public spaces

#### `cu space view <space-id>`
- [ ] Show space details
- [ ] List folders and lists
- [ ] Show space members
- [ ] Display space settings

#### `cu folder list`
- [ ] List folders in space
- [ ] Show folder hierarchy
- [ ] Include list count
- [ ] Filter by space

#### `cu folder view <folder-id>`
- [ ] Show folder details
- [ ] List contained lists
- [ ] Show task count

### 4. User & Team Commands
#### `cu me`
- [ ] Show current user info:
  - [ ] Name, email, ID
  - [ ] Workspace membership
  - [ ] Assigned task count
  - [ ] API rate limit status
- [ ] Include workspace details

#### `cu user list`
- [ ] List workspace members
- [ ] Filter by:
  - [ ] `--space <space>`
  - [ ] `--email <email>`
  - [ ] `--role <role>`
- [ ] Show user status (active/invited)

### 5. Search Command
#### `cu search <query>`
- [ ] Search across tasks, lists, folders
- [ ] Filter by type:
  - [ ] `--type task|list|folder|doc`
- [ ] Scope search:
  - [ ] `--space <space>`
  - [ ] `--list <list>`
  - [ ] `--assignee <user>`
- [ ] Full-text search in:
  - [ ] Task titles
  - [ ] Task descriptions
  - [ ] Comments
- [ ] Sort results by relevance
- [ ] Highlight matching terms

### 6. API Passthrough Command
#### `cu api <endpoint>`
- [ ] Direct API access for advanced users
- [ ] HTTP methods:
  - [ ] `--method GET|POST|PUT|DELETE`
  - [ ] `-X GET` (short form)
- [ ] Request body:
  - [ ] `--data <json>`
  - [ ] `--data-file <file>`
- [ ] Headers:
  - [ ] `--header <key:value>`
- [ ] Show request/response with `--verbose`
- [ ] Pretty-print JSON output
- [ ] Save response to file

### 7. Shell Completion Enhancement
- [ ] Dynamic completion for:
  - [ ] Space names
  - [ ] List names
  - [ ] User names
  - [ ] Task IDs (recent)
  - [ ] Status values
- [ ] Cache completion data
- [ ] Fast completion response (<100ms)

### 8. Performance Optimizations
- [ ] Implement caching layer:
  - [ ] Workspace structure cache (1 hour)
  - [ ] User list cache (1 hour)
  - [ ] Recent tasks cache (5 minutes)
  - [ ] Cache invalidation commands
- [ ] Parallel API requests where possible
- [ ] Lazy loading for large datasets
- [ ] Progress indicators for long operations

### 9. Interactive Mode Enhancements
- [ ] Use survey/promptui for better UX:
  - [ ] Fuzzy search for selections
  - [ ] Multi-select where applicable
  - [ ] Syntax highlighting
  - [ ] Input validation
- [ ] Remember recent selections
- [ ] Keyboard shortcuts
- [ ] Cancel operation handling

### 10. Error Recovery
- [ ] Retry failed requests automatically
- [ ] Resume interrupted operations
- [ ] Offline mode for cached data
- [ ] Graceful degradation
- [ ] Clear error messages with solutions

## Testing Requirements
- [ ] Unit tests for each command
- [ ] Integration tests with mock API
- [ ] End-to-end tests with test workspace
- [ ] Performance benchmarks
- [ ] Error scenario testing
- [ ] Multi-workspace testing

## Documentation Requirements
- [ ] Command reference for each command
- [ ] Example usage scenarios
- [ ] Common workflows guide
- [ ] API mapping reference
- [ ] Troubleshooting guide

## Validation Checklist
- [ ] All task commands functional
- [ ] List management complete
- [ ] Search returns relevant results
- [ ] Output formats working (table, JSON, YAML)
- [ ] Interactive mode smooth and intuitive
- [ ] Performance targets met (<500ms for most operations)
- [ ] Error messages helpful and actionable
- [ ] Documentation complete and accurate

## Next Steps
Once Phase 2 is complete, proceed to [Phase 3: Distribution & Packaging](./PHASE_3_DISTRIBUTION.md) to prepare for release.