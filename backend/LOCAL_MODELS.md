# Local Models Guide

This guide explains how to use Arandu with local LLM models without requiring paid API keys.

## Recommended Local Models

For best results with Arandu, use models that support:
- Function/tool calling OR
- Reliable JSON output
- At least 7B parameters (13B+ recommended for complex tasks)

### Best Models for Coding Tasks

| Model | Size | Best For | Tool Calling |
|-------|------|----------|--------------|
| **Qwen2.5-Coder** | 7B/14B/32B | Coding tasks | Yes |
| **DeepSeek-Coder-V2** | 16B/236B | Complex coding | Yes |
| **CodeLlama** | 7B/13B/34B | Code generation | No (JSON mode) |
| **Mistral** | 7B | General + coding | Yes |
| **Mixtral** | 8x7B | Complex reasoning | Yes |
| **Llama 3.1** | 8B/70B | General purpose | Yes |
| **Phi-3** | 3.8B/14B | Lightweight tasks | Yes |

## Setup Options

### Option 1: Ollama (Recommended for beginners)

[Ollama](https://ollama.ai) is the easiest way to run local models.

```bash
# Install Ollama
curl -fsSL https://ollama.ai/install.sh | sh

# Pull a recommended model
ollama pull qwen2.5-coder:14b

# Or for smaller systems
ollama pull qwen2.5-coder:7b
ollama pull codellama:7b
ollama pull mistral:7b
```

**Configuration:**
```bash
# .env file
OLLAMA_MODEL=qwen2.5-coder:14b
OLLAMA_SERVER_URL=http://localhost:11434
```

### Option 2: LM Studio (Best GUI experience)

[LM Studio](https://lmstudio.ai) provides a user-friendly interface for running local models.

1. Download and install LM Studio
2. Download a model (search for "qwen2.5-coder" or "deepseek-coder")
3. Start the local server (default: http://localhost:1234)

**Configuration:**
```bash
# .env file
LMSTUDIO_MODEL=qwen2.5-coder-14b-instruct
LMSTUDIO_SERVER_URL=http://localhost:1234/v1
```

### Option 3: LocalAI (Docker-based)

[LocalAI](https://localai.io) runs models in Docker containers.

```bash
# Run LocalAI with a model
docker run -p 8080:8080 --name local-ai \
  -v $PWD/models:/models \
  localai/localai:latest-cpu

# Or with GPU support
docker run -p 8080:8080 --gpus all --name local-ai \
  -v $PWD/models:/models \
  localai/localai:latest-gpu-nvidia-cuda-12
```

**Configuration:**
```bash
# .env file
LOCALAI_MODEL=gpt-4
LOCALAI_SERVER_URL=http://localhost:8080/v1
```

### Option 4: Generic OpenAI-Compatible Server

For other servers like vLLM, text-generation-webui, or llama.cpp:

**vLLM example:**
```bash
python -m vllm.entrypoints.openai.api_server \
  --model Qwen/Qwen2.5-Coder-14B-Instruct \
  --port 8000
```

**Configuration:**
```bash
# .env file
OPENAI_COMPATIBLE_MODEL=Qwen/Qwen2.5-Coder-14B-Instruct
OPENAI_COMPATIBLE_SERVER_URL=http://localhost:8000/v1
OPENAI_COMPATIBLE_API_KEY=not-needed
```

## Hardware Requirements

| Model Size | RAM Required | VRAM (GPU) | Recommended |
|------------|--------------|------------|-------------|
| 3B-7B | 8GB | 6GB | Entry level |
| 7B-14B | 16GB | 8-12GB | Good balance |
| 14B-34B | 32GB | 16-24GB | Best quality |
| 70B+ | 64GB+ | 48GB+ | Maximum quality |

### Tips for Limited Hardware

1. **Use quantized models** (Q4, Q5, Q8):
   ```bash
   ollama pull qwen2.5-coder:7b-instruct-q4_K_M
   ```

2. **Reduce context length** if you run out of memory

3. **Use CPU-only** mode if you don't have a GPU (slower but works)

## Troubleshooting

### "Connection refused" error
- Make sure the local server is running
- Check the port is correct
- For Docker: use `host.docker.internal` instead of `localhost`

### Slow responses
- Use a smaller/quantized model
- Enable GPU acceleration if available
- Reduce max tokens in responses

### Poor quality responses
- Try a larger model (14B+ recommended)
- Use a model specifically trained for coding
- Make sure you're using the instruct/chat variant

### Model doesn't follow instructions
- Some models don't support tool calling - the system will fall back to JSON mode
- Try Qwen2.5-Coder or Mistral which have good instruction following

## Example .env for Local-Only Setup

```bash
# No paid APIs needed!
# Using Ollama with Qwen2.5-Coder

OLLAMA_MODEL=qwen2.5-coder:14b
OLLAMA_SERVER_URL=http://localhost:11434

# Security settings
CORS_ALLOWED_ORIGINS=http://localhost:5173
ALLOW_ANY_DOCKER_IMAGE=true

# Optional: Use a different port
PORT=8080
```

## Performance Comparison

Based on typical coding tasks:

| Provider | Model | Speed | Quality | Cost |
|----------|-------|-------|---------|------|
| OpenAI | GPT-4o | Fast | Excellent | $$$ |
| Ollama | Qwen2.5-Coder:14B | Medium | Very Good | Free |
| Ollama | Qwen2.5-Coder:7B | Fast | Good | Free |
| LM Studio | DeepSeek-Coder-V2 | Medium | Excellent | Free |
| Ollama | CodeLlama:13B | Medium | Good | Free |
| Ollama | Mistral:7B | Fast | Good | Free |

## Need Help?

- Ollama Discord: https://discord.gg/ollama
- LM Studio Discord: https://discord.gg/lmstudio
- LocalAI: https://github.com/go-skynet/LocalAI
