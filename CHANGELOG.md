# Changelog

Todos los cambios notables de este proyecto serán documentados en este archivo.

El formato está basado en [Keep a Changelog](https://keepachangelog.com/es-ES/1.0.0/),
y este proyecto adhiere a [Semantic Versioning](https://semver.org/lang/es/).

## [Unreleased]

### Added
- Próximas funcionalidades...

---

## [1.0.0] - 2025-01-17

### Added
- **Backend en Go** con API GraphQL (gqlgen)
- **Frontend en React** con TypeScript y Vite
- **Ejecución aislada** en contenedores Docker
- **Navegador headless** integrado con Rod (CDP)
- **Terminal interactiva** con WebSocket y XTerm.js
- **Soporte multi-proveedor LLM**:
  - OpenAI (API de pago)
  - Ollama (local, gratis)
  - LM Studio (local, gratis)
  - LocalAI (Docker, gratis)
  - Servidores compatibles con OpenAI
- **Persistencia** con SQLite y sqlc
- **Subscripciones GraphQL** en tiempo real
- **UI moderna** con Vanilla Extract y Radix UI
- **Configuración de seguridad** (CORS, rate limiting, modo producción)
- **Selector automático** de imagen Docker según la tarea
- **CI/CD completo** con GitHub Actions

### Security
- Ejecución sandboxed en Docker
- Variables de entorno para configuración sensible
- Soporte para modo producción con hardening

---

## Tipos de Cambios

- `Added` - Nuevas funcionalidades
- `Changed` - Cambios en funcionalidades existentes
- `Deprecated` - Funcionalidades que serán removidas
- `Removed` - Funcionalidades removidas
- `Fixed` - Corrección de bugs
- `Security` - Correcciones de vulnerabilidades

[Unreleased]: https://github.com/skyvanguard/arandu/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/skyvanguard/arandu/releases/tag/v1.0.0
