package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
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

	// Обновляем данные в shop_subscriptions перед миграцией (если есть старые данные)
	log.Println("🔄 Preparing shop_subscriptions data before migration...")
	if err := prepareShopSubscriptionsForMigration(); err != nil {
		log.Printf("⚠️ Warning: Failed to prepare shop_subscriptions: %v", err)
		// Не прерываем работу, но логируем предупреждение
	}

	// Выполнение миграций
	log.Println("🔄 Running database migrations...")
	if err := runMigrations(); err != nil {
		log.Printf("❌ Failed to run migrations: %v", err)
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("✅ Database migrations completed")

	// Очистка лишних пробелов из device_id в лицензиях
	log.Println("🔄 Cleaning device_id whitespace in licenses...")
	if err := cleanDeviceIDWhitespace(); err != nil {
		log.Printf("⚠️ Warning: Failed to clean device_id whitespace: %v", err)
		// Не прерываем работу, но логируем предупреждение
	}

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

	// Создание городов по умолчанию
	log.Println("🔄 Checking and creating default cities...")
	if err := createDefaultCities(); err != nil {
		log.Printf("⚠️ Warning: Failed to create default cities: %v", err)
		// Не прерываем работу при ошибке
	} else {
		log.Println("✅ Default cities checked/created")
	}

	// Создание планов подписки по умолчанию
	log.Println("🔄 Checking and creating default subscription plans...")
	if err := createDefaultSubscriptionPlans(); err != nil {
		log.Printf("⚠️ Warning: Failed to create default subscription plans: %v", err)
		// Не прерываем работу при ошибке
	} else {
		log.Println("✅ Default subscription plans checked/created")
	}

	// Миграция данных: создание shops из существующих shop_owners
	log.Println("🔄 Migrating shop owners to shops table...")
	if err := migrateShopsFromUsers(); err != nil {
		log.Printf("⚠️ Warning: Failed to migrate shops from users: %v", err)
		// Не прерываем работу при ошибке миграции
	} else {
		log.Println("✅ Shops migration completed")
	}

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
	// Сначала выполняем GORM AutoMigrate для автоматического создания/обновления таблиц
	if err := DB.AutoMigrate(
		&models.Role{},
		&models.User{},
		&models.City{}, // Таблица городов
		&models.Shop{}, // Новая таблица магазинов
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
		&models.ShopSubscription{},
		&models.DeviceToken{},
		&models.SubscriptionPlan{}, // Планы подписки
		&models.License{},          // Лицензии
		&models.UpdateRelease{},    // Обновления приложений/сервера
	); err != nil {
		return fmt.Errorf("failed to run GORM AutoMigrate: %w", err)
	}

	// Затем выполняем SQL миграции из папки migrations
	if err := runSQLMigrations(); err != nil {
		log.Printf("⚠️ Warning: Failed to run SQL migrations: %v", err)
		// Не прерываем работу, но логируем предупреждение
	}

	return nil
}

// runSQLMigrations выполняет SQL миграции из папки database/migrations
func runSQLMigrations() error {
	migrationsDir := "database/migrations"
	
	// Проверяем существование папки
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		log.Printf("ℹ️ Migrations directory not found: %s", migrationsDir)
		return nil
	}

	// Получаем список SQL файлов
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	if len(files) == 0 {
		log.Printf("ℹ️ No SQL migration files found in %s", migrationsDir)
		return nil
	}

	log.Printf("📋 Found %d SQL migration files", len(files))

	// Получаем базовое подключение для выполнения SQL
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// Выполняем каждую миграцию
	for _, file := range files {
		fileName := filepath.Base(file)
		log.Printf("🔄 Running SQL migration: %s", fileName)

		// Читаем содержимое файла
		sqlContent, err := os.ReadFile(file)
		if err != nil {
			log.Printf("❌ Failed to read migration file %s: %v", fileName, err)
			continue
		}

		// Разбиваем на отдельные команды (разделитель - точка с запятой)
		statements := strings.Split(string(sqlContent), ";")
		
		for _, statement := range statements {
			statement = strings.TrimSpace(statement)
			// Пропускаем пустые строки и комментарии
			if statement == "" || strings.HasPrefix(statement, "--") {
				continue
			}

			// Выполняем SQL команду
			if _, err := sqlDB.Exec(statement); err != nil {
				// Игнорируем ошибки "уже существует" (IF NOT EXISTS)
				if strings.Contains(err.Error(), "already exists") || 
				   strings.Contains(err.Error(), "duplicate") {
					log.Printf("ℹ️ Migration %s: %s (already applied)", fileName, err.Error())
					continue
				}
				log.Printf("❌ Failed to execute migration %s: %v", fileName, err)
				log.Printf("   Statement: %s", statement[:min(100, len(statement))])
				// Продолжаем выполнение других миграций
			}
		}

		log.Printf("✅ Migration %s completed", fileName)
	}

	return nil
}

// min возвращает минимальное из двух чисел
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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

// prepareShopSubscriptionsForMigration обновляет shop_id в shop_subscriptions перед миграцией
// Это нужно, чтобы внешний ключ мог быть добавлен успешно
func prepareShopSubscriptionsForMigration() error {
	// Проверяем, есть ли таблица shop_subscriptions
	var tableExists bool
	if err := DB.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = CURRENT_SCHEMA() AND table_name = 'shop_subscriptions')").Scan(&tableExists).Error; err != nil {
		return fmt.Errorf("failed to check shop_subscriptions table: %w", err)
	}

	if !tableExists {
		log.Println("ℹ️ shop_subscriptions table doesn't exist yet, skipping preparation")
		return nil
	}

	// Проверяем, есть ли таблица shops
	var shopsTableExists bool
	if err := DB.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = CURRENT_SCHEMA() AND table_name = 'shops')").Scan(&shopsTableExists).Error; err != nil {
		return fmt.Errorf("failed to check shops table: %w", err)
	}

	if !shopsTableExists {
		log.Println("ℹ️ shops table doesn't exist yet, skipping preparation")
		return nil
	}

	// Находим все shop_subscriptions, где shop_id не существует в shops
	// и обновляем их, создавая shops из users если нужно
	var subscriptions []struct {
		ShopID uuid.UUID
		UserID uuid.UUID
	}

	// Находим подписки, где shop_id не существует в shops
	if err := DB.Raw(`
		SELECT ss.shop_id, ss.user_id 
		FROM shop_subscriptions ss
		WHERE NOT EXISTS (
			SELECT 1 FROM shops s WHERE s.id = ss.shop_id
		)
	`).Scan(&subscriptions).Error; err != nil {
		return fmt.Errorf("failed to find invalid shop_subscriptions: %w", err)
	}

	if len(subscriptions) == 0 {
		log.Println("✅ All shop_subscriptions are valid")
		return nil
	}

	log.Printf("🔄 Found %d shop_subscriptions with invalid shop_id, fixing...", len(subscriptions))

	// Для каждой подписки создаем shop из user, если его нет
	for _, sub := range subscriptions {
		// Проверяем, существует ли shop с таким ID
		var shop models.Shop
		if err := DB.Where("id = ?", sub.ShopID).First(&shop).Error; err == nil {
			// Shop уже существует, пропускаем
			continue
		}

		// Проверяем, существует ли user с таким ID и является ли он shop_owner
		var user models.User
		if err := DB.Preload("Role").Where("id = ?", sub.ShopID).First(&user).Error; err != nil {
			log.Printf("⚠️ User %s not found for shop_subscription, skipping", sub.ShopID)
			continue
		}

		// Проверяем, является ли пользователь shop_owner
		if user.Role == nil || user.Role.Name != "shop_owner" {
			log.Printf("⚠️ User %s is not a shop_owner, skipping", sub.ShopID)
			continue
		}

		// Создаем shop из user
		shop = models.Shop{
			ID:        user.ID, // Используем тот же ID
			Name:      user.Name,
			INN:       user.INN,
			Email:     user.Email,
			Phone:     user.Phone,
			Logo:      user.Avatar,
			IsActive:  user.IsActive,
			OwnerID:   user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}

		if err := DB.Create(&shop).Error; err != nil {
			log.Printf("⚠️ Failed to create shop for user %s: %v", user.ID, err)
			continue
		}

		log.Printf("✅ Created shop %s for user %s (from shop_subscription)", shop.ID, user.ID)
	}

	return nil
}

// createDefaultCities создает города по умолчанию
func createDefaultCities() error {
	// Список городов Таджикистана с координатами
	defaultCities := []struct {
		name      string
		latitude  float64
		longitude float64
	}{
		{"Душанбе", 38.5598, 68.7870},
		{"Худжанд", 40.2833, 69.6167},
		{"Куляб", 37.9097, 69.7844},
		{"Бохтар", 37.8364, 68.7803},
		{"Истаравшан", 39.9108, 69.0064},
		{"Пенджикент", 39.4953, 67.6094},
		{"Хорог", 37.4897, 71.5531},
		{"Исфара", 40.1264, 70.6253},
		{"Канибадам", 40.2833, 70.4167}, // Канибадам
	}

	for _, cityData := range defaultCities {
		var existingCity models.City
		if err := DB.Where("name = ?", cityData.name).First(&existingCity).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				city := models.City{
					Name:      cityData.name,
					Latitude:  cityData.latitude,
					Longitude: cityData.longitude,
					IsActive:  true,
				}
				if err := DB.Create(&city).Error; err != nil {
					log.Printf("⚠️ Failed to create city %s: %v", cityData.name, err)
					continue
				}
				log.Printf("✅ City created: %s", cityData.name)
			} else {
				log.Printf("⚠️ Error checking city %s: %v", cityData.name, err)
			}
		} else {
			log.Printf("✅ City already exists: %s", cityData.name)
		}
	}

	return nil
}

// migrateShopsFromUsers мигрирует данные из users (shop_owner) в shops
func migrateShopsFromUsers() error {
	// Получаем всех пользователей с ролью shop_owner
	var shopOwners []models.User
	if err := DB.Preload("Role").Where("role_id IN (SELECT id FROM roles WHERE name = 'shop_owner')").Find(&shopOwners).Error; err != nil {
		return fmt.Errorf("failed to find shop owners: %w", err)
	}

	log.Printf("📦 Found %d shop owners to migrate", len(shopOwners))

	for _, owner := range shopOwners {
		// Проверяем, существует ли уже shop для этого owner
		var existingShop models.Shop
		if err := DB.Where("owner_id = ?", owner.ID).First(&existingShop).Error; err == nil {
			log.Printf("✅ Shop already exists for owner %s (%s), skipping", owner.ID, owner.Email)
			continue
		}

		// Создаем shop из данных owner
		shop := models.Shop{
			ID:        owner.ID, // Используем тот же ID для обратной совместимости
			Name:      owner.Name,
			INN:       owner.INN,
			Email:     owner.Email,
			Phone:     owner.Phone,
			Logo:      owner.Avatar, // Avatar -> Logo
			IsActive:  owner.IsActive,
			OwnerID:   owner.ID,
			CreatedAt: owner.CreatedAt,
			UpdatedAt: owner.UpdatedAt,
		}

		if err := DB.Create(&shop).Error; err != nil {
			log.Printf("❌ Failed to create shop for owner %s: %v", owner.ID, err)
			continue
		}

		log.Printf("✅ Created shop %s for owner %s", shop.ID, owner.ID)

		// Обновляем продукты: owner_id -> shop_id
		result := DB.Model(&models.Product{}).
			Where("owner_id = ? AND shop_id IS NULL", owner.ID).
			Update("shop_id", shop.ID)
		if result.Error != nil {
			log.Printf("⚠️ Failed to update products for shop %s: %v", shop.ID, result.Error)
		} else {
			log.Printf("✅ Updated %d products for shop %s", result.RowsAffected, shop.ID)
		}

		// Обновляем order_items: shop_owner_id -> shop_id
		result = DB.Model(&models.OrderItem{}).
			Where("shop_owner_id = ? AND shop_id IS NULL", owner.ID).
			Update("shop_id", shop.ID)
		if result.Error != nil {
			log.Printf("⚠️ Failed to update order items for shop %s: %v", shop.ID, result.Error)
		} else {
			log.Printf("✅ Updated %d order items for shop %s", result.RowsAffected, shop.ID)
		}

		// Обновляем shop_subscriptions: shop_id должен ссылаться на shops, а не users
		result = DB.Model(&models.ShopSubscription{}).
			Where("shop_id = ?", owner.ID).
			Update("shop_id", shop.ID)
		if result.Error != nil {
			log.Printf("⚠️ Failed to update shop_subscriptions for shop %s: %v", shop.ID, result.Error)
		} else {
			log.Printf("✅ Updated %d shop_subscriptions for shop %s", result.RowsAffected, shop.ID)
		}
	}

	return nil
}

// cleanDeviceIDWhitespace очищает лишние пробелы и переносы строк из device_id в таблице licenses
func cleanDeviceIDWhitespace() error {
	// Используем raw SQL для обновления всех записей
	result := DB.Exec(`
		UPDATE licenses 
		SET device_id = TRIM(REGEXP_REPLACE(device_id, E'[\\n\\r\\t]+', '', 'g'))
		WHERE device_id IS NOT NULL 
		  AND device_id != TRIM(REGEXP_REPLACE(device_id, E'[\\n\\r\\t]+', '', 'g'))
	`)
	
	if result.Error != nil {
		return result.Error
	}
	
	if result.RowsAffected > 0 {
		log.Printf("✅ Очищено %d записей с лишними пробелами в device_id", result.RowsAffected)
	} else {
		log.Println("✅ Нет записей с лишними пробелами в device_id")
	}
	
	return nil
}
