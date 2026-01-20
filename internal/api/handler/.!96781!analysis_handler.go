package handler

import (
	"context"
	"encoding/json"
	"log"

	"copycat/internal/core/llm"
	"copycat/internal/model"
	"copycat/internal/repository"
	"copycat/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

