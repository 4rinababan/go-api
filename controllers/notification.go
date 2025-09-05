package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ary/go-api/config"
	"github.com/ary/go-api/models"
	"github.com/ary/go-api/sse"
	"github.com/ary/go-api/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrderInput struct {
	User struct {
		ID       string  `json:"id" example:"4cd64505-f56f-421b-9243-e6bc9cbbaa7b"`
		Name     string  `json:"name" example:"Budi"`
		Phone    string  `json:"phone" example:"08123456789"`
		Email    string  `json:"email" example:"budi@email.com"`
		Address  string  `json:"address" example:"Jl. Mawar No. 123"`
		Regency  string  `json:"regency" example:"Cengkareng"`
		District string  `json:"district" example:"Duri Kosambi"`
		Lang     float64 `json:"lang" example:""`
		Lat      float64 `json:"lat" example:""`
	} `json:"user"`
	ProductID   string `json:"product_id" example:"4cd64505-f56f-421b-9243-e6bc9cbbaa7b"`
	CompanyName string `json:"company_name" example:"PT. ABC"`
	Priority    string `json:"priority" example:"normal"`
	Details     string `json:"details" example:"Pesanan baru dari Budi"`
	Address     string `json:"address" example:"Jl. Mawar No. 123"`
	Quantity    int    `json:"quantity" example:"2"`
}

// CreateOrderAndNotify godoc
// @Summary      Create a new order and notify via WebSocket
// @Description  Create a new order, save user if not exists, and send notification via WebSocket
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Param        order body OrderInput true "Order Input"
// @Success      201 {object} utils.SuccessResponse
// @Failure      400 {object} utils.ErrorResponse
// @Failure      500 {object} utils.ErrorResponse
// @Router       /orders [post]
func CreateOrderAndNotify(c *gin.Context) {
	var input OrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid input data", nil)
		return
	}

	var user models.User
	var err error

	// ✅ 1. Cek berdasarkan UserID jika ada
	if input.User.ID != "" {
		userID := utils.ParseUUID(input.User.ID)
		if userID != uuid.Nil {
			err = config.DB.First(&user, "id = ?", userID).Error
		}
	}

	// ✅ 2. Kalau tidak ketemu, cek berdasarkan Phone
	if (err != nil || user.ID == uuid.Nil) && input.User.Phone != "" {
		err = config.DB.Where("phone = ?", input.User.Phone).First(&user).Error
	}

	// ✅ 3. Kalau tetap tidak ketemu → buat user baru
	if err != nil || user.ID == uuid.Nil {
		user = models.User{
			Name:     input.User.Name,
			Email:    input.User.Email,
			Phone:    input.User.Phone,
			Address:  input.User.Address,
			Regency:  input.User.Regency,
			District: input.User.District,
			Lang:     input.User.Lang,
			Lat:      input.User.Lat,
			IsActive: false,
		}
		if err := config.DB.Create(&user).Error; err != nil {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Gagal membuat user", nil)
			return
		}
	} else {
		// ✅ Kalau user ditemukan, log info
		fmt.Println("User ditemukan, pakai data existing:", user.ID)
	}

	// ✅ Validasi product_id
	productUUID := utils.ParseUUID(input.ProductID)
	if productUUID == uuid.Nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid product ID", nil)
		return
	}

	// ✅ Buat order
	order := models.Order{
		UserID:      user.ID, // Ambil dari user existing atau baru
		ProductID:   productUUID,
		CompanyName: input.CompanyName,
		Priority:    input.Priority,
		Details:     input.Details,
		Address:     input.Address,
		Quantity:    input.Quantity,
		Status:      "Menunggu konfirmasi",
	}
	if err := config.DB.Create(&order).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Gagal menyimpan order", nil)
		return
	}

	// ✅ Simpan status awal ke OrderStatusUpdate
	statusUpdate := models.OrderStatusUpdate{
		OrderID: order.ID,
		Status:  order.Status,
	}
	_ = config.DB.Create(&statusUpdate)

	// ✅ Buat notifikasi dan simpan ke DB
	notif := models.Notification{
		ID:      uuid.New(),
		UserID:  user.ID,
		Message: "📦 Pesanan Dibuat",
		OrderID: order.ID,
		Read:    false,
	}
	_ = config.DB.Create(&notif)

	// ✅ Broadcast ke admin via SSE / WebSocket
	notifJson, _ := json.Marshal(notif)
	sse.BroadcastToRole("admin", string(notifJson))

	utils.SendSuccessResponse(c, http.StatusCreated, "Success", map[string]interface{}{
		"order": order,
		"user":  user,
	})
}

// UpdateOrderStatusAndNotify godoc
// @Summary      Update order status and notify user
// @Description  Update status order dan kirim notifikasi ke user terkait
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Param        id path string true "Order ID"
// @Param        status body struct{ Status string `json:"status"` } true "Status baru"
// @Success      200 {object} utils.SuccessResponse
// @Failure      400 {object} utils.ErrorResponse
// @Failure      404 {object} utils.ErrorResponse
// @Failure      500 {object} utils.ErrorResponse
// @Router       /orders/{id}/status [put]
func UpdateOrderStatusAndNotify(c *gin.Context) {
	// Ambil ID order dari URL
	orderIDStr := c.Param("id")
	orderID := utils.ParseUUID(orderIDStr)
	if orderID == uuid.Nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid order ID", nil)
		return
	}

	// Ambil status dari body
	var input struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	// Mulai transaction
	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Cari order
	var order models.Order
	if err := tx.Preload("User").First(&order, "id = ?", orderID).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.SendErrorResponse(c, http.StatusNotFound, "Order not found", nil)
			return
		}
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Database error", nil)
		return
	}

	// Update hanya kolom status
	// Hanya update kolom status, tidak memengaruhi user_id
	if err := tx.Model(&models.Order{}).Where("id = ?", order.ID).
		UpdateColumn("status", input.Status).Error; err != nil {
		tx.Rollback()
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to update order status", nil)
		return
	}

	// Simpan riwayat status
	statusUpdate := models.OrderStatusUpdate{
		OrderID: order.ID,
		Status:  input.Status,
	}
	if err := tx.Create(&statusUpdate).Error; err != nil {
		tx.Rollback()
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to save order status history", nil)
		return
	}

	// Buat notifikasi
	notif := models.Notification{
		ID:      uuid.New(),
		UserID:  order.UserID,
		Message: fmt.Sprintf("📢  %s", input.Status),
		OrderID: order.ID,
		Read:    false,
	}
	if err := tx.Create(&notif).Error; err != nil {
		tx.Rollback()
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to create notification", nil)
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to commit transaction", nil)
		return
	}

	// Kirim notifikasi via SSE
	notifJson, _ := json.Marshal(notif)
	sse.BroadcastToUser(order.User.ID.String(), string(notifJson))
	log.Println("order ID:", order.User.ID.String())

	utils.SendSuccessResponse(c, http.StatusOK, "Status order berhasil diperbarui", gin.H{
		"order_id": order.ID,
		"status":   input.Status,
	})
}

func GetNotification(c *gin.Context) {
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "user_id is required", nil)
		return
	}

	userID := utils.ParseUUID(userIDStr)
	if userID == uuid.Nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid user_id", nil)
		return
	}

	var notifications []models.Notification

	// Ambil notifikasi + relasi Order & Product
	if err := config.DB.
		Preload("Order.Product").
		Preload("Order.User").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&notifications).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get notifications", nil)
		return
	}

	// Transform ke response agar tidak kirim semua field (misal hide DeletedAt)
	type NotificationResponse struct {
		ID        uuid.UUID `json:"id"`
		Message   string    `json:"message"`
		Read      bool      `json:"read"`
		CreatedAt time.Time `json:"created_at"`
		Order     struct {
			ID        uuid.UUID `json:"id"`
			OrderCode string    `json:"order_code"`
			Status    string    `json:"status"`
			Quantity  int       `json:"quantity"`
			Company   string    `json:"company_name"`
			Product   string    `json:"product_name"`
			UserName  string    `json:"user_name"`
			UserPhone string    `json:"user_phone"`
			Detail    string    `json:"detail"`
			Address   string    `json:"address"`
		} `json:"order"`
	}

	var response []NotificationResponse
	for _, n := range notifications {
		item := NotificationResponse{
			ID:        n.ID,
			Message:   n.Message,
			Read:      n.Read,
			CreatedAt: n.CreatedAt,
		}
		item.Order.ID = n.Order.ID
		item.Order.OrderCode = n.Order.OrderCode
		item.Order.Status = n.Order.Status
		item.Order.Quantity = n.Order.Quantity
		item.Order.Company = n.Order.CompanyName
		item.Order.Product = n.Order.Product.Name // pastikan model Product punya field Name
		item.Order.UserName = n.Order.User.Name
		item.Order.UserPhone = n.Order.User.Phone
		item.Order.Detail = n.Order.Details
		item.Order.Address = n.Order.Address

		response = append(response, item)
	}

	utils.SendSuccessResponse(c, http.StatusOK, "Success", response)
}

func GetAdminNotifications(c *gin.Context) {
	var notifications []models.Notification

	// Query: hanya Read = true, message mengandung "📦 Pesanan Dibuat", urut created_at desc, limit 20
	if err := config.DB.
		Preload("Order.Product").
		Preload("Order.User").
		Where("message LIKE ?", "%📦 Pesanan Dibuat%").
		Order("CASE WHEN read = false THEN 0 ELSE 1 END, created_at DESC").
		Limit(20).
		Find(&notifications).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get notifications", nil)
		return
	}

	// Transform ke response
	type NotificationResponse struct {
		ID        uuid.UUID `json:"id"`
		Message   string    `json:"message"`
		Read      bool      `json:"read"`
		CreatedAt time.Time `json:"created_at"`
		Order     struct {
			ID        uuid.UUID `json:"id"`
			OrderCode string    `json:"order_code"`
			Status    string    `json:"status"`
			Quantity  int       `json:"quantity"`
			Company   string    `json:"company_name"`
			Product   string    `json:"product_name"`
			UserName  string    `json:"user_name"`
			UserPhone string    `json:"user_phone"`
			Detail    string    `json:"detail"`
			Address   string    `json:"address"`
		} `json:"order"`
	}

	var response []NotificationResponse
	for _, n := range notifications {
		item := NotificationResponse{
			ID:        n.ID,
			Message:   n.Message,
			Read:      n.Read,
			CreatedAt: n.CreatedAt,
		}
		item.Order.ID = n.Order.ID
		item.Order.OrderCode = n.Order.OrderCode
		item.Order.Status = n.Order.Status
		item.Order.Quantity = n.Order.Quantity
		item.Order.Company = n.Order.CompanyName
		item.Order.Product = n.Order.Product.Name
		item.Order.UserName = n.Order.User.Name
		item.Order.UserPhone = n.Order.User.Phone
		item.Order.Detail = n.Order.Details
		item.Order.Address = n.Order.Address

		response = append(response, item)
	}

	utils.SendSuccessResponse(c, http.StatusOK, "Success", response)
}

func MarkNotificationAsRead(c *gin.Context) {
	// Ambil notification_id dari param URL
	notifIDStr := c.Param("id")
	if notifIDStr == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "notification id is required", nil)
		return
	}

	notifID := utils.ParseUUID(notifIDStr)
	if notifID == uuid.Nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid notification id", nil)
		return
	}

	// Cari notifikasi di DB
	var notif models.Notification
	if err := config.DB.First(&notif, "id = ?", notifID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.SendErrorResponse(c, http.StatusNotFound, "Notification not found", nil)
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to find notification", nil)
		}
		return
	}

	// Update kolom read
	notif.Read = true
	if err := config.DB.Save(&notif).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to update notification", nil)
		return
	}

	// Response sukses
	utils.SendSuccessResponse(c, http.StatusOK, "Notification marked as read", gin.H{
		"id":      notif.ID,
		"read":    notif.Read,
		"message": notif.Message,
	})
}
