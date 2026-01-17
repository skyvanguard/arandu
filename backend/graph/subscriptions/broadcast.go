package subscriptions

import (
	gmodel "github.com/arandu-ai/arandu/graph/model"
)

// BroadcastTaskAdded envía una tarea a todos los suscriptores del flow
func BroadcastTaskAdded(flowID int64, task *gmodel.Task) {
	taskAddedManager.Broadcast(flowID, task)
}

// BroadcastTaskUpdated envía una actualización de tarea a todos los suscriptores
func BroadcastTaskUpdated(flowID int64, task *gmodel.Task) {
	taskUpdatedManager.Broadcast(flowID, task)
}

// BroadcastFlowUpdated envía una actualización de flow a todos los suscriptores
func BroadcastFlowUpdated(flowID int64, flow *gmodel.Flow) {
	flowUpdatedManager.Broadcast(flowID, flow)
}

// BroadcastTerminalLogsAdded envía logs de terminal a todos los suscriptores
func BroadcastTerminalLogsAdded(flowID int64, log *gmodel.Log) {
	terminalLogsAddedManager.Broadcast(flowID, log)
}

// BroadcastBrowserUpdated envía actualizaciones del browser a todos los suscriptores
func BroadcastBrowserUpdated(flowID int64, browser *gmodel.Browser) {
	browserManager.Broadcast(flowID, browser)
}
