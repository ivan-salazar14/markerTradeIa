package main

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// Work representa el "paquete" de trabajo que cada goroutine procesará
type Work struct {
	UID    string
	APIKey string
}

// TradeResponse simula la respuesta de la API de trade
type TradeResponse struct {
	UID         string
	ExecutedQty float64
	Status      string
	Error       error
}

// simulateTrade simula la operación de comercio con la API Key
func simulateTrade(work Work, amount float64) TradeResponse {
	// Simulamos un retraso para imitar una llamada de red o a la base de datos
	time.Sleep(500 * time.Millisecond)

	// Simulamos un fallo aleatorio
	if rand.Float64() < 0.3 { // 30% de probabilidad de fallo
		return TradeResponse{
			UID:    work.UID,
			Status: "FAILED",
			Error:  errors.New("API connection failed"),
		}
	}

	// Simulamos una ejecución parcial
	if rand.Float64() < 0.5 { // 50% de probabilidad de ejecución parcial
		executedQty := amount * rand.Float64() // Ejecuta una porción del monto
		return TradeResponse{
			UID:         work.UID,
			ExecutedQty: executedQty,
			Status:      "PARTIALLY_FILLED",
			Error:       nil,
		}
	}

	// Si todo va bien, se ejecuta completamente
	return TradeResponse{
		UID:         work.UID,
		ExecutedQty: amount,
		Status:      "FILLED",
		Error:       nil,
	}
}

// worker procesa trabajos del canal de entrada y envía los resultados a otro canal
func worker(id int, inputCh <-chan Work, resultsCh chan<- TradeResponse, wg *sync.WaitGroup, totalBalance *float64, balanceMu *sync.Mutex) {
	defer wg.Done()

	for work := range inputCh {
		amountToTrade := 10.0

		// Bloqueamos el mutex ANTES de leer y modificar el saldo
		balanceMu.Lock()
		if *totalBalance < amountToTrade {
			balanceMu.Unlock() // Desbloqueamos antes de continuar
			continue
		}
		*totalBalance -= amountToTrade
		balanceMu.Unlock() // Desbloqueamos después de la modificación

		// Realizamos el trade simulado y enviamos el resultado al canal de fan-in
		tradeResult := simulateTrade(work, amountToTrade)
		resultsCh <- tradeResult
	}
}

func main() {
	rand.Seed(time.Now().UnixNano()) // Inicializamos el generador de números aleatorios

	numWorkers := 5
	numUsers := 1000
	totalBalance := 100.0
	var balanceMu sync.Mutex

	// Canales y WaitGroup
	inputCh := make(chan Work, numUsers)
	resultsCh := make(chan TradeResponse, numUsers)
	var wg sync.WaitGroup

	// FAN-OUT: Lanzar workers
	wg.Add(numWorkers)
	for i := 1; i <= numWorkers; i++ {
		go worker(i, inputCh, resultsCh, &wg, &totalBalance, &balanceMu)
	}

	// Enviar trabajos al canal de entrada
	go func() {
		for i := 0; i < numUsers; i++ {
			uid := fmt.Sprintf("uid%d", i)
			apiKey := fmt.Sprintf("api-key-%d", i)
			inputCh <- Work{UID: uid, APIKey: apiKey}
		}
		close(inputCh) // Señal para que los workers dejen de leer
	}()

	// FAN-IN: Esperar a que todos los workers terminen y luego cerrar el canal de resultados
	go func() {
		wg.Wait()
		close(resultsCh) // Señal para que la goroutine de lectura de resultados termine
	}()

	// Medición del tiempo de ejecución
	startTime := time.Now()

	// Leer y procesar los resultados
	fmt.Println("Processing trades and collecting results...")
	tradesCompleted := 0
	tradesPartiallyFilled := 0
	tradesFailed := 0

	for result := range resultsCh {
		fmt.Printf("Received result for UID %s: Status: %s, Executed: %.2f\n",
			result.UID, result.Status, result.ExecutedQty)

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
}
