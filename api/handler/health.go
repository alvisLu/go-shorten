package handler

import (
	"net/http"

	"github.com/alvisLu/go-shorten/internal/service"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service service.Service
}

func NewHandler(s service.Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) Health(c *gin.Context) {
	response := h.service.Health()
	c.JSON(http.StatusOK, gin.H{"message": response})
}
