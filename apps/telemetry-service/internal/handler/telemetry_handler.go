package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"telemetry-service/internal/model"
	"telemetry-service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type TelemetryHandler struct {
	service *service.TelemetryService
}

func NewTelemetryHandler(svc *service.TelemetryService) *TelemetryHandler {
	return &TelemetryHandler{service: svc}
}

func (h *TelemetryHandler) RegisterRoutes(router *gin.RouterGroup) {
	telemetry := router.Group("/telemetry")
	{
		telemetry.GET("/:deviceId", h.GetTelemetry)
		telemetry.GET("/:deviceId/history", h.GetHistory)
		telemetry.POST("", h.ReceiveTelemetry)
		telemetry.GET("/stats", h.GetStats)
	}
}

func (h *TelemetryHandler) GetTelemetry(c *gin.Context) {
	deviceID, err := strconv.Atoi(c.Param("deviceId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid device ID"})
		return
	}

	limit := 100
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	telemetry, err := h.service.GetTelemetry(c.Request.Context(), deviceID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, telemetry)
}

func (h *TelemetryHandler) GetHistory(c *gin.Context) {
	deviceID, err := strconv.Atoi(c.Param("deviceId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid device ID"})
		return
	}

	fromStr := c.Query("from")
	toStr := c.Query("to")

	if fromStr == "" || toStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from and to query parameters are required"})
		return
	}

	from, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid from parameter"})
		return
	}

	to, err := time.Parse(time.RFC3339, toStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid to parameter"})
		return
	}

	telemetry, err := h.service.GetHistory(c.Request.Context(), deviceID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, telemetry)
}

func (h *TelemetryHandler) ReceiveTelemetry(c *gin.Context) {
	var t model.TelemetryCreate
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	telemetry, err := h.service.ReceiveTelemetry(c.Request.Context(), t)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, telemetry)
}

func (h *TelemetryHandler) GetStats(c *gin.Context) {
	deviceIDStr := c.Query("deviceId")
	if deviceIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "deviceId query parameter is required"})
		return
	}

	deviceID, err := strconv.Atoi(deviceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid device ID"})
		return
	}

	stats, err := h.service.GetStats(c.Request.Context(), deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *TelemetryHandler) HandleDeviceWebhook(c *gin.Context) {
	var webhook model.DeviceWebhook
	if err := c.ShouldBindJSON(&webhook); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.ProcessDeviceWebhook(c.Request.Context(), webhook); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "webhook processed"})
}

func (h *TelemetryHandler) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	deviceIDStr := c.Query("deviceId")
	if deviceIDStr == "" {
		conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"deviceId query parameter is required"}`))
		return
	}

	deviceID, _ := strconv.Atoi(deviceIDStr)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case <-ticker.C:
			telemetry, err := h.service.GetTelemetry(c.Request.Context(), deviceID, 1)
			if err != nil {
				continue
			}
			if len(telemetry) > 0 {
				data, _ := json.Marshal(telemetry[0])
				conn.WriteMessage(websocket.TextMessage, data)
			}
		}
	}
}
