package controllers

import (
	"fmt"
	"math"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ary/go-api/config"
	"github.com/ary/go-api/models"
	"github.com/ary/go-api/utils"
	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
)

// GetCategories godoc
// @Summary      Get all categories (paginated)
// @Description  Retrieve a list of categories with pagination support
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Param        page   query     int false "Page number (default is 1)"
// @Param        limit  query     int false "Number of items per page (default is 10)"
// @Success      200 {object} utils.SuccessResponse
// @Failure      400 {object} utils.ErrorResponse
// @Failure      500 {object} utils.ErrorResponse
// @Router       /categories [get]
func GetCategories(c *gin.Context) {
	var categories []models.Category

	// Ambil query parameter: page & limit
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	// Konversi ke integer
	page, err1 := strconv.Atoi(pageStr)
	limit, err2 := strconv.Atoi(limitStr)
	if err1 != nil || err2 != nil || page <= 0 || limit <= 0 {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid pagination parameters", nil)
		return
	}

	offset := (page - 1) * limit

	// Ambil total data untuk info total halaman
	var total int64
	if err := config.DB.Model(&models.Category{}).Count(&total).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to count categories", nil)
		return
	}

	// Ambil data berdasarkan offset dan limit
	if err := config.DB.Limit(limit).Offset(offset).Find(&categories).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to fetch categories", nil)
		return
	}

	// Response dengan metadata
	response := gin.H{
		"data":       categories,
		"page":       page,
		"limit":      limit,
		"total":      total,
		"totalPages": int(math.Ceil(float64(total) / float64(limit))),
	}

	utils.SendSuccessResponse(c, http.StatusOK, "Success", response)
}

// CreateCategory godoc
// @Summary      Create a new category
// @Description  Create a new category with name, detail, image_path, and icon
// @Tags         Categories
// @Accept       multipart/form-data
// @Produce      json
// @Param        name        formData string true "Category Name"
// @Param        detail      formData string true "Category Detail"
// @Param        image_path  formData file   true "Category Image"
// @Param        icon        formData file   true "Category Icon"
// @Success      201 {object} utils.SuccessResponse
// @Failure      400 {object} utils.ErrorResponse
// @Failure      500 {object} utils.ErrorResponse
// @Router       /categories [post]
func CreateCategory(c *gin.Context) {
	name := c.PostForm("name")
	detail := c.PostForm("detail")

	imageFile, err := c.FormFile("image_path")
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "image_path is required", nil)
		return
	}
	iconFile, err := c.FormFile("icon")
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "icon is required", nil)
		return
	}

	ext := filepath.Ext(imageFile.Filename)
	cleanUUID := strings.ReplaceAll(uuid.New().String(), "-", "")
	imageName := fmt.Sprintf("%s%s", cleanUUID, ext)
	imagePath := "uploads/" + imageName

	if err := c.SaveUploadedFile(imageFile, imagePath); err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to save image", nil)
		return
	}

	iconExt := filepath.Ext(iconFile.Filename)
	cleanUUID2 := strings.ReplaceAll(uuid.New().String(), "-", "")
	iconName := fmt.Sprintf("%s%s", cleanUUID2, iconExt)
	iconPath := "uploads/" + iconName

	if err := c.SaveUploadedFile(iconFile, iconPath); err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to save icon", nil)
		return
	}

	category := models.Category{
		Name:      name,
		Detail:    detail,
		ImagePath: imagePath,
		Icon:      iconPath,
	}

	if err := config.DB.Create(&category).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to create category", nil)
		return
	}

	utils.SendSuccessResponse(c, http.StatusCreated, "Success", category)
}

// DeleteCategory godoc
// @Summary      Delete a category
// @Description  Delete a category by ID and remove associated image/icon files
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Param        id   path      string false "Category ID (UUID)"
// @Success      200  {object}  utils.SuccessResponse
// @Failure      400  {object}  utils.ErrorResponse
// @Failure      404  {object}  utils.ErrorResponse
// @Failure      500  {object}  utils.ErrorResponse
// @Router       /categories/{id} [delete]
func DeleteCategory(c *gin.Context) {
	id := c.Param("id")

	var category models.Category
	// if err := config.DB.First(&category, id).Error; err != nil {
	// 	utils.SendErrorResponse(c, http.StatusNotFound, "Category not found", nil)
	// 	return
	// }

	if err := config.DB.First(&category, "id = ?", id).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusNotFound, "Category not found", nil)
		return
	}

	// Hapus file gambar dan icon jika ada
	if category.ImagePath != "" {
		if err := utils.DeleteFile(category.ImagePath); err != nil {
			fmt.Println("Failed to delete image:", err)
		}
	}

	if category.Icon != "" {
		if err := utils.DeleteFile(category.Icon); err != nil {
			fmt.Println("Failed to delete icon:", err)
		}
	}

	if err := config.DB.Delete(&category).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to delete category", nil)
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, "Category deleted", nil)
}

// UpdateCategory godoc
// @Summary      Update a category
// @Description  Update a category with optional image/icon replacement
// @Tags         Categories
// @Accept       multipart/form-data
// @Produce      json
// @Param        id          path     string    true  "Category ID"
// @Param        name        formData string false "Category Name"
// @Param        detail      formData string false "Category Detail"
// @Param        image_path  formData file   false "Category Image"
// @Param        icon        formData file   false "Category Icon"
// @Success      200 {object} utils.SuccessResponse
// @Failure      400 {object} utils.ErrorResponse
// @Failure      404 {object} utils.ErrorResponse
// @Failure      500 {object} utils.ErrorResponse
// @Router       /categories/{id} [patch]
func UpdateCategory(c *gin.Context) {
	id := c.Param("id")
	var category models.Category

	if err := config.DB.First(&category, "id = ?", id).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusNotFound, "Category not found", nil)
		return
	}

	name := c.PostForm("name")
	detail := c.PostForm("detail")

	// Jika ada file image baru
	imageFile, _ := c.FormFile("image_path")
	if imageFile != nil {
		// Hapus gambar lama
		utils.DeleteFile(category.ImagePath)

		ext := filepath.Ext(imageFile.Filename)
		newImageName := fmt.Sprintf("%s%s", strings.ReplaceAll(uuid.New().String(), "-", ""), ext)
		newImagePath := "uploads/" + newImageName

		if err := c.SaveUploadedFile(imageFile, newImagePath); err != nil {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to save new image", nil)
			return
		}
		category.ImagePath = newImagePath
	}

	// Jika ada file icon baru
	iconFile, _ := c.FormFile("icon")
	if iconFile != nil {
		// Hapus icon lama
		utils.DeleteFile(category.Icon)

		ext := filepath.Ext(iconFile.Filename)
		newIconName := fmt.Sprintf("%s%s", strings.ReplaceAll(uuid.New().String(), "-", ""), ext)
		newIconPath := "uploads/" + newIconName

		if err := c.SaveUploadedFile(iconFile, newIconPath); err != nil {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to save new icon", nil)
			return
		}
		category.Icon = newIconPath
	}

	// Update field jika ada
	if name != "" {
		category.Name = name
	}
	if detail != "" {
		category.Detail = detail
	}

	if err := config.DB.Save(&category).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to update category", nil)
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, "Category updated successfully", category)
}
