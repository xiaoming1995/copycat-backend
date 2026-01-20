package handler

import (
	"errors"

	"copycat/config"
	"copycat/internal/model"
	"copycat/internal/repository"
	"copycat/pkg/response"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserHandler 用户处理器
type UserHandler struct {
	userRepo repository.UserRepository
}

// NewUserHandler 创建用户处理器
func NewUserHandler(userRepo repository.UserRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Nickname string `json:"nickname"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token string      `json:"token"`
	User  *model.User `json:"user"`
}

// Register 用户注册
// @Summary 用户注册
// @Tags User
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "注册信息"
// @Success 200 {object} response.Response
// @Router /api/v1/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	// 检查邮箱是否已存在
	_, err := h.userRepo.GetByEmail(c.Request.Context(), req.Email)
	if err == nil {
		response.BadRequest(c, "email already exists")
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) && err != nil {
		response.ServerError(c, "failed to check email")
		return
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.ServerError(c, "failed to hash password")
		return
	}

	user := &model.User{
		Email:    req.Email,
		Password: string(hashedPassword),
		Nickname: req.Nickname,
	}

	if err := h.userRepo.Create(c.Request.Context(), user); err != nil {
		response.ServerError(c, "failed to create user")
		return
	}

	response.SuccessWithMessage(c, "user registered successfully", nil)
}

// Login 用户登录
// @Summary 用户登录
// @Tags User
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录信息"
// @Success 200 {object} response.Response{data=LoginResponse}
// @Router /api/v1/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	// 查找用户
	user, err := h.userRepo.GetByEmail(c.Request.Context(), req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Unauthorized(c, "invalid email or password")
			return
		}
		response.ServerError(c, "failed to get user")
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		response.Unauthorized(c, "invalid email or password")
		return
	}

	// 生成 JWT token
	token, err := generateToken(user.ID)
	if err != nil {
		response.ServerError(c, "failed to generate token")
		return
	}

	response.Success(c, LoginResponse{
		Token: token,
		User:  user,
	})
}

// GetProfile 获取当前用户信息
// @Summary 获取当前用户信息
// @Tags User
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=model.User}
// @Router /api/v1/user/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	user, err := h.userRepo.GetByID(c.Request.Context(), userID.(int64))
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}

	response.Success(c, user)
}

// generateToken 生成 JWT Token
func generateToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * time.Duration(config.AppCfg.JWT.ExpireHours)).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.AppCfg.JWT.Secret))
}

// UpdateProfileRequest 更新资料请求
type UpdateProfileRequest struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Bio      string `json:"bio"`
}

// UpdateProfile 更新用户资料
// @Summary 更新用户资料
// @Tags User
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body UpdateProfileRequest true "资料信息"
// @Success 200 {object} response.Response{data=model.User}
// @Router /api/v1/user/profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "用户未登录")
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	user, err := h.userRepo.GetByID(c.Request.Context(), userID.(int64))
	if err != nil {
		response.NotFound(c, "用户不存在")
		return
	}

	// 更新字段
	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	if req.Bio != "" {
		user.Bio = req.Bio
	}

	if err := h.userRepo.Update(c.Request.Context(), user); err != nil {
		response.ServerError(c, "更新失败")
		return
	}

	response.SuccessWithMessage(c, "资料更新成功", user)
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Tags User
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body ChangePasswordRequest true "密码信息"
// @Success 200 {object} response.Response
// @Router /api/v1/user/password [put]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "用户未登录")
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	user, err := h.userRepo.GetByID(c.Request.Context(), userID.(int64))
	if err != nil {
		response.NotFound(c, "用户不存在")
		return
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		response.BadRequest(c, "原密码错误")
		return
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		response.ServerError(c, "密码加密失败")
		return
	}

	user.Password = string(hashedPassword)
	if err := h.userRepo.Update(c.Request.Context(), user); err != nil {
		response.ServerError(c, "密码更新失败")
		return
	}

	response.SuccessWithMessage(c, "密码修改成功", nil)
}
