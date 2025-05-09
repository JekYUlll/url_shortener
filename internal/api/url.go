package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jekyulll/url_shortener/internal/dto"
	"github.com/jekyulll/url_shortener/internal/service"
)

type URLService interface {
	CreateURL(ctx context.Context, req dto.CreateURLRequest) (*dto.CreateURLResponse, error)
	GetURL(ctx context.Context, shortCode string) (string, error)
}

type URLHandler struct {
	urlService URLService
}

func NewURLHandler(urlService URLService) *URLHandler {
	return &URLHandler{
		urlService: urlService,
	}
}

// POST /api/url original_url, custom_code, duration -> 短url, 过期时间
func (h *URLHandler) CreateURL(c *gin.Context) {
	// 1.提取数据
	var req dto.CreateURLRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	// 2.验证数据格式
	validate := validator.New()
	err := validate.Struct(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	// 3. 调用业务函数
	resp, err := h.urlService.CreateURL(c.Request.Context(), req)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, service.ErrShortCodeTaken) {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	// 4. 返回响应
	c.JSON(http.StatusCreated, resp)
}

// GET /:code 把短url重定向到长URL
// TODO 每次访问时，统计该短链接访问次数
func (h *URLHandler) RedirectURL(c *gin.Context) {
	// 取出 code
	shortCode := c.Param("code")
	// shortcode -> url
	originalURL, err := h.urlService.GetURL(c.Request.Context(), shortCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	if originalURL == "" {
		c.JSON(404, gin.H{
			"error": "no such short code",
		})
		return
	}
	// 永久重定向（浏览器会缓存）
	c.Redirect(http.StatusPermanentRedirect, originalURL)
}

// TODO
// GET /api/url/:code
// 获取该 url 的浏览量
// 1. 通过短链接 url，到数据库中获取 views1
// 2. 去 redis 缓存中查看浏览量 views2
// 3. 返回 views1 + views2
