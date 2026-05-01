package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// WebhookHandler handles webhook-related requests.
type WebhookHandler struct{}

// NewWebhookHandler creates a new WebhookHandler.
func NewWebhookHandler() *WebhookHandler {
	return &WebhookHandler{}
}

// ReceiveWebhook handles POST /api/webhooks
// It accepts any arbitrary JSON payload, logs it, and returns an acknowledgement.
func (h *WebhookHandler) ReceiveWebhook(c *gin.Context) {
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid JSON payload",
		})
		return
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal webhook payload: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to process payload",
		})
		return
	}

	log.Printf("Webhook received: %s", string(data))

	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "webhook received",
	})
}
