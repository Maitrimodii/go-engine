package client

// Client defines the interface for protocol clients (SSH, SNMP, etc.)
type Client interface {

	// SetContext configures the client with context data (ip, port, timeout)
	SetContext(context map[string]interface{})

	// Connect tries credentials, returns connection, credential name, and error
	Connect(credentials []interface{}) (interface{}, string, error)

	// Execute runs a command and returns stdout, stderr, exit code, and error
	Execute(command string) (string, string, int, error)

	// Close terminates the connection
	Close()
}
