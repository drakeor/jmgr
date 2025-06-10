// internal/utils/ch.go
package utils

// NonBlockingSend tries to send v on ch without blocking.
func NonBlockingSend[T any](ch chan T, v T) {
	select {
	case ch <- v:
	default:
	}
}
