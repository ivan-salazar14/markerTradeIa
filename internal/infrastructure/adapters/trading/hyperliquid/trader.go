// Adaptador de salida que implementa el puerto Trader para HyperLiquid.
package hyperliquid

import (
	"context"
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

// HyperLiquidTrader es un adaptador que implementa el puerto Trader
// para interactuar con el exchange de HyperLiquid.
type HyperLiquidTrader struct {
	// client *hyperliquid.Client // Aquí iría el cliente de la SDK de HyperLiquid
	address string
	key     string
	mu      sync.Mutex
}

// NewHyperLiquidTrader crea un nuevo adaptador de trader para HyperLiquid.
func NewHyperLiquidTrader(address string, key string) out.Trader {
	// Aquí se inicializaría el cliente con las llaves o configuración necesaria
	return &HyperLiquidTrader{
		address: address,
		key:     key,
	}
}

// executeTrade implementa la lógica para ejecutar una orden en HyperLiquid.
func (a *HyperLiquidTrader) executeTrade(ctx context.Context, user domain.User, signal domain.TradingSignal) (domain.TradeExecution, error) {
	select {
	case <-ctx.Done():
		log.Printf("[HyperLiquid] Contexto cancelado para usuario %s", user.UID)
		return domain.TradeExecution{
			ExecutionID: user.UID,
			SignalID:    signal.ID,
			Status:      domain.Failed,
			Error:       ctx.Err(),
		}, ctx.Err()
	default:
	}

	log.Printf("[HyperLiquid] Enviando orden a HyperLiquid: %s %s a precio %.4f", signal.Type, signal.Symbol, signal.Price)

	// Simulación de los parámetros de HyperLiquid
	// En una implementación real usaríamos algo como:
	// order := &hyperliquid.OrderRequest{
	//     Coin:     signal.Symbol,
	//     IsBuy:    signal.Type == domain.Buy,
	//     Sz:       1.0, // Cantidad basada en el balance/señal
	//     LimitPx:  signal.Price,
	//     OrderType: "Limit",
	// }
	// res, err := a.client.Exchange.PlaceOrder(order)

	// Simulamos latencia de red de HyperLiquid (L1)
	time.Sleep(300 * time.Millisecond)

	// Simulación de éxito/fallo
	if rand.Float64() < 0.05 { // 5% de probabilidad de fallo en HyperLiquid
		return domain.TradeExecution{
			ExecutionID: user.UID,
			SignalID:    signal.ID,
			Status:      domain.Failed,
			ExecutedAt:  time.Now(),
			Details:     "HyperLiquid API Error: insufficient balance or slippage",
			Error:       fmt.Errorf("api error"),
		}, fmt.Errorf("api error")
	}

	return domain.TradeExecution{
		ExecutionID: user.UID,
		SignalID:    signal.ID,
		Status:      domain.Success,
		ExecutedAt:  time.Now(),
		ExecutedQty: 1.0, // Cantidad simulada
		Details:     fmt.Sprintf("Order executed on HyperLiquid for %s", signal.Symbol),
	}, nil
}

func (a *HyperLiquidTrader) worker(
	ctx context.Context,
	id int,
	taskCh <-chan task,
	resultsCh chan<- domain.TradeExecution,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case t, ok := <-taskCh:
			if !ok {
				return
			}
			log.Printf("[HyperLiquid] Worker %d: Procesando señal %s", id, t.signal.ID)
			execution, _ := a.executeTrade(ctx, t.user, t.signal)
			resultsCh <- execution
		}
	}
}

// ProcessBatch procesa un lote de señales en paralelo.
func (a *HyperLiquidTrader) ProcessBatch(ctx context.Context, users []domain.User, signals []domain.TradingSignal) []domain.TradeExecution {
	tasks := make([]task, 0, len(users)*len(signals))
	for _, user := range users {
		for _, signal := range signals {
			tasks = append(tasks, task{user: user, signal: signal})
		}
	}

	taskCh := make(chan task, len(tasks))
	resultsCh := make(chan domain.TradeExecution, len(tasks))
	var wg sync.WaitGroup

	numWorkers := 10 // HyperLiquid es muy rápido, podemos usar más workers si el rate limit lo permite
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go a.worker(ctx, i, taskCh, resultsCh, &wg)
	}

	go func() {
		defer close(taskCh)
		for _, t := range tasks {
			taskCh <- t
		}
	}()

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	var executions []domain.TradeExecution
	for res := range resultsCh {
		executions = append(executions, res)
	}

	return executions
}
