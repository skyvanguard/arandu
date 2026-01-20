package websocket

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/arandu-ai/arandu/config"
	"github.com/arandu-ai/arandu/logging"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// ConnectionManager maneja las conexiones WebSocket de forma thread-safe
type ConnectionManager struct {
	mu          sync.RWMutex
	connections map[int64]*websocket.Conn
}

var (
	connManager = &ConnectionManager{
		connections: make(map[int64]*websocket.Conn),
	}
	upgrader websocket.Upgrader
)

func init() {
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     checkOrigin,
	}
}

// checkOrigin validates the WebSocket connection origin against allowed origins
func checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}

	allowedOrigins := config.Config.CORSAllowedOrigins
	if allowedOrigins == "" {
		return false
	}

	origins := strings.Split(allowedOrigins, ",")
	for _, allowed := range origins {
		allowed = strings.TrimSpace(allowed)
		if allowed == "*" || allowed == origin {
			return true
		}
	}

	logging.Warn("Terminal WebSocket connection rejected", "origin", origin)
	return false
}

// HandleWebsocket maneja nuevas conexiones WebSocket
func HandleWebsocket(c *gin.Context) {
	id := c.Param("id")

	parsedID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		_ = c.AbortWithError(400, fmt.Errorf("failed to parse id: %w", err))
		return
	}

	// Upgrade HTTP connection to WebSocket with origin validation
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logging.Error("WebSocket upgrade failed", "error", err.Error())
		_ = c.AbortWithError(400, err)
		return
	}

	// Cerrar conexión anterior si existe
	connManager.CloseConnection(parsedID)

	// Guardar la nueva conexión
	connManager.AddConnection(parsedID, conn)

	// Configurar handler de cierre
	conn.SetCloseHandler(func(code int, text string) error {
		connManager.RemoveConnection(parsedID)
		return nil
	})
}

// AddConnection agrega una conexión de forma thread-safe
func (cm *ConnectionManager) AddConnection(id int64, conn *websocket.Conn) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.connections[id] = conn
	logging.Debug("WebSocket connection added", "id", id)
}

// RemoveConnection elimina una conexión de forma thread-safe
func (cm *ConnectionManager) RemoveConnection(id int64) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.connections, id)
	logging.Debug("WebSocket connection removed", "id", id)
}

// CloseConnection cierra y elimina una conexión
func (cm *ConnectionManager) CloseConnection(id int64) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if conn, ok := cm.connections[id]; ok {
		conn.Close()
		delete(cm.connections, id)
		logging.Debug("WebSocket connection closed", "id", id)
	}
}

// GetConnection obtiene una conexión de forma thread-safe
func GetConnection(id int64) (*websocket.Conn, error) {
	connManager.mu.RLock()
	defer connManager.mu.RUnlock()

	conn, ok := connManager.connections[id]
	if !ok {
		return nil, fmt.Errorf("connection not found for id %d", id)
	}
	return conn, nil
}

// SendToChannel envía un mensaje a un canal específico
func SendToChannel(id int64, message string) error {
	connManager.mu.RLock()
	conn, ok := connManager.connections[id]
	connManager.mu.RUnlock()

	if !ok {
		return fmt.Errorf("connection not found for id %d", id)
	}

	// Usar mutex para escritura serializada
	return conn.WriteMessage(websocket.BinaryMessage, []byte(message))
}

// ANSI color codes for terminal formatting
const (
	ANSIYellow = "\033[33m"
	ANSIBlue   = "\033[34m"
	ANSIReset  = "\033[0m"
)

// FormatTerminalInput formatea la entrada del terminal con color amarillo
func FormatTerminalInput(text string) string {
	return fmt.Sprintf("$ %s%s%s\r\n", ANSIYellow, text, ANSIReset)
}

// FormatTerminalSystemOutput formatea la salida del sistema con color azul
func FormatTerminalSystemOutput(text string) string {
	return fmt.Sprintf("%s%s%s\r\n", ANSIBlue, text, ANSIReset)
}

// CloseAll cierra todas las conexiones (útil para shutdown)
func CloseAll() {
	connManager.mu.Lock()
	defer connManager.mu.Unlock()

	for id, conn := range connManager.connections {
		conn.Close()
		logging.Debug("WebSocket connection closed (shutdown)", "id", id)
	}
	connManager.connections = make(map[int64]*websocket.Conn)
}
