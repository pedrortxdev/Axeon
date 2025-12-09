package events

// EventType define os tipos de eventos do sistema.
type EventType string

const (
	JobUpdate   EventType = "job_update"
	StateChange EventType = "state_change"
)

// Event representa uma mensagem no barramento de eventos.
type Event struct {
	Type      EventType   `json:"type"`
	JobID     string      `json:"job_id,omitempty"`
	Target    string      `json:"target,omitempty"`
	Payload   interface{} `json:"payload"`
	Timestamp int64       `json:"timestamp"`
}

// GlobalBus é o canal onde todos os eventos são publicados.
// O buffer de 1000 evita bloqueios se o consumidor (WebSocket) for lento.
var GlobalBus = make(chan Event, 1000)

// Publish envia um evento para o barramento.
func Publish(evt Event) {
	// Non-blocking publish para não travar o emissor se o bus estiver cheio
	select {
	case GlobalBus <- evt:
	default:
		// Logar drop de evento em produção
	}
}
