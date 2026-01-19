# Arandu Frontend

React-based web interface for the Arandu AI coding assistant.

## Tech Stack

- **React 18** - UI framework
- **TypeScript** - Type safety
- **Vite** - Build tool and dev server
- **urql** - GraphQL client with normalized caching
- **vanilla-extract** - CSS-in-TypeScript styling
- **xterm.js** - Terminal emulation

## Project Structure

```
src/
├── components/         # Reusable UI components
│   ├── Browser/       # Web browser preview
│   ├── Button/        # Button component
│   ├── Dropdown/      # Dropdown menu
│   ├── Icon/          # SVG icons
│   ├── Messages/      # Chat message display
│   ├── Panel/         # Panel container
│   ├── Sidebar/       # Navigation sidebar
│   ├── Tabs/          # Tab navigation
│   ├── Terminal/      # Terminal emulator
│   └── Tooltip/       # Tooltip component
├── hooks/             # Custom React hooks
├── layouts/           # Page layouts
├── pages/             # Page components
│   └── ChatPage/      # Main chat interface
├── styles/            # Global styles and theme
├── App.tsx            # Root component
├── graphql.ts         # GraphQL client setup
└── main.tsx           # Entry point

generated/             # Auto-generated GraphQL types
```

## Development

### Prerequisites

- Node.js 22+
- Yarn

### Setup

```bash
# Install dependencies
yarn install

# Copy environment file
cp .env.example .env.local

# Edit .env.local with your backend URL
# VITE_API_URL=localhost:8080
```

### Commands

```bash
# Start development server
yarn dev

# Build for production
yarn build

# Preview production build
yarn preview

# Run linter
yarn lint

# Format code
yarn format:fix

# Generate GraphQL types
yarn codegen

# Run tests
yarn test

# Run tests with coverage
yarn test:coverage
```

## GraphQL

The frontend uses urql for GraphQL operations with:

- **graphcache** - Normalized caching with automatic updates
- **WebSocket subscriptions** - Real-time updates for tasks and terminal

### Generating Types

GraphQL types are generated from the backend schema:

```bash
yarn codegen
```

This reads `../backend/graph/schema.graphqls` and generates:
- `generated/graphql.ts` - TypeScript types and urql hooks
- `generated/graphql.schema.json` - Schema introspection

### Available Operations

**Queries:**
- `availableModels` - List configured LLM models
- `flows` - List all conversation flows
- `flow(id)` - Get single flow with tasks

**Mutations:**
- `createFlow` - Start new conversation
- `createTask` - Send user message
- `finishFlow` - End conversation

**Subscriptions:**
- `taskAdded` / `taskUpdated` - Task changes
- `browserUpdated` - Browser screenshots
- `terminalLogsAdded` - Terminal output

## Styling

Uses vanilla-extract for type-safe CSS:

```typescript
// Button.css.ts
import { style } from '@vanilla-extract/css';

export const button = style({
  padding: '8px 16px',
  borderRadius: '4px',
});
```

Theme variables are defined in `src/styles/theme.css.ts`.

## Path Aliases

Configured in `vite.config.ts`:

- `@/` → `./src`
- `@/generated` → `./generated`

Example:
```typescript
import { useFlowData } from '@/hooks';
import { FlowQuery } from '@/generated/graphql';
```
