package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) Health(c *gin.Context) {
	response := h.service.Health()
	c.JSON(http.StatusOK, gin.H{"message": response})
}
