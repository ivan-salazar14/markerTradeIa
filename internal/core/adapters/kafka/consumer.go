// Adaptador de entrada que implementa el puerto EventReceiver.
package kafka

import (
	"context"
	"log"
	"time" // Se agrega la importación de time

	"github.com/ivan-salazar14/markerTradeIa/internal/core/domain"
	"github.com/ivan-salazar14/markerTradeIa/internal/core/ports/in"
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
	for {
		select {
		case <-ctx.Done():
			log.Println("Consumidor de Kafka detenido.")
			return ctx.Err()
		default:
			// Simulación de recibir un mensaje de Kafka
			// En un caso real, el mensaje vendría del cliente de Kafka
			signal := domain.TradingSignal{
				ID:        "SIG-12345",
				Symbol:    "BTCUSDT",
				Type:      domain.Buy,
				Price:     25000.0,
				Timestamp: time.Now(),
			}

			// Lógica para deserializar y pasar la señal al servicio de la aplicación
			if err := a.tradingService.ProcessSignal(ctx, signal); err != nil {
				log.Printf("Fallo al procesar señal: %v", err)
			}
			time.Sleep(5 * time.Second) // Simulación de intervalo entre mensajes
		}
	}
}
