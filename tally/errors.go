package tally

import "fmt"

// TallyError represents an error from Tally XML-RPC
type TallyError struct {
	Code    int
	Message string
	Type    string
}

func (e *TallyError) Error() string {
	return fmt.Sprintf("Tally error (code %d): %s", e.Code, e.Message)
}

func NewTallyError(code int, message string) *TallyError {
	return &TallyError{
		Code:    code,
		Message: message,
		Type:    "TallyError",
	}
}

// ConnectionError represents a connection error to Tally
type ConnectionError struct {
	Address string
	Details string
	Type    string
}

func (e *ConnectionError) Error() string {
	return fmt.Sprintf("Cannot connect to Tally at %s: %s", e.Address, e.Details)
}

func NewConnectionError(address, details string) *ConnectionError {
	return &ConnectionError{
		Address: address,
		Details: details,
		Type:    "ConnectionError",
	}
}

// ValidationError represents input validation errors
type ValidationError struct {
	Field   string
	Message string
	Type    string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("Validation error in %s: %s", e.Field, e.Message)
}

func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
		Type:    "ValidationError",
	}
}
