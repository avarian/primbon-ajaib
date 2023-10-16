package controllers

import (
	"context"
	"errors"
	"net/http"

	"github.com/avarian/primbon-ajaib-backend/model"
	"github.com/avarian/primbon-ajaib-backend/service/repository"
	"github.com/avarian/primbon-ajaib-backend/util"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type PostChatboxRequest struct {
	ChatboxCode string `json:"chatbox_code"`
	Message     string `json:"message" validate:"required"`
}

type OpenaiChatboxController struct {
	db        *gorm.DB
	validator *util.Validator
	apiKey    string
}

func NewOpenaiChatboxController(db *gorm.DB, validator *util.Validator, apiKey string) *OpenaiChatboxController {
	return &OpenaiChatboxController{
		db:        db,
		validator: validator,
		apiKey:    apiKey,
	}
}

func (s *OpenaiChatboxController) PostChatbox(c *gin.Context) {
	// bind data
	var req PostChatboxRequest
	if err := c.ShouldBind(&req); err != nil {
		log.WithField("reason", err).Error("error Binding")
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	// validate
	if err := s.validator.Validate.Struct(&req); err != nil {
		log.WithField("reason", err).Error("invalid Request")
		errs := err.(validator.ValidationErrors)
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{"error": errs.Translate(s.validator.Trans)})
		return
	}

	// log
	logCtx := log.WithFields(log.Fields{
		"api": "PostChatbox",
	})

	username := c.GetString("username")
	accountRepo := repository.NewAccountRepository(s.db)
	account, result := accountRepo.OneByEmail(username)
	if result.Error != nil {
		err := errors.New("error find account")
		if result.Error != nil {
			err = result.Error
		}
		logCtx.WithField("reason", err).Error("error find account")
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	} else if result.RowsAffected == 0 {
		account.ID = 1
	}

	chatboxRepo := repository.NewChatboxRepository(s.db)
	chatbox, result := chatboxRepo.OneByCodeAndAccountID(req.ChatboxCode, int(account.ID))
	if result.Error != nil {
		logCtx.WithField("reason", result.Error).Error("error find chatbox")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "error find chatbox"})
		return
	} else if result.RowsAffected == 0 {
		lenMsg := len(req.Message)
		if len(req.Message) > 250 {
			lenMsg = 250
		}
		code := uuid.New()
		chatbox, _ = chatboxRepo.Create(model.Chatbox{
			AccountID: account.ID,
			Code:      code.String(),
			Name:      req.Message[:lenMsg],
			// CreatedBy: username
		})
	}

	chatboxMessageRepo := repository.NewChatboxMessageRepository(s.db)
	chatboxMessage, result := chatboxMessageRepo.AllByChatboxCode(chatbox.Code)
	if result.Error != nil {
		logCtx.WithField("reason", result.Error).Error("error find chatbox message")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "error find chatbox message"})
		return
	}

	client := openai.NewClient(s.apiKey)

	base := openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "From now you are Primon Ajab!",
			},
		},
	}

	for _, v := range chatboxMessage {
		base.Messages = append(base.Messages, openai.ChatCompletionMessage{
			Role:    v.Role,
			Content: v.Content,
		})
	}

	base.Messages = append(base.Messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: req.Message,
	})

	resp, err := client.CreateChatCompletion(context.Background(), base)
	if err != nil {
		logCtx.WithFields(log.Fields{
			"reason": err.Error(),
		}).Error("failed get response chatbox")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "error generate chatbox"})
		return
	}

	chatboxMessageRepo.Create(model.ChatboxMessage{
		ChatboxCode: chatbox.Code,
		Role:        openai.ChatMessageRoleUser,
		Content:     req.Message,
	})

	chatboxMessageRepo.Create(model.ChatboxMessage{
		ChatboxCode: chatbox.Code,
		Role:        resp.Choices[0].Message.Role,
		Content:     resp.Choices[0].Message.Content,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Success!",
		"data": gin.H{
			"chatbox_code": chatbox.Code,
			"result":       resp.Choices[0].Message,
		},
	})
}

func (s *OpenaiChatboxController) GetListChatbox(c *gin.Context) {
	// log
	logCtx := log.WithFields(log.Fields{
		"api": "GetListChatbox",
	})
	username := c.GetString("username")
	accountRepo := repository.NewAccountRepository(s.db)
	account, result := accountRepo.OneByEmail(username)
	if result.Error != nil {
		err := errors.New("error find account")
		if result.Error != nil {
			err = result.Error
		}
		logCtx.WithField("reason", err).Error("error find account")
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	}

	chatboxRepo := repository.NewChatboxRepository(s.db)
	chatbox, result := chatboxRepo.AllByAccountID(int(account.ID))
	if result.Error != nil && !errors.Is(gorm.ErrRecordNotFound, result.Error) {
		logCtx.WithField("reason", result.Error).Error("error find chatbox message")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "error find chatbox message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success!",
		"data":    chatbox,
	})
}

func (s *OpenaiChatboxController) GetChatboxMessages(c *gin.Context) {
	// log
	logCtx := log.WithFields(log.Fields{
		"api": "GetListChatbox",
	})

	code := c.Param("code")
	username := c.GetString("username")
	accountRepo := repository.NewAccountRepository(s.db)
	account, result := accountRepo.OneByEmail(username)
	if result.Error != nil {
		err := errors.New("error find account")
		if result.Error != nil {
			err = result.Error
		}
		logCtx.WithField("reason", err).Error("error find account")
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	}

	chatboxRepo := repository.NewChatboxRepository(s.db)
	chatbox, result := chatboxRepo.OneByCodeAndAccountID(code, int(account.ID))
	if result.Error != nil || result.RowsAffected == 0 {
		logCtx.WithField("reason", result.Error).Error("error find chatbox")
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "error find chatbox"})
		return
	}

	chatboxMessageRepo := repository.NewChatboxMessageRepository(s.db)
	chatboxMessage, result := chatboxMessageRepo.AllByChatboxCode(chatbox.Code)
	if result.Error != nil {
		logCtx.WithField("reason", result.Error).Error("error find chatbox message")
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "error find chatbox message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success!",
		"data":    chatboxMessage,
	})
}
