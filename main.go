package main

import (
	"NMS/constants"
	"NMS/ssh"
	"NMS/utils"
	"fmt"
	"os"
	"sync"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Missing context JSON")
		os.Exit(1)
	}

	input, err := utils.Decode(os.Args[1])
	if err != nil {
		fmt.Printf("Failed to decode context: %v\n", err)
		os.Exit(1)
	}

	if len(input.Contexts) == 0 {
		fmt.Println("No contexts provided")
		os.Exit(1)
	}

	requestType := input.RequestType

	if requestType != constants.Discovery && requestType != constants.Collect {

		fmt.Printf("Invalid request type: %s\n", requestType)

		os.Exit(1)
	}

	channel := make(chan map[string]interface{}, len(input.Contexts))

	operation := ssh.Discovery

	if requestType == constants.Collect {

		operation = ssh.Collect

	}

	var wg sync.WaitGroup

	for _, context := range input.Contexts {
		wg.Add(1)
		go func(ctx map[string]interface{}) {
			defer wg.Done()
			operation(ctx, channel)
		}(context)
	}

	// Close channel after all goroutines finish
	go func() {
		wg.Wait()
		close(channel)
	}()

	// Collect results
	results := make([]map[string]interface{}, 0)
	for result := range channel {
		results = append(results, result)
	}

	// Encode results
	encoded, err := utils.Encode(results)
	if err != nil {
		fmt.Printf("Failed to encode results: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(encoded)
}
