package handler

import (
	"net/http"

	"github.com/alvisLu/go-shorten/internal/service"
	"github.com/gin-gonic/gin"
)

type URLHandler struct {
	urlService service.URLService
}

func NewURLHandler(s service.URLService) *URLHandler {
	return &URLHandler{urlService: s}
}

func (h *URLHandler) CreateShortUrl(c *gin.Context) {
	var request struct {
		LongURL string `json:"url"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	shortURL, err := h.urlService.CreateShortURL(request.LongURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create short URL"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"short_url": shortURL})
}

func (h *URLHandler) GetOriginalURL(c *gin.Context) {
	code := c.Param("code")
	originalURL, err := h.urlService.GetOriginalURL(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get original URL"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"original_url": originalURL})
}
