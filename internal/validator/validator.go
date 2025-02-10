package validator

import (
	"regexp"
)

var (
	// This regex allows email addresses with no dot in the DNS, eg. user@gmail, because there are some valid DNS's without a dot so it is technically a valid email address.
	EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

type Validator struct {
	Errors map[string]string
}

// Function to create new Validator
func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

// If the Validator conatins no errors, the value is valid
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// If an error doesn't already exist at the provided key, create one with the provided message
func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

// Same as Python "if _ in []:"
func In(value string, list ...string) bool {
	for i := range list {
		if value == list[i] {
			return true
		}
	}
	return false
}

// Returns true if value matches regex pattern
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

// If there are duplicate values in the provided slice of strings, the duplicate value(s) will just be set twice in uniqueValues rather than creating a new key-value pair, so the length of uniqueValues will be shorter than the length of values
func Unique(values []string) bool {
	uniqueValues := make(map[string]bool)

	for _, value := range values {
		uniqueValues[value] = true
	}

	return len(values) == len(uniqueValues)
}

// Instead of checking whether string is empty, it is likely that later we will implement the In function to test provided genres against a slice of valid genres. For now, this is a lovely little proof of concept
func NoEmptyString(values []string) bool {
	for _, value := range values {
		if value == "" {
			return false
		}
	}

	return true
}
