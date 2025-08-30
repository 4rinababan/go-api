package controllers

import (
	"net/http"
	"path/filepath"
	"time"

	"github.com/ary/go-api/config"
	"github.com/ary/go-api/models"
	"github.com/ary/go-api/utils"
	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/gin-gonic/gin"
)

// @Summary Get all products
// @Description Retrieve a list of all products
// @Tags Products
// @Produce json
// @Success 200 {array} models.Product
// @Failure 500 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Router /products [get]
func GetProducts(c *gin.Context) {
	var products []models.Product
	if err := config.DB.Preload("Category").Find(&products).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to fetch products", nil)
		return
	}
	if len(products) == 0 {
		utils.SendErrorResponse(c, http.StatusNotFound, "No product found", nil)
		return
	}
	utils.SendSuccessResponse(c, http.StatusOK, "Success", products)
}

// @Summary Get product by ID
// @Description Retrieve a product by its ID
// @Tags Products
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} models.Product
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /products/{id} [get]
func GetProductByID(c *gin.Context) {
	id := c.Param("id")
	var product models.Product

	if err := config.DB.Where("id = ?", id).First(&product).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusNotFound, "product not found", nil)
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, "Success", product)
}

// @Summary Get products by category ID
// @Description Retrieve products that belong to a specific category
// @Tags Products
// @Produce json
// @Param category_id path string true "Category ID"
// @Success 200 {array} models.Product
// @Failure 404 {object} utils.ErrorResponse
// @Router /categories/{category_id}/products [get]
func GetProductByCategoryID(c *gin.Context) {
	categoryID := c.Param("category_id")
	var products []models.Product

	if err := config.DB.Where("category_id = ?", categoryID).Find(&products).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusNotFound, "Products not found", nil)
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, "Success", products)
}

// @Summary      Create a new product
// @Description  Create a new product with name, detail, category_id, and multiple images
// @Tags         Products
// @Accept       multipart/form-data
// @Produce      json
// @Param        name         formData string  true  "Product Name"
// @Param        detail       formData string  true  "Product Detail (as integer)"
// @Param        category_id  formData string  true  "Category ID (UUID)"
// @Param        images       formData file    true  "Product Images (multiple upload allowed)"
// @Success      201 {object} models.Product
// @Failure      400 {object} utils.ErrorResponse
// @Failure      500 {object} utils.ErrorResponse
// @Router       /products [post]
func CreateProduct(c *gin.Context) {
	var product models.Product

	name := c.PostForm("name")
	detail := c.PostForm("detail")
	categoryID := c.PostForm("category_id")

	product.Name = name
	product.Detail = detail

	parsedCategoryID, err := uuid.Parse(categoryID)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid category ID", nil)
		return
	}
	product.CategoryID = parsedCategoryID

	// Ambil semua file dengan nama "images"
	form, err := c.MultipartForm()
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Failed to read multipart form", nil)
		return
	}

	files := form.File["images"] // harus pakai array: name=\"images\"
	if len(files) == 0 {
		utils.SendErrorResponse(c, http.StatusBadRequest, "No image files provided", nil)
		return
	}

	var imagePaths []string
	uploadDir := "uploads"
	if _, err := utils.EnsureDir(uploadDir); err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to create upload directory", nil)
		return
	}

	for _, file := range files {
		ext := filepath.Ext(file.Filename)
		filename := uuid.New().String() + ext
		filepath := filepath.Join(uploadDir, filename)

		if err := c.SaveUploadedFile(file, filepath); err != nil {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to upload image", nil)
			return
		}

		imagePaths = append(imagePaths, filepath)
	}

	product.Images = pq.StringArray(imagePaths)

	if err := config.DB.Create(&product).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to create product", nil)
		return
	}

	utils.SendSuccessResponse(c, http.StatusCreated, "Success", product)
}

// @Summary Get best selling products
// @Description Get top 10 best-selling products for the current month
// @Tags Products
// @Produce json
// @Success 200 {array} object
// @Failure 500 {object} utils.ErrorResponse
// @Router /products/best-selling [get]
func GetBestSellingProducts(c *gin.Context) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second) // akhir bulan 23:59:59

	type Result struct {
		ProductID uuid.UUID `json:"product_id"`
		Name      string    `json:"name"`
		Detail    string    `json:"detail"`
		TotalSold int       `json:"total_sold"`
	}

	var results []Result

	err := config.DB.Table("orders").
		Select("products.id as product_id, products.name, products.detail, SUM(orders.quantity) as total_sold").
		Joins("JOIN products ON products.id = orders.product_id").
		Where("orders.created_at BETWEEN ? AND ?", startOfMonth, endOfMonth).
		Group("products.id, products.name, products.detail").
		Order("total_sold DESC").
		Limit(5).
		Scan(&results).Error

	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get best selling products", nil)
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, "Success", results)
}

// @Summary Update a product
// @Description Update product details and optionally replace images
// @Tags Products
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Product ID"
// @Param name formData string false "Product Name"
// @Param detail formData string false "Product Detail (as integer)"
// @Param category_id formData string false "Category ID (UUID)"
// @Param images formData file false "Product Images (multiple upload allowed)"
// @Success 200 {object} models.Product
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /products/{id} [patch]
func UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	var product models.Product

	// Cek apakah produk ada
	if err := config.DB.First(&product, "id = ?", id).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusNotFound, "Product not found", nil)
		return
	}

	name := c.PostForm("name")
	detail := c.PostForm("detail")
	categoryID := c.PostForm("category_id")

	if name != "" {
		product.Name = name
	}
	if detail != "" {
		product.Detail = detail
	}
	if categoryID != "" {
		parsedCategoryID, err := uuid.Parse(categoryID)
		if err != nil {
			utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid category ID", nil)
			return
		}
		product.CategoryID = parsedCategoryID
	}

	// Ganti gambar jika diberikan
	form, err := c.MultipartForm()
	if err == nil && form.File["images"] != nil && len(form.File["images"]) > 0 {
		// Hapus gambar lama
		for _, old := range product.Images {
			utils.DeleteFile(old)
		}

		files := form.File["images"]
		var newImagePaths []string
		uploadDir := "uploads"
		if _, err := utils.EnsureDir(uploadDir); err != nil {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to create upload directory", nil)
			return
		}

		for _, file := range files {
			ext := filepath.Ext(file.Filename)
			filename := uuid.New().String() + ext
			path := filepath.Join(uploadDir, filename)

			if err := c.SaveUploadedFile(file, path); err != nil {
				utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to upload image", nil)
				return
			}
			newImagePaths = append(newImagePaths, path)
		}
		product.Images = pq.StringArray(newImagePaths)
	}

	if err := config.DB.Save(&product).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to update product", nil)
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, "Product updated", product)
}

// @Summary Delete a product
// @Description Delete a product by ID and its images
// @Tags Products
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} utils.SuccessResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /products/{id} [delete]
func DeleteProduct(c *gin.Context) {
	id := c.Param("id")
	var product models.Product

	if err := config.DB.First(&product, "id = ?", id).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusNotFound, "Product not found", nil)
		return
	}

	// Hapus file gambar
	for _, img := range product.Images {
		utils.DeleteFile(img)
	}

	if err := config.DB.Delete(&product).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to delete product", nil)
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, "Product deleted", nil)
}
