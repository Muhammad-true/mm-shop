package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/mm-api/mm-api/controllers"
	"github.com/mm-api/mm-api/middleware"
)

// SetupRoutes настраивает все маршруты API
func SetupRoutes() *gin.Engine {
	r := gin.Default()

	// Отключаем автоматическое перенаправление для API маршрутов
	r.RedirectTrailingSlash = false
	r.RedirectFixedPath = false
	// Middleware
	r.Use(middleware.CORS())
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Инициализация контроллеров
	authController := &controllers.AuthController{}
	productController := &controllers.ProductController{}
	cartController := &controllers.CartController{}
	categoryController := &controllers.CategoryController{}
	favoriteController := &controllers.FavoriteController{}
	addressController := &controllers.AddressController{}
	notificationController := &controllers.NotificationController{}
	settingsController := &controllers.SettingsController{}
	userController := &controllers.UserController{}
	orderController := &controllers.OrderController{}
	roleController := &controllers.RoleController{}
	uploadController := &controllers.UploadController{}
	imageController := &controllers.ImageController{}

	// API группа
	api := r.Group("/api/v1")

	// Публичные маршруты (без аутентификации)
	public := api.Group("/")
	{
		// Конфигурация (только для разработки)
		config := public.Group("config")
		{
			config.GET("/health", controllers.GetConfig)
		}

		// Аутентификация
		auth := public.Group("auth")
		{
			auth.POST("/register", authController.Register)
			auth.POST("/login", authController.Login)
			auth.POST("/refresh", authController.RefreshToken)
			// auth.POST("/forgot-password", authController.ForgotPassword) // TODO: Реализовать
			// auth.POST("/logout", authController.Logout) // TODO: Реализовать
		}

		// Продукты (требуют аутентификации для изоляции данных)
		products := public.Group("products")
		products.Use(middleware.AuthRequired())
		{
			products.GET("/", productController.GetProducts)
			products.GET("/:id", productController.GetProduct)
			products.GET("/featured", productController.GetProducts)                      // TODO: Добавить логику для рекомендуемых
			products.GET("/search", productController.GetProducts)                        // Используем тот же метод с параметром search
			products.GET("/with-variations", productController.GetProductsWithVariations) // Новый endpoint с JOIN запросом
		}

		// Категории (публичный доступ)
		categories := public.Group("categories")
		{
			categories.GET("/", categoryController.GetCategories)
			categories.GET("/:id", categoryController.GetCategory)
			categories.GET("/:id/products", categoryController.GetCategoryProducts)
		}
	}

	// Защищенные маршруты (требуют аутентификации)
	protected := api.Group("/")
	protected.Use(middleware.AuthRequired())
	{
		// Пользователи
		users := protected.Group("users")
		{
			users.GET("/profile", authController.Profile)
			users.PUT("/profile", authController.UpdateProfile)
			// users.POST("/avatar", authController.UploadAvatar) // TODO: Реализовать

			// Адреса пользователя
			addresses := users.Group("addresses")
			{
				addresses.GET("/", addressController.GetAddresses)
				addresses.POST("/", addressController.CreateAddress)
				addresses.PUT("/:id", addressController.UpdateAddress)
				addresses.DELETE("/:id", addressController.DeleteAddress)
				addresses.PUT("/:id/default", addressController.SetDefaultAddress)
			}
		}

		// Корзина
		cart := protected.Group("cart")
		{
			cart.GET("/", cartController.GetCart)
			cart.POST("/items", cartController.AddToCart)
			cart.PUT("/items/:id", cartController.UpdateCartItem)
			cart.DELETE("/items/:id", cartController.RemoveFromCart)
			cart.DELETE("/", cartController.ClearCart)
		}

		// Избранное
		favorites := protected.Group("favorites")
		{
			favorites.GET("/", favoriteController.GetFavorites)
			favorites.POST("/:productId", favoriteController.AddToFavorites)
			favorites.DELETE("/:productId", favoriteController.RemoveFromFavorites)
			favorites.GET("/sync", favoriteController.SyncFavorites)
			favorites.GET("/:productId/check", favoriteController.CheckFavorite)
		}

		// TODO: Заказы (не реализованы)
		// orders := protected.Group("orders")
		// {
		// 	orders.GET("/", orderController.GetOrders)
		// 	orders.POST("/", orderController.CreateOrder)
		// 	orders.GET("/:id", orderController.GetOrder)
		// 	orders.GET("/:id/tracking", orderController.GetOrderTracking)
		// }

		// Уведомления
		notifications := protected.Group("notifications")
		{
			notifications.GET("/", notificationController.GetNotifications)
			notifications.PUT("/:id/read", notificationController.MarkAsRead)
			notifications.PUT("/read-all", notificationController.MarkAllAsRead)
			notifications.DELETE("/:id", notificationController.DeleteNotification)
			notifications.GET("/unread-count", notificationController.GetUnreadCount)
		}

		// Настройки
		settings := protected.Group("settings")
		{
			settings.GET("/", settingsController.GetSettings)
			settings.PUT("/", settingsController.UpdateSettings)
			settings.POST("/reset", settingsController.ResetSettings)
		}
	}

	// Админские маршруты (для админов и супер админов)
	admin := api.Group("/admin")
	admin.Use(middleware.AuthRequired())
	admin.Use(middleware.AdminOrSuperAdminRequired())
	{
		// Управление пользователями (админы и супер админы)
		adminUsers := admin.Group("users")
		{
			adminUsers.GET("/", userController.GetUsers)
			adminUsers.POST("/", userController.CreateUser)
			adminUsers.GET("/:id", userController.GetUser)
			adminUsers.PUT("/:id", userController.UpdateUser)
			adminUsers.DELETE("/:id", userController.DeleteUser)
		}

		// Управление ролями (админы и супер админы)
		adminRoles := admin.Group("roles")
		{
			adminRoles.GET("/", roleController.GetRoles)
			adminRoles.GET("/:id", roleController.GetRole)
			adminRoles.POST("/", roleController.CreateRole)
			adminRoles.PUT("/:id", roleController.UpdateRole)
			adminRoles.DELETE("/:id", roleController.DeleteRole)
		}

		// Управление уведомлениями
		adminNotifications := admin.Group("notifications")
		{
			adminNotifications.POST("/", notificationController.CreateNotification)
		}

		// Управление категориями (админы и супер админы)
		adminCategories := admin.Group("categories")
		{
			// Создание категорий доступно только супер админам
			adminCategories.POST("/", middleware.SuperAdminRequired(), categoryController.CreateCategory)
			adminCategories.PUT("/:id", categoryController.UpdateCategory)
			adminCategories.DELETE("/:id", categoryController.DeleteCategory)
		}

		// Управление заказами (админы и супер админы)
		adminOrders := admin.Group("orders")
		{
			adminOrders.GET("/", orderController.GetOrders)
			adminOrders.GET("/:id", orderController.GetOrder)
			adminOrders.PUT("/:id/status", orderController.UpdateOrderStatus)
		}

		// Управление продуктами (админы и супер админы)
		adminProducts := admin.Group("products")
		{
			adminProducts.GET("/", productController.GetAllProducts)
			adminProducts.GET("/:id", productController.GetProductAdmin)
		}
	}

	// Маршруты для владельцев магазинов (админы + владельцы магазинов)
	shop := api.Group("/shop")
	shop.Use(middleware.AuthRequired())
	shop.Use(middleware.AdminOrShopOwnerRequired())
	{
		// Управление продуктами
		shopProducts := shop.Group("products")
		{
			shopProducts.GET("/", productController.GetShopProducts)     // Получение товаров владельца
			shopProducts.POST("/", productController.CreateProduct)      // Создание товара
			shopProducts.PUT("/:id", productController.UpdateProduct)    // Обновление товара
			shopProducts.DELETE("/:id", productController.DeleteProduct) // Удаление товара
		}

		// Управление категориями
		shopCategories := shop.Group("categories")
		{
			// Создание категорий УБРАНО из shop маршрутов - доступно только супер админам
			// shopCategories.POST("/", categoryController.CreateCategory) // УДАЛЕНО
			shopCategories.PUT("/:id", categoryController.UpdateCategory)
			shopCategories.DELETE("/:id", categoryController.DeleteCategory)
		}

		// Управление заказами (только заказы клиентов владельца)
		shopOrders := shop.Group("orders")
		{
			shopOrders.GET("/", orderController.GetShopOrders) // Только заказы клиентов
			shopOrders.GET("/:id", orderController.GetShopOrder)
			shopOrders.PUT("/:id/status", orderController.UpdateOrderStatus)
		}

		// Клиенты владельца магазина
		shopCustomers := shop.Group("customers")
		{
			shopCustomers.GET("/", userController.GetShopCustomers)             // Только клиенты
			shopCustomers.GET("/:id/orders", orderController.GetCustomerOrders) // Заказы клиента
		}
	}

	// Загрузка файлов
	upload := api.Group("upload")
	{
		upload.POST("/image", uploadController.UploadImage)
		upload.DELETE("/image/:filename", uploadController.DeleteImage)
	}

	// Работа с изображениями
	images := api.Group("images")
	{
		images.GET("/fix-urls", imageController.FixImageURLs)
		images.GET("/url/:filename", imageController.GetImageURL)
	}

	// Статические файлы для изображений
	r.Static("/images", "./images")

	// Обслуживание админ панели (если файлы присутствуют)
	r.Static("/admin", "./admin")

	// Информационные маршруты
	healthHandler := func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "MM API is running",
			"version": "1.1.0",
		})
	}
	r.GET("/health", healthHandler)
	r.HEAD("/health", healthHandler) // Добавляем поддержку HEAD запросов для Docker health check

	r.GET("/version", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"version": "1.1.0",
			"name":    "MM API",
			"build":   "development",
		})
	})

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to MM API",
			"version": "1.1.0",
			"docs":    "/api/v1/docs",
			"health":  "/health",
		})
	})

	return r
}
