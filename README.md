<div align="center">
  <img src=".github/logo.png" alt="Arandu Logo" width="400"/>
  <p>Agente de IA completamente autónomo que puede realizar tareas y proyectos complejos<br/>usando terminal, navegador y editor.</p>

  [![CI](https://github.com/skyvanguard/arandu/actions/workflows/ci.yml/badge.svg)](https://github.com/skyvanguard/arandu/actions/workflows/ci.yml)
  [![codecov](https://codecov.io/gh/skyvanguard/arandu/branch/master/graph/badge.svg)](https://codecov.io/gh/skyvanguard/arandu)
  [![Release](https://img.shields.io/github/v/release/skyvanguard/arandu)](https://github.com/skyvanguard/arandu/releases)
  [![License](https://img.shields.io/github/license/skyvanguard/arandu)](LICENSE)
  [![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev)
  [![React](https://img.shields.io/badge/React-18+-61DAFB?logo=react&logoColor=black)](https://react.dev)
  [![Docker](https://img.shields.io/badge/Docker-Required-2496ED?logo=docker&logoColor=white)](https://docker.com)
</div>

---

## 📋 Tabla de Contenidos

- [Características](#-características)
- [¿Por qué Arandu?](#-por-qué-arandu)
- [Ejemplos de Uso](#-ejemplos-de-uso)
- [Arquitectura](#-arquitectura)
- [Inicio Rápido](#-inicio-rápido)
- [Configuración](#-configuración)
- [Desarrollo](#-desarrollo)
- [Tecnologías](#-tecnologías)
- [Licencia](#-licencia)

---

## ✨ Características

| Característica | Descripción |
|----------------|-------------|
| 🔓 **Seguro** | Todo se ejecuta en un entorno Docker sandboxed |
| 🤖 **Autónomo** | Detecta automáticamente el siguiente paso y lo ejecuta |
| 🔍 **Navegador integrado** | Obtiene información actualizada de la web cuando es necesario |
| 📝 **Editor integrado** | Visualiza todos los archivos modificados en tu navegador |
| 🧠 **Persistencia** | Historial de comandos y salidas guardado en SQLite |
| 📦 **Auto-selección** | Elige la imagen Docker óptima según la tarea |
| 🏠 **LLMs locales** | Soporte para Ollama, LM Studio, LocalAI y más |
| 💅 **UI moderna** | Interfaz limpia y responsive |

---

## 🤔 ¿Por qué Arandu?

### El Problema

Los agentes de IA actuales tienen limitaciones importantes:

- **Dependencia de APIs caras** → Sin OpenAI no funcionan
- **Sin aislamiento real** → Ejecutan comandos directamente en tu sistema
- **Interfaces limitadas** → Solo terminal, sin visualización
- **Difíciles de extender** → Código cerrado o arquitecturas complejas

### La Solución

Arandu está diseñado para ser el agente de IA que realmente puedes usar en producción:

| Problema | Otras herramientas | Arandu |
|----------|-------------------|--------|
| **Costo** | Solo APIs de pago | LLMs locales gratuitos (Ollama, LM Studio) |
| **Seguridad** | Ejecución directa en host | Todo en containers Docker aislados |
| **Visibilidad** | Terminal básica | UI web con editor, terminal y navegador |
| **Navegación** | Limitada o inexistente | Navegador headless integrado |
| **Autonomía** | Requiere confirmación constante | Detecta y ejecuta pasos automáticamente |
| **Persistencia** | Sin historial | SQLite con logs y sesiones guardadas |

### Comparación con Alternativas

| Característica | Arandu | Open Interpreter | Aider | AutoGPT |
|---------------|--------|------------------|-------|---------|
| LLMs locales (Ollama) | ✅ | ✅ | ✅ | ❌ |
| Sandbox Docker | ✅ | ❌ | ❌ | Parcial |
| UI Web | ✅ | ❌ | ❌ | ✅ |
| Navegador integrado | ✅ | ❌ | ❌ | ✅ |
| Editor visual | ✅ | ❌ | ❌ | ❌ |
| Terminal integrada | ✅ | ✅ | ✅ | ❌ |
| 100% Open Source | ✅ | ✅ | ✅ | ✅ |
| Self-hosted | ✅ | ✅ | ✅ | ✅ |

> **Arandu** significa "sabiduría" en Guaraní. Representa la filosofía del proyecto: un agente que actúa con inteligencia y prudencia, ejecutando tareas de forma segura y autónoma.

---

## 💡 Ejemplos de Uso

### Desarrollo de Software

```
> Crea una API REST en Go con endpoints CRUD para gestionar usuarios.
  Usa Gin, GORM con PostgreSQL, y agrega autenticación JWT.
```

Arandu automáticamente:
1. Selecciona una imagen Docker con Go
2. Crea la estructura del proyecto
3. Implementa los endpoints
4. Configura la base de datos
5. Agrega middleware de autenticación

### Web Scraping

```
> Extrae los títulos y precios de los primeros 20 productos de
  https://example-store.com/laptops y guárdalos en un CSV.
```

Arandu:
1. Abre el navegador headless
2. Navega a la página
3. Extrae los datos con selectores CSS
4. Genera el archivo CSV

### Análisis de Datos

```
> Descarga el dataset de Kaggle sobre ventas de videojuegos,
  analiza las tendencias por región y genera gráficos en Python.
```

### DevOps y Automatización

```
> Crea un Dockerfile optimizado para una aplicación Next.js,
  con multi-stage build y configuración de nginx.
```

### Investigación

```
> Busca los últimos 5 papers sobre transformers en arXiv,
  resume cada uno y crea una tabla comparativa en Markdown.
```

### Casos de Uso Avanzados

| Tarea | Herramientas que usa |
|-------|---------------------|
| Crear proyecto full-stack | Terminal + Editor |
| Debuggear código existente | Editor + Terminal |
| Investigar competencia | Navegador |
| Automatizar tareas repetitivas | Terminal |
| Generar documentación | Editor |
| Hacer deploy | Terminal + Docker |
| Scrapear datos | Navegador + Terminal |

---

## 🏗 Arquitectura

```
┌─────────────────────────────────────────────────────────────────┐
│                         FRONTEND                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │
│  │   React     │  │   urql      │  │   XTerm.js  │              │
│  │   + Vite    │  │   GraphQL   │  │   Terminal  │              │
│  └─────────────┘  └─────────────┘  └─────────────┘              │
└────────────────────────────┬────────────────────────────────────┘
                             │ GraphQL + WebSocket
┌────────────────────────────▼────────────────────────────────────┐
│                         BACKEND                                  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │
│  │   Router    │  │   GraphQL   │  │  Providers  │              │
│  │   (Chi)     │  │   (gqlgen)  │  │  (LLM API)  │              │
│  └─────────────┘  └─────────────┘  └─────────────┘              │
│                                                                  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │
│  │  Executor   │  │  Database   │  │  WebSocket  │              │
│  │  (Tasks)    │  │  (SQLite)   │  │  (Logs)     │              │
│  └─────────────┘  └─────────────┘  └─────────────┘              │
└────────────────────────────┬────────────────────────────────────┘
                             │ Docker API
┌────────────────────────────▼────────────────────────────────────┐
│                      CONTAINERS                                  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │
│  │   Task      │  │   Browser   │  │   Custom    │              │
│  │  Container  │  │  (Rod/CDP)  │  │   Images    │              │
│  └─────────────┘  └─────────────┘  └─────────────┘              │
└─────────────────────────────────────────────────────────────────┘
```

### Componentes Principales

| Componente | Tecnología | Descripción |
|------------|------------|-------------|
| **Frontend** | React + TypeScript | UI con Vanilla Extract, Radix UI |
| **API** | Go + gqlgen | GraphQL con subscripciones en tiempo real |
| **Executor** | Go + Docker SDK | Orquestación de tareas en contenedores |
| **Browser** | Rod (CDP) | Automatización de navegador headless |
| **Database** | SQLite + sqlc | Persistencia de flows, tasks y logs |
| **Providers** | OpenAI/Ollama/etc | Abstracción de proveedores LLM |

---

## 🚀 Inicio Rápido

> [!IMPORTANT]
> Necesitas configurar al menos un proveedor LLM usando variables de entorno.

### Con OpenAI

```bash
docker run \
  -e OPEN_AI_KEY=your_open_ai_key \
  -e OPEN_AI_MODEL=gpt-4o \
  -p 3000:8080 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  ghcr.io/skyvanguard/arandu:latest
```

### Con Ollama (Gratis, Local)

```bash
# 1. Instala Ollama desde https://ollama.ai
# 2. Descarga un modelo
ollama pull qwen2.5-coder:14b

# 3. Ejecuta Arandu
docker run \
  -e OLLAMA_MODEL=qwen2.5-coder:14b \
  -e OLLAMA_SERVER_URL=http://host.docker.internal:11434 \
  -p 3000:8080 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  ghcr.io/skyvanguard/arandu:latest
```

Visita [http://localhost:3000](http://localhost:3000) para comenzar.

---

## ⚙ Configuración

<details>
<summary><strong>🤖 Proveedores LLM</strong></summary>

### OpenAI (Pago)
| Variable | Descripción | Default |
|----------|-------------|---------|
| `OPEN_AI_KEY` | API key de OpenAI | - |
| `OPEN_AI_MODEL` | Modelo a usar | `gpt-4o` |
| `OPEN_AI_SERVER_URL` | URL de la API | `https://api.openai.com/v1` |

### Ollama (Gratis, Local) ⭐ Recomendado
| Variable | Descripción | Default |
|----------|-------------|---------|
| `OLLAMA_MODEL` | Nombre del modelo | - |
| `OLLAMA_SERVER_URL` | URL del servidor | `http://localhost:11434` |

### LM Studio (Gratis, Local)
| Variable | Descripción | Default |
|----------|-------------|---------|
| `LMSTUDIO_MODEL` | Nombre del modelo | - |
| `LMSTUDIO_SERVER_URL` | URL del servidor | `http://localhost:1234/v1` |

### LocalAI (Gratis, Docker)
| Variable | Descripción | Default |
|----------|-------------|---------|
| `LOCALAI_MODEL` | Nombre del modelo | - |
| `LOCALAI_SERVER_URL` | URL del servidor | - |

### Compatible con OpenAI (Genérico)
Funciona con vLLM, text-generation-webui, llama.cpp, etc.

| Variable | Descripción | Default |
|----------|-------------|---------|
| `OPENAI_COMPATIBLE_MODEL` | Nombre del modelo | - |
| `OPENAI_COMPATIBLE_SERVER_URL` | URL del servidor | - |
| `OPENAI_COMPATIBLE_API_KEY` | API key (opcional) | - |

</details>

<details>
<summary><strong>🔒 Seguridad</strong></summary>

| Variable | Descripción | Default |
|----------|-------------|---------|
| `CORS_ALLOWED_ORIGINS` | Orígenes permitidos (separados por coma) | `*` |
| `PRODUCTION_MODE` | Habilitar modo producción | `false` |
| `DISABLE_INTROSPECTION` | Deshabilitar introspección GraphQL | `false` |
| `RATE_LIMIT_PER_MINUTE` | Límite de peticiones por minuto/IP | `60` |
| `ALLOW_ANY_DOCKER_IMAGE` | Permitir cualquier imagen Docker | `false` |

</details>

<details>
<summary><strong>🐳 Docker</strong></summary>

| Variable | Descripción | Default |
|----------|-------------|---------|
| `CHROME_DEBUG_URL` | URL de Chrome para debugging | Auto-detect |
| `DEFAULT_DOCKER_IMAGE` | Imagen Docker por defecto | `debian:latest` |

</details>

Ver [backend/.env.example](./backend/.env.example) para todas las opciones.

---

## 🛠 Desarrollo

### Requisitos

- Go 1.22+
- Node.js 22+
- Yarn
- Docker

### Instalación

```bash
# Clonar repositorio
git clone https://github.com/skyvanguard/arandu.git
cd arandu

# Backend
cd backend
cp .env.example .env  # Configurar variables
go mod download
go run .

# Frontend (nueva terminal)
cd frontend
yarn install
yarn dev
```

### Estructura del Proyecto

```
arandu/
├── backend/
│   ├── config/          # Configuración y variables de entorno
│   ├── database/         # SQLite + queries sqlc
│   ├── executor/         # Orquestación de tareas Docker
│   ├── graph/            # Schema y resolvers GraphQL
│   ├── providers/        # Integraciones LLM
│   ├── router/           # HTTP router (Chi)
│   └── websocket/        # WebSocket para logs
├── frontend/
│   ├── src/
│   │   ├── components/   # Componentes React
│   │   ├── pages/        # Páginas de la aplicación
│   │   ├── hooks/        # Custom hooks
│   │   └── generated/    # Código GraphQL generado
│   └── public/
└── Dockerfile
```

---

## 🔧 Tecnologías

### Backend
- **Go** - Lenguaje principal
- **gqlgen** - Servidor GraphQL
- **Chi** - Router HTTP
- **sqlc** - Queries SQL type-safe
- **Docker SDK** - Gestión de contenedores
- **Rod** - Automatización de navegador

### Frontend
- **React 18** - Framework UI
- **TypeScript** - Type safety
- **Vite** - Build tool
- **Vanilla Extract** - CSS-in-JS type-safe
- **urql** - Cliente GraphQL
- **XTerm.js** - Terminal embebida
- **Radix UI** - Componentes accesibles

---

## 📄 Licencia

Este proyecto está bajo la licencia MIT. Ver [LICENSE](LICENSE) para más detalles.

---

<div align="center">
  <p>Hecho con ❤️ en Paraguay</p>
  <p><sub>Arandu - "Sabiduría" en Guaraní</sub></p>
</div>
