// ===== APP.JS - Главный модуль приложения =====

// Инициализация приложения
async function initializeApp() {
    console.log('🚀 Инициализация админ панели...');
    
    // Загружаем сохраненный токен и роль
    const adminToken = localStorage.getItem('adminToken');
    const userRole = localStorage.getItem('userRole');
    
    console.log('🔑 Загруженный токен:', adminToken ? 'Присутствует' : 'Отсутствует');
    console.log('👤 Загруженная роль:', userRole);
    
    // Устанавливаем токен и роль в storage для использования другими модулями
    if (window.storage) {
        if (adminToken) window.storage.setAdminToken(adminToken);
        if (userRole) window.storage.setUserRole(userRole);
    }
    
    // Если есть токен, проверяем его валидность
    if (adminToken && userRole) {
        console.log('🔑 Проверяем валидность токена...');
        
        // Проверяем время последней активности (24 часа)
        const lastActivity = localStorage.getItem('lastActivity');
        const now = Date.now();
        const twentyFourHours = 24 * 60 * 60 * 1000;
        
        if (lastActivity && (now - parseInt(lastActivity)) > twentyFourHours) {
            console.log('⏰ Токен истек, очищаем...');
            clearAllStorage();
            showLoginForm();
        } else {
            try {
                console.log('🌐 Проверяем API доступность...');
                
                // Проверяем токен через API
                const response = await fetch(getApiUrl(CONFIG.API.ENDPOINTS.AUTH.PROFILE), {
                    headers: {
                        'Authorization': `Bearer ${adminToken}`
                    }
                });
                
                console.log('📡 Ответ API:', response.status, response.statusText);
                
                if (response.ok) {
                    const data = await response.json();
                    console.log('✅ Токен валиден, пользователь:', data.data?.user);
                    
                    // Обновляем время последней активности
                    localStorage.setItem('lastActivity', now.toString());
                    
                    // Показываем админ панель
                    showAdminPanel();
                    
                    // Обновляем информацию о пользователе
                    if (window.auth && window.auth.updateUserInfo) {
                        window.auth.updateUserInfo(data.data?.user, userRole);
                    }
                    
                    // Настройка навигации
                    if (window.navigation && window.navigation.setupNavigation) {
                        window.navigation.setupNavigation(userRole);
                    }
                    
                    // Загружаем данные
                    setTimeout(() => {
                        loadInitialData(userRole);
                        // Обрабатываем deep link из push-уведомления (если есть)
                        handleDeepLink();
                    }, 100);
                } else {
                    console.log('❌ Токен невалиден, очищаем...');
                    clearAllStorage();
                    showLoginForm();
                }
            } catch (error) {
                console.error('❌ Ошибка проверки токена:', error);
                clearAllStorage();
                showLoginForm();
            }
        }
    } else {
        showLoginForm();
    }
    
    // Настройка форм
    setupForms();
    
    // Настройка фильтров
    setupFilters();
    
    // Настройка фильтров для заказов
    if (window.orders && window.orders.setupOrderFilters) {
        window.orders.setupOrderFilters();
    }
    
    // Настройка UI
    setupUIFeatures();
    
    console.log('✅ Админ панель инициализирована');
}

// Очистка хранилища
function clearAllStorage() {
    localStorage.removeItem('adminToken');
    localStorage.removeItem('userRole');
    localStorage.removeItem('lastActivity');
    localStorage.removeItem('userData');
}

// Показ формы входа
function showLoginForm() {
    document.getElementById('login-modal').style.display = 'block';
    document.getElementById('admin-content').style.display = 'none';
}

// Показ админ панели
function showAdminPanel() {
    document.getElementById('login-modal').style.display = 'none';
    document.getElementById('admin-content').style.display = 'flex';
}

// Загрузка начальных данных
function loadInitialData(userRole) {
    console.log('📊 Загружаем начальные данные для роли:', userRole);
    
    const roleName = typeof userRole === 'object' ? userRole.name : userRole;
    
    if (roleName === 'super_admin' || roleName === 'admin') {
        console.log('🔱 Загружаем данные для админа...');
        if (window.dashboard && window.dashboard.loadDashboard) window.dashboard.loadDashboard(userRole);
        if (window.categories && window.categories.loadCategories) window.categories.loadCategories();
        if (window.users && window.users.loadUsers) window.users.loadUsers();
        if (window.products && window.products.loadProducts) window.products.loadProducts();
        if (window.orders && window.orders.loadOrders) window.orders.loadOrders();
    } else if (roleName === 'shop_owner') {
        console.log('🏪 Загружаем данные для владельца магазина...');
        if (window.dashboard && window.dashboard.loadDashboard) window.dashboard.loadDashboard(userRole);
        if (window.products && window.products.loadProducts) window.products.loadProducts();
        if (window.orders && window.orders.loadOrders) window.orders.loadOrders();
    } else {
        console.log('👤 Загружаем данные для пользователя...');
        if (window.dashboard && window.dashboard.loadDashboard) window.dashboard.loadDashboard(userRole);
        if (window.categories && window.categories.loadCategories) window.categories.loadCategories();
        if (window.products && window.products.loadProducts) window.products.loadProducts();
    }
}

// Настройка форм
function setupForms() {
    // Форма входа
    const loginForm = document.getElementById('login-form');
    if (loginForm) {
        loginForm.addEventListener('submit', function(e) {
            e.preventDefault();
            if (window.auth && window.auth.handleLogin) {
                window.auth.handleLogin(e);
                // После успешного входа перезагрузим вкладку заказов согласно роли
                setTimeout(() => {
                    if (window.orders && window.orders.loadOrders) {
                        window.orders.loadOrders(1, {});
                    }
                }, 200);
            }
        });
    }
    
    // Форма товара
    const productForm = document.getElementById('product-form');
    if (productForm) {
        productForm.addEventListener('submit', function(e) {
            e.preventDefault();
            if (window.products && window.products.handleProductSubmit) {
                window.products.handleProductSubmit(e);
            }
        });
    }
    
    // Форма категории
    const categoryForm = document.getElementById('category-form');
    if (categoryForm) {
        categoryForm.addEventListener('submit', function(e) {
            e.preventDefault();
            if (window.categories && window.categories.handleCategorySubmit) {
                window.categories.handleCategorySubmit(e);
            }
        });
    }
    
    // Форма создания пользователя
    const userForm = document.getElementById('user-form');
    if (userForm) {
        userForm.addEventListener('submit', function(e) {
            e.preventDefault();
            if (window.users && window.users.handleUserSubmit) {
                window.users.handleUserSubmit(e);
            }
        });
    }
}


// Настройка фильтров
function setupFilters() {
    const searchInput = document.getElementById('product-search');
    const categoryFilter = document.getElementById('category-filter');
    
    if (searchInput) {
        searchInput.addEventListener('input', filterProducts);
    }
    
    if (categoryFilter) {
        categoryFilter.addEventListener('change', filterProducts);
    }
}

// Фильтрация товаров
function filterProducts() {
    console.log('🔄 Фильтрация товаров...');
    
    const searchTerm = document.getElementById('product-search')?.value?.toLowerCase() || '';
    const categoryFilter = document.getElementById('category-filter')?.value || '';
    
    console.log('🔍 Параметры фильтрации:', { searchTerm, categoryFilter });
    
    // Получаем все товары из storage
    let allProducts = [];
    if (window.storage && window.storage.getAllProducts) {
        allProducts = window.storage.getAllProducts();
        console.log(`📦 Получено ${allProducts.length} товаров из storage`);
    } else {
        console.warn('⚠️ Storage недоступен, не можем фильтровать');
        return;
    }
    
    if (!allProducts || allProducts.length === 0) {
        console.warn('⚠️ Нет товаров для фильтрации');
        return;
    }
    
    // Фильтруем товары
    const filteredProducts = allProducts.filter(product => {
        const matchesSearch = !searchTerm || 
            product.name?.toLowerCase().includes(searchTerm) ||
            product.brand?.toLowerCase().includes(searchTerm);
        
        const matchesCategory = !categoryFilter || 
            product.categoryId === categoryFilter || 
            (product.category && product.category.id === categoryFilter);
        
        return matchesSearch && matchesCategory;
    });
    
    console.log(`📊 Отфильтровано ${filteredProducts.length} из ${allProducts.length} товаров`);
    
    // Отображаем отфильтрованные товары
    if (window.products && window.products.displayProducts) {
        window.products.displayProducts(filteredProducts);
    } else {
        console.error('❌ window.products.displayProducts недоступен');
    }
}

// Настройка UI функций
function setupUIFeatures() {
    // Настройка мобильной навигации
    if (window.navigation && window.navigation.setupMobileNavigation) {
        window.navigation.setupMobileNavigation();
    }
    
    // Настройка адаптивности
    if (window.navigation && window.navigation.setupResponsiveTables) {
        window.navigation.setupResponsiveTables();
    }
    
    if (window.navigation && window.navigation.optimizeForMobile) {
        window.navigation.optimizeForMobile();
    }
    
    // Проверка подключения
    if (window.api && window.api.testConnection) {
        window.api.testConnection();
    }
}

// Обработка deep links из push-уведомлений
function handleDeepLink() {
    const hash = window.location.hash;
    if (hash && hash.includes('?')) {
        const [tabName, params] = hash.substring(1).split('?');
        const urlParams = new URLSearchParams(params);
        const orderId = urlParams.get('orderId');
        
        // Переключаемся на нужную вкладку
        if (tabName && window.navigation && window.navigation.showTab) {
            window.navigation.showTab(tabName);
        }
        
        // Если есть orderId, открываем детали заказа
        if (orderId && tabName === 'orders' && window.orders && window.orders.viewOrderDetails) {
            setTimeout(() => {
                window.orders.viewOrderDetails(orderId);
            }, 1500); // Задержка для загрузки списка заказов
        }
    }
}

// Функции для обратной совместимости
window.allProducts = [];
window.filterProducts = filterProducts;
window.showTab = (tabName) => {
    if (window.navigation && window.navigation.showTab) {
        window.navigation.showTab(tabName);
    }
};

window.viewProductVariations = (id) => {
    if (window.products && window.products.viewProductVariations) {
        window.products.viewProductVariations(id);
    }
};

window.deleteProduct = (id) => {
    if (window.products && window.products.deleteProduct) {
        window.products.deleteProduct(id);
    }
};

window.loadProducts = () => {
    if (window.products && window.products.loadProducts) {
        window.products.loadProducts();
    }
};

// Инициализация при загрузке страницы
document.addEventListener('DOMContentLoaded', function() {
    console.log('🚀 MM Admin Panel v3.0 загружена!');
    console.log('📅 Время загрузки:', new Date().toLocaleString());
    console.log('🌐 User Agent:', navigator.userAgent);
    initializeApp();
});

// Экспорт
window.app = {
    initializeApp,
    loadInitialData
};

