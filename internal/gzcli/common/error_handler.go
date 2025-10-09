package common

import (
	"fmt"
	"runtime"

	"github.com/dimasma0305/gzcli/internal/gzcli/errors"
	"github.com/dimasma0305/gzcli/internal/log"
)

// ErrorHandler provides centralized error handling
type ErrorHandler struct {
	logErrors bool
}

// NewErrorHandler creates a new ErrorHandler
func NewErrorHandler(logErrors bool) *ErrorHandler {
	return &ErrorHandler{
		logErrors: logErrors,
	}
}

// HandleError handles an error with optional logging and context
func (h *ErrorHandler) HandleError(err error, context string) error {
	if err == nil {
		return nil
	}

	// Add context to the error
	wrappedErr := errors.Wrap(err, context)

	// Log error if enabled
	if h.logErrors {
		h.logError(wrappedErr, context)
	}

	return wrappedErr
}

// HandleErrorf handles an error with formatted context
func (h *ErrorHandler) HandleErrorf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	// Add formatted context to the error
	wrappedErr := errors.Wrapf(err, format, args...)

	// Log error if enabled
	if h.logErrors {
		h.logError(wrappedErr, fmt.Sprintf(format, args...))
	}

	return wrappedErr
}

// HandlePanic recovers from panics and converts them to errors
func (h *ErrorHandler) HandlePanic() error {
	if r := recover(); r != nil {
		// Get stack trace
		stack := make([]byte, 4096)
		length := runtime.Stack(stack, false)
		
		// Create error with stack trace
		err := fmt.Errorf("panic recovered: %v\nStack trace:\n%s", r, stack[:length])
		
		// Log the panic
		if h.logErrors {
			log.Error("Panic recovered: %v", r)
			log.Debug("Stack trace: %s", stack[:length])
		}
		
		return err
	}
	return nil
}

// logError logs an error with additional context
func (h *ErrorHandler) logError(err error, context string) {
	// Get caller information
	_, file, line, ok := runtime.Caller(2)
	if ok {
		log.Error("Error in %s:%d - %s: %v", file, line, context, err)
	} else {
		log.Error("Error in %s: %v", context, err)
	}
}