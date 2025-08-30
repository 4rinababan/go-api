package controllers

import (
	"net/http"

	"github.com/ary/go-api/config"
	"github.com/ary/go-api/models"
	"github.com/ary/go-api/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// @Summary Create a new user account
// @Description Create a new user account with name, password, and user ID
// @Tags User Accounts
// @Accept json
// @Produce json
// @Param user_account body models.UserAccount true "User Account Input"
// @Success 201 {object} models.UserAccount
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /user-accounts [post]
func CreateUserAccount(c *gin.Context) {
	var input struct {
		Phone    string    `json:"phone" binding:"required"`
		Password string    `json:"password" binding:"required"`
		Role     string    `json:"role" binding:"required"`
		UserID   uuid.UUID `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	// Hash password
	passwordHash := utils.HashPassword(input.Password)

	// Cek apakah user_id sudah punya account
	var existing models.UserAccount
	if err := config.DB.Where("user_id = ?", input.UserID).First(&existing).Error; err == nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Account already exists", nil)
		return
	}

	userAccount := models.UserAccount{
		Phone:        input.Phone,
		PasswordHash: passwordHash,
		Role:         input.Role,
		UserID:       input.UserID,
	}

	if err := config.DB.Create(&userAccount).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to create account", nil)
		return
	}

	utils.SendSuccessResponse(c, http.StatusCreated, "Account created successfully", userAccount)
}

// @Summary Update user account password
// @Description Update the password for a user account
// @Tags User Accounts
// @Accept json
// @Produce json
// @Param input body object{user_account_id=string,old_password=string,new_password=string} true "Password Update Input"
// @Success 200 {string} string "Password berhasil diperbarui"
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /user-accounts/password [patch]
func UpdatePasswordUserAccount(c *gin.Context) {
	var input struct {
		UserAccountID uuid.UUID `json:"user_account_id" binding:"required"`
		OldPassword   string    `json:"old_password"`
		NewPassword   string    `json:"new_password" binding:"required"`
		IsActive      bool      `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Permintaan tidak valid", nil)
		return
	}

	var account models.UserAccount
	if err := config.DB.Preload("User").Where("user_id = ?", input.UserAccountID).First(&account).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusNotFound, "Akun tidak ditemukan", nil)
		return
	}

	// Jika IsActive true → wajib validasi old password
	if input.IsActive {
		if input.OldPassword == "" {
			utils.SendErrorResponse(c, http.StatusBadRequest, "Password lama diperlukan", nil)
			return
		}
		if !utils.CheckPasswordHash(account.PasswordHash, input.OldPassword) {
			utils.SendErrorResponse(c, http.StatusUnauthorized, "Password lama salah", nil)
			return
		}
	}

	newHashed := utils.HashPassword(input.NewPassword)
	tx := config.DB.Begin()

	// ✅ Hanya update password_hash
	if err := tx.Model(&models.UserAccount{}).
		Where("user_id = ?", input.UserAccountID).
		Update("password_hash", newHashed).Error; err != nil {
		tx.Rollback()
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Gagal update password", nil)
		return
	}

	// Jika user belum aktif → aktifkan
	if !account.User.IsActive {
		if err := tx.Model(&models.User{}).
			Where("id = ?", account.User.ID).
			Update("is_active", true).Error; err != nil {
			tx.Rollback()
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Gagal mengaktifkan user", nil)
			return
		}
	}

	tx.Commit()
	utils.SendSuccessResponse(c, http.StatusOK, "Password berhasil diperbarui", nil)
}
