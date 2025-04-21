package ssh

import (
	"NMS/client"
	"NMS/constants"
	"NMS/logger"
	"NMS/utils"
	"strings"
)

// Discovery handles discovery requests by running hostname
func Discovery(context map[string]interface{}, channel chan map[string]interface{}) {

	logger := logger.NewLogger("plugins", "ssh")

	logger.Info("Inside Discovery")

	errorArray := make([]map[string]interface{}, 0)

	result := make(map[string]interface{})

	sshClient := &client.SSHClient{}

	sshClient.SetContext(context)

	credentials, ok := context["credentials"].([]interface{})

	if !ok {

		errorArray = append(errorArray, utils.ErrorHandler("INVALID_CREDENTIALS", "credentials missing or invalid"))

		utils.SendResult(context, constants.Fail, result, errorArray, channel)

		return

	}

	_, credName, err := sshClient.Connect(credentials)
	if err != nil {

		errorArray = append(errorArray, utils.ErrorHandler(constants.CONNECTIONERROR, err.Error()))

		utils.SendResult(context, constants.Fail, result, errorArray, channel)

		return
	}

	command := "hostname"

	output, errorOutput, exitCode, err := sshClient.Execute(command)

	if err != nil || exitCode != 0 {

		errorArray = append(errorArray, utils.ErrorHandler(constants.COMMANDERROR, errorOutput))

		utils.SendResult(context, constants.Fail, result, errorArray, channel)

		return
	}

	context["credential.name"] = credName

	result[constants.IP] = context[constants.IP]

	result[constants.Hostname] = strings.Trim(output, "\r\n")

	utils.SendResult(context, constants.Success, result, errorArray, channel)
}
