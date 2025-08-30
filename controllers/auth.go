package controllers

import (
	"net/http"
	"strconv"

	"github.com/ary/go-api/config"
	"github.com/ary/go-api/models"
	"github.com/ary/go-api/utils"
	"github.com/gin-gonic/gin"
)

func Login(c *gin.Context) {
	var input struct {
		Phone    string `json:"phone" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var account models.UserAccount
	if err := config.DB.Preload("User").First(&account, "phone = ?", input.Phone).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// cek password
	if !utils.CheckPasswordHash(account.PasswordHash, input.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Wrong password"})
		return
	}

	// generate token dengan data user
	token, err := utils.GenerateJWT(
		account.User.ID.String(),
		account.Phone,
		account.User.Name,
		account.User.Email,
		account.User.Address,
		account.User.Regency,
		account.User.District,
		strconv.FormatFloat(account.User.Lat, 'f', -1, 64),
		strconv.FormatFloat(account.User.Lang, 'f', -1, 64),
		account.User.PhotoUrl,
		account.Role,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login success",
		"token":   token,
	})
}
