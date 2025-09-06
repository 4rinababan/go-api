package controllers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ary/go-api/config"
	"github.com/ary/go-api/models"
	"github.com/ary/go-api/utils"
)

func GetDashboard(c *gin.Context) {
	var totalOrders, menungguKonfirmasi, diproses, selesai int64

	// Hitung total semua order
	config.DB.Model(&models.Order{}).Count(&totalOrders)
	config.DB.Model(&models.Order{}).Where("status = ?", "Menunggu konfirmasi").Count(&menungguKonfirmasi)
	config.DB.Model(&models.Order{}).Where("status = ?", "Diproses").Count(&diproses)
	config.DB.Model(&models.Order{}).Where("status = ?", "Selesai").Count(&selesai)

	// Hitung 3 bulan terakhir + breakdown status
	var chartData []struct {
		Month              string
		TotalOrders        int64
		MenungguKonfirmasi int64
		Diproses           int64
		Selesai            int64
	}

	err := config.DB.
		Model(&models.Order{}).
		Select(`
			TO_CHAR(created_at, 'Mon') AS month,
			COUNT(*) AS total_orders,
			COUNT(*) FILTER (WHERE status = 'Menunggu konfirmasi') AS menunggu_konfirmasi,
			COUNT(*) FILTER (WHERE status = 'Diproses') AS diproses,
			COUNT(*) FILTER (WHERE status = 'Selesai') AS selesai
		`).
		Where("created_at >= date_trunc('month', NOW()) - interval '2 months'").
		Group("month").
		Order("MIN(created_at) ASC").
		Scan(&chartData).Error

	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get chart data", nil)
		return
	}

	// Response
	response := gin.H{
		"total_orders":        totalOrders,
		"menunggu_konfirmasi": menungguKonfirmasi,
		"diproses":            diproses,
		"selesai":             selesai,
		"chart":               chartData,
	}

	utils.SendSuccessResponse(c, http.StatusOK, "Success", response)
}

// GetAllOrders mengambil semua order dengan preload user dan product
// @Summary      Get all orders
// @Description  Retrieve all orders with user and product details
// @Tags         Orders
// @Produce      json
// @Success      200 {array} models.Order
// @Failure      500 {object} utils.ErrorResponse
// @Router       /orders [get]
func GetAllOrders(c *gin.Context) {
	var orders []models.Order
	var total int64

	// Ambil parameter query, default page=1 dan limit=10
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "5")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Hitung total record
	if err := config.DB.Model(&models.Order{}).Count(&total).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to count orders", nil)
		return
	}

	// Ambil data dengan pagination dan preload relasi
	if err := config.DB.Preload("User").Preload("Product").
		Order(`CASE WHEN status = 'Menunggu konfirmasi' THEN 0 ELSE 1 END`).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&orders).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get orders", nil)
		return
	}

	// Buat response pagination
	response := map[string]interface{}{
		"data":       orders,
		"page":       page,
		"limit":      limit,
		"total":      total,
		"totalPages": (total + int64(limit) - 1) / int64(limit), // hitung total pages

	}

	utils.SendSuccessResponse(c, http.StatusOK, "Success", response)
}

// GetOrderByID mengambil satu order berdasarkan ID dengan preload user dan product
// @Summary      Get order by ID
// @Description  Retrieve an order by its ID with user and product details
// @Tags         Orders
// @Produce      json
// @Param        id   path      string true "Order ID"
// @Success      200  {object}  models.Order
// @Failure      404  {object}  utils.ErrorResponse
// @Failure      500  {object}  utils.ErrorResponse
// @Router       /orders/{id} [get]
func GetOrderByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid order ID", nil)
		return
	}

	var order models.Order
	err = config.DB.Preload("User").Preload("Product").First(&order, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		utils.SendErrorResponse(c, http.StatusNotFound, "Order not found", nil)
		return
	} else if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get order", nil)
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, "Success", order)
}

// GetOrdersByUserID mengambil semua order berdasarkan UserID dengan preload user dan product
// @Summary      Get orders by user ID
// @Description  Retrieve all orders for a specific user by user ID
// @Tags         Orders
// @Produce      json
// @Param        user_id path      string true "User ID"
// @Success      200     {array}   models.Order
// @Failure      404     {object}  utils.ErrorResponse
// @Failure      500     {object}  utils.ErrorResponse
// @Router       /orders/user/{userid} [get]
func GetOrdersByUserID(c *gin.Context) {
	userIDParam := c.Param("userid")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}

	var orders []models.Order
	err = config.DB.Preload("User").Preload("Product").Where("user_id = ?", userID).Find(&orders).Error
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get orders", nil)
		return
	}
	if len(orders) == 0 {
		utils.SendErrorResponse(c, http.StatusNotFound, "Orders not found for user", nil)
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, "Success", orders)
}

func GetOrderHistoryByUserID(c *gin.Context) {
	userIDParam := c.Param("userid")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}

	var orders []models.Order
	err = config.DB.
		Preload("Product").
		Preload("Updates", func(db *gorm.DB) *gorm.DB {
			return db.Order("timestamp ASC") // histori status urut waktu
		}).
		Where("user_id = ?", userID).
		Find(&orders).Error

	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get order history", nil)
		return
	}

	if len(orders) == 0 {
		utils.SendErrorResponse(c, http.StatusNotFound, "Orders not found for user", nil)
		return
	}

	// Bentuk respons sesuai format yang kamu mau
	var response []map[string]interface{}
	for _, o := range orders {
		orderData := map[string]interface{}{
			"orderId":   o.OrderCode,
			"product":   o.Product.Name,
			"createdAt": o.CreatedAt,
			"status":    o.Status,
			"updates":   o.Updates,
		}
		response = append(response, orderData)
	}

	utils.SendSuccessResponse(c, http.StatusOK, "Success", response)
}
