package api

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jekyulll/url_shortener/internal/dto"
	"github.com/jekyulll/url_shortener/internal/service"
	"net/http"
)

type UserService interface {
	Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error)
	IsEmailAvailable(ctx context.Context, email string) error
	Register(ctx context.Context, req dto.RegisterReqeust) (*dto.LoginResponse, error)
	SendEmailCode(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, req dto.ForgetPasswordReqeust) (*dto.LoginResponse, error)
}

// UserHandler 处理用户相关的HTTP请求
type UserHandler struct {
	userService UserService
}

func NewUserHandler(userService UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) Login(c *gin.Context) error {
	var req dto.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return err
	}

	resp, err := h.userService.Login(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrUserNameOrPasswordFailed) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			// TODO check
			return err
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return err
	}
	c.JSON(http.StatusOK, resp)
	return nil
}
