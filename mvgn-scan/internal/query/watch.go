package query

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// ████████████████████████████████████████████████████████████████████████████
// Watcher — Polling loop para observación en vivo (P0-04, P0-05)
// ████████████████████████████████████████████████████████████████████████████

// Watcher maneja el ciclo de observación continua de un proyecto MVGN.
type Watcher struct {
	source      StateSource
	interval    time.Duration
	noColor     bool
	renderer    Renderer

	mu          sync.Mutex
	lastState   *ObservedState
	cycleCount  int
	errorCount  int
}

// Renderer define el contrato para renderizar el estado observado.
type Renderer interface {
	Render(state ObservedState, noColor bool) string
}

// NewWatcher crea un nuevo Watcher para observación continua.
func NewWatcher(source StateSource, interval time.Duration, noColor bool, renderer Renderer) *Watcher {
	if interval < 500*time.Millisecond {
		interval = 500 * time.Millisecond
	}
	return &Watcher{
		source:   source,
		interval: interval,
		noColor:  noColor,
		renderer: renderer,
	}
}

// Run inicia el ciclo de observación. Bloquea hasta que se recibe una señal de terminación.
func (w *Watcher) Run(ctx context.Context) error {
	// Configurar captura de señales para terminación limpia
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Goroutine para manejar señales
	go func() {
		select {
		case <-sigCh:
			fmt.Fprintf(os.Stderr, "\nShutdown signal received. Exiting...\n")
			cancel()
		case <-ctx.Done():
		}
	}()

	// bucle principal de polling
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			w.poll(ctx)
			time.Sleep(w.interval)
		}
	}
}

// poll ejecuta un ciclo de polling: lee estado, renderiza, actualiza contadores.
func (w *Watcher) poll(ctx context.Context) {
	w.mu.Lock()
	defer w.mu.Unlock()

	start := time.Now()

	// Contexto con timeout por ciclo para evitar bloqueos indefinidos (ADR-001 riesgo #6)
	pollCtx, cancel := context.WithTimeout(ctx, w.interval)
	defer cancel()

	// Leer estado
	state, err := w.source.Current(pollCtx)
	if err != nil && state.Err == nil {
		state.Err = err
	}

	w.cycleCount++

	// Si no hay estado previo válido y hay error, inicializar
	if w.lastState == nil {
		w.lastState = &state
	} else if err == nil {
		// Nuevo estado válido
		w.lastState = &state
		w.errorCount = 0
	} else {
		// Error transitorio
		w.errorCount++
	}

	// Renderizar
	output := w.renderer.Render(*w.lastState, w.noColor)
	fmt.Print(output)

	// Mostrar metadata del ciclo solo si hubo error o cada 10 ciclos
	elapsed := time.Since(start)
	if w.errorCount > 0 || w.cycleCount%10 == 0 {
		cycleInfo := fmt.Sprintf("Cycle %d · Poll: %v · Errors: %d",
			w.cycleCount, elapsed.Round(time.Millisecond), w.errorCount)
		fmt.Fprintf(os.Stderr, "  %s\n", cycleInfo)
	}
}
