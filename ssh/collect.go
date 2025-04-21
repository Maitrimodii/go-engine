package ssh

import (
	"NMS/client"
	"NMS/constants"
	"NMS/logger"
	"NMS/utils"
	"fmt"
	"strconv"
	"strings"
)

// Collect handles collect requests by running batched metric commands
func Collect(context map[string]interface{}, channel chan map[string]interface{}) {

	logger := logger.NewLogger("plugins", "ssh")

	logger.Info("Inside Collect")

	errorArray := make([]map[string]interface{}, 0)

	result := make(map[string]interface{})

	sshClient := &client.SSHClient{}

	sshClient.SetContext(context)

	credentials, ok := context["credentials"].([]interface{})

	if !ok || len(credentials) == 0 {

		errorArray = append(errorArray, utils.ErrorHandler("INVALID_CREDENTIALS", "credentials missing or invalid"))

		utils.SendResult(context, constants.Fail, result, errorArray, channel)

		return
	}

	_, _, err := sshClient.Connect(credentials)
	if err != nil {

		errorArray = append(errorArray, utils.ErrorHandler(constants.CONNECTIONERROR, err.Error()))

		utils.SendResult(context, constants.Fail, result, errorArray, channel)

		return
	}

	commands := []string{
		`top -bn1 | grep "Cpu(s)" | awk '{print "system.cpu.user.percent:" $2 "\nsystem.cpu.idle.percent:" $8}'`,
		`free -b | awk '/Mem:/ {print "system.memory.total.bytes:" $2 "\nsystem.memory.used.bytes:" $3}'`,
		`df -B1 / | tail -1 | awk '{print "system.disk.total.bytes:" $2 "\nsystem.disk.used.bytes:" $3}'`,
		`cat /proc/net/dev | grep eth0 | awk '{print "system.network.in.bytes:" $2}'`,
		`echo "system.os.name:$(uname -s)"`,
	}

	batches := make([]string, 0)

	batchSize := 10

	for i := 0; i < len(commands); i += batchSize {

		end := i + batchSize

		if end > len(commands) {

			end = len(commands)

		}

		batches = append(batches, strings.Join(commands[i:end], "; "))

	}

	for i, batch := range batches {

		logger.Info(fmt.Sprintf("Executing batch %d", i+1))

		output, errorOutput, exitCode, err := sshClient.Execute(batch)

		if err != nil || exitCode != 0 {

			errorArray = append(errorArray, utils.ErrorHandler(constants.COMMANDERROR, errorOutput))

			logger.Error(fmt.Sprintf("Batch %d error: %v", i+1, errorOutput))

			continue

		}

		lines := strings.Split(strings.TrimSpace(output), "\n")

		for _, line := range lines {

			if line == "" {

				continue

			}

			parts := strings.SplitN(line, ":", 2)

			if len(parts) != 2 {

				continue

			}

			key := strings.TrimSpace(parts[0])

			value := strings.TrimSpace(parts[1])

			if utils.MetricsMap[key] == "Count" {

				floatVal, err := strconv.ParseFloat(value, 64)

				if err != nil {

					logger.Error(fmt.Sprintf("Failed to parse %s: %v", key, err))

					continue

				}

				result[key] = floatVal

			} else {

				result[key] = value

			}
		}
	}

	sshClient.Close()

	if len(errorArray) > 0 {

		context[constants.Status] = constants.Fail

	} else {

		context[constants.Status] = constants.Success

	}

	utils.SendResult(context, constants.Success, result, errorArray, channel)

}
