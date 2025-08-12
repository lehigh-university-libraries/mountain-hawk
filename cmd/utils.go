package cmd

import "fmt"

// getSeverityIcon returns an appropriate icon for the severity level
func getSeverityIcon(severity interface{}) string {
	switch fmt.Sprintf("%v", severity) {
	case "error":
		return "🚨"
	case "warning":
		return "⚠️"
	case "info":
		return "💡"
	default:
		return "📝"
	}
}
