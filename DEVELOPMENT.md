# Development

## Prerequisites
- Go 1.24+
- Node.js 22+
- Docker
- Ollama (optional, for local models)

## Environment variables
First, run `cp ./backend/.env.example ./backend/.env && cp ./frontend/.env.example ./frontend/.env.local` to generate the env files for both backend and frontend.

### Backend
Edit the `.env` file in `backend` folder

**Required (choose one LLM provider):**
- `OPEN_AI_KEY` - OpenAI API key (for OpenAI provider)
- `OLLAMA_MODEL` - Ollama model name (for local Ollama provider, recommended)

**Optional:**
- `PORT` - Port to run the server (default: `8080`)
- `DATABASE_URL` - SQLite database file (default: `database.db`)
- `OPEN_AI_MODEL` - OpenAI model (default: `gpt-4o`)
- `OLLAMA_SERVER_URL` - Ollama server URL (default: `http://localhost:11434`)
- `DOCKER_HOST` - Docker SDK API (eg. `DOCKER_HOST=unix:///Users/<my-user>/Library/Containers/com.docker.docker/Data/docker.raw.sock`) [more info](https://stackoverflow.com/a/62757128/5922857)

See [backend/.env.example](./backend/.env.example) for all configuration options including LM Studio, LocalAI, and other OpenAI-compatible providers.

### Frontend
Edit the `.env.local` file in `frontend` folder
- `VITE_API_URL` - Backend API URL. *Omit* the URL scheme (e.g., `localhost:8080` *NOT* `http://localhost:8080`).

## Steps

### Backend
Run the command(s) in `backend` folder:
```bash
# Install dependencies
go mod download

# Run the server
go run .
```

>The first run can be a long wait because the dependencies and the docker images need to be download to setup the backend environment.
When you see output below, the server has started successfully:
```
<your-date> <your-time> connect to http://localhost:<your-port>/playground for GraphQL playground
```

### Frontend
Run the command(s) in `frontend` folder:
```bash
# Install dependencies
yarn

# Run the web app
yarn dev
```

Open your browser and visit the web app URL.

## Running Tests

```bash
# Backend tests
cd backend && go test ./...
```

## Building for Production

```bash
# Build Docker image
docker build -t arandu .
```
