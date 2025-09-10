package database

import (
	"fmt"
	"log"

	"github.com/mm-api/mm-api/config"
	"github.com/mm-api/mm-api/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Connect подключается к базе данных и выполняет миграции
func Connect() error {
	log.Println("🔧 Connect function started")

	var err error
	cfg := config.GetConfig()

	log.Println("🔗 Connecting to PostgreSQL database...")

	// Используем DATABASE_URL из конфигурации
	dsn := cfg.DatabaseURL
	log.Printf("📊 Database URL configured: %s", maskDatabaseURL(dsn))

	// Настройка логирования GORM
	var gormLogger logger.Interface
	if cfg.IsDevelopment() {
		gormLogger = logger.Default.LogMode(logger.Info)
		log.Println("📝 GORM logging enabled (development mode)")
	} else {
		gormLogger = logger.Default.LogMode(logger.Silent)
		log.Println("📝 GORM logging disabled (production mode)")
	}

	// Подключение к базе данных
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})

	if err != nil {
		log.Printf("❌ Failed to connect to database: %v", err)
		return fmt.Errorf("failed to connect to PostgreSQL database: %w", err)
	}

	log.Println("✅ Connected to PostgreSQL database successfully")

	// Получение базового подключения
	sqlDB, err := DB.DB()
	if err != nil {
		log.Printf("❌ Failed to get database instance: %v", err)
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	// Настройка пула соединений
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	log.Println("✅ Database connection pool configured")

	// Проверка подключения
	if err := sqlDB.Ping(); err != nil {
		log.Printf("❌ Failed to ping database: %v", err)
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("✅ Database ping successful")

	// Выполнение миграций
	log.Println("🔄 Running database migrations...")
	if err := runMigrations(); err != nil {
		log.Printf("❌ Failed to run migrations: %v", err)
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("✅ Database migrations completed")

	// Проверка и создание ролей
	log.Println("🔄 Checking and creating default roles...")
	if err := createDefaultRoles(); err != nil {
		log.Printf("❌ Failed to create default roles: %v", err)
		return fmt.Errorf("failed to create default roles: %w", err)
	}

	log.Println("✅ Default roles checked/created")

	// Создание администратора по умолчанию
	log.Println("🔄 Checking and creating default admin...")
	if err := createDefaultAdmin(); err != nil {
		log.Printf("❌ Failed to create default admin: %v", err)
		return fmt.Errorf("failed to create default admin: %w", err)
	}

	log.Println("✅ Default admin checked/created")

	// Создание владельца магазина по умолчанию
	log.Println("🔄 Checking and creating default shop owner...")
	if err := createDefaultShopOwner(); err != nil {
		log.Printf("❌ Failed to create default shop owner: %v", err)
		return fmt.Errorf("failed to create default shop owner: %w", err)
	}

	log.Println("✅ Default shop owner checked/created")

	// Создание тестовых данных только в режиме разработки
	if cfg.IsDevelopment() {
		log.Println("🔄 Creating sample data (development mode)...")
		if err := createSampleData(); err != nil {
			log.Printf("⚠️ Warning: Failed to create sample data: %v", err)
			// Не прерываем работу при ошибке создания тестовых данных
		} else {
			log.Println("✅ Sample data created")
		}
	}

	log.Println("🎉 Database initialization completed successfully")
	return nil
}

// maskDatabaseURL маскирует пароль в URL базы данных для логирования
func maskDatabaseURL(url string) string {
	// Простая маскировка для безопасности
	if len(url) > 20 {
		return url[:20] + "***"
	}
	return "***"
}

// runMigrations выполняет миграции базы данных
func runMigrations() error {
	return DB.AutoMigrate(
		&models.Role{},
		&models.User{},
		&models.Category{},
		&models.Product{},
		&models.ProductVariation{},
		&models.CartItem{},
		&models.Favorite{},
		&models.Address{},
		&models.Order{},
		&models.OrderItem{},
		&models.Notification{},
		&models.UserSettings{},
	)
}

// createDefaultRoles создает роли по умолчанию, если они не существуют
func createDefaultRoles() error {
	// Создаем роль супер админа
	var superAdminRole models.Role
	if err := DB.Where("name = ?", "super_admin").First(&superAdminRole).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			superAdmin := models.Role{
				Name:        "super_admin",
				DisplayName: "Супер Администратор",
				Description: "Максимальные права доступа, включая создание категорий",
				Permissions: `{"dashboard": true, "users": true, "products": true, "categories": true, "create_categories": true, "orders": true, "settings": true, "roles": true}`,
				IsActive:    true,
				IsSystem:    true,
			}
			if err := DB.Create(&superAdmin).Error; err != nil {
				return fmt.Errorf("failed to create super admin role: %w", err)
			}
			log.Println("✅ Super admin role created")
		} else {
			return fmt.Errorf("failed to check super admin role: %w", err)
		}
	} else {
		log.Printf("✅ Super admin role already exists: %s", superAdminRole.Name)
	}

	var adminRole models.Role
	if err := DB.Where("name = ?", "admin").First(&adminRole).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			admin := models.Role{
				Name:        "admin",
				DisplayName: "Администратор",
				Description: "Полный доступ ко всем функциям системы",
				Permissions: `{"dashboard": true, "users": true, "products": true, "categories": true, "orders": true, "settings": true}`,
				IsActive:    true,
				IsSystem:    true,
			}
			if err := DB.Create(&admin).Error; err != nil {
				return fmt.Errorf("failed to create admin role: %w", err)
			}
			log.Println("✅ Admin role created")
		} else {
			return fmt.Errorf("failed to check admin role: %w", err)
		}
	} else {
		log.Printf("✅ Admin role already exists: %s", adminRole.Name)
	}

	var shopOwnerRole models.Role
	if err := DB.Where("name = ?", "shop_owner").First(&shopOwnerRole).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			shopOwner := models.Role{
				Name:        "shop_owner",
				DisplayName: "Владелец магазина",
				Description: "Управление товарами, категориями и просмотр заказов клиентов",
				Permissions: `{"dashboard": true, "products": true, "categories": true, "orders": true, "settings": true}`,
				IsActive:    true,
				IsSystem:    true,
			}
			if err := DB.Create(&shopOwner).Error; err != nil {
				return fmt.Errorf("failed to create shop owner role: %w", err)
			}
			log.Println("✅ Shop owner role created")
		} else {
			return fmt.Errorf("failed to check shop owner role: %w", err)
		}
	} else {
		log.Printf("✅ Shop owner role already exists: %s", shopOwnerRole.Name)
	}

	var userRole models.Role
	if err := DB.Where("name = ?", "user").First(&userRole).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			user := models.Role{
				Name:        "user",
				DisplayName: "Пользователь",
				Description: "Обычный пользователь с доступом к покупкам",
				Permissions: `{"profile": true, "orders": true, "favorites": true}`,
				IsActive:    true,
				IsSystem:    true,
			}
			if err := DB.Create(&user).Error; err != nil {
				return fmt.Errorf("failed to create user role: %w", err)
			}
			log.Println("✅ User role created")
		} else {
			return fmt.Errorf("failed to check user role: %w", err)
		}
	} else {
		log.Printf("✅ User role already exists: %s", userRole.Name)
	}

	return nil
}

// createDefaultAdmin создает администратора по умолчанию, если он не существует
func createDefaultAdmin() error {
	var adminUser models.User
	if err := DB.Where("email = ?", "admin@mm.com").First(&adminUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Получаем роль админа
			var adminRole models.Role
			if err := DB.Where("name = ?", "admin").First(&adminRole).Error; err != nil {
				return fmt.Errorf("failed to find admin role: %w", err)
			}

			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
			admin := models.User{
				Email:           "admin@mm.com",
				Password:        string(hashedPassword),
				Name:            "Администратор",
				RoleID:          &adminRole.ID,
				IsActive:        true,
				IsEmailVerified: true,
			}
			if err := DB.Create(&admin).Error; err != nil {
				return fmt.Errorf("failed to create default admin user: %w", err)
			}
			log.Println("✅ Default admin user created")
		} else {
			return fmt.Errorf("failed to check default admin user: %w", err)
		}
	} else {
		log.Printf("✅ Default admin user already exists: %s", adminUser.Email)
	}

	return nil
}

// createSampleData создает начальные тестовые данные, если они не существуют
func createSampleData() error {
	// Проверяем, есть ли уже продукты
	var count int64
	DB.Model(&models.Product{}).Count(&count)

	if count > 0 {
		log.Println("✅ Sample data already seeded")
		return nil // Данные уже есть
	}

	// Создаем тестовые категории
	categories := []models.Category{
		{
			Name:        "Мужская одежда",
			Description: "Одежда для мужчин",
			IconURL:     "https://example.com/icons/men.png",
			IsActive:    true,
			SortOrder:   1,
		},
		{
			Name:        "Женская одежда",
			Description: "Одежда для женщин",
			IconURL:     "https://example.com/icons/women.png",
			IsActive:    true,
			SortOrder:   2,
		},
	}

	for _, category := range categories {
		if err := DB.Create(&category).Error; err != nil {
			return err
		}
	}

	// Получаем созданные категории
	var menCategory, womenCategory models.Category
	DB.Where("name = ?", "Мужская одежда").First(&menCategory)
	DB.Where("name = ?", "Женская одежда").First(&womenCategory)

	// Создаем тестовые продукты
	products := []models.Product{
		{
			Name:        "Джинсы классические",
			Description: "Классические джинсы из 100% хлопка",
			Gender:      "unisex",
			CategoryID:  menCategory.ID,
			Brand:       "Levi's",
			IsAvailable: true,
		},
		{
			Name:        "Футболка базовая",
			Description: "Базовая футболка из хлопка",
			Gender:      "unisex",
			CategoryID:  womenCategory.ID,
			Brand:       "Nike",
			IsAvailable: true,
		},
		{
			Name:        "Кроссовки спортивные",
			Description: "Удобные кроссовки для спорта",
			Gender:      "unisex",
			CategoryID:  menCategory.ID,
			Brand:       "Adidas",
			IsAvailable: true,
		},
		{
			Name:        "Платье летнее",
			Description: "Легкое летнее платье",
			Gender:      "female",
			CategoryID:  womenCategory.ID,
			Brand:       "Zara",
			IsAvailable: true,
		},
		{
			Name:        "Рубашка офисная",
			Description: "Классическая офисная рубашка",
			Gender:      "male",
			CategoryID:  menCategory.ID,
			Brand:       "H&M",
			IsAvailable: true,
		},
	}

	// Создаем продукты
	for i := range products {
		if err := DB.Create(&products[i]).Error; err != nil {
			log.Printf("❌ Failed to create product %d: %v", i+1, err)
			continue
		}

		// Создаем вариации для каждого продукта
		variations := []models.ProductVariation{
			{
				ProductID:     products[i].ID,
				Sizes:         []string{"S", "M", "L"},
				Colors:        []string{"Черный", "Синий"},
				Price:         2999.0,
				ImageURLs:     []string{"/images/products/jeans1.jpg", "/images/products/jeans1_2.jpg"},
				StockQuantity: 10,
				IsAvailable:   true,
				SKU:           "LEVI-001-BLACK-BLUE",
			},
			{
				ProductID:     products[i].ID,
				Sizes:         []string{"M", "L", "XL"},
				Colors:        []string{"Белый", "Серый"},
				Price:         2999.0,
				ImageURLs:     []string{"/images/products/jeans2.jpg", "/images/products/jeans2_2.jpg"},
				StockQuantity: 15,
				IsAvailable:   true,
				SKU:           "LEVI-001-WHITE-GRAY",
			},
		}

		for _, variation := range variations {
			if err := DB.Create(&variation).Error; err != nil {
				log.Printf("❌ Failed to create variation for product %s: %v", products[i].Name, err)
			}
		}
	}

	return nil
}

// createDefaultShopOwner создает владельца магазина по умолчанию, если он не существует
func createDefaultShopOwner() error {
	var shopOwnerUser models.User
	if err := DB.Where("email = ?", "shopowner@mm.com").First(&shopOwnerUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Получаем роль владельца магазина
			var shopOwnerRole models.Role
			if err := DB.Where("name = ?", "shop_owner").First(&shopOwnerRole).Error; err != nil {
				return fmt.Errorf("failed to find shop owner role: %w", err)
			}

			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("shopowner123"), bcrypt.DefaultCost)
			shopOwner := models.User{
				Email:           "shopowner@mm.com",
				Password:        string(hashedPassword),
				Name:            "Владелец магазина",
				RoleID:          &shopOwnerRole.ID,
				IsActive:        true,
				IsEmailVerified: true,
			}
			if err := DB.Create(&shopOwner).Error; err != nil {
				return fmt.Errorf("failed to create default shop owner user: %w", err)
			}
			log.Println("✅ Default shop owner user created")
		} else {
			return fmt.Errorf("failed to check default shop owner user: %w", err)
		}
	} else {
		log.Printf("✅ Default shop owner user already exists: %s", shopOwnerUser.Email)
	}

	return nil
}
