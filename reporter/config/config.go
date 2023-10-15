// Package config provides types and functions for working with configuration data.
package config

// Config represents the configuration for a program.
type Config struct {
	Input    string
	Output   string
	Root     string
	Cutlines *Cutlines
}

// Cutlines represents the values for safe, warning and danger.
type Cutlines struct {
	Safe    float64
	Warning float64
}
