package logger

import (
	"fmt"
	"log"
)

// Logger provides logging functionality
type Logger struct {
	prefix string
}

// NewLogger creates a new logger
func NewLogger(category, subcategory string) *Logger {

	return &Logger{prefix: fmt.Sprintf("[%s/%s] ", category, subcategory)}

}

// Info logs an info message
func (l *Logger) Info(msg string) {

	log.Printf("%sINFO: %s\n", l.prefix, msg)

}

// Error logs an error message
func (l *Logger) Error(msg string) {

	log.Printf("%sERROR: %s\n", l.prefix, msg)

}
