package consumer

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

const apiGatewayWebhookURL = "http://api-gateway_devcontainer-app-1:8080/api/webhooks"

func (c *Consumer) callWebhook(payload webhookPayload) {
	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal payload for webhook: %v", err)
		return
	}

	resp, err := http.Post(apiGatewayWebhookURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Failed to call api-gateway webhook for payment %s: %v", payload.PaymentID, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Webhook returned non-200 status %d for payment %s", resp.StatusCode, payload.PaymentID)
		return
	}

	log.Printf("Successfully called webhook for payment %s", payload.PaymentID)
}
