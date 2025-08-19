// Entidad del dominio que representa el resultado de ejecutar una orden.
package domain

import "time"

// ExecutionStatus representa el estado de una ejecución de orden.
type ExecutionStatus string

const (
	Success ExecutionStatus = "SUCCESS"
	Failed  ExecutionStatus = "FAILED"
)

// TradeExecution es el objeto de dominio que encapsula el resultado de una orden.
type TradeExecution struct {
	ExecutionID string
	SignalID    string
	Status      ExecutionStatus
	ExecutedAt  time.Time
	ExecutedQty float64 // Cantidad ejecutada
	Details     string
	Error       error // Error opcional en caso de fallo
}
