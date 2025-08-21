// Adaptador de salida que implementa el puerto Trader para Binance.
package binance

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/ivan-salazar14/markerTradeIa/internal/domain"
	"github.com/ivan-salazar14/markerTradeIa/internal/domain/ports/out"
)

// task representa una combinación de usuario y señal a procesar
type task struct {
	user   domain.User
	signal domain.TradingSignal
}

// BinanceTrader es un adaptador que implementa el puerto Trader
// para interactuar con el exchange de Binance.
type BinanceTrader struct {
	// client *binance.Client // Cliente de la API de Binance
	mu sync.Mutex
}

// NewBinanceTrader crea un nuevo adaptador de trader para Binance.
func NewBinanceTrader() out.Trader {
	// Inicialización del cliente de Binance aquí
	return &BinanceTrader{}
}

var amount = 10.0

// ExecuteTrade implementa la lógica para ejecutar una orden en Binance.
func (a *BinanceTrader) executeTrade(ctx context.Context, user domain.User, signal domain.TradingSignal) (domain.TradeExecution, error) {

	select {
	case <-ctx.Done():
		log.Printf("Contexto cancelado. No se puede ejecutar la orden para el usuario %s", user.UID)
		return domain.TradeExecution{
			ExecutionID: user.UID,
			Status:      "CANCELED",
			Error:       ctx.Err(),
		}, ctx.Err()
	default:
		// Continuar con la ejecución de la orden
	}
	log.Printf("Ejecutando orden '%s' en Binance para el símbolo '%s' a precio %.2f...", signal.Type, signal.Symbol, signal.Price)

	// Simulación de una llamada a la API de Binance
	time.Sleep(5000 * time.Millisecond)

	if rand.Float64() < 0.3 { // 30% de probabilidad de fallo
		return domain.TradeExecution{
			ExecutionID: user.UID,
			Status:      "FAILED",
			Error:       errors.New("API connection failed"),
		}, errors.New("API connection failed")
	}

	// Simulamos una ejecución parcial
	if rand.Float64() < 0.5 { // 50% de probabilidad de ejecución parcial
		executedQty := amount * rand.Float64() // Ejecuta una porción del monto
		return domain.TradeExecution{
			ExecutionID: user.UID,
			ExecutedQty: executedQty,
			Status:      "PARTIALLY_FILLED",
			Error:       nil,
		}, nil
	}

	return domain.TradeExecution{
		ExecutionID: user.UID,
		ExecutedQty: amount,
		Status:      "FILLED",
		Error:       nil,
	}, nil

}

func (p *BinanceTrader) worker(
	ctx context.Context,
	id int,
	taskCh <-chan task,
	resultsCh chan<- domain.TradeExecution,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done(): // Si el contexto es cancelado, salimos del bucle
			log.Printf("Worker %d: Contexto cancelado. Deteniendo operaciones.", id)
			return
		case t, ok := <-taskCh: // Leemos del canal de tareas
			if !ok {
				return // El canal se ha cerrado, salimos
			}
			log.Printf("Worker %d: Procesando señal '%s' para user '%s'", id, t.signal.ID, t.user.UID)

			// Llamamos a la función de ejecución con el contexto
			tradeExecution, err := p.executeTrade(ctx, t.user, t.signal)
			if err != nil {
				log.Printf("Worker %d: Error al ejecutar orden para señal %s y user %s: %v", id, t.signal.ID, t.user.UID, err)
			}
			resultsCh <- tradeExecution
		}
	}
}

// ProcessBatch procesa un lote de señales de trading en paralelo.
func (a *BinanceTrader) ProcessBatch(ctx context.Context, users []domain.User, signals []domain.TradingSignal) []domain.TradeExecution {
	tasks := make([]task, 0, len(users)*len(signals))
	for _, user := range users {
		for _, signal := range signals {
			tasks = append(tasks, task{user: user, signal: signal})
		}
	}

	taskCh := make(chan task, len(tasks))
	resultsCh := make(chan domain.TradeExecution, len(tasks))
	var wg sync.WaitGroup

	// FAN-OUT: Lanzar workers
	numWorkers := 5
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go a.worker(ctx, i, taskCh, resultsCh, &wg)
	}

	// Llenar el canal de tareas y cerrarlo cuando termine
	go func() {
		defer close(taskCh)
		for _, t := range tasks {
			taskCh <- t
		}
	}()

	// Esperar a que los workers terminen y cerrar el canal de resultados
	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	// FAN-IN: Leer los resultados del canal y consolidarlos, y hacer el summary en el mismo ciclo
	var executions []domain.TradeExecution
	tradesCompleted := 0
	tradesPartiallyFilled := 0
	tradesFailed := 0
	startTime := time.Now()

	fmt.Println("Processing trades and collecting results...")
	for result := range resultsCh {
		executions = append(executions, result)
		fmt.Printf("Received result for UID %s: Status: %s, Executed: %.2f\n",
			result.ExecutionID, result.Status, result.ExecutedQty)

		switch result.Status {
		case "FILLED":
			tradesCompleted++
		case "PARTIALLY_FILLED":
			tradesPartiallyFilled++
		case "FAILED":
			tradesFailed++
		}
	}

	elapsedTime := time.Since(startTime)

	fmt.Println("\nAll results processed.")
	fmt.Printf("Summary: Completed: %d, Partially Filled: %d, Failed: %d\n",
		tradesCompleted, tradesPartiallyFilled, tradesFailed)
	fmt.Printf("Total execution time: %s\n", elapsedTime)
	return executions
}
