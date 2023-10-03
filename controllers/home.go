package controllers

import (
	"github.com/gin-gonic/gin"
)

type HomeController struct {
	// db *gorm.DB
}

func NewHomeController() *HomeController {
	return &HomeController{
		// db: db,
	}
}

func (s *HomeController) GetHome(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Welcome home!",
	})
}
