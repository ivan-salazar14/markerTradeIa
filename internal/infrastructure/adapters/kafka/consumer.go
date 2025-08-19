// Adaptador de entrada que implementa el puerto EventReceiver.
package kafka

import (
	"context"
	"log"
	"time" // Se agrega la importación de time

	"github.com/ivan-salazar14/markerTradeIa/internal/domain"
	"github.com/ivan-salazar14/markerTradeIa/internal/domain/ports/in"
)

// ConsumerAdapter implementa la interfaz in.EventReceiver
type ConsumerAdapter struct {
	// cliente de Kafka, tópico, etc.
	tradingService in.TradingServicePort
}

// NewConsumerAdapter crea un nuevo adaptador de consumidor de Kafka.
func NewConsumerAdapter(ts in.TradingServicePort) *ConsumerAdapter {
	// Inicialización del cliente de Kafka
	return &ConsumerAdapter{tradingService: ts}
}

// StartConsuming inicia la escucha de mensajes de Kafka.
func (a *ConsumerAdapter) StartConsuming(ctx context.Context) error {
	log.Println("Consumidor de Kafka iniciado. Escuchando en el tópico...")
	// Simulación de un bucle de consumo de Kafka
	signalsToProcess := make(chan domain.TradingSignal, 5) // Canal intermedio

	// Goroutine que procesa los lotes del canal intermedio
	go a.processSignalsInBatches(ctx, signalsToProcess)

	for {
		select {
		case <-ctx.Done():
			log.Println("Consumidor de Kafka detenido.")
			return ctx.Err()
		default:
			// Lógica para recibir un mensaje de Kafka (un solo mensaje por vez)
			signal := domain.TradingSignal{
				ID:       "1",
				Strategy: "conservative",
				Symbol:   "BTCUSDT",
				Price:    114844,
				Type:     "BUY",
			}

			// Enviamos la señal al canal intermedio, no la procesamos aquí
			signalsToProcess <- signal
		}
	}
}

// processSignalsInBatches es una nueva goroutine que lee del canal
func (a *ConsumerAdapter) processSignalsInBatches(ctx context.Context, signalsCh <-chan domain.TradingSignal) {
	batchSize := 10
	var signalsBatch []domain.TradingSignal

	for {
		select {
		case signal := <-signalsCh:
			signalsBatch = append(signalsBatch, signal)
			if len(signalsBatch) >= batchSize {
				log.Printf("Recibida batch de señales por batch de trading: %+v", len(signalsBatch))
				a.tradingService.ProcessSignalsInBatch(ctx, signalsBatch)
				signalsBatch = []domain.TradingSignal{} // Limpiamos el batch
			}
		case <-time.After(5 * time.Second): // Procesar el batch incluso si no está lleno
			log.Println("Procesando señales en batch...", signalsBatch)
			if len(signalsBatch) > 0 {
				a.tradingService.ProcessSignalsInBatch(ctx, signalsBatch)
				signalsBatch = []domain.TradingSignal{}
			}
		case <-ctx.Done():
			return
		}
	}
}
