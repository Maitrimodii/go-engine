package utils

import (
	"NMS/constants"
	"encoding/json"
	"fmt"
	"time"
)

// MetricsMap defines metric types
var MetricsMap = map[string]string{
	"system.cpu.user.percent":   "Count",
	"system.cpu.idle.percent":   "Count",
	"system.memory.total.bytes": "Count",
	"system.memory.used.bytes":  "Count",
	"system.disk.total.bytes":   "Count",
	"system.disk.used.bytes":    "Count",
	"system.network.in.bytes":   "Count",
	"system.os.name":            "String",
}

// ToString converts interface to string
func ToString(val interface{}) string {
	if val == nil {

		return ""

	}

	return fmt.Sprintf("%v", val)
}

// ValidatePort returns the port from context or default
func ValidatePort(context map[string]interface{}) int {

	if port, ok := context[constants.Port].(float64); ok {

		return int(port)

	}

	return 22
}

// ValidateTimeOut returns the timeout from context or default
func ValidateTimeOut(context map[string]interface{}) time.Duration {
	if timeout, ok := context["timeout"].(float64); ok && timeout > 0 {
		return time.Duration(timeout) * time.Second
	}
	return 10 * time.Second // Default timeout
}

// ErrorHandler creates an error map
func ErrorHandler(code, message string) map[string]interface{} {
	return map[string]interface{}{
		"code":    code,
		"message": message,
	}
}

// Decode parses JSON context
func Decode(input string) ([]map[string]interface{}, error) {
	var contexts []map[string]interface{}
	if err := json.Unmarshal([]byte(input), &contexts); err != nil {
		return nil, err
	}
	return contexts, nil
}

// Encode
func Encode(results []map[string]interface{}) (string, error) {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func SendResult(context map[string]interface{}, status string, result map[string]interface{}, errors []map[string]interface{}, channel chan map[string]interface{}) {
	context[constants.Status] = status
	context[constants.Result] = result
	context[constants.Error] = errors
	channel <- context
}
