package queue

import (
	"encoding/json"
	"fmt"
	"log"
	"tsimserver/config"

	"github.com/rabbitmq/amqp091-go"
)

var Connection *amqp091.Connection
var Channel *amqp091.Channel

const (
	SMSQueue            = "sms_queue"
	USSDQueue           = "ussd_queue"
	AlarmQueue          = "alarm_queue"
	DeviceQueue         = "device_queue"
	DeliveryReportQueue = "deliveryreport"
)

// Connect establishes RabbitMQ connection
func Connect() error {
	cfg := config.AppConfig.RabbitMQ

	var err error
	Connection, err = amqp091.Dial(cfg.URL)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	Channel, err = Connection.Channel()
	if err != nil {
		return fmt.Errorf("failed to open RabbitMQ channel: %v", err)
	}

	// Declare queues
	if err := declareQueues(); err != nil {
		return fmt.Errorf("failed to declare queues: %v", err)
	}

	log.Println("RabbitMQ connection established successfully")
	return nil
}

// Close closes RabbitMQ connection
func Close() error {
	if Channel != nil {
		Channel.Close()
	}
	if Connection != nil {
		return Connection.Close()
	}
	return nil
}

// declareQueues declares all necessary queues
func declareQueues() error {
	queues := []string{SMSQueue, USSDQueue, AlarmQueue, DeviceQueue, DeliveryReportQueue}

	for _, queueName := range queues {
		_, err := Channel.QueueDeclare(
			queueName, // name
			true,      // durable
			false,     // delete when unused
			false,     // exclusive
			false,     // no-wait
			nil,       // arguments
		)
		if err != nil {
			return fmt.Errorf("failed to declare queue %s: %v", queueName, err)
		}
	}

	return nil
}

// PublishMessage publishes a message to specified queue
func PublishMessage(queueName string, message interface{}) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	err = Channel.Publish(
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish message: %v", err)
	}

	return nil
}

// ConsumeMessages starts consuming messages from specified queue
func ConsumeMessages(queueName string, handler func([]byte) error) error {
	msgs, err := Channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %v", err)
	}

	go func() {
		for d := range msgs {
			if err := handler(d.Body); err != nil {
				log.Printf("Error handling message from %s: %v", queueName, err)
				d.Nack(false, true) // Negative acknowledgment, requeue
			} else {
				d.Ack(false) // Acknowledge message
			}
		}
	}()

	log.Printf("Started consuming messages from queue: %s", queueName)
	return nil
}

// PublishSMSCommand publishes SMS command to queue
func PublishSMSCommand(deviceID string, target string, message string, simSlot int, internalLogID int) error {
	smsCommand := map[string]interface{}{
		"device_id":     deviceID,
		"type":          "send_sms",
		"target":        target,
		"message":       message,
		"simSlot":       simSlot,
		"internalLogId": internalLogID,
	}

	return PublishMessage(SMSQueue, smsCommand)
}

// PublishUSSDCommand publishes USSD command to queue
func PublishUSSDCommand(deviceID string, ussdCode string, simSlot int, internalLogID int) error {
	ussdCommand := map[string]interface{}{
		"device_id":     deviceID,
		"type":          "ussd_command",
		"ussdCode":      ussdCode,
		"simSlot":       simSlot,
		"internalLogId": internalLogID,
	}

	return PublishMessage(USSDQueue, ussdCommand)
}

// PublishAlarm publishes alarm to queue
func PublishAlarm(deviceID string, alarmType string, message string, severity string) error {
	alarm := map[string]interface{}{
		"device_id":  deviceID,
		"type":       "alarm",
		"alarm_type": alarmType,
		"message":    message,
		"severity":   severity,
	}

	return PublishMessage(AlarmQueue, alarm)
}

// PublishDeviceCommand publishes device command to queue
func PublishDeviceCommand(deviceID string, command string, data interface{}) error {
	deviceCommand := map[string]interface{}{
		"device_id": deviceID,
		"command":   command,
		"data":      data,
	}

	return PublishMessage(DeviceQueue, deviceCommand)
}

// PublishToQueue publishes message to specified queue (alias for PublishMessage)
func PublishToQueue(queueName string, message interface{}) error {
	return PublishMessage(queueName, message)
}

// GetChannel returns RabbitMQ channel
func GetChannel() *amqp091.Channel {
	return Channel
}
