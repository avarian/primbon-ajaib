package controllers

import (
	"errors"
	"net/http"
	"time"

	"github.com/avarian/primbon-ajaib-backend/model"
	"github.com/avarian/primbon-ajaib-backend/service/repository"
	"github.com/avarian/primbon-ajaib-backend/util"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type PostRegisterRequest struct {
	Name        string `json:"name"  validate:"required"`
	Email       string `json:"email"  validate:"required,email"`
	PhoneNumber string `json:"phone_number"  validate:"required"`
	Password    string `json:"password"  validate:"required"`
	Address     string `json:"address"  validate:"required"`
}

type PostLoginRequest struct {
	Email    string `json:"email"  validate:"required,email"`
	Password string `json:"password"  validate:"required"`
}

type JWTClaim struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Type      string `json:"type"`
	IsPremium bool   `json:"is_premium"`
	jwt.StandardClaims
}

type AccountController struct {
	db        *gorm.DB
	validator *util.Validator
	jwtSecret string
}

func NewAccountController(db *gorm.DB, validator *util.Validator, jwtSecret string) *AccountController {
	return &AccountController{
		db:        db,
		validator: validator,
		jwtSecret: jwtSecret,
	}
}

// RegisterAccount	goDocs
// @Summary      register an account
// @Description  register account with type CUSTOMER
// @Tags         Account
// @Produce      application/json
// @Param        tags body PostRegisterRequest true "Body Request"
// @Router       /register [post]
func (s *AccountController) PostRegister(c *gin.Context) {
	// bind data
	var req PostRegisterRequest
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
		"email": req.Email,
		"api":   "PostRegister",
	})

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 5)
	if err != nil {
		logCtx.WithField("reason", err).Error("error hash password")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	account := model.Account{
		Address:     req.Address,
		Email:       req.Email,
		Name:        req.Name,
		PhoneNumber: req.PhoneNumber,
		Password:    string(hashedPassword),
	}

	accountRepo := repository.NewAccountRepository(s.db)
	account, result := accountRepo.Create(account)
	if result.Error != nil {
		logCtx.WithField("reason", err).Error("error create account")
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Sucess!",
		"data":    account,
	})
}

// LoginAccount	goDocs
// @Summary      login an account
// @Description  login account with return JWT token
// @Tags         Account
// @Produce      application/json
// @Param        tags body PostLoginRequest true "Body Request"
// @Router       /login [post]
func (s *AccountController) PostLogin(c *gin.Context) {
	// bind data
	var req PostLoginRequest
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
		"email": req.Email,
		"api":   "PostLogin",
	})

	accountRepo := repository.NewAccountRepository(s.db)
	account, result := accountRepo.OneByEmail(req.Email)
	if result.Error != nil || result.RowsAffected == 0 {
		err := errors.New("not found")
		if result.Error != nil {
			err = result.Error
		}
		logCtx.WithField("reason", err).Error("error find account")
		c.AbortWithStatusJSON(http.StatusUnauthorized, nil)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(req.Password)); err != nil {
		// If the two passwords don't match, return a 401 status
		logCtx.WithField("reason", err).Error("error compare password")
		c.AbortWithStatusJSON(http.StatusUnauthorized, nil)
		return
	}

	isPremium := false
	if time.Now().Before(time.Time(account.ValidUntil)) {
		isPremium = true
	}
	expirationTime := time.Now().Add(7 * 24 * time.Hour)
	claims := &JWTClaim{
		Email:     account.Email,
		Username:  account.Email,
		Type:      account.Type,
		IsPremium: isPremium,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		logCtx.WithField("reason", err).Error("error generate jwt")
		c.AbortWithStatusJSON(http.StatusUnauthorized, nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
	})
}
