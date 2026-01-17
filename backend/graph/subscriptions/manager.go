package subscriptions

import (
	"sync"

	gmodel "github.com/arandu-ai/arandu/graph/model"
)

// SubscriptionManager maneja las suscripciones de forma thread-safe
type SubscriptionManager[T any] struct {
	mu            sync.RWMutex
	subscriptions map[int64][]chan T
}

// Managers globales para cada tipo de suscripción
// Estos managers son genéricos y thread-safe para manejar suscripciones GraphQL
var (
	taskAddedManager         = NewSubscriptionManager[*gmodel.Task]()
	taskUpdatedManager       = NewSubscriptionManager[*gmodel.Task]()
	flowUpdatedManager       = NewSubscriptionManager[*gmodel.Flow]()
	terminalLogsAddedManager = NewSubscriptionManager[*gmodel.Log]()
	browserManager           = NewSubscriptionManager[*gmodel.Browser]()
)

// NewSubscriptionManager crea un nuevo manager de suscripciones
func NewSubscriptionManager[T any]() *SubscriptionManager[T] {
	return &SubscriptionManager[T]{
		subscriptions: make(map[int64][]chan T),
	}
}

// Subscribe agrega un canal a las suscripciones
func (sm *SubscriptionManager[T]) Subscribe(flowID int64) (chan T, func()) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	ch := make(chan T, 10) // Buffer para evitar bloqueos

	sm.subscriptions[flowID] = append(sm.subscriptions[flowID], ch)

	// Función de unsubscribe
	unsubscribe := func() {
		sm.mu.Lock()
		defer sm.mu.Unlock()

		channels := sm.subscriptions[flowID]
		for i, c := range channels {
			if c == ch {
				// Cerrar el canal y removerlo
				close(ch)
				sm.subscriptions[flowID] = append(channels[:i], channels[i+1:]...)
				break
			}
		}

		// Limpiar si no hay más suscriptores
		if len(sm.subscriptions[flowID]) == 0 {
			delete(sm.subscriptions, flowID)
		}
	}

	return ch, unsubscribe
}

// Broadcast envía un mensaje a todos los suscriptores de un flow
func (sm *SubscriptionManager[T]) Broadcast(flowID int64, msg T) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	channels, ok := sm.subscriptions[flowID]
	if !ok {
		return
	}

	for _, ch := range channels {
		select {
		case ch <- msg:
		default:
			// Canal lleno, saltear para no bloquear
		}
	}
}

