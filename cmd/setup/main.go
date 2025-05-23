package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/blackarbiter/go-sac/pkg/mq/rabbitmq" // Updated to match the module name
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	log.Println("Starting RabbitMQ setup...")

	// Get connection parameters from environment variables or use defaults
	rabbitURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")

	// Connect to RabbitMQ
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	// Setup RabbitMQ exchanges, queues, and bindings
	if err := rabbitmq.Setup(conn); err != nil {
		log.Fatalf("Failed to setup RabbitMQ: %v", err)
	}

	log.Println("RabbitMQ setup complete!")

	// For manual setup, wait for signal to exit
	if getEnv("SETUP_MODE", "oneshot") == "daemon" {
		log.Println("Running in daemon mode. Press CTRL+C to exit.")
		waitForSignal()
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func waitForSignal() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
}
