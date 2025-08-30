package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/ary/go-api/config"
	"github.com/ary/go-api/models"
	"github.com/ary/go-api/utils"

	"github.com/gin-gonic/gin"
)

// @Summary Get all users
// @Description Retrieve a list of all registered users
// @Tags Users
// @Produce json
// @Success 200 {array} models.User
// @Failure 500 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Router /users [get]
func GetUsers(c *gin.Context) {
	var users []models.User
	if err := config.DB.Find(&users).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to fetch users", nil)
		return
	}
	if len(users) == 0 {
		utils.SendErrorResponse(c, http.StatusNotFound, "No users found", nil)
		return
	}
	utils.SendSuccessResponse(c, http.StatusOK, "Success", users)
}

// @Summary Get user by ID
// @Description Retrieve user details by their ID
// @Tags Users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} models.User
// @Failure 404 {object} utils.ErrorResponse
// @Router /users/{id} [get]
func GetUserByID(c *gin.Context) {
	id := c.Param("id")
	var user models.User

	// Cari user berdasarkan ID
	if err := config.DB.First(&user, "id = ?", id).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusNotFound, "User not found", nil)
		return
	}

	// Generate token untuk user ini
	token, err := utils.GenerateJWT(
		user.ID.String(),
		user.Phone,
		user.Name,
		user.Email,
		user.Address,
		user.Regency,
		user.District,
		strconv.FormatFloat(user.Lat, 'f', -1, 64),
		strconv.FormatFloat(user.Lang, 'f', -1, 64),
		user.PhotoUrl,
		"user",
	)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to generate token", nil)
		return
	}

	// Buat response
	response := struct {
		User  models.User `json:"user"`
		Token string      `json:"token"`
	}{
		User:  user,
		Token: token,
	}

	utils.SendSuccessResponse(c, http.StatusOK, "Success", response)
}

// func CreateUser(c *gin.Context) {
// 	var user models.User
// 	if err := c.ShouldBindJSON(&user); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}
// 	config.DB.Create(&user)
// 	c.JSON(http.StatusOK, user)
// }

// @Summary Check if user is active
// @Description Check whether a user account is active by their ID
// @Tags Users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} map[string]bool
// @Failure 404 {object} utils.ErrorResponse
// @Router /users/{id}/is-active [get]
func CheckUserIsActive(c *gin.Context) {
	id := c.Param("id")
	var user models.User

	// Cari user berdasarkan ID
	if err := config.DB.First(&user, "id = ?", id).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusNotFound, "User not found", nil)
		return
	}

	// Return status aktif
	utils.SendSuccessResponse(c, http.StatusOK, "Success", gin.H{
		"is_active": user.IsActive,
	})
}

// @Summary Register new user
// @Description Create a new user account
// @Tags Users
// @Accept json
// @Produce json
// @Param user body models.User true "User data"
// @Success 201 {object} models.User
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /users [post]
func CreateUser(c *gin.Context) {
	var user models.User

	if err := c.ShouldBindJSON(&user); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := config.DB.Create(&user).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to create user")
		return
	}

	utils.SendSuccessResponse(c, http.StatusCreated, "user created successfully", user)
}

// @Router /users/{id} [patch]
func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User

	if err := config.DB.First(&user, "id = ?", id).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusNotFound, "User not found", nil)
		return
	}

	// Parse multipart form (supaya PostForm dan FormFile bisa dibaca)
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		fmt.Println("Failed to parse multipart form:", err)
	}

	// Ambil form-data
	name := c.PostForm("name")
	email := c.PostForm("email")
	phone := c.PostForm("phone")
	addressLocation := c.PostForm("address_location")
	addressCity := c.PostForm("address_city")
	addressDistrict := c.PostForm("address_district")
	latitude := c.PostForm("latitude")
	longitude := c.PostForm("longitude")

	// **Log semua value yang diterima**
	fmt.Println("===== RECEIVED FORM DATA =====")
	fmt.Println("Name:", name)
	fmt.Println("Email:", email)
	fmt.Println("Phone:", phone)
	fmt.Println("Address Location:", addressLocation)
	fmt.Println("Address City:", addressCity)
	fmt.Println("Address District:", addressDistrict)
	fmt.Println("Latitude:", latitude)
	fmt.Println("Longitude:", longitude)

	// Handle file upload
	file, err := c.FormFile("photo")
	if err != nil {
		fmt.Println("Photo not received or error:", err)
	} else {
		fmt.Println("Photo received:", file.Filename, "Size:", file.Size)
	}

	// parsing string ke float64
	_lat, _ := strconv.ParseFloat(latitude, 64) // ignore error sementara
	_lang, _ := strconv.ParseFloat(longitude, 64)

	// Update field kalau ada value
	if name != "" {
		user.Name = name
	}
	if email != "" {
		user.Email = email
	}
	if phone != "" {
		user.Phone = phone
	}
	if addressLocation != "" {
		user.Address = addressLocation
	}
	if addressCity != "" {
		user.Regency = addressCity
	}
	if addressDistrict != "" {
		user.District = addressDistrict
	}
	if latitude != "" {
		user.Lat = _lat
	}
	if longitude != "" {
		user.Lang = _lang
	}

	// Kalau ada foto, simpan
	if file != nil {
		filePath := "uploads/" + file.Filename
		if err := c.SaveUploadedFile(file, filePath); err != nil {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to upload photo", nil)
			return
		}
		user.PhotoUrl = filePath
		fmt.Println("Photo saved to:", filePath)
	}

	// Simpan ke DB
	if err := config.DB.Save(&user).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to update user", nil)
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, "Success Update", user)
}

// @Summary Delete user
// @Description Delete a user by their ID
// @Tags Users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {string} string "User deleted successfully"
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /users/{id} [delete]
func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User

	if err := config.DB.First(&user, id).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusNotFound, "User not found", nil)
		return
	}

	if err := config.DB.Delete(&user).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to delete user", nil)
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, "User deleted successfully", nil)
}
