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
	log.Printf("Ejecutando orden '%s' en Binance para el símbolo '%s' a precio %.2f...",
		signal.Type, signal.Symbol, signal.Price)
	// Simulación de una llamada a la API de Binance
	time.Sleep(500 * time.Millisecond)

	// Simulamos un fallo aleatorio
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

	// Si todo va bien, se ejecuta completamente
	return domain.TradeExecution{
		ExecutionID: user.UID,
		ExecutedQty: amount,
		Status:      "FILLED",
		Error:       nil,
	}, nil
}

func (a *BinanceTrader) worker(
	ctx context.Context,
	id int,
	inputCh <-chan domain.TradingSignal,
	resultsCh chan<- domain.TradeExecution,
	wg *sync.WaitGroup,
	users []domain.User, // Pasamos la información de usuarios aquí
) {
	defer wg.Done()

	for signal := range inputCh {
		// Aquí puedes buscar el usuario por alguna clave en la lista 'users'.
		// Por simplicidad, simularemos que encontramos el usuario.
		user := users[id%len(users)]

		log.Printf("Worker %d: Procesando señal '%s' para user '%s'", id, signal.ID, user.UID)

		// La lógica original de ExecuteTrade va aquí
		tradeExecution, err := a.executeTrade(ctx, user, signal)
		if err != nil {
			log.Printf("Worker %d: Error al ejecutar orden para señal %s: %v", id, signal.ID, err)
			resultsCh <- tradeExecution
		}
		resultsCh <- tradeExecution
	}
}

// ProcessBatch procesa un lote de señales de trading en paralelo.
func (a *BinanceTrader) ProcessBatch(ctx context.Context, users []domain.User, signals []domain.TradingSignal) []domain.TradeExecution {
	inputCh := make(chan domain.TradingSignal, len(signals))
	resultsCh := make(chan domain.TradeExecution, len(signals))
	var wg sync.WaitGroup

	// FAN-OUT: Lanzar workers
	numWorkers := 5 // O el número que decidas
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go a.worker(ctx, i, inputCh, resultsCh, &wg, users)
	}

	// Llenar el canal de entrada y cerrarlo cuando termine
	go func() {
		defer close(inputCh)
		for _, signal := range signals {
			inputCh <- signal
		}
	}()

	// Esperar a que los workers terminen y cerrar el canal de resultados
	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	// FAN-IN: Leer los resultados del canal y consolidarlos
	var executions []domain.TradeExecution
	for result := range resultsCh {
		executions = append(executions, result)
	}
	startTime := time.Now()

	// Leer y procesar los resultados
	fmt.Println("Processing trades and collecting results...")
	tradesCompleted := 0
	tradesPartiallyFilled := 0
	tradesFailed := 0

	for result := range resultsCh {
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
