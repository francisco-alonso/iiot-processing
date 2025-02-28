package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/pubsub"
)

// SensorData represents the JSON structure received from sensors.
type SensorData struct {
	SensorID    int     `json:"sensor_id"`
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Timestamp   string  `json:"timestamp"`
}

func main() {
	ctx := context.Background()

	// Read environment variables
	projectID := os.Getenv("PROJECT_ID")
	subscriptionID := os.Getenv("SUBSCRIPTION_ID")
	credentialsFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")

	if projectID == "" || subscriptionID == "" || credentialsFile == "" {
		log.Fatalf("Missing required environment variables: PROJECT_ID, SUBSCRIPTION_ID, or GOOGLE_APPLICATION_CREDENTIALS")
	}

	// Create Pub/Sub client
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create Pub/Sub client: %v", err)
	}
	defer client.Close()

	// Reference the subscription
	sub := client.Subscription(subscriptionID)

	fmt.Println("Listening for Pub/Sub messages...")

	// Process messages
	err = sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		var data SensorData
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			log.Printf("Error parsing message: %v", err)
			msg.Nack()
			return
		}

		// Add timestamp
		data.Timestamp = time.Now().Format(time.RFC3339)

		// Print received message
		fmt.Printf("Received data: %+v\n", data)

		msg.Ack() // Acknowledge successful processing
	})

	if err != nil {
		log.Fatalf("Error receiving messages: %v", err)
	}
}
