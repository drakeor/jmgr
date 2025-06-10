package alert

import (
	"fmt"
	"os"
	"time"
)

// TODO: Move this to another package

// Alerter interface defines the methods required for sending alerts.
// This allows for different implementations, such as sending to a file, email, etc.
// The implementation can be swapped out without changing the alerting logic.
type Alerter interface {
	Send(subject, body string)
}

// We use a file-based alerter for testing
type FileAlerter struct {
	filePath string
}

// NewFileAlerter creates a new FileAlerter instance with the specified file path.
func NewFileAlerter(filePath string) *FileAlerter {
	return &FileAlerter{filePath: filePath}
}

// "Sending" an alert to a file.
func (f *FileAlerter) Send(subject, body string) {
	fpath := f.filePath
	fhandle, err := os.OpenFile(fpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Failed to write alert: %v\n", err)
		return
	}
	defer fhandle.Close()

	fullMsg := fmt.Sprintf("=== ALERT %s ===\nSubject: %s\n%s\n\n", time.Now().Format(time.RFC3339), subject, body)
	_, err = fhandle.WriteString(fullMsg)
	if err != nil {
		fmt.Printf("Failed to write alert: %v\n", err)
	}
}
