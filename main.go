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

	contexts, err := utils.Decode(os.Args[1])
	if err != nil {
		fmt.Printf("Failed to decode context: %v\n", err)
		os.Exit(1)
	}

	channel := make(chan map[string]interface{}, len(contexts))
	var wg sync.WaitGroup

	for _, context := range contexts {
		requestType, ok := context[constants.RequestType].(string)
		if !ok {
			fmt.Println("Missing request type")
			continue
		}

		wg.Add(1)

		switch requestType {
		case constants.Discovery:
			go func(ctx map[string]interface{}) {
				defer wg.Done()
				ssh.Discovery(ctx, channel)
			}(context)
		case constants.Collect:
			go func(ctx map[string]interface{}) {
				defer wg.Done()
				ssh.Collect(ctx, channel)
			}(context)
		default:
			fmt.Printf("Unknown request type: %s\n", requestType)
			wg.Done()
			continue
		}
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
