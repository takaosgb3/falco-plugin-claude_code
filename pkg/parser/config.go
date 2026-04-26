package parser

// Config holds the parser configuration.
type Config struct {
	LogFormat        string // "auto", "combined", "common", "json", "custom"
	CustomFormat     string // Custom regex pattern (for "custom" format)
	SecurityPatterns bool   // Enable security threat detection
	MaxFieldLength   int    // Threshold for large field truncation (bytes, default: 10KB)
}
