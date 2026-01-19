# Arandu GraphQL API

This document describes the GraphQL API for Arandu.

## Endpoint

- **HTTP**: `POST /graphql`
- **WebSocket**: `ws://host:port/graphql` (for subscriptions)
- **Playground**: `GET /playground` (development only)

## Authentication

If `REQUIRE_API_KEY=true` is set, all requests must include:

```
X-API-Key: your-api-key
```

## Custom Scalars

| Scalar | Description | Example |
|--------|-------------|---------|
| `JSON` | Arbitrary JSON data | `{"key": "value"}` |
| `Uint` | Unsigned integer | `123` |
| `Time` | ISO 8601 timestamp | `"2024-01-15T10:30:00Z"` |

## Types

### Model

Represents a configured LLM provider.

```graphql
type Model {
  provider: String!  # Provider type: openai, ollama, lmstudio, localai, openai-compatible
  id: String!        # Model identifier (e.g., "gpt-4o", "qwen2.5-coder:14b")
}
```

### Flow

A conversation session with the AI agent.

```graphql
type Flow {
  id: Uint!
  name: String!
  tasks: [Task!]!
  terminal: Terminal!
  browser: Browser!
  status: FlowStatus!
  model: Model!
}

enum FlowStatus {
  inProgress
  finished
}
```

### Task

A single action within a flow.

```graphql
type Task {
  id: Uint!
  message: String!       # Human-readable description
  createdAt: Time!
  type: TaskType!
  status: TaskStatus!
  args: JSON!            # Task-specific arguments
  results: JSON!         # Execution results
}

enum TaskType {
  input     # User message
  terminal  # Shell command execution
  browser   # Web browsing action
  code      # File read/write/patch
  ask       # Request for user input
  done      # Task completion marker
}

enum TaskStatus {
  inProgress
  finished
  stopped
  failed
}
```

### Terminal

Container terminal state.

```graphql
type Terminal {
  containerName: String!
  connected: Boolean!
  logs: [Log!]!
}

type Log {
  id: Uint!
  text: String!
}
```

### Browser

Browser automation state.

```graphql
type Browser {
  url: String!
  screenshotUrl: String!
}
```

## Queries

### availableModels

List all configured LLM models.

```graphql
query {
  availableModels {
    provider
    id
  }
}
```

**Response:**
```json
{
  "data": {
    "availableModels": [
      { "provider": "ollama", "id": "qwen2.5-coder:14b" },
      { "provider": "openai", "id": "gpt-4o" }
    ]
  }
}
```

### flows

List all conversation flows.

```graphql
query {
  flows {
    id
    name
    status
    model {
      provider
      id
    }
  }
}
```

### flow

Get a single flow with full details.

```graphql
query GetFlow($id: Uint!) {
  flow(id: $id) {
    id
    name
    status
    model {
      provider
      id
    }
    tasks {
      id
      type
      message
      status
      args
      results
      createdAt
    }
    terminal {
      containerName
      connected
      logs {
        id
        text
      }
    }
    browser {
      url
      screenshotUrl
    }
  }
}
```

## Mutations

### createFlow

Start a new conversation with a specific model.

```graphql
mutation CreateFlow($modelProvider: String!, $modelId: String!) {
  createFlow(modelProvider: $modelProvider, modelId: $modelId) {
    id
    name
    status
    model {
      provider
      id
    }
  }
}
```

**Variables:**
```json
{
  "modelProvider": "ollama",
  "modelId": "qwen2.5-coder:14b"
}
```

### createTask

Send a user message to start task processing.

```graphql
mutation CreateTask($flowId: Uint!, $query: String!) {
  createTask(flowId: $flowId, query: $query) {
    id
    type
    message
    status
  }
}
```

**Variables:**
```json
{
  "flowId": 1,
  "query": "Create a Python script that prints Hello World"
}
```

### finishFlow

End a conversation and clean up resources.

```graphql
mutation FinishFlow($flowId: Uint!) {
  finishFlow(flowId: $flowId) {
    id
    status
  }
}
```

## Subscriptions

All subscriptions require a `flowId` parameter and return real-time updates.

### taskAdded

Notifies when a new task is created.

```graphql
subscription OnTaskAdded($flowId: Uint!) {
  taskAdded(flowId: $flowId) {
    id
    type
    message
    status
    args
  }
}
```

### taskUpdated

Notifies when a task status or results change.

```graphql
subscription OnTaskUpdated($flowId: Uint!) {
  taskUpdated(flowId: $flowId) {
    id
    status
    results
  }
}
```

### flowUpdated

Notifies when flow status changes.

```graphql
subscription OnFlowUpdated($flowId: Uint!) {
  flowUpdated(flowId: $flowId) {
    id
    status
  }
}
```

### browserUpdated

Notifies when browser takes a new screenshot.

```graphql
subscription OnBrowserUpdated($flowId: Uint!) {
  browserUpdated(flowId: $flowId) {
    url
    screenshotUrl
  }
}
```

### terminalLogsAdded

Notifies when new terminal output is available.

```graphql
subscription OnTerminalLogs($flowId: Uint!) {
  terminalLogsAdded(flowId: $flowId) {
    id
    text
  }
}
```

## Task Arguments

Each task type has specific arguments stored in the `args` field.

### Terminal Task

```json
{
  "command": "ls -la",
  "message": "List files in current directory"
}
```

### Browser Task

```json
{
  "url": "https://example.com",
  "action": "read",  // or "url" to extract links
  "message": "Reading documentation"
}
```

### Code Task

```json
{
  "action": "read",   // read, write, or patch
  "path": "/app/main.py",
  "content": "...",   // for write/patch
  "message": "Reading main.py"
}
```

### Ask Task

```json
{
  "question": "What programming language should I use?",
  "message": "Asking user for clarification"
}
```

### Done Task

```json
{
  "summary": "Created Python script with Hello World output",
  "message": "Task completed successfully"
}
```

## Error Handling

Errors are returned in the standard GraphQL format:

```json
{
  "errors": [
    {
      "message": "flow not found",
      "path": ["flow"],
      "extensions": {
        "code": "NOT_FOUND"
      }
    }
  ]
}
```

## WebSocket Protocol

Subscriptions use the `graphql-ws` protocol. Connect with:

```javascript
import { createClient } from 'graphql-ws';

const client = createClient({
  url: 'ws://localhost:8080/graphql',
});
```
