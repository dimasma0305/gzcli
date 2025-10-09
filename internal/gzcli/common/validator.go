package common

import (
	"fmt"
	"strings"
)

// Validator provides common validation utilities
type Validator struct{}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateRequired validates that a field is not empty
func (v *Validator) ValidateRequired(value, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	return nil
}

// ValidateLength validates that a string has the correct length
func (v *Validator) ValidateLength(value, fieldName string, min, max int) error {
	length := len(strings.TrimSpace(value))
	if length < min {
		return fmt.Errorf("%s must be at least %d characters long", fieldName, min)
	}
	if length > max {
		return fmt.Errorf("%s must be no more than %d characters long", fieldName, max)
	}
	return nil
}

// ValidateRange validates that a number is within a range
func (v *Validator) ValidateRange(value int, fieldName string, min, max int) error {
	if value < min {
		return fmt.Errorf("%s must be at least %d", fieldName, min)
	}
	if value > max {
		return fmt.Errorf("%s must be no more than %d", fieldName, max)
	}
	return nil
}

// ValidateEmail validates that a string is a valid email format
func (v *Validator) ValidateEmail(email, fieldName string) error {
	if err := v.ValidateRequired(email, fieldName); err != nil {
		return err
	}
	
	// Basic email validation
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return fmt.Errorf("%s must be a valid email address", fieldName)
	}
	
	return nil
}

// ValidateURL validates that a string is a valid URL format
func (v *Validator) ValidateURL(url, fieldName string) error {
	if err := v.ValidateRequired(url, fieldName); err != nil {
		return err
	}
	
	// Basic URL validation
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("%s must be a valid URL (http:// or https://)", fieldName)
	}
	
	return nil
}