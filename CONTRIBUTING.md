# Contribuyendo a Arandu

Gracias por tu interés en contribuir a Arandu. Este documento proporciona guías y pasos para contribuir.

## Código de Conducta

Este proyecto y todos sus participantes están regidos por nuestro [Código de Conducta](CODE_OF_CONDUCT.md). Al participar, se espera que respetes este código.

## Cómo Contribuir

### Reportar Bugs

Si encuentras un bug, por favor [abre un issue](https://github.com/skyvanguard/arandu/issues/new?template=bug_report.yml) con:

- Descripción clara del problema
- Pasos para reproducirlo
- Comportamiento esperado vs actual
- Versión de Arandu y método de deployment
- Logs relevantes

### Sugerir Features

Para sugerir nuevas funcionalidades, [abre un feature request](https://github.com/skyvanguard/arandu/issues/new?template=feature_request.yml) describiendo:

- El problema que resolvería
- Tu solución propuesta
- Alternativas consideradas

### Pull Requests

1. **Fork** el repositorio
2. **Crea una rama** para tu feature (`git checkout -b feature/amazing-feature`)
3. **Haz commit** de tus cambios (`git commit -m 'Add amazing feature'`)
4. **Push** a la rama (`git push origin feature/amazing-feature`)
5. **Abre un Pull Request**

## Configuración del Entorno de Desarrollo

### Requisitos

- Go 1.22+
- Node.js 22+
- Yarn
- Docker

### Backend

```bash
cd backend
cp .env.example .env
# Edita .env con tu configuración
go mod download
go run .
```

### Frontend

```bash
cd frontend
yarn install
yarn dev
```

### Tests

```bash
# Backend
cd backend
go test ./...

# Frontend
cd frontend
yarn test
```

### Linting

```bash
# Backend
cd backend
golangci-lint run

# Frontend
cd frontend
yarn lint
```

## Guías de Estilo

### Go

- Seguir las convenciones de [Effective Go](https://golang.org/doc/effective_go)
- Usar `gofmt` para formatear
- Documentar funciones públicas
- Manejar todos los errores

### TypeScript/React

- Usar TypeScript estricto
- Componentes funcionales con hooks
- Nombrar componentes en PascalCase
- Usar CSS-in-JS con Vanilla Extract

### Commits

Usamos commits descriptivos en inglés:

```
feat: add new LLM provider support
fix: resolve container cleanup issue
docs: update installation instructions
refactor: simplify task executor logic
test: add unit tests for providers
```

### Estructura de Archivos

- **Backend**: Organizado por responsabilidad (`config/`, `executor/`, `providers/`, etc.)
- **Frontend**: Organizado por feature (`components/`, `pages/`, `hooks/`)

## Proceso de Review

1. Todos los PRs requieren al menos una aprobación
2. Los tests deben pasar en CI
3. El código debe pasar el linting
4. Los cambios significativos necesitan documentación

## Preguntas

Si tienes preguntas, abre una [discusión](https://github.com/skyvanguard/arandu/discussions) o contacta a los maintainers.

---

¡Gracias por contribuir a Arandu! 🙏
