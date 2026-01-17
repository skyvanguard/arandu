# Política de Seguridad

## Versiones Soportadas

| Versión | Soportada          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reportar una Vulnerabilidad

La seguridad de Arandu es una prioridad. Si descubres una vulnerabilidad de seguridad, te pedimos que la reportes de manera responsable.

### Cómo Reportar

**Por favor NO reportes vulnerabilidades de seguridad a través de issues públicos de GitHub.**

En su lugar:

1. **Email**: Envía un reporte detallado a los maintainers del proyecto
2. **GitHub Security Advisories**: Usa la [función de advisories privados](https://github.com/skyvanguard/arandu/security/advisories/new)

### Qué Incluir en tu Reporte

- Tipo de vulnerabilidad (ej: inyección SQL, XSS, RCE)
- Ubicación del código fuente afectado (archivo/línea)
- Pasos para reproducir el issue
- Impacto potencial de la vulnerabilidad
- Posible solución (si la tienes)

### Proceso de Respuesta

1. **Confirmación**: Responderemos dentro de 48 horas confirmando la recepción
2. **Evaluación**: Evaluaremos la vulnerabilidad y su severidad
3. **Solución**: Trabajaremos en un fix y te mantendremos informado
4. **Disclosure**: Coordinaremos contigo el disclosure público

### Alcance

Las siguientes áreas están en alcance:

- Código del backend (Go)
- Código del frontend (React/TypeScript)
- Configuración de Docker
- Dependencias directas del proyecto

Las siguientes áreas están **fuera de alcance**:

- Vulnerabilidades en dependencias de terceros (reportar directamente al proyecto)
- Ataques de ingeniería social
- Ataques físicos
- Denial of Service (DoS)

## Mejores Prácticas de Seguridad

Al desplegar Arandu en producción:

### Variables de Entorno

```bash
# Habilitar modo producción
PRODUCTION_MODE=true

# Restringir CORS
CORS_ALLOWED_ORIGINS=https://tu-dominio.com

# Deshabilitar introspección GraphQL
DISABLE_INTROSPECTION=true

# Configurar rate limiting
RATE_LIMIT_PER_MINUTE=30

# NO permitir cualquier imagen Docker
ALLOW_ANY_DOCKER_IMAGE=false
```

### Docker

- No exponer el socket de Docker directamente a internet
- Usar redes Docker aisladas
- Limitar recursos de contenedores
- Mantener imágenes actualizadas

### Red

- Usar HTTPS con certificados válidos
- Configurar un reverse proxy (nginx, Caddy)
- Implementar autenticación adicional si es necesario

## Reconocimientos

Agradecemos a todos los investigadores de seguridad que ayudan a mantener Arandu seguro. Los contribuyentes serán reconocidos en nuestros release notes (con su permiso).
