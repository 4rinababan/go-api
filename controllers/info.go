package controllers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ary/go-api/config"
	"github.com/ary/go-api/models"
	"github.com/ary/go-api/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetInfo godoc
// @Summary      Get site information
// @Description  Retrieve site information such as phone, address, and location coordinates
// @Tags         Info
// @Produce      json
// @Success      200 {object} utils.SuccessResponse
// @Failure      404 {object} utils.ErrorResponse
// @Router       /info [get]
func GetInfo(c *gin.Context) {
	var info models.Info

	if err := config.DB.First(&info).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusNotFound, "Info not found", nil)
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, "Success", info)
}

// UpsertInfo godoc
// @Summary      Upsert site information
// @Description  Create or update site information such as phone, address, and location coordinates
// @Tags         Info
// @Accept       json
// @Produce      json
// @Param        info body models.Info true "Site Information"
// @Success      201 {object} utils.SuccessResponse
// @Success      200 {object} utils.SuccessResponse
// @Failure      400 {object} utils.ErrorResponse
// @Failure      500 {object} utils.ErrorResponse
// @Router       /info [post]
func UpsertInfo(c *gin.Context) {
	var input models.Info

	// Bind form data (support multipart/form-data)
	if err := c.ShouldBind(&input); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid form data", nil)
		return
	}

	// Handle file upload
	file, err := c.FormFile("photo")
	if err != nil {
		fmt.Println("Photo not received or error:", err)
	} else {
		// Simpan file ke folder 'uploads'
		filePath := "uploads/" + file.Filename
		if err := c.SaveUploadedFile(file, filePath); err != nil {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to upload photo", nil)
			return
		}
		input.ImagePath = filePath
		fmt.Println("Photo saved to:", filePath)
	}

	var existing models.Info
	err = config.DB.First(&existing).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Insert baru
		if err := config.DB.Create(&input).Error; err != nil {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to create info", nil)
			return
		}
		utils.SendSuccessResponse(c, http.StatusCreated, "Created successfully", input)
		return
	} else if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Database error", nil)
		return
	}

	// Update record lama
	if err := config.DB.Model(&existing).Updates(input).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to update info", nil)
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, "Updated successfully", existing)
}
