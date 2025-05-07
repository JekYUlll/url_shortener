package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jekyulll/url_shortener/internal/dto"
)

// TODO
// POST /api/url original_url, custom_code, duration -> 短url, 过期时间
// GET /:code 把短url重定向到长URL

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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	// 4. 返回响应
	c.JSON(http.StatusCreated, resp)
}

// GET /:code 重定向
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
	// 永久重定向（浏览器会缓存）
	c.Redirect(http.StatusPermanentRedirect, originalURL)
}
