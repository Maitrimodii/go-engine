package client

import (
	"NMS/constants"
	"NMS/logger"
	"NMS/utils"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"golang.org/x/crypto/ssh"
	"time"
)

// SSHClient implements the Client interface for SSH
type SSHClient struct {
	ip string

	port int

	timeout time.Duration

	logger *logger.Logger

	client *ssh.Client
}

// SetContext configures the client with context data
func (client *SSHClient) SetContext(context map[string]interface{}) {

	client.logger = logger.NewLogger("client", "SSHClient")

	client.ip = utils.ToString(context[constants.IP])

	client.port = utils.ValidatePort(context)

	client.timeout = utils.ValidateTimeOut(context)

}

// Connect tries each credential, returns the first successful connection
func (client *SSHClient) Connect(credentials []interface{}) (interface{}, string, error) {

	for _, cred := range credentials {

		credential, ok := cred.(map[string]interface{})

		if !ok {

			client.logger.Error("Invalid credential format")

			continue
		}

		credType, _ := credential["credential.type"].(string)

		if credType != "ssh" {

			continue

		}

		attributes, ok := credential["attributes"].(map[string]interface{})

		if !ok {

			client.logger.Error("Attributes missing")

			continue

		}

		username := utils.ToString(attributes[constants.Username])

		password := utils.ToString(attributes[constants.Password])

		if username == "" || password == "" {

			client.logger.Error("Username or password missing")

			continue
		}

		config := &ssh.ClientConfig{

			User: username,

			Auth: []ssh.AuthMethod{ssh.Password(password)},

			HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: Use knownhosts.

			Timeout: client.timeout,
		}

		address := fmt.Sprintf("%s:%d", client.ip, client.port)

		var sshClient *ssh.Client

		operation := func() error {

			var err error

			sshClient, err = ssh.Dial("tcp", address, config)

			return err
		}

		retry := backoff.NewExponentialBackOff()

		retry.MaxElapsedTime = 3 * client.timeout

		if err := backoff.Retry(operation, retry); err != nil {

			client.logger.Error(fmt.Sprintf("Failed to connect to %s: %v", address, err))

			continue

		}

		client.client = sshClient

		client.logger.Info(fmt.Sprintf("Connected to %s with username %s", address, username))

		credName, _ := credential["credential.name"].(string)

		return sshClient, credName, nil
	}

	return nil, "", fmt.Errorf("no valid SSH credentials")
}

// Execute runs a command and returns output
func (client *SSHClient) Execute(command string) (string, string, int, error) {

	if client.client == nil {

		return "", "", -1, fmt.Errorf("no active connection")

	}

	session, err := client.client.NewSession()

	if err != nil {

		client.logger.Error(fmt.Sprintf("Failed to create session: %v", err))

		return "", "", -1, err

	}

	defer session.Close()

	output, err := session.CombinedOutput(command)

	if err != nil {

		client.logger.Error(fmt.Sprintf("Failed to run command '%s': %v", command, err))

		if exitErr, ok := err.(*ssh.ExitError); ok {

			return string(output), err.Error(), exitErr.ExitStatus(), err

		}

		return string(output), err.Error(), -1, err

	}

	client.logger.Info(fmt.Sprintf("Executed command '%s' on %s", command, client.ip))

	return string(output), "", 0, nil
}

// Close terminates the connection
func (client *SSHClient) Close() {

	if client.client != nil {

		client.client.Close()

		client.client = nil

		client.logger.Info("Closed SSH connection")

	}
}
