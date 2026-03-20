package shorturl

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

func (h *Handler) CreateShortUrl(c *gin.Context) {
	var request struct {
		LongURL string `json:"url"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	shortURL, err := h.service.CreateShortURL(request.LongURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create short URL"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"short_url": shortURL})
}

func (h *Handler) GetOriginalURL(c *gin.Context) {
	code := c.Param("code")
	originalURL, err := h.service.GetOriginalURL(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get original URL"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"original_url": originalURL})
}
