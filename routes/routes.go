package routes

import (
	"github.com/ary/go-api/controllers"
	_ "github.com/ary/go-api/docs"
	"github.com/ary/go-api/middlewares"
	"github.com/ary/go-api/sse"

	// "github.com/ary/go-api/ws"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		// Auth routes
		api.POST("/login", controllers.Login)

		// Public User routes (misalnya register user baru)
		api.POST("/users", controllers.CreateUser)
		api.POST("/user-accounts", controllers.CreateUserAccount)

		// Public Product routes
		api.GET("/products", controllers.GetProducts)
		api.GET("/products/:id", controllers.GetProductByID)
		api.GET("/products/best-selling", controllers.GetBestSellingProducts)
		api.GET("/categories", controllers.GetCategories)
		api.GET("/categories/:category_id/products", controllers.GetProductByCategoryID)

		// Public Orders (opsional kalau boleh pesan tanpa login)
		api.POST("/orders", controllers.CreateOrderAndNotify)

		api.GET("/info", controllers.GetInfo)

		// Swagger
		api.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

		// Protected routes (harus pakai token JWT)
		protected := api.Group("/")
		protected.Use(middlewares.AuthMiddleware())
		{
			// User
			protected.GET("/users", controllers.GetUsers)
			protected.GET("/users/:id", controllers.GetUserByID)
			protected.GET("/users/:id/is-active", controllers.CheckUserIsActive)
			protected.PATCH("/users/:id", controllers.UpdateUser)
			protected.DELETE("/users/:id", controllers.DeleteUser)

			//user-account
			protected.PATCH("/user-accounts/update-password", controllers.UpdatePasswordUserAccount)

			// Categories (CRUD penuh)
			protected.POST("/categories", controllers.CreateCategory)
			protected.DELETE("/categories/:id", controllers.DeleteCategory)
			protected.PATCH("/categories/:id", controllers.UpdateCategory)

			// Products (CRUD penuh)
			protected.POST("/products", controllers.CreateProduct)
			protected.DELETE("/products/:id", controllers.DeleteProduct)
			protected.PATCH("/products/:id", controllers.UpdateProduct)

			// Orders
			protected.GET("/orders", controllers.GetAllOrders)
			protected.GET("/orders/dashboard", controllers.GetDashboard)
			protected.GET("/orders/user/:userid", controllers.GetOrdersByUserID)
			protected.GET("/orders/:orderid", controllers.GetOrderByID)
			protected.GET("/orders/history/:userid", controllers.GetOrderHistoryByUserID)
			protected.PATCH("/orders/:id/status", controllers.UpdateOrderStatusAndNotify)

			//notification
			protected.GET("/notification", controllers.GetNotification)
			protected.PATCH("/notification/:id/read", controllers.MarkNotificationAsRead)
			protected.GET("/notification/admin", controllers.GetAdminNotifications)

			// Info

			protected.PUT("/info", controllers.UpsertInfo)

		}
	}

	// SSE
	r.GET("/events", sse.GinHandler)
}
