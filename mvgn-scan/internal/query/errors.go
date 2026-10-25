package query

import "errors"

// ████████████████████████████████████████████████████████████████████████████
// Errores tipificados — P0-07, P0-08, P0-10
// ████████████████████████████████████████████████████████████████████████████

var (
	// ErrBlackboardNotFound ocurre cuando no existe .mvgn/blackboard/state.json.
	ErrBlackboardNotFound = errors.New("BLACKBOARD_NOT_FOUND")

	// ErrReadFailed ocurre cuando falla la lectura del archivo state.json.
	ErrReadFailed = errors.New("READ_FAILED")

	// ErrCorruptJSON ocurre cuando state.json tiene JSON malformado.
	ErrCorruptJSON = errors.New("INVALID_JSON")

	// ErrInvalidSchema ocurre cuando state.json tiene un esquema inválido o incompleto.
	ErrInvalidSchema = errors.New("INVALID_SCHEMA")

	// ErrStaleState ocurre cuando el estado superó el umbral de frescura.
	ErrStaleState = errors.New("STALE_STATE")

	// ErrNoBlackboard es un alias de ErrBlackboardNotFound para mensajes al usuario.
	ErrNoBlackboard = ErrBlackboardNotFound
)

// ErrorType devuelve una representación legible del error para mostrar en la consola.
func ErrorType(err error) string {
	switch {
	case errors.Is(err, ErrBlackboardNotFound):
		return "BLACKBOARD_NOT_FOUND"
	case errors.Is(err, ErrReadFailed):
		return "READ_FAILED"
	case errors.Is(err, ErrCorruptJSON):
		return "INVALID_JSON"
	case errors.Is(err, ErrInvalidSchema):
		return "INVALID_SCHEMA"
	case errors.Is(err, ErrStaleState):
		return "STALE_STATE"
	default:
		return "UNKNOWN_ERROR"
	}
}

// ActionForError devuelve una acción sugerida para el usuario según el tipo de error.
func ActionForError(err error) string {
	switch {
	case errors.Is(err, ErrBlackboardNotFound):
		return "Run: .\\NYX.ps1 -Mode init  or  read NYX.md to initialize the project first."
	case errors.Is(err, ErrCorruptJSON):
		return "The blackboard state file is corrupt. Run recovery: See mvgn-recovery.md"
	case errors.Is(err, ErrInvalidSchema):
		return "The state.json schema is not recognized. Verify MVGN version compatibility."
	default:
		return "Check the project state and try again."
	}
}
