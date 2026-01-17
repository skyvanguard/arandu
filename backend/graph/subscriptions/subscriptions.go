package subscriptions

import (
	"context"

	gmodel "github.com/arandu-ai/arandu/graph/model"
)

// TaskAdded crea una suscripción para nuevas tareas de un flow
func TaskAdded(ctx context.Context, flowId int64) (<-chan *gmodel.Task, error) {
	ch, unsubscribe := taskAddedManager.Subscribe(flowId)
	go handleUnsubscribe(ctx, unsubscribe)
	return ch, nil
}

// TaskUpdated crea una suscripción para actualizaciones de tareas
func TaskUpdated(ctx context.Context, flowId int64) (<-chan *gmodel.Task, error) {
	ch, unsubscribe := taskUpdatedManager.Subscribe(flowId)
	go handleUnsubscribe(ctx, unsubscribe)
	return ch, nil
}

// FlowUpdated crea una suscripción para actualizaciones de flow
func FlowUpdated(ctx context.Context, flowId int64) (<-chan *gmodel.Flow, error) {
	ch, unsubscribe := flowUpdatedManager.Subscribe(flowId)
	go handleUnsubscribe(ctx, unsubscribe)
	return ch, nil
}

// TerminalLogsAdded crea una suscripción para logs de terminal
func TerminalLogsAdded(ctx context.Context, flowId int64) (<-chan *gmodel.Log, error) {
	ch, unsubscribe := terminalLogsAddedManager.Subscribe(flowId)
	go handleUnsubscribe(ctx, unsubscribe)
	return ch, nil
}

// BrowserUpdated crea una suscripción para actualizaciones del browser
func BrowserUpdated(ctx context.Context, flowId int64) (<-chan *gmodel.Browser, error) {
	ch, unsubscribe := browserManager.Subscribe(flowId)
	go handleUnsubscribe(ctx, unsubscribe)
	return ch, nil
}

// handleUnsubscribe espera a que el contexto se cancele y luego ejecuta unsubscribe
func handleUnsubscribe(ctx context.Context, unsubscribe func()) {
	<-ctx.Done()
	unsubscribe()
}
