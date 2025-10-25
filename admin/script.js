// Конфигурация
// Используем централизованный конфиг
let API_BASE_URL = CONFIG.API.BASE_URL;
let currentProductId = null;
let currentCategoryId = null;
let adminToken = null;
let userRole = null;

// Глобальные переменные для загрузки изображений
let uploadedImages = [];
let imageUrls = [];

// Глобальные переменные для вариаций
let variations = [];

// Инициализация при загрузке страницы
document.addEventListener('DOMContentLoaded', function() {
    console.log('🚀 MM Admin Panel v3.0 загружена!');
    console.log('📅 Время загрузки:', new Date().toLocaleString());
    console.log('🌐 User Agent:', navigator.userAgent);
    initializeApp();
});

// Основная инициализация
async function initializeApp() {
    console.log('🚀 Инициализация админ панели...');
    
    // Загружаем сохраненный токен и роль
    adminToken = localStorage.getItem('adminToken');
    userRole = localStorage.getItem('userRole');
    
    console.log('🔑 Загруженный токен:', adminToken ? 'Присутствует' : 'Отсутствует');
    console.log('👤 Загруженная роль:', userRole);
    console.log('🔍 localStorage adminToken:', localStorage.getItem('adminToken'));
    console.log('🔍 localStorage userRole:', localStorage.getItem('userRole'));
    console.log('🔍 localStorage lastActivity:', localStorage.getItem('lastActivity'));
    
    // Если есть токен, проверяем его валидность и время жизни
    if (adminToken && userRole) {
        console.log('🔑 Проверяем валидность токена...');
        
        // Проверяем время последней активности (24 часа)
        const lastActivity = localStorage.getItem('lastActivity');
        const now = Date.now();
        const twentyFourHours = 24 * 60 * 60 * 1000; // 24 часа в миллисекундах
        
        if (lastActivity && (now - parseInt(lastActivity)) > twentyFourHours) {
            console.log('⏰ Токен истек (прошло больше 24 часов), очищаем...');
            localStorage.removeItem('adminToken');
            localStorage.removeItem('userRole');
            localStorage.removeItem('lastActivity');
            adminToken = null;
            userRole = null;
            
            // Показываем форму входа
            document.getElementById('login-modal').style.display = 'flex';
            document.getElementById('admin-content').style.display = 'none';
    } else {
            try {
                console.log('🌐 Проверяем API доступность...');
                console.log('🔗 URL:', `${API_BASE_URL}/api/v1/users/profile`);
                console.log('🔑 Токен для проверки:', adminToken ? 'Присутствует' : 'Отсутствует');
                
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
                    console.log('🔍 Полный ответ API:', data);
                    
                    // Обновляем время последней активности
                    localStorage.setItem('lastActivity', now.toString());
                    
                    // Показываем админ панель
                    document.getElementById('login-modal').style.display = 'none';
                    document.getElementById('admin-content').style.display = 'flex';
                    
                    // Обновляем информацию о пользователе
                    console.log('🔄 Вызываем updateUserInfo с данными:', { user: data.data?.user, userRole });
                    updateUserInfo(data.data?.user, userRole);
                    setupNavigation(userRole);
                    loadSettings();
                    
                    // Загружаем данные
                    setTimeout(() => {
                        loadDashboard(userRole);
                        loadInitialData(userRole);
                    }, 100);
                } else {
                    console.log('❌ Токен невалиден, очищаем...');
                    // Токен невалиден, очищаем
                    localStorage.removeItem('adminToken');
                    localStorage.removeItem('userRole');
                    localStorage.removeItem('lastActivity');
                    adminToken = null;
                    userRole = null;
                    
                    // Показываем форму входа
                    document.getElementById('login-modal').style.display = 'flex';
                    document.getElementById('admin-content').style.display = 'none';
                }
            } catch (error) {
                console.error('❌ Ошибка проверки токена:', error);
                // При ошибке тоже очищаем
                localStorage.removeItem('adminToken');
                localStorage.removeItem('userRole');
                localStorage.removeItem('lastActivity');
                adminToken = null;
                userRole = null;
                
                // Показываем форму входа
                document.getElementById('login-modal').style.display = 'flex';
                document.getElementById('admin-content').style.display = 'none';
            }
        }
    } else {
        // Показываем форму входа
        document.getElementById('login-modal').style.display = 'flex';
        document.getElementById('admin-content').style.display = 'none';
    }
    
    // Настройка форм
    setupLoginForm();
    setupForms();
    
    // Настройка мобильной навигации
    setupMobileNavigation();
    // setupMobileTabbar(); // Отключено - убираем нижний таббар
    
    // Дополнительная настройка мобильной навигации после загрузки DOM
    setTimeout(() => {
        console.log('🔄 Дополнительная настройка мобильной навигации...');
        setupMobileNavigation();
    }, 500);
    
    // Настройка адаптивности
    setupResponsiveTables();
    optimizeForMobile();
    
    // Настройка фильтров
    setupFilters();
    
    // Настройка загрузки изображений
    setupImageUpload();
    
    // Проверка подключения
    testConnection();
    
    console.log('✅ Админ панель инициализирована');
}

// Функция обновления времени последней активности
function updateLastActivity() {
    localStorage.setItem('lastActivity', Date.now().toString());
    console.log('🕐 Время активности обновлено');
}

// Функция загрузки начальных данных в зависимости от роли
function loadInitialData(userRole) {
    console.log('📊 Загружаем начальные данные для роли:', userRole);
    
    const roleName = typeof userRole === 'object' ? userRole.name : userRole;
    
    if (roleName === 'super_admin') {
        console.log('🔱 Загружаем данные для супер админа...');
        loadCategories();
        loadUsers();
        loadRoles();
        loadProducts();
        loadOrders();
    } else if (roleName === 'admin') {
        console.log('👑 Загружаем данные для админа...');
        loadCategories();
        loadUsers();
        loadProducts();
        loadOrders();
    } else if (roleName === 'shop_owner') {
        console.log('🏪 Загружаем данные для владельца магазина...');
        loadProducts();
        loadOrders();
    } else {
        console.log('👤 Загружаем данные для обычного пользователя...');
        loadCategories();
        loadProducts();
    }
}

// Настройка формы входа
function setupLoginForm() {
    const loginForm = document.getElementById('login-form');
    loginForm.addEventListener('submit', handleLogin);
}

// Обработка входа
async function handleLogin(e) {
    e.preventDefault();
    
    const phone = document.getElementById('login-phone').value;
    const password = document.getElementById('login-password').value;
    
    // Показываем сообщение о загрузке
    showMessage('Проверяем данные...', 'info');
    
    try {
        const response = await fetch(getApiUrl(CONFIG.API.ENDPOINTS.AUTH.LOGIN), {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ phone, password })
        });
        
        const data = await response.json();
        console.log('Ответ сервера:', data);
        
        if (response.ok && data.success && data.data && data.data.token) {
            adminToken = data.data.token;
            // Получаем роль как строку из объекта роли
            userRole = data.data.user?.role?.name || 'user';
            console.log('✅ Успешный вход, токен получен, роль:', userRole);
            
            // Скрываем форму входа и показываем админ панель
            document.getElementById('login-modal').style.display = 'none';
            document.getElementById('admin-content').style.display = 'flex';
            
            // Сохраняем токен, роль и данные пользователя
            localStorage.setItem('adminToken', data.data.token);
            localStorage.setItem('userRole', userRole);
            localStorage.setItem('lastActivity', Date.now().toString());
            localStorage.setItem('userData', JSON.stringify(data.data.user));
            
            console.log('💾 Сохранено в localStorage:');
            console.log('  - adminToken:', localStorage.getItem('adminToken') ? 'Присутствует' : 'Отсутствует');
            console.log('  - userRole:', localStorage.getItem('userRole'));
            console.log('  - lastActivity:', localStorage.getItem('lastActivity'));
            console.log('  - userData:', localStorage.getItem('userData'));
            
            // Обновляем информацию о пользователе
            updateUserInfo(data.data.user, userRole);
            
            // Инициализируем панель в зависимости от роли
            setupNavigation(userRole);
            loadSettings();
            
            // Настройка форм
            setupForms();
            setupFilters();
            
                // Настройка мобильной навигации
    setupMobileNavigation();
    // setupMobileTabbar(); // Отключено - убираем нижний таббар
            
            // Загружаем данные с небольшой задержкой, чтобы токен успел установиться
            setTimeout(() => {
                console.log('🔄 Начинаем загрузку данных после входа...');
                console.log('👤 Роль пользователя:', userRole);
                
                loadDashboard(userRole);
                
                // Загружаем данные в зависимости от роли
                const roleName = typeof userRole === 'object' ? userRole.name : userRole;
                
                if (roleName === 'super_admin') {
                    console.log('🔱 Загружаем данные для супер админа...');
                    loadCategories();
                    loadUsers();
                    loadRoles();
                    loadProducts();
                    loadOrders();
                } else if (roleName === 'admin') {
                    console.log('👑 Загружаем данные для админа...');
                    loadCategories();
                    loadUsers();
                    loadProducts();
                    loadOrders();
                } else if (roleName === 'shop_owner') {
                    console.log('🏪 Загружаем данные для владельца магазина...');
                    loadProducts();
                    loadOrders();
                } else {
                    console.log('👤 Загружаем данные для обычного пользователя...');
                    loadCategories();
                    loadProducts();
                }
            }, 100);
            
            const roleText = (userRole === 'super_admin' || userRole === 'admin') ? 'админ панель' : 'панель управления магазином';
            showMessage(`Успешный вход в ${roleText}!`, 'success');
        } else {
            // Обрабатываем разные типы ошибок
            let errorMessage = 'Неверный email или пароль';
            
            if (response.status === 401) {
                errorMessage = '❌ Неверный email или пароль. Проверьте данные и попробуйте снова.';
            } else if (response.status === 403) {
                errorMessage = '🚫 У вас нет прав доступа к админ панели.';
            } else if (response.status === 429) {
                errorMessage = '⏳ Слишком много попыток входа. Попробуйте позже.';
            } else if (response.status >= 500) {
                errorMessage = '🔧 Ошибка сервера. Попробуйте позже или обратитесь к администратору.';
            } else if (data.message) {
                errorMessage = data.message;
            } else if (data.error) {
                errorMessage = data.error;
            }
            
            showMessage(errorMessage, 'error');
            
            // Встряхиваем форму для визуального эффекта
            const loginModal = document.getElementById('login-modal');
            loginModal.classList.add('shake');
            setTimeout(() => loginModal.classList.remove('shake'), 600);
        }
        } catch (error) {
        console.error('Ошибка входа:', error);
        let errorMessage = 'Ошибка подключения к серверу.';
        
        if (error.name === 'TypeError' && error.message.includes('fetch')) {
            errorMessage = '🌐 Не удается подключиться к серверу. Проверьте подключение к интернету.';
        } else if (error.message.includes('timeout')) {
            errorMessage = '⏰ Время ожидания истекло. Попробуйте снова.';
        }
        
        showMessage(errorMessage, 'error');
    }
}

// Обновление информации о пользователе
function updateUserInfo(user, userRole) {
    console.log('🔄 updateUserInfo вызвана с данными:', { user, userRole });
    
    // Получаем имя пользователя
    let userName = 'Пользователь';
    let userEmail = '';
    
    if (user) {
        if (user.name && user.name.trim() !== '') {
            userName = user.name;
        } else if (user.email && user.email.trim() !== '') {
            userName = user.email.split('@')[0]; // Берем часть до @
        }
        userEmail = user.email || '';
    } else {
        // Если нет данных пользователя, пробуем взять из localStorage
        const savedUserData = localStorage.getItem('userData');
        if (savedUserData) {
            try {
                const userData = JSON.parse(savedUserData);
                console.log('🔄 Используем сохраненные данные пользователя:', userData);
                
                if (userData.name && userData.name.trim() !== '') {
                    userName = userData.name;
                } else if (userData.email && userData.email.trim() !== '') {
                    userName = userData.email.split('@')[0];
                }
                userEmail = userData.email || '';
            } catch (error) {
                console.error('❌ Ошибка парсинга сохраненных данных:', error);
            }
        }
    }
    
    console.log('🔍 Данные пользователя для отображения:', { 
        originalUser: user, 
        userName, 
        userEmail 
    });
    
    // Определяем роль для отображения
    let roleDisplay = '';
    // Проверяем, является ли userRole объектом или строкой
    const roleName = typeof userRole === 'object' ? userRole.name : userRole;
    
    switch (roleName) {
        case 'admin':
            roleDisplay = 'Администратор';
            break;
        case 'shop_owner':
            roleDisplay = 'Владелец магазина';
            break;
        case 'user':
            roleDisplay = 'Пользователь';
            break;
        default:
            roleDisplay = 'Пользователь';
    }
    
    // Обновляем элементы header по их ID
    const headerUserName = document.getElementById('header-user-name');
    const headerUserEmail = document.getElementById('header-user-email');
    const headerUserRole = document.getElementById('header-user-role');
    
    if (headerUserName) headerUserName.textContent = userName;
    if (headerUserEmail) headerUserEmail.textContent = userEmail;
    if (headerUserRole) headerUserRole.textContent = roleDisplay;
    
    console.log('✅ Информация о пользователе обновлена:', { userName, userEmail, roleDisplay });
}

// Переключение выпадающего меню пользователя
function toggleUserDropdown() {
    updateLastActivity(); // Обновляем время активности
    const dropdown = document.getElementById('user-dropdown');
    if (dropdown) {
        dropdown.classList.toggle('show');
    }
}

// Закрытие dropdown при клике вне его
document.addEventListener('click', function(event) {
    const userMenu = document.querySelector('.user-menu');
    const dropdown = document.getElementById('user-dropdown');
    
    if (dropdown && userMenu && !userMenu.contains(event.target)) {
        dropdown.classList.remove('show');
    }
});

// Выход из системы
function logout() {
    // Очищаем токен и данные пользователя
    adminToken = null;
    userRole = null;
    
    // Скрываем основной контент
    document.getElementById('admin-content').style.display = 'none';
        
        // Показываем форму входа
    document.getElementById('login-modal').style.display = 'block';
    
    // Очищаем форму входа
    document.getElementById('login-form').reset();
    
    // Скрываем выпадающее меню
    const dropdown = document.getElementById('user-dropdown');
    if (dropdown) {
        dropdown.classList.remove('show');
    }
    
    showMessage('Вы успешно вышли из системы', 'success');
}

// Тестирование подключения к API
async function testConnection() {
    try {
        showMessage('Проверяем подключение к API...', 'info');
        
        const response = await fetch(getApiUrl(CONFIG.API.ENDPOINTS.HEALTH));
        const data = await response.json();
        
        if (response.ok) {
            showMessage('✅ API сервер работает! Статус: ' + (data.status || 'OK'), 'success');
        } else {
            showMessage('❌ API сервер не отвечает', 'error');
        }
    } catch (error) {
        console.error('Ошибка подключения:', error);
        showMessage('❌ Не удается подключиться к API серверу. Убедитесь, что сервер запущен.', 'error');
    }
}

// Навигация между вкладками
function setupNavigation(userRole = 'admin') {
    console.log('🔧 Настройка навигации для роли:', userRole);
    
    // Настройка сайдбара
    const navItems = document.querySelectorAll('.nav-item');
    
    // Скрываем все элементы навигации
    navItems.forEach(item => {
        item.style.display = 'none';
    });
    
    // Показываем элементы в зависимости от роли
    if (userRole === 'super_admin') {
        // Супер админ видит все разделы включая роли и пользователей
        navItems.forEach(item => {
            item.style.display = 'flex';
        });
    } else if (userRole === 'admin') {
        // Обычный админ видит все кроме ролей
        const allowedTabs = ['dashboard', 'products', 'categories', 'users', 'orders', 'settings'];
        
        navItems.forEach(item => {
            const tabName = item.dataset.tab;
            if (allowedTabs.includes(tabName)) {
                item.style.display = 'flex';
            } else {
                item.style.display = 'none';
            }
        });
    } else if (userRole === 'shop_owner') {
        // Владелец магазина видит только свои разделы (БЕЗ категорий)
        const allowedTabs = ['dashboard', 'products', 'orders', 'settings'];
        
        navItems.forEach(item => {
            const tabName = item.dataset.tab;
            if (allowedTabs.includes(tabName)) {
                item.style.display = 'flex';
            } else {
                item.style.display = 'none';
            }
        });
    }
    
    // Обработчики для элементов навигации
    navItems.forEach(item => {
        item.addEventListener('click', function() {
            // Убираем активный класс у всех элементов
            navItems.forEach(nav => nav.classList.remove('active'));
            
            // Добавляем активный класс к выбранному элементу
            this.classList.add('active');
            
            // Показываем соответствующую вкладку
            const tabName = this.dataset.tab;
            showTab(tabName, userRole);
        });
    });
    
    // Показываем первую доступную вкладку
    const firstVisibleItem = Array.from(navItems).find(item => item.style.display !== 'none');
    if (firstVisibleItem) {
        firstVisibleItem.classList.add('active');
        const tabName = firstVisibleItem.dataset.tab;
        showTab(tabName, userRole);
    }
}

// Функция для программного переключения вкладок
function showTab(tabName, userRole = 'admin') {
    updateLastActivity(); // Обновляем время активности
    console.log(`🔄 Переключаемся на вкладку: ${tabName}, роль: ${userRole}`);
    console.log('📊 Состояние DOM при переключении:', document.readyState);
    console.log('⏰ Время переключения:', new Date().toISOString());
    
    // Проверяем, является ли userRole объектом или строкой
    const roleName = typeof userRole === 'object' ? userRole.name : userRole;
    console.log('🔍 Используем роль в showTab:', roleName);
    
    const navItems = document.querySelectorAll('.nav-item');
    const tabContents = document.querySelectorAll('.tab-content');
    const currentPage = document.getElementById('current-page');
    
    console.log('🔍 Найдено navItems:', navItems.length);
    console.log('🔍 Найдено tabContents:', tabContents.length);
    
    // Убираем активный класс со всех элементов
    navItems.forEach(nav => nav.classList.remove('active'));
    tabContents.forEach(tab => tab.classList.remove('active'));
    
    // Находим нужную вкладку и активируем её
    const targetNav = document.querySelector(`[data-tab="${tabName}"]`);
    const targetTab = document.getElementById(tabName);
    
    console.log(`📦 Найденные элементы:`, { targetNav, targetTab });
    console.log(`🔍 targetNav существует:`, !!targetNav);
    console.log(`🔍 targetTab существует:`, !!targetTab);
    
    if (targetNav && targetTab) {
        targetNav.classList.add('active');
        targetTab.classList.add('active');
        
        // Обновляем заголовок
        const navText = targetNav.querySelector('span').textContent;
        currentPage.textContent = navText;
        
        console.log(`✅ Вкладка ${tabName} активирована`);
        
        // Проверяем контейнер перед загрузкой данных
        if (tabName === 'products') {
            const container = document.getElementById('products-table');
            console.log('🔍 Контейнер products-table при активации вкладки:', container);
            console.log('🔍 Контейнер существует:', !!container);
            console.log('🔍 Контейнер innerHTML доступен:', container ? typeof container.innerHTML : 'N/A');
        }
        
        // Загружаем данные для активной вкладки (всегда, даже если уже активна)
        setTimeout(() => {
            console.log(`🔄 Загружаем данные для вкладки: ${tabName}`);
            switch (tabName) {
                case 'dashboard':
                    loadDashboard(userRole);
                    break;
                case 'products':
                    // Увеличиваем задержку для товаров, чтобы DOM точно был готов
                    setTimeout(() => {
                        console.log('🔄 Загружаем товары с дополнительной задержкой...');
                        loadProducts();
                        // Загружаем категории для заполнения селекта
                        loadCategories();
                    }, 200);
                    break;
                case 'categories':
                    loadCategories();
                    break;
                        case 'users':
            if (userRole === 'super_admin' || userRole === 'admin') {
                loadUsers();
            } else if (userRole === 'shop_owner') {
                loadShopCustomers();
            }
            break;
        case 'roles':
            if (userRole === 'super_admin') {
                loadRoles();
            }
            break;
        case 'orders':
            if (userRole === 'super_admin' || userRole === 'admin') {
                loadOrders();
            } else if (userRole === 'shop_owner') {
                loadShopOrders();
            }
            break;
                case 'settings':
                    loadSettings();
                    break;
                default:
                    console.warn(`⚠️ Неизвестная вкладка: ${tabName}`);
            }
        }, 150); // Увеличиваем задержку для стабильности
    } else {
        console.error(`❌ Не удалось найти элементы для вкладки ${tabName}`);
    }
}

// Загрузка дашборда
async function loadDashboard(userRole = 'admin') {
    updateLastActivity(); // Обновляем время активности
    try {
        console.log('🔄 Загружаем данные дашборда...');
        console.log('🔑 Токен админа:', adminToken ? 'Присутствует' : 'Отсутствует');
        console.log('👑 Роль пользователя:', userRole);
        
        // Проверяем наличие токена
        if (!adminToken) {
            console.warn('⚠️ Токен админа отсутствует, откладываем загрузку дашборда...');
            setTimeout(() => loadDashboard(userRole), 500);
        return;
    }
    
        // Загружаем данные с обработкой ошибок для каждого запроса
        let products = { data: [] };
        let users = { data: { users: [] } };
        let orders = { data: { orders: [] } };
        
        try {
                    // Выбираем правильный эндпоинт для товаров в зависимости от роли
        let productsEndpoint;
        // Все роли используют общий эндпоинт, но товары фильтруются по ownerId на сервере
        productsEndpoint = CONFIG.API.ENDPOINTS.PRODUCTS.LIST;
        if (userRole === 'super_admin' || userRole === 'admin') {
            console.log('👑 Админ загружает все товары для дашборда');
        } else {
            console.log('🏪 Владелец магазина загружает свои товары для дашборда (фильтруется на сервере)');
        }
            
            products = await fetchData(productsEndpoint);
            console.log('✅ Товары загружены:', products.products?.length || 0);
        } catch (error) {
            console.warn('⚠️ Ошибка загрузки товаров:', error.message);
        }
        
        // Проверяем, является ли userRole объектом или строкой
        const roleName = typeof userRole === 'object' ? userRole.name : userRole;
        console.log('🔍 Используем роль в loadDashboard:', roleName);
        
        try {
            if (roleName === 'super_admin' || roleName === 'admin') {
                users = await fetchData(CONFIG.API.ENDPOINTS.USERS.LIST);
                console.log('✅ Пользователи загружены:', users.data?.users?.length || 0);
            } else if (roleName === 'shop_owner') {
                users = await fetchData('/api/v1/shop/customers/');
                console.log('✅ Клиенты загружены:', users.data?.customers?.length || 0);
            }
        } catch (error) {
            console.warn('⚠️ Ошибка загрузки пользователей/клиентов:', error.message);
        }
        
        try {
            if (roleName === 'super_admin' || roleName === 'admin') {
                orders = await fetchData(CONFIG.API.ENDPOINTS.ORDERS.LIST);
                console.log('✅ Заказы загружены:', orders.data?.orders?.length || 0);
            } else if (roleName === 'shop_owner') {
                orders = await fetchData('/api/v1/shop/orders/');
                console.log('✅ Заказы магазина загружены:', orders.data?.orders?.length || 0);
            }
        } catch (error) {
            console.warn('⚠️ Ошибка загрузки заказов:', error.message);
        }
        
        console.log('📊 Данные получены:', {
            products: products.products?.length || 0,
            users: (roleName === 'super_admin' || roleName === 'admin') ? (users.data?.users?.length || 0) : (users.data?.customers?.length || 0),
            orders: orders.data?.orders?.length || 0
        });
        
        // Обновляем статистику с красивым форматированием
        const totalProducts = products.products?.length || 0;
        const totalUsers = (roleName === 'super_admin' || roleName === 'admin') ? (users.data?.users?.length || 0) : (users.data?.customers?.length || 0);
        const totalOrders = orders.data?.orders?.length || 0;
        const revenue = orders.data?.orders?.reduce((sum, order) => sum + (order.total_amount || 0), 0) || 0;
        
        // Анимация счетчиков
        console.log('🎯 Обновляем счетчики:', {
            products: totalProducts,
            users: totalUsers,
            orders: totalOrders,
            revenue: revenue
        });
        
        animateCounter('total-products', totalProducts);
        animateCounter('total-users', totalUsers);
        animateCounter('total-orders', totalOrders);
        animateRevenue('total-revenue', revenue);
        
        console.log('📊 Статистика обновлена:', {
            products: totalProducts,
            users: totalUsers,
            orders: totalOrders,
            revenue: revenue
        });
        
        // Показываем последние заказы
        displayRecentOrders(orders.data?.orders?.slice(0, 5) || []);
        
        console.log('✅ Дашборд загружен успешно');
        
    } catch (error) {
        console.error('❌ Ошибка загрузки дашборда:', error);
        showMessage('Ошибка загрузки данных дашборда: ' + error.message, 'error');
    }
}

// Анимация счетчиков
function animateCounter(elementId, targetValue) {
    const element = document.getElementById(elementId);
    if (!element) {
        console.warn(`⚠️ Элемент ${elementId} не найден для анимации счетчика`);
        return;
    }
    
    console.log(`🎯 Анимируем счетчик ${elementId}: ${targetValue}`);
    
    const startValue = 0;
    const duration = 1000; // 1 секунда
    const startTime = performance.now();
    
    function updateCounter(currentTime) {
        const elapsed = currentTime - startTime;
        const progress = Math.min(elapsed / duration, 1);
        
        // Плавная анимация с easeOutQuart
        const easeProgress = 1 - Math.pow(1 - progress, 4);
        const currentValue = Math.floor(startValue + (targetValue - startValue) * easeProgress);
        
        element.textContent = currentValue.toLocaleString();
        
        if (progress < 1) {
            requestAnimationFrame(updateCounter);
            } else {
            element.textContent = targetValue.toLocaleString();
        }
    }
    
    requestAnimationFrame(updateCounter);
}

// Анимация дохода
function animateRevenue(elementId, targetValue) {
    const element = document.getElementById(elementId);
    if (!element) return;
    
    const startValue = 0;
    const duration = 1000; // 1 секунда
    const startTime = performance.now();
    
    function updateRevenue(currentTime) {
        const elapsed = currentTime - startTime;
        const progress = Math.min(elapsed / duration, 1);
        
        // Плавная анимация с easeOutQuart
        const easeProgress = 1 - Math.pow(1 - progress, 4);
        const currentValue = Math.floor(startValue + (targetValue - startValue) * easeProgress);
        
        element.textContent = `₽${currentValue.toLocaleString()}`;
        
        if (progress < 1) {
            requestAnimationFrame(updateRevenue);
            } else {
            element.textContent = `₽${targetValue.toLocaleString()}`;
        }
    }
    
    requestAnimationFrame(updateRevenue);
}

// Отображение последних заказов
function displayRecentOrders(orders) {
    const container = document.getElementById('recent-orders');
    
    if (!container) {
        console.warn('⚠️ Контейнер recent-orders не найден');
        return;
    }
    
    if (orders.length === 0) {
        container.innerHTML = `
            <div class="empty-state">
                <i class="fas fa-shopping-cart"></i>
                <p>Заказов пока нет</p>
                <small>Когда появятся заказы, они отобразятся здесь</small>
            </div>
        `;
        return;
    }
    
    const table = `
        <table>
            <thead>
                <tr>
                    <th><i class="fas fa-hashtag"></i> ID</th>
                    <th><i class="fas fa-user"></i> Пользователь</th>
                    <th><i class="fas fa-info-circle"></i> Статус</th>
                    <th><i class="fas fa-ruble-sign"></i> Сумма</th>
                    <th><i class="fas fa-calendar"></i> Дата</th>
                    <th><i class="fas fa-cog"></i> Действия</th>
                </tr>
            </thead>
            <tbody>
                ${orders.map(order => `
                    <tr>
                        <td><code>${order.id?.substring(0, 8)}...</code></td>
                        <td>
                            <div class="user-info">
                                <i class="fas fa-user-circle"></i>
                                <span>${order.user_id?.substring(0, 8)}...</span>
                            </div>
                        </td>
                        <td>
                            <span class="status-badge ${getStatusClass(order.status)}">
                                <i class="fas ${getStatusIcon(order.status)}"></i>
                                ${order.status || 'Новый'}
                            </span>
                        </td>
                        <td class="amount">
                            <strong>₽${(order.total_amount || 0).toLocaleString()}</strong>
                        </td>
                        <td>
                            <div class="date-info">
                                <div class="date">${new Date(order.created_at).toLocaleDateString()}</div>
                                <div class="time">${new Date(order.created_at).toLocaleTimeString()}</div>
                            </div>
                        </td>
                        <td>
                            <button class="action-btn view" onclick="viewOrder('${order.id}')" title="Просмотр">
                                <i class="fas fa-eye"></i>
                            </button>
                        </td>
                    </tr>
                `).join('')}
            </tbody>
        </table>
    `;
    
    container.innerHTML = table;
}

// Получение класса статуса
function getStatusClass(status) {
    switch (status?.toLowerCase()) {
        case 'completed':
        case 'завершен':
            return 'status-completed';
        case 'processing':
        case 'обработка':
            return 'status-processing';
        case 'shipped':
        case 'отправлен':
            return 'status-shipped';
        case 'cancelled':
        case 'отменен':
            return 'status-cancelled';
        default:
            return 'status-new';
    }
}

// Получение иконки статуса
function getStatusIcon(status) {
    switch (status?.toLowerCase()) {
        case 'completed':
        case 'завершен':
            return 'fa-check-circle';
        case 'processing':
        case 'обработка':
            return 'fa-clock';
        case 'shipped':
        case 'отправлен':
            return 'fa-shipping-fast';
        case 'cancelled':
        case 'отменен':
            return 'fa-times-circle';
        default:
            return 'fa-circle';
    }
}

// Загрузка товаров с учетом роли пользователя
async function loadProducts() {
    console.log('🔄 Начинаем загрузку товаров...');
    
    // Объявляем переменную container в начале функции
    const container = document.getElementById('products-table');
    
    try {
        // Показываем индикатор загрузки
        if (container) {
            container.innerHTML = `
                <div class="table-container">
                    <h3><i class="fas fa-box"></i> Список товаров</h3>
                    <div class="text-center" style="padding: 40px 20px;">
                        <div class="loading" style="margin: 0 auto 20px;">
                            <i class="fas fa-spinner fa-spin" style="font-size: 32px; color: #667eea;"></i>
                        </div>
                        <p style="color: #666; font-size: 14px;">Загрузка товаров...</p>
                    </div>
                </div>
            `;
        }
        
        // Определяем роль пользователя и выбираем правильный эндпоинт
        const userRole = localStorage.getItem('userRole') || 'admin';
        let endpoint;
        let title = 'Список товаров';
        
        if (userRole === 'super_admin' || userRole === 'admin') {
            // Админы видят все товары
            endpoint = CONFIG.API.ENDPOINTS.PRODUCTS.LIST;
            title = 'Список всех товаров (Админ)';
            console.log('👑 Админ загружает все товары');
        } else {
            // Владельцы магазинов и обычные пользователи используют общий эндпоинт
            endpoint = CONFIG.API.ENDPOINTS.PRODUCTS.LIST;
            title = 'Мои товары';
            console.log('🏪 Владелец магазина загружает все товары, фильтруем по ownerId');
        }
        
        console.log(`🔗 Используем эндпоинт: ${endpoint}`);
        
        const response = await fetchData(endpoint);
        console.log('📡 Ответ API товаров:', response);
        
        // Проверяем разные возможные форматы ответа
        let products = [];
        console.log('🔍 Анализируем ответ API:', response);
        
        if (response.data && response.data.products && Array.isArray(response.data.products)) {
            products = response.data.products;
            console.log('✅ Используем response.data.products');
        } else if (response.products && Array.isArray(response.products)) {
            products = response.products;
            console.log('✅ Используем response.products');
        } else if (response.data && Array.isArray(response.data)) {
            products = response.data;
            console.log('✅ Используем response.data');
        } else if (Array.isArray(response)) {
            products = response;
            console.log('✅ Используем response напрямую');
        } else {
            console.warn('⚠️ Неожиданный формат ответа:', response);
            products = [];
        }
        
        console.log(`📦 Получено ${products.length} товаров из API`);
        
        // Фильтруем товары по ownerId для владельцев магазинов
        if (userRole === 'shop_owner' || userRole === 'user') {
            const userData = JSON.parse(localStorage.getItem('userData'));
            const userId = userData?.id;
            
            if (userId) {
                console.log(`🔍 Фильтруем товары по ownerId: ${userId}`);
                const originalCount = products.length;
                products = products.filter(product => product.ownerId === userId);
                console.log(`✅ Отфильтровано: ${originalCount} → ${products.length} товаров`);
            } else {
                console.warn('⚠️ Не удалось получить ID пользователя для фильтрации');
            }
        }
        
        console.log(`📦 Получено ${products.length} товаров для роли ${userRole}`);
        
        // Проверяем, есть ли данные категорий
        products.forEach((product, index) => {
            console.log(`📦 Товар ${index + 1}: "${product.name}" - categoryId: ${product.categoryId}, category:`, product.category);
            console.log(`📦 Товар ${index + 1} вариации:`, product.variations);
            console.log(`📦 Товар ${index + 1} вариации тип:`, typeof product.variations);
            console.log(`📦 Товар ${index + 1} вариации длина:`, product.variations?.length);
            
            // Для владельцев магазинов проверяем ownerId
            if (userRole === 'shop_owner' || userRole === 'user') {
                console.log(`📦 Товар ${index + 1} ownerId:`, product.ownerId);
            }
        });
        
        // Сохраняем товары в глобальную переменную для фильтрации
        allProducts = products;
        
        // Проверяем, что контейнер существует перед отображением
        if (!container) {
            console.error('❌ Контейнер products-table не найден при загрузке товаров!');
            
            // Проверяем, активна ли вкладка товаров
            const productsTab = document.getElementById('products');
            if (!productsTab || !productsTab.classList.contains('active')) {
                console.warn('⚠️ Вкладка товаров не активна, переключаемся...');
                showTab('products');
                return;
            }
            
            console.error('❌ Контейнер не найден даже на активной вкладке товаров!');
            return;
        }
        
        // Обновляем заголовок в зависимости от роли
        if (container) {
            const titleElement = container.querySelector('h3');
            if (titleElement) {
                titleElement.innerHTML = `<i class="fas fa-box"></i> ${title}`;
            }
        }
        
        displayProducts(products);
        
        // Показываем уведомление об успешной загрузке
        if (products.length > 0) {
            const roleText = userRole === 'shop_owner' ? 'ваших' : '';
            showMessage(`Успешно загружено ${products.length} ${roleText} товаров`, 'success');
        }
        
    } catch (error) {
        console.error('❌ Ошибка загрузки товаров:', error);
        
        // Показываем ошибку в контейнере
        if (container) {
            container.innerHTML = `
                <div class="table-container">
                    <h3><i class="fas fa-box"></i> Список товаров</h3>
                    <div class="text-center" style="padding: 40px 20px;">
                        <div style="font-size: 48px; color: #e74c3c; margin-bottom: 20px;">
                            <i class="fas fa-exclamation-triangle"></i>
                        </div>
                        <h4 style="color: #e74c3c; margin-bottom: 10px;">Ошибка загрузки</h4>
                        <p style="color: #666; font-size: 14px; margin-bottom: 20px;">${error.message}</p>
                        <button class="btn btn-primary" onclick="loadProducts()">
                            <i class="fas fa-redo"></i> Попробовать снова
                        </button>
                    </div>
                </div>
            `;
        }
        
        showMessage('Ошибка загрузки товаров: ' + error.message, 'error');
    }
}

// Принудительное обновление списка товаров (для использования после добавления/изменения)
async function refreshProductsList() {
    console.log('🔄 Принудительное обновление списка товаров...');
    
    // Проверяем, что мы на вкладке товаров
    const productsTab = document.getElementById('products');
    if (!productsTab || !productsTab.classList.contains('active')) {
        console.warn('⚠️ Вкладка товаров не активна, переключаемся...');
        showTab('products');
        // Даем время на переключение
        await new Promise(resolve => setTimeout(resolve, 150));
    }
    
    try {
        await loadProducts();
        console.log('✅ Список товаров успешно обновлен');
        return true;
    } catch (error) {
        console.error('❌ Ошибка обновления списка товаров:', error);
        showMessage('Не удалось обновить список товаров: ' + error.message, 'error');
        return false;
    }
}

// Отображение товаров
function displayProducts(products) {
    console.log('🔄 Отображение товаров:', products);
    
    const container = document.getElementById('products-table');
    
    if (!container) {
        console.error('❌ Контейнер products-table не найден!');
        showMessage('Ошибка: контейнер товаров не найден. Попробуйте обновить страницу.', 'error');
        return;
    }
    
    if (!Array.isArray(products)) {
        console.warn('⚠️ products не является массивом:', products);
        products = [];
    }
    
    if (products.length === 0) {
        container.innerHTML = `
            <div class="table-container">
                <h3><i class="fas fa-box"></i> Список товаров</h3>
                <div class="text-center" style="padding: 60px 20px;">
                    <div style="font-size: 64px; color: #ddd; margin-bottom: 20px;">
                        <i class="fas fa-box-open"></i>
                    </div>
                    <h4 style="color: #666; margin-bottom: 10px;">Товаров пока нет</h4>
                    <p style="color: #999; font-size: 14px;">Добавьте первый товар, чтобы начать работу</p>
                </div>
            </div>
        `;
        return;
    }
    
    console.log(`📊 Отображаем ${products.length} товаров`);
    
    // Отладочная информация для категорий
    products.forEach((product, index) => {
        console.log(`📦 Товар ${index + 1} категория:`, product.category);
        if (product.category) {
            console.log(`📦 Товар ${index + 1} категория детали:`, {
                id: product.category.id,
                name: product.category.name,
                Name: product.category.Name,
                displayName: product.category.displayName,
                keys: Object.keys(product.category)
            });
        }
    });
    
    try {
        const table = `
            <div class="table-container">
                <h3><i class="fas fa-box"></i> Список товаров</h3>
                <div class="table-responsive">
                    <table class="data-table">
                        <thead>
                            <tr>
                                <th><i class="fas fa-tag"></i> Товар</th>
                                <th><i class="fas fa-building"></i> Бренд</th>
                                <th><i class="fas fa-venus-mars"></i> Пол</th>
                                <th><i class="fas fa-folder"></i> Категория</th>
                                <th><i class="fas fa-layer-group"></i> Вариации</th>
                                <th><i class="fas fa-calendar"></i> Дата создания</th>
                                <th><i class="fas fa-cogs"></i> Действия</th>
                            </tr>
                        </thead>
                        <tbody>
                            ${products.map((product, index) => `
                                <tr style="animation-delay: ${index * 0.1}s;">
                                    <td data-label="Товар">
                                        <div style="display: flex; align-items: center; gap: 12px;">
                                            <div style="width: 50px; height: 50px; border-radius: 12px; background: linear-gradient(135deg, #667eea, #764ba2); display: flex; align-items: center; justify-content: center; color: white; font-weight: bold; font-size: 20px; box-shadow: 0 4px 15px rgba(102, 126, 234, 0.3); position: relative; overflow: hidden;">
                                                <i class="fas fa-box"></i>
                                                <div style="position: absolute; top: 0; left: 0; right: 0; bottom: 0; background: linear-gradient(45deg, transparent 30%, rgba(255,255,255,0.1) 50%, transparent 70%); animation: shine 2s infinite;"></div>
                                            </div>
                                            <div>
                                                <div style="font-weight: 700; color: #333; font-size: 16px; margin-bottom: 4px;">${product.name}</div>
                                                <div style="font-size: 12px; color: #888; font-family: monospace; background: #f8f9fa; padding: 2px 6px; border-radius: 4px; display: inline-block;">
                                                    ${product.id?.substring(0, 8)}...
                                                </div>
                                            </div>
                                        </div>
                                    </td>
                                    <td data-label="Бренд">
                                        <div style="display: flex; align-items: center; gap: 8px;">
                                            <i class="fas fa-building" style="color: #4ecdc4; font-size: 16px;"></i>
                                            <span style="font-weight: 600; color: #2c3e50;">${product.brand || 'Не указан'}</span>
                                        </div>
                                    </td>
                                    <td data-label="Пол">
                                        <span class="badge" style="background: ${getGenderColor(product.gender)}; font-size: 12px; padding: 8px 12px;">
                                            <i class="fas ${getGenderIcon(product.gender)}" style="margin-right: 4px;"></i>
                                            ${getGenderText(product.gender)}
                                        </span>
                                    </td>
                                    <td data-label="Категория">
                                        <div style="display: flex; align-items: center; gap: 8px;">
                                            <i class="fas fa-folder" style="color: #f093fb; font-size: 16px;"></i>
                                            <span style="font-weight: 600; color: #2c3e50;">${product.category?.name || 'Без категории'}</span>
                                        </div>
                                    </td>
                                    <td data-label="Вариации">
                                        <div style="display: flex; align-items: center; gap: 8px;">
                                            <i class="fas fa-layer-group" style="color: #45b7d1; font-size: 16px;"></i>
                                            <span class="badge" style="background: linear-gradient(135deg, #45b7d1, #96ceb4); color: white; font-size: 12px; padding: 8px 12px;">
                                                ${product.variations?.length || 0} ${product.variations?.length === 1 ? 'вариация' : product.variations?.length < 5 ? 'вариации' : 'вариаций'}
                                            </span>
                                        </div>
                                    </td>
                                    <td data-label="Дата создания">
                                        <div style="display: flex; align-items: center; gap: 8px;">
                                            <i class="fas fa-calendar" style="color: #f093fb; font-size: 16px;"></i>
                                            <div>
                                                <div style="font-size: 14px; color: #2c3e50; font-weight: 600;">
                                                    ${product.createdAt ? new Date(product.createdAt).toLocaleDateString('ru-RU', {
                                                        day: '2-digit',
                                                        month: '2-digit',
                                                        year: 'numeric'
                                                    }) : 'N/A'}
                                                </div>
                                                <div style="font-size: 11px; color: #7f8c8d;">
                                                    ${product.createdAt ? new Date(product.createdAt).toLocaleTimeString('ru-RU', {
                                                        hour: '2-digit',
                                                        minute: '2-digit'
                                                    }) : ''}
                                                </div>
                                            </div>
                                        </div>
                                    </td>
                                    <td data-label="Действия">
                                        <div style="display: flex; gap: 8px; justify-content: center;">
                                            <button class="btn-sm btn-info" onclick="viewProductVariations('${product.id}')" title="Просмотр вариаций" style="min-width: 44px; min-height: 44px;">
                                                <i class="fas fa-eye"></i>
                                            </button>
                                            <button class="btn-sm btn-primary" onclick="editProduct('${product.id}')" title="Редактировать товар" style="min-width: 44px; min-height: 44px;">
                                                <i class="fas fa-edit"></i>
                                            </button>
                                            <button class="btn-sm btn-danger" onclick="deleteProduct('${product.id}')" title="Удалить товар" style="min-width: 44px; min-height: 44px;">
                                                <i class="fas fa-trash"></i>
                                            </button>
                                        </div>
                                    </td>
                                </tr>
                            `).join('')}
                        </tbody>
                    </table>
                </div>
            </div>
        `;
        
        container.innerHTML = table;
        
        console.log('✅ Таблица товаров обновлена');
    } catch (error) {
        console.error('❌ Ошибка при установке innerHTML:', error);
        showMessage('Ошибка при обновлении списка товаров: ' + error.message, 'error');
    }
}

// Загрузка категорий
async function loadCategories() {
    try {
        const response = await fetchData(CONFIG.API.ENDPOINTS.CATEGORIES.LIST);
        console.log('📡 Ответ API категорий:', response);
        
        // Проверяем разные возможные форматы ответа
        let categories = [];
        if (response.success && response.data) {
            categories = response.data;
        } else if (response.categories) {
            categories = response.categories;
        } else if (response.data && Array.isArray(response.data)) {
            categories = response.data;
        } else if (Array.isArray(response)) {
            categories = response;
    } else {
            console.warn('⚠️ Неожиданный формат ответа категорий:', response);
            categories = [];
        }
        
        console.log(`📦 Получено ${categories.length} категорий:`, categories);
        
        displayCategories(categories);
        console.log('🔄 Вызываем populateCategorySelects с категориями:', categories);
        populateCategorySelects(categories);
        
        console.log('✅ Категории загружены и селекты обновлены');
    } catch (error) {
        console.error('Ошибка загрузки категорий:', error);
        showMessage('Ошибка загрузки категорий', 'error');
    }
}

// Отображение категорий
function displayCategories(categories) {
    const container = document.getElementById('categories-table');
    
    if (!container) {
        console.warn('Контейнер categories-table не найден');
        return;
    }
    
    if (categories.length === 0) {
        container.innerHTML = `
            <div class="table-container">
                <h3><i class="fas fa-tags"></i> Список категорий</h3>
                <div class="text-center" style="padding: 60px 20px;">
                    <div style="font-size: 64px; color: #ddd; margin-bottom: 20px;">
                        <i class="fas fa-tags"></i>
                    </div>
                    <h4 style="color: #666; margin-bottom: 10px;">Категорий пока нет</h4>
                    <p style="color: #999; font-size: 14px;">Добавьте первую категорию, чтобы начать работу</p>
                </div>
            </div>
        `;
        return;
    }
    
    const table = `
        <div class="table-container">
            <h3><i class="fas fa-tags"></i> Список категорий</h3>
            <div class="table-responsive">
                <table class="data-table">
                    <thead>
                        <tr>
                            <th><i class="fas fa-tag"></i> Категория</th>
                            <th><i class="fas fa-info-circle"></i> Описание</th>
                            <th><i class="fas fa-sitemap"></i> Родительская</th>
                            <th><i class="fas fa-sort"></i> Порядок</th>
                            <th><i class="fas fa-calendar"></i> Дата создания</th>
                            <th><i class="fas fa-cogs"></i> Действия</th>
                        </tr>
                    </thead>
                    <tbody>
                        ${categories.map((category, index) => `
                            <tr style="animation-delay: ${index * 0.1}s;">
                                <td data-label="Категория">
                                    <div style="display: flex; align-items: center; gap: 12px;">
                                        <div style="width: 50px; height: 50px; border-radius: 12px; background: linear-gradient(135deg, #f093fb, #f5576c); display: flex; align-items: center; justify-content: center; color: white; font-weight: bold; font-size: 20px; box-shadow: 0 4px 15px rgba(240, 147, 251, 0.3); position: relative; overflow: hidden;">
                                            <i class="fas fa-tag"></i>
                                            <div style="position: absolute; top: 0; left: 0; right: 0; bottom: 0; background: linear-gradient(45deg, transparent 30%, rgba(255,255,255,0.1) 50%, transparent 70%); animation: shine 2s infinite;"></div>
                                        </div>
                                        <div>
                                            <div style="font-weight: 700; color: #333; font-size: 16px; margin-bottom: 4px;">${category.name}</div>
                                            <div style="font-size: 12px; color: #888; font-family: monospace; background: #f8f9fa; padding: 2px 6px; border-radius: 4px; display: inline-block;">
                                                ${category.id?.substring(0, 8)}...
                                            </div>
                                        </div>
                                    </div>
                                </td>
                                <td data-label="Описание">
                                    <div style="display: flex; align-items: center; gap: 8px;">
                                        <i class="fas fa-info-circle" style="color: #4ecdc4; font-size: 16px;"></i>
                                        <span style="font-weight: 600; color: #2c3e50; max-width: 200px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;" title="${category.description || 'Нет описания'}">
                                            ${category.description || 'Нет описания'}
                                        </span>
                                    </div>
                                </td>
                                <td data-label="Родительская">
                                    <div style="display: flex; align-items: center; gap: 8px;">
                                        <i class="fas fa-sitemap" style="color: #45b7d1; font-size: 16px;"></i>
                                        <span style="font-weight: 600; color: #2c3e50;">${category.parent?.name || 'Корневая'}</span>
                                    </div>
                                </td>
                                <td data-label="Порядок">
                                    <div style="display: flex; align-items: center; gap: 8px;">
                                        <i class="fas fa-sort" style="color: #f093fb; font-size: 16px;"></i>
                                        <span class="badge" style="background: linear-gradient(135deg, #45b7d1, #96ceb4); color: white; font-size: 12px; padding: 8px 12px;">
                                            ${category.sortOrder || 0}
                                        </span>
                                    </div>
                                </td>
                                <td data-label="Дата создания">
                                    <div style="display: flex; align-items: center; gap: 8px;">
                                        <i class="fas fa-calendar" style="color: #f093fb; font-size: 16px;"></i>
                                        <div>
                                            <div style="font-size: 14px; color: #2c3e50; font-weight: 600;">
                                                ${category.createdAt ? new Date(category.createdAt).toLocaleDateString('ru-RU', {
                                                    day: '2-digit',
                                                    month: '2-digit',
                                                    year: 'numeric'
                                                }) : 'N/A'}
                                            </div>
                                            <div style="font-size: 11px; color: #7f8c8d;">
                                                ${category.createdAt ? new Date(category.createdAt).toLocaleTimeString('ru-RU', {
                                                    hour: '2-digit',
                                                    minute: '2-digit'
                                                }) : ''}
                                            </div>
                                        </div>
                                    </div>
                                </td>
                                <td data-label="Действия">
                                    <div style="display: flex; gap: 8px; justify-content: center;">
                                        <button class="btn-sm btn-primary" onclick="editCategory('${category.id}')" title="Редактировать категорию" style="min-width: 44px; min-height: 44px;">
                                            <i class="fas fa-edit"></i>
                                        </button>
                                        <button class="btn-sm btn-danger" onclick="deleteCategory('${category.id}')" title="Удалить категорию" style="min-width: 44px; min-height: 44px;">
                                            <i class="fas fa-trash"></i>
                                        </button>
                                    </div>
                                </td>
                            </tr>
                        `).join('')}
                    </tbody>
                </table>
            </div>
        </div>
    `;
    
    container.innerHTML = table;
    
    // Добавляем data-label атрибуты для мобильных устройств
    
}

// Загрузка пользователей
async function loadUsers() {
    console.log('🔄 Загружаем пользователей...');
    try {
        const response = await fetchData('/api/v1/admin/users/');
        console.log('📡 Ответ API для пользователей:', response);
        if (response.success) {
            console.log('✅ Пользователи загружены успешно:', response.data.users);
            displayUsers(response.data.users);
        } else {
            console.error('❌ Ошибка в ответе API для пользователей:', response);
                }
            } catch (error) {
        console.error('❌ Ошибка загрузки пользователей:', error);
        showMessage('Ошибка загрузки пользователей', 'error');
    }
}

// Отображение пользователей
function displayUsers(users) {
    console.log('🔍 displayUsers вызвана с данными:', users);
    
    const tbody = document.getElementById('users-table-body');
    console.log('🔍 Найден tbody:', tbody);
    
    if (!tbody) {
        console.error('❌ Элемент users-table-body не найден!');
        return;
    }

    tbody.innerHTML = '';
    console.log('🔍 Очистили tbody');
    
    if (!users || users.length === 0) {
        console.log('⚠️ Нет пользователей для отображения');
        tbody.innerHTML = `
            <tr>
                <td colspan="7" class="text-center">
                    <div style="padding: 40px 20px;">
                        <i class="fas fa-users" style="font-size: 48px; color: #ccc; margin-bottom: 15px;"></i>
                        <div style="font-size: 18px; color: #666; margin-bottom: 10px;">Пользователей не найдено</div>
                        <div style="font-size: 14px; color: #999;">Создайте первого пользователя, нажав кнопку "Добавить пользователя"</div>
                    </div>
                </td>
            </tr>
        `;
        return;
    }
    
    users.forEach((user, index) => {
        console.log(`🔍 Обрабатываем пользователя ${index + 1}:`, user);
        
        const row = document.createElement('tr');
        row.style.animationDelay = `${index * 0.1}s`;
        
        const rowHtml = `
            <td>
                <div style="display: flex; align-items: center; gap: 12px;">
                    <div style="width: 45px; height: 45px; border-radius: 50%; background: linear-gradient(135deg, #667eea, #764ba2); display: flex; align-items: center; justify-content: center; color: white; font-weight: bold; font-size: 18px; box-shadow: 0 4px 15px rgba(102, 126, 234, 0.3);">
                        ${user.name ? user.name.charAt(0).toUpperCase() : 'U'}
                    </div>
                    <div>
                        <div style="font-weight: 700; color: #333; font-size: 15px;">${user.name || 'Не указано'}</div>
                        <div style="font-size: 11px; color: #888; font-family: monospace;">${user.id ? user.id.substring(0, 8) + '...' : 'N/A'}</div>
                    </div>
                </div>
            </td>
            <td>
                <div style="display: flex; align-items: center; gap: 10px;">
                    <i class="fas fa-envelope" style="color: #667eea; font-size: 16px;"></i>
                    <span style="font-weight: 500;">${user.email || 'N/A'}</span>
                </div>
            </td>
            <td>
                <div style="display: flex; align-items: center; gap: 10px;">
                    <i class="fas fa-phone" style="color: #4ecdc4; font-size: 16px;"></i>
                    <span style="font-weight: 500;">${user.phone || 'Не указан'}</span>
                </div>
            </td>
            <td>
                <span class="badge role-${user.role?.name || 'user'}">
                    <i class="fas fa-user-shield"></i>
                    ${user.role?.displayName || 'Пользователь'}
                </span>
            </td>
            <td>
                <span class="badge ${user.isActive ? 'role-user' : 'role-admin'}" style="background: ${user.isActive ? 'linear-gradient(135deg, #4ecdc4, #44a08d)' : 'linear-gradient(135deg, #ff6b6b, #ee5a24)'};">
                    <i class="fas ${user.isActive ? 'fa-check-circle' : 'fa-times-circle'}"></i>
                    ${user.isActive ? 'Активен' : 'Неактивен'}
                </span>
            </td>
            <td>
                <div style="display: flex; align-items: center; gap: 8px;">
                    <i class="fas fa-calendar" style="color: #f093fb; font-size: 14px;"></i>
                    <span style="font-size: 13px; color: #666; font-weight: 500;">
                        ${user.createdAt ? new Date(user.createdAt).toLocaleDateString('ru-RU', {
                            day: '2-digit',
                            month: '2-digit',
                            year: 'numeric'
                        }) : 'N/A'}
                    </span>
                </div>
            </td>
            <td>
                <div style="display: flex; gap: 6px;">
                    <button class="btn-sm btn-primary" onclick="viewUser('${user.id}')" title="Просмотр">
                        <i class="fas fa-eye"></i>
                    </button>
                    <button class="btn-sm btn-secondary" onclick="editUser('${user.id}')" title="Редактировать">
                        <i class="fas fa-edit"></i>
                    </button>
                    <button class="btn-sm btn-danger" onclick="deleteUser('${user.id}')" title="Удалить">
                        <i class="fas fa-trash"></i>
                    </button>
                </div>
            </td>
        `;
        
        console.log(`🔍 HTML для строки ${index + 1}:`, rowHtml);
        row.innerHTML = rowHtml;
        tbody.appendChild(row);
        console.log(`✅ Добавлена строка для пользователя ${index + 1}`);
        console.log(`🔍 Количество строк в таблице:`, tbody.children.length);
    });
    
    console.log('✅ displayUsers завершена');
    
    // Дополнительная отладка
    console.log('🔍 Финальное состояние tbody:', tbody);
    console.log('🔍 Количество дочерних элементов:', tbody.children.length);
    console.log('🔍 HTML содержимое tbody:', tbody.innerHTML);
    
    // Проверяем видимость
    const computedStyle = window.getComputedStyle(tbody);
    console.log('🔍 CSS display:', computedStyle.display);
    console.log('🔍 CSS visibility:', computedStyle.visibility);
    console.log('🔍 CSS opacity:', computedStyle.opacity);
}

// Получение класса роли
function getRoleClass(role) {
    switch (role?.toLowerCase()) {
        case 'admin':
            return 'role-admin';
        case 'moderator':
            return 'role-moderator';
        default:
            return 'role-user';
    }
}

// Получение иконки роли
function getRoleIcon(role) {
    switch (role?.toLowerCase()) {
        case 'admin':
            return 'fa-crown';
        case 'moderator':
            return 'fa-user-shield';
        default:
            return 'fa-user';
    }
}

// Функции для работы с полом товара
function getGenderColor(gender) {
    switch (gender?.toLowerCase()) {
        case 'male':
            return 'linear-gradient(135deg, #4ecdc4, #44a08d)';
        case 'female':
            return 'linear-gradient(135deg, #f093fb, #f5576c)';
        default:
            return 'linear-gradient(135deg, #45b7d1, #96ceb4)';
    }
}

function getGenderIcon(gender) {
    switch (gender?.toLowerCase()) {
        case 'male':
            return 'fa-mars';
        case 'female':
            return 'fa-venus';
        default:
            return 'fa-venus-mars';
    }
}

function getGenderText(gender) {
    switch (gender?.toLowerCase()) {
        case 'male':
            return 'Мужской';
        case 'female':
            return 'Женский';
        default:
            return 'Унисекс';
    }
}

// Функции для работы с пользователями
async function viewUser(id) {
    try {
        const response = await fetchData(`/api/v1/admin/users/${id}`);
        const user = response.data;
        
        // Показываем информацию о пользователе в модальном окне
        const modal = `
            <div class="modal" style="display: block;">
                <div class="modal-content" style="max-width: 600px;">
                    <div class="modal-header">
                        <h3><i class="fas fa-user"></i> Информация о пользователе</h3>
                        <span class="close" onclick="this.parentElement.parentElement.parentElement.remove()">&times;</span>
                    </div>
                    <div class="user-details-modal">
                        <div class="user-avatar-large">
                            ${user.avatar ? 
                                `<img src="${user.avatar}" alt="${user.name}">` : 
                                `<i class="fas fa-user-circle"></i>`
                            }
                        </div>
                        <div class="user-info-grid">
                            <div class="info-item">
                                <label>Имя:</label>
                                <span>${user.name || 'Не указано'}</span>
                            </div>
                            <div class="info-item">
                                <label>Email:</label>
                                <span>${user.email}</span>
                            </div>
                            <div class="info-item">
                                <label>Телефон:</label>
                                <span>${user.phone || 'Не указан'}</span>
                            </div>
                            <div class="info-item">
                                <label>Роль:</label>
                                <span class="role-badge ${getRoleClass(user.role?.name || 'user')}">${user.role?.displayName || 'Пользователь'}</span>
                            </div>
                            <div class="info-item">
                                <label>Статус:</label>
                                <span class="status-badge ${user.isActive ? 'status-active' : 'status-inactive'}">
                                    ${user.isActive ? 'Активен' : 'Неактивен'}
                                </span>
                            </div>
                            <div class="info-item">
                                <label>Дата регистрации:</label>
                                <span>${new Date(user.created_at).toLocaleString()}</span>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        `;
        
        document.body.insertAdjacentHTML('beforeend', modal);
        
    } catch (error) {
        console.error('Ошибка загрузки данных пользователя:', error);
        showMessage('Ошибка загрузки данных пользователя', 'error');
    }
}

async function editUser(id) {
    try {
        const response = await fetchData(`/api/v1/admin/users/${id}`);
        const user = response.data;
        
        // Показываем форму редактирования
        const modal = `
            <div class="modal" style="display: block;">
                <div class="modal-content" style="max-width: 500px;">
                    <div class="modal-header">
                        <h3><i class="fas fa-edit"></i> Редактировать пользователя</h3>
                        <span class="close" onclick="this.parentElement.parentElement.parentElement.remove()">&times;</span>
                    </div>
                    <form id="edit-user-form">
                        <div class="form-group">
                            <label>Имя</label>
                            <input type="text" id="edit-user-name" value="${user.name || ''}" class="form-input">
                        </div>
                        <div class="form-group">
                            <label>Email</label>
                            <input type="email" id="edit-user-email" value="${user.email}" class="form-input" readonly>
                        </div>
                        <div class="form-group">
                            <label>Телефон</label>
                            <input type="tel" id="edit-user-phone" value="${user.phone || ''}" class="form-input">
                        </div>
                        <div class="form-group">
                            <label>Роль</label>
                            <select id="edit-user-role" class="form-input">
                                <option value="user" ${user.role === 'user' ? 'selected' : ''}>Пользователь</option>
                                <option value="moderator" ${user.role === 'moderator' ? 'selected' : ''}>Модератор</option>
                                <option value="admin" ${user.role === 'admin' ? 'selected' : ''}>Администратор</option>
                            </select>
                        </div>
                        <div class="form-group">
                            <label>
                                <input type="checkbox" id="edit-user-active" ${user.isActive ? 'checked' : ''}>
                                Активен
                            </label>
                        </div>
                        <div class="modal-actions">
                            <button type="submit" class="btn btn-primary">Сохранить</button>
                            <button type="button" class="btn btn-secondary" onclick="this.parentElement.parentElement.parentElement.parentElement.remove()">Отмена</button>
                        </div>
                    </form>
                </div>
            </div>
        `;
        
        document.body.insertAdjacentHTML('beforeend', modal);
        
        // Обработчик формы
        document.getElementById('edit-user-form').addEventListener('submit', async (e) => {
            e.preventDefault();
            
            const formData = {
                name: document.getElementById('edit-user-name').value,
                phone: document.getElementById('edit-user-phone').value,
                role: document.getElementById('edit-user-role').value,
                isActive: document.getElementById('edit-user-active').checked
            };
            
            try {
                await fetchData(`/api/v1/admin/users/${id}`, {
                    method: 'PUT',
                    body: JSON.stringify(formData)
                });
                
                showMessage('Пользователь успешно обновлен', 'success');
                document.querySelector('.modal').remove();
                loadUsers(); // Перезагружаем список
                
            } catch (error) {
                console.error('Ошибка обновления пользователя:', error);
                showMessage('Ошибка обновления пользователя', 'error');
            }
        });
        
    } catch (error) {
        console.error('Ошибка загрузки данных пользователя:', error);
        showMessage('Ошибка загрузки данных пользователя', 'error');
    }
}

async function deleteUser(id) {
    if (!confirm('Вы уверены, что хотите удалить этого пользователя? Это действие нельзя отменить.')) {
        return;
    }
    
    try {
        await fetchData(`/api/v1/admin/users/${id}`, {
            method: 'DELETE'
        });
        
        showMessage('Пользователь успешно удален', 'success');
        loadUsers(); // Перезагружаем список
        
    } catch (error) {
        console.error('Ошибка удаления пользователя:', error);
        showMessage('Ошибка удаления пользователя', 'error');
    }
}

// Загрузка заказов
// Глобальные переменные для заказов
let currentOrdersPage = 1;
let currentOrdersFilters = {};
let ordersStats = {};

async function loadOrders(page = 1, filters = {}) {
    try {
        currentOrdersPage = page;
        currentOrdersFilters = filters;
        
        // Определяем роль пользователя
        const userRole = localStorage.getItem('userRole') || 'admin';
        
        let endpoint;
        if (userRole === 'super_admin' || userRole === 'admin') {
            endpoint = '/api/v1/admin/orders';
        } else if (userRole === 'shop_owner') {
            endpoint = '/api/v1/shop/orders/';
        } else {
            endpoint = '/api/v1/admin/orders';
        }
        
        // Добавляем параметры фильтрации
        const params = new URLSearchParams({
            page: page,
            limit: 20,
            ...filters
        });
        
        const fullEndpoint = `${endpoint}?${params.toString()}`;
        console.log('📡 Загрузка заказов:', fullEndpoint);
        
        const response = await fetchData(fullEndpoint);
        
        if (response.data) {
            // Сохраняем список владельцев магазинов для фильтра
            if (response.data.shop_owners) {
                window.shopOwners = response.data.shop_owners;
            }
            displayOrders(response.data.orders || [], response.data.pagination, response.data.stats);
        } else {
            displayOrders([], {}, {});
        }
    } catch (error) {
        console.error('Ошибка загрузки заказов:', error);
        showMessage('Ошибка загрузки заказов', 'error');
    }
}

// Отображение заказов с расширенной информацией
function displayOrders(orders, pagination = {}, stats = {}) {
    const container = document.getElementById('orders-table');
    
    if (!container) {
        console.warn('Контейнер orders-table не найден');
        return;
    }
    
    // Сохраняем статистику
    ordersStats = stats;
    
    // Создаём фильтры и поиск
    const filtersHTML = `
        <div class="orders-filters" style="margin-bottom: 20px; padding: 20px; background: white; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
            <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 15px; margin-bottom: 15px;">
                <input type="text" id="order-search" placeholder="Поиск по имени, телефону, номеру..." 
                    style="padding: 10px; border: 1px solid #ddd; border-radius: 4px;" 
                    value="${currentOrdersFilters.search || ''}">
                <select id="order-status-filter" style="padding: 10px; border: 1px solid #ddd; border-radius: 4px;">
                    <option value="">Все статусы</option>
                    <option value="pending" ${currentOrdersFilters.status === 'pending' ? 'selected' : ''}>Ожидают (${stats.pending || 0})</option>
                    <option value="confirmed" ${currentOrdersFilters.status === 'confirmed' ? 'selected' : ''}>Подтверждены (${stats.confirmed || 0})</option>
                    <option value="preparing" ${currentOrdersFilters.status === 'preparing' ? 'selected' : ''}>Готовятся (${stats.preparing || 0})</option>
                    <option value="inDelivery" ${currentOrdersFilters.status === 'inDelivery' ? 'selected' : ''}>В доставке (${stats.inDelivery || 0})</option>
                    <option value="delivered" ${currentOrdersFilters.status === 'delivered' ? 'selected' : ''}>Доставлены (${stats.delivered || 0})</option>
                    <option value="completed" ${currentOrdersFilters.status === 'completed' ? 'selected' : ''}>Завершены (${stats.completed || 0})</option>
                    <option value="cancelled" ${currentOrdersFilters.status === 'cancelled' ? 'selected' : ''}>Отменены (${stats.cancelled || 0})</option>
                </select>
                <select id="order-shop-filter" style="padding: 10px; border: 1px solid #ddd; border-radius: 4px;">
                    <option value="">Все магазины</option>
                    ${(window.shopOwners || []).map(shop => 
                        `<option value="${shop.id}" ${currentOrdersFilters.shop_owner_id === shop.id ? 'selected' : ''}>${shop.name} (${shop.phone})</option>`
                    ).join('')}
                </select>
                <input type="date" id="order-date-from" placeholder="Дата от" 
                    style="padding: 10px; border: 1px solid #ddd; border-radius: 4px;"
                    value="${currentOrdersFilters.date_from || ''}">
                <input type="date" id="order-date-to" placeholder="Дата до" 
                    style="padding: 10px; border: 1px solid #ddd; border-radius: 4px;"
                    value="${currentOrdersFilters.date_to || ''}">
            </div>
            <div style="display: flex; gap: 10px;">
                <button onclick="applyOrdersFilters()" class="btn btn-primary">
                    <i class="fas fa-filter"></i> Применить фильтры
                </button>
                <button onclick="resetOrdersFilters()" class="btn btn-secondary">
                    <i class="fas fa-times"></i> Сбросить
                </button>
            </div>
        </div>
    `;
    
    if (orders.length === 0) {
        container.innerHTML = filtersHTML + '<p style="text-align: center; padding: 40px;">Заказов не найдено</p>';
        return;
    }
    
    // Функция для перевода статусов
    const statusLabels = {
        'pending': 'Ожидает',
        'confirmed': 'Подтвержден',
        'preparing': 'Готовится',
        'inDelivery': 'В доставке',
        'delivered': 'Доставлен',
        'completed': 'Завершен',
        'cancelled': 'Отменен'
    };
    
    const table = `
        ${filtersHTML}
        <div style="background: white; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); overflow: hidden;">
            <table class="data-table">
                <thead>
                    <tr>
                        <th>№ Заказа</th>
                        <th>Клиент</th>
                        <th>Телефон</th>
                        <th>Магазин</th>
                        <th>Товары</th>
                        <th>Сумма</th>
                        <th>Статус</th>
                        <th>Дата</th>
                        <th>Желаемая доставка</th>
                        <th>Действия</th>
                    </tr>
                </thead>
                <tbody>
                    ${orders.map(order => `
                        <tr>
                            <td><strong>${order.order_number || order.id?.substring(0, 8)}</strong></td>
                            <td>
                                <div style="line-height: 1.4;">
                                    <div><strong>${order.recipient_name || order.user?.name || 'N/A'}</strong></div>
                                    ${order.user?.is_guest ? '<small style="color: #999;">🎭 Гость</small>' : ''}
                                </div>
                            </td>
                            <td><a href="tel:${order.phone}" style="color: #667eea;">${order.phone}</a></td>
                            <td>
                                <div style="line-height: 1.4;">
                                    <div><strong>${order.shop_owner?.name || 'N/A'}</strong></div>
                                    <small style="color: #999;">${order.shop_owner?.phone || ''}</small>
                                </div>
                            </td>
                            <td>${order.order_items?.length || 0} шт.</td>
                            <td><strong>${order.total_amount || 0} ${order.currency || 'TJS'}</strong></td>
                            <td>
                                <span class="status-badge ${order.status}" style="padding: 5px 10px; border-radius: 12px; font-size: 12px; font-weight: 600;">
                                    ${statusLabels[order.status] || order.status}
                                </span>
                            </td>
                            <td>${new Date(order.created_at).toLocaleString('ru-RU', {day: '2-digit', month: '2-digit', year: 'numeric', hour: '2-digit', minute: '2-digit'})}</td>
                            <td>${order.desired_at ? new Date(order.desired_at).toLocaleString('ru-RU', {day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit'}) : '-'}</td>
                            <td>
                                <div class="action-buttons" style="display: flex; gap: 5px; flex-wrap: wrap;">
                                    <button class="action-btn view" onclick="viewOrderDetails('${order.id}')" title="Просмотр">
                                        <i class="fas fa-eye"></i>
                                    </button>
                                    ${order.status === 'pending' ? `
                                        <button class="action-btn success" onclick="confirmOrder('${order.id}')" title="Подтвердить">
                                            <i class="fas fa-check"></i>
                                        </button>
                                        <button class="action-btn danger" onclick="rejectOrder('${order.id}')" title="Отклонить">
                                            <i class="fas fa-times"></i>
                                        </button>
                                    ` : ''}
                                    ${order.status !== 'cancelled' && order.status !== 'completed' ? `
                                        <button class="action-btn edit" onclick="changeOrderStatus('${order.id}', '${order.status}')" title="Изменить статус">
                                            <i class="fas fa-edit"></i>
                                        </button>
                                    ` : ''}
                                </div>
                            </td>
                        </tr>
                    `).join('')}
                </tbody>
            </table>
        </div>
        ${pagination.totalPages > 1 ? createPagination(pagination) : ''}
    `;
    
    container.innerHTML = table;
}

// Применить фильтры
function applyOrdersFilters() {
    const search = document.getElementById('order-search')?.value || '';
    const status = document.getElementById('order-status-filter')?.value || '';
    const shopOwnerId = document.getElementById('order-shop-filter')?.value || '';
    const dateFrom = document.getElementById('order-date-from')?.value || '';
    const dateTo = document.getElementById('order-date-to')?.value || '';
    
    const filters = {};
    if (search) filters.search = search;
    if (status) filters.status = status;
    if (shopOwnerId) filters.shop_owner_id = shopOwnerId;
    if (dateFrom) filters.date_from = dateFrom;
    if (dateTo) filters.date_to = dateTo;
    
    loadOrders(1, filters);
}

// Сбросить фильтры
function resetOrdersFilters() {
    document.getElementById('order-search').value = '';
    document.getElementById('order-status-filter').value = '';
    document.getElementById('order-shop-filter').value = '';
    document.getElementById('order-date-from').value = '';
    document.getElementById('order-date-to').value = '';
    loadOrders(1, {});
}

// Создать пагинацию
function createPagination(pagination) {
    const { page, totalPages } = pagination;
    let pages = '';
    
    for (let i = 1; i <= totalPages; i++) {
        if (i === page) {
            pages += `<button class="pagination-btn active">${i}</button>`;
        } else if (i === 1 || i === totalPages || (i >= page - 2 && i <= page + 2)) {
            pages += `<button class="pagination-btn" onclick="loadOrders(${i}, currentOrdersFilters)">${i}</button>`;
        } else if (i === page - 3 || i === page + 3) {
            pages += `<span>...</span>`;
        }
    }
    
    return `
        <div class="pagination" style="display: flex; justify-content: center; gap: 5px; margin-top: 20px; padding: 20px;">
            ${page > 1 ? `<button class="pagination-btn" onclick="loadOrders(${page - 1}, currentOrdersFilters)"><i class="fas fa-chevron-left"></i></button>` : ''}
            ${pages}
            ${page < totalPages ? `<button class="pagination-btn" onclick="loadOrders(${page + 1}, currentOrdersFilters)"><i class="fas fa-chevron-right"></i></button>` : ''}
        </div>
    `;
}

// Подтвердить заказ
async function confirmOrder(orderId) {
    if (!confirm('Подтвердить этот заказ?')) return;
    
    try {
        const response = await fetch(getApiUrl(`/api/v1/admin/orders/${orderId}/confirm`), {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${adminToken}`,
                'Content-Type': 'application/json'
            }
        });
        
        const data = await response.json();
        
        if (data.success) {
            showMessage('Заказ подтвержден!', 'success');
            loadOrders(currentOrdersPage, currentOrdersFilters);
        } else {
            showMessage(data.message || 'Ошибка при подтверждении заказа', 'error');
        }
    } catch (error) {
        console.error('Ошибка:', error);
        showMessage('Ошибка при подтверждении заказа', 'error');
    }
}

// Отклонить заказ
async function rejectOrder(orderId) {
    if (!confirm('Отклонить этот заказ?')) return;
    
    try {
        const response = await fetch(getApiUrl(`/api/v1/admin/orders/${orderId}/reject`), {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${adminToken}`,
                'Content-Type': 'application/json'
            }
        });
        
        const data = await response.json();
        
        if (data.success) {
            showMessage('Заказ отклонен', 'success');
            loadOrders(currentOrdersPage, currentOrdersFilters);
        } else {
            showMessage(data.message || 'Ошибка при отклонении заказа', 'error');
        }
    } catch (error) {
        console.error('Ошибка:', error);
        showMessage('Ошибка при отклонении заказа', 'error');
    }
}

// Изменить статус заказа
async function changeOrderStatus(orderId, currentStatus) {
    const statuses = [
        { value: 'pending', label: 'Ожидает подтверждения' },
        { value: 'confirmed', label: 'Подтвержден' },
        { value: 'preparing', label: 'Готовится' },
        { value: 'inDelivery', label: 'В доставке' },
        { value: 'delivered', label: 'Доставлен' },
        { value: 'completed', label: 'Завершен' },
        { value: 'cancelled', label: 'Отменен' }
    ];
    
    const options = statuses.map(s => 
        `<option value="${s.value}" ${s.value === currentStatus ? 'selected' : ''}>${s.label}</option>`
    ).join('');
    
    const newStatus = prompt(`Выберите новый статус:\n\n${statuses.map((s, i) => `${i+1}. ${s.label}`).join('\n')}\n\nВведите номер или название:`, currentStatus);
    
    if (!newStatus || newStatus === currentStatus) return;
    
    // Находим статус по номеру или названию
    let selectedStatus = newStatus;
    const num = parseInt(newStatus);
    if (!isNaN(num) && num >= 1 && num <= statuses.length) {
        selectedStatus = statuses[num - 1].value;
    }
    
    try {
        const response = await fetch(getApiUrl(`/api/v1/admin/orders/${orderId}/status`), {
            method: 'PUT',
            headers: {
                'Authorization': `Bearer ${adminToken}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ status: selectedStatus })
        });
        
        const data = await response.json();
        
        if (data.success) {
            showMessage('Статус заказа изменен!', 'success');
            loadOrders(currentOrdersPage, currentOrdersFilters);
        } else {
            showMessage(data.message || 'Ошибка при изменении статуса', 'error');
        }
    } catch (error) {
        console.error('Ошибка:', error);
        showMessage('Ошибка при изменении статуса', 'error');
    }
}

// Просмотр деталей заказа
async function viewOrderDetails(orderId) {
    try {
        const response = await fetchData(`/api/v1/admin/orders/${orderId}`);
        
        if (response.data) {
            const order = response.data;
            
            const statusLabels = {
                'pending': 'Ожидает подтверждения',
                'confirmed': 'Подтвержден',
                'preparing': 'Готовится',
                'inDelivery': 'В доставке',
                'delivered': 'Доставлен',
                'completed': 'Завершен',
                'cancelled': 'Отменен'
            };
            
            const itemsHTML = order.order_items?.map(item => `
                <tr>
                    <td>${item.name || 'N/A'}</td>
                    <td>${item.size || '-'}</td>
                    <td>${item.color || '-'}</td>
                    <td>${item.quantity}</td>
                    <td>${item.price} ${order.currency || 'TJS'}</td>
                    <td><strong>${item.subtotal || (item.price * item.quantity)} ${order.currency || 'TJS'}</strong></td>
                </tr>
            `).join('') || '<tr><td colspan="6">Нет товаров</td></tr>';
            
            const detailsHTML = `
                <div style="max-height: 70vh; overflow-y: auto; padding: 20px;">
                    <h3 style="margin-bottom: 20px;">Заказ №${order.order_number || order.id?.substring(0, 8)}</h3>
                    
                    <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 20px; margin-bottom: 20px;">
                        <div>
                            <h4>Информация о клиенте</h4>
                            <p><strong>Имя:</strong> ${order.recipient_name || 'N/A'}</p>
                            <p><strong>Телефон:</strong> <a href="tel:${order.phone}">${order.phone}</a></p>
                            <p><strong>Адрес:</strong> ${order.shipping_address || 'N/A'}</p>
                            ${order.notes ? `<p><strong>Примечания:</strong> ${order.notes}</p>` : ''}
                        </div>
                        <div>
                            <h4>Информация о заказе</h4>
                            <p><strong>Статус:</strong> <span class="status-badge ${order.status}">${statusLabels[order.status] || order.status}</span></p>
                            <p><strong>Способ оплаты:</strong> ${order.payment_method === 'cash' ? 'Наличные' : 'Карта'}</p>
                            <p><strong>Способ доставки:</strong> ${order.shipping_method === 'courier' ? 'Курьер' : 'Самовывоз'}</p>
                            <p><strong>Дата создания:</strong> ${new Date(order.created_at).toLocaleString('ru-RU')}</p>
                            ${order.desired_at ? `<p><strong>Желаемое время:</strong> ${new Date(order.desired_at).toLocaleString('ru-RU')}</p>` : ''}
                            ${order.confirmed_at ? `<p><strong>Подтвержден:</strong> ${new Date(order.confirmed_at).toLocaleString('ru-RU')}</p>` : ''}
                        </div>
                    </div>
                    
                    <h4>Товары в заказе</h4>
                    <table class="data-table" style="margin-bottom: 20px;">
                        <thead>
                            <tr>
                                <th>Название</th>
                                <th>Размер</th>
                                <th>Цвет</th>
                                <th>Количество</th>
                                <th>Цена</th>
                                <th>Сумма</th>
                            </tr>
                        </thead>
                        <tbody>
                            ${itemsHTML}
                        </tbody>
                    </table>
                    
                    <div style="text-align: right; padding: 15px; background: #f5f5f5; border-radius: 8px;">
                        <p><strong>Стоимость товаров:</strong> ${order.items_subtotal || 0} ${order.currency || 'TJS'}</p>
                        <p><strong>Доставка:</strong> ${order.delivery_fee || 0} ${order.currency || 'TJS'}</p>
                        <h3 style="margin-top: 10px; color: #667eea;"><strong>Итого:</strong> ${order.total_amount || 0} ${order.currency || 'TJS'}</h3>
                    </div>
                </div>
            `;
            
            // Показываем в модальном окне (нужно создать универсальное модальное окно)
            showModal('Детали заказа', detailsHTML);
        }
    } catch (error) {
        console.error('Ошибка загрузки деталей заказа:', error);
        showMessage('Ошибка загрузки деталей заказа', 'error');
    }
}

// Универсальное модальное окно
function showModal(title, content) {
    const modal = document.createElement('div');
    modal.className = 'modal';
    modal.style.display = 'flex';
    modal.innerHTML = `
        <div class="modal-content" style="max-width: 900px;">
            <div class="modal-header">
                <h3>${title}</h3>
                <span class="close" onclick="this.closest('.modal').remove()">&times;</span>
            </div>
            ${content}
        </div>
    `;
    document.body.appendChild(modal);
    
    // Закрытие при клике вне модального окна
    modal.addEventListener('click', (e) => {
        if (e.target === modal) {
            modal.remove();
        }
    });
}

// Настройка форм
function setupForms() {
    // Форма товара
    document.getElementById('product-form').addEventListener('submit', handleProductSubmit);
    
    // Форма категории
    document.getElementById('category-form').addEventListener('submit', handleCategorySubmit);
}

// Обработка отправки формы товара
async function handleProductSubmit(e) {
    e.preventDefault();
    
    console.log('🚀 Начинаем обработку формы товара...');
    console.log('📦 Текущие вариации:', variations);
    console.log('📦 Количество вариаций:', variations.length);
    
    // Получаем элементы формы
    const nameInput = document.getElementById('product-name');
    const descriptionInput = document.getElementById('product-description');
    const genderInput = document.getElementById('product-gender');
    const categoryInput = document.getElementById('product-category');
    const brandInput = document.getElementById('product-brand');
    
    // Валидация обязательных полей
    const errors = [];
    
    if (!nameInput.value.trim()) {
        errors.push('Название товара');
        nameInput.style.borderColor = '#e74c3c';
    } else {
        nameInput.style.borderColor = '#ddd';
    }
    
    if (!descriptionInput.value.trim()) {
        errors.push('Описание товара');
        descriptionInput.style.borderColor = '#e74c3c';
    } else {
        descriptionInput.style.borderColor = '#ddd';
    }
    
    if (!genderInput.value) {
        errors.push('Пол товара');
        genderInput.style.borderColor = '#e74c3c';
    } else {
        genderInput.style.borderColor = '#ddd';
    }
    
    if (!categoryInput.value) {
        errors.push('Категория товара');
        categoryInput.style.borderColor = '#e74c3c';
    } else {
        categoryInput.style.borderColor = '#ddd';
    }
    
    if (variations.length === 0) {
        errors.push('Хотя бы одна вариация товара');
    }
    
    // Если есть ошибки, показываем их все сразу
    if (errors.length > 0) {
        const errorMessage = `Пожалуйста, заполните обязательные поля:\n• ${errors.join('\n• ')}`;
        showMessage(errorMessage, 'error');
        
        // Фокусируемся на первом поле с ошибкой
        if (!nameInput.value.trim()) nameInput.focus();
        else if (!descriptionInput.value.trim()) descriptionInput.focus();
        else if (!genderInput.value) genderInput.focus();
        else if (!categoryInput.value) categoryInput.focus();
        
        return;
    }
    
    // Проверяем вариации
    const variationErrors = [];
    for (let i = 0; i < variations.length; i++) {
        const variation = variations[i];
        const variationNumber = i + 1;
        
        // Проверяем размеры
        if (!variation.sizes || variation.sizes.length === 0) {
            variationErrors.push(`Вариация ${variationNumber}: размеры не выбраны`);
        }
        
        // Проверяем цвета
        if (!variation.colors || variation.colors.length === 0) {
            variationErrors.push(`Вариация ${variationNumber}: цвета не выбраны`);
        }
        
        // Проверяем цену
        if (!variation.price || variation.price <= 0) {
            variationErrors.push(`Вариация ${variationNumber}: цена должна быть больше 0`);
        }
        
        // Проверяем количество на складе
        if (!variation.stockQuantity || variation.stockQuantity < 0) {
            variationErrors.push(`Вариация ${variationNumber}: количество на складе должно быть 0 или больше`);
        }
        
        console.log(`✅ Вариация ${variationNumber} валидна:`, variation);
    }
    
    // Если есть ошибки в вариациях, показываем их
    if (variationErrors.length > 0) {
        const errorMessage = `Ошибки в вариациях:\n• ${variationErrors.join('\n• ')}`;
        showMessage(errorMessage, 'error');
        return;
    }
    
    // Показываем сообщение об успешной валидации
    showMessage('✅ Все поля заполнены корректно! Отправляем данные на сервер...', 'success');

    const formData = {
        name: nameInput.value.trim(),
        description: descriptionInput.value.trim(),
        gender: genderInput.value,
        categoryId: categoryInput.value,
        brand: brandInput.value.trim(),
        variations: variations
    };

    console.log('📦 Отправляем данные товара:', formData);
    console.log('📦 Количество вариаций:', variations.length);
    console.log('📦 Вариации:', variations);
    
    // Показываем индикатор загрузки
    const submitBtn = e.target.querySelector('button[type="submit"]');
    const originalText = submitBtn.innerHTML;
    submitBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Сохранение...';
    submitBtn.disabled = true;

    try {
        if (currentProductId) {
            // Обновление существующего товара
            const result = await fetchData(CONFIG.API.ENDPOINTS.PRODUCTS.UPDATE(currentProductId), {
                method: 'PUT',
                body: JSON.stringify(formData)
            });
            
            console.log('✅ Товар успешно обновлен:', result);
            showMessage(`✅ Товар "${formData.name}" успешно обновлен!`, 'success');
        } else {
            // Создание нового товара
            const result = await fetchData(CONFIG.API.ENDPOINTS.PRODUCTS.CREATE, {
                method: 'POST',
                body: JSON.stringify(formData)
            });
            
            console.log('✅ Товар успешно создан:', result);
            showMessage(`✅ Товар "${formData.name}" успешно создан!`, 'success');
        }
        
        // Закрываем модальное окно
        closeProductModal();
        
        // Переключаемся на вкладку товаров и обновляем список
        showTab('products');
        await refreshProductsList();
        
    } catch (error) {
        console.error('💥 Ошибка сохранения товара:', error);
        
        let errorMessage = 'Неизвестная ошибка';
        if (error.error && error.error.message) {
            errorMessage = error.error.message;
        } else if (error.error && error.error.details) {
            errorMessage = `${error.error.message}: ${error.error.details}`;
        } else if (error.error) {
            errorMessage = error.error;
        } else if (error.message) {
            errorMessage = error.message;
        }
        
        showMessage(`Ошибка: ${errorMessage}`, 'error');
    } finally {
        // Восстанавливаем кнопку
        submitBtn.innerHTML = originalText;
        submitBtn.disabled = false;
    }
}

// Обработка отправки формы категории
async function handleCategorySubmit(e) {
    e.preventDefault();
    
    const formData = {
        name: document.getElementById('category-name').value,
        description: document.getElementById('category-description').value,
        parent_id: document.getElementById('category-parent').value || null
    };
    
    try {
        if (currentCategoryId) {
            await updateCategory(currentCategoryId, formData);
        } else {
            await createCategory(formData);
        }
        
        closeCategoryModal();
        loadCategories();
        showMessage('Категория успешно сохранена', 'success');
    } catch (error) {
        console.error('Ошибка сохранения категории:', error);
        showMessage('Ошибка сохранения категории', 'error');
    }
}

// Модальные окна
function openProductModal(productId = null) {
    currentProductId = productId;
    const modal = document.getElementById('product-modal');
    const title = document.getElementById('product-modal-title');
    
    if (productId) {
        title.textContent = 'Редактировать товар';
        loadProductData(productId);
    } else {
        title.textContent = 'Добавить товар';
        document.getElementById('product-form').reset();
        clearVariationsForm(); // Очищаем форму вариаций при открытии модального окна
    }
    
    // Загружаем категории для заполнения селекта
    loadCategories();
    
    modal.style.display = 'block';
}

function closeProductModal() {
    document.getElementById('product-modal').style.display = 'none';
    currentProductId = null;
    
    // Очищаем форму
    document.getElementById('product-form').reset();
    clearImageForm();
    clearVariationsForm();
    
    // Сбрасываем глобальные переменные
    variations = [];
    uploadedImages = [];
    imageUrls = [];
}

function openCategoryModal(categoryId = null) {
    currentCategoryId = categoryId;
    const modal = document.getElementById('category-modal');
    const title = document.getElementById('category-modal-title');
    
    if (categoryId) {
        title.textContent = 'Редактировать категорию';
        loadCategoryData(categoryId);
    } else {
        title.textContent = 'Добавить категорию';
        document.getElementById('category-form').reset();
    }
    
    modal.style.display = 'block';
}

function closeCategoryModal() {
    document.getElementById('category-modal').style.display = 'none';
    currentCategoryId = null;
}

// API функции
async function fetchData(endpoint, options = {}) {
    // Убираем лишний слеш если endpoint начинается с /
    const cleanEndpoint = endpoint.startsWith('/') ? endpoint : `/${endpoint}`;
    const url = `${API_BASE_URL}${cleanEndpoint}`;
    
    // Логируем API запросы для отладки
    console.log(`🌐 API Request: ${options.method || 'GET'} ${url}`);
    
    const headers = {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
        'X-Requested-With': 'XMLHttpRequest',
        ...options.headers
    };
    
    // Добавляем токен для ВСЕХ запросов, если он есть
    if (adminToken) {
        headers['Authorization'] = `Bearer ${adminToken}`;
        console.log('🔑 Добавляем токен авторизации для запроса:', endpoint);
    } else {
        console.log('⚠️ Токен отсутствует для запроса:', endpoint);
    }
    
    const response = await fetch(url, {
        headers,
        ...options
    });
    
    if (!response.ok) {
        let errorMessage = `HTTP error! status: ${response.status}`;
        try {
            const errorData = await response.json();
            errorMessage = errorData.message || errorData.error || errorMessage;
        } catch (e) {
            // Если не удалось распарсить JSON, используем текст ответа
            try {
                const errorText = await response.text();
                if (errorText) errorMessage = errorText;
            } catch (e2) {
                // Игнорируем ошибки парсинга
            }
        }
        throw new Error(errorMessage);
    }
    
    return await response.json();
}

async function createProduct(data) {
            return await fetchData('/api/v1/shop/products/', {
        method: 'POST',
        body: JSON.stringify(data)
    });
}

async function updateProduct(id, data) {
            return await fetchData(`/api/v1/shop/products/${id}/`, {
        method: 'PUT',
        body: JSON.stringify(data)
    });
}

async function deleteProduct(id) {
    // Показываем красивое модальное окно подтверждения
    const confirmed = await showConfirmDialog(
        'Удаление товара',
        'Вы уверены, что хотите удалить этот товар?',
        'Это действие нельзя отменить.',
        'Удалить',
        'Отмена'
    );
    
    if (!confirmed) {
        return;
    }
    
    try {
        // Показываем индикатор загрузки
        showMessage('Удаление товара...', 'info');
        
        await fetchData(`/api/v1/shop/products/${id}/`, { method: 'DELETE' });
        
        showMessage('✅ Товар успешно удален', 'success');
        
        // Обновляем список товаров
        await refreshProductsList();
        
    } catch (error) {
        console.error('Ошибка удаления товара:', error);
        
        let errorMessage = 'Неизвестная ошибка';
        if (error.error && error.error.message) {
            errorMessage = error.error.message;
        } else if (error.error) {
            errorMessage = error.error;
        } else if (error.message) {
            errorMessage = error.message;
        }
        
        showMessage(`❌ Ошибка удаления товара: ${errorMessage}`, 'error');
    }
}

// Функция для показа диалога подтверждения
function showConfirmDialog(title, message, description, confirmText, cancelText) {
    return new Promise((resolve) => {
        // Создаем модальное окно
        const modal = document.createElement('div');
        modal.className = 'modal';
        modal.style.display = 'block';
        modal.style.zIndex = '9999';
        
        // Создаем обработчики событий
        const handleCancel = () => {
            modal.remove();
            resolve(false);
        };
        
        const handleConfirm = () => {
            modal.remove();
            resolve(true);
        };
        
        const handleOutsideClick = (e) => {
            if (e.target === modal) {
                modal.remove();
                resolve(false);
            }
        };
        
        modal.innerHTML = `
            <div class="modal-content" style="max-width: 400px; margin: 100px auto;">
                <div class="modal-header">
                    <h3 style="color: #e74c3c;">
                        <i class="fas fa-exclamation-triangle"></i> ${title}
                    </h3>
                </div>
                <div style="padding: 20px;">
                    <p style="font-size: 16px; margin-bottom: 10px; color: #2c3e50;">${message}</p>
                    <p style="font-size: 14px; color: #7f8c8d; margin-bottom: 20px;">${description}</p>
                    <div style="display: flex; gap: 10px; justify-content: flex-end;">
                        <button class="btn btn-secondary" id="cancel-btn">
                            ${cancelText}
                        </button>
                        <button class="btn btn-danger" id="confirm-btn">
                            <i class="fas fa-trash"></i> ${confirmText}
                        </button>
                    </div>
                </div>
            </div>
        `;
        
        document.body.appendChild(modal);
        
        // Добавляем обработчики событий
        const cancelBtn = modal.querySelector('#cancel-btn');
        const confirmBtn = modal.querySelector('#confirm-btn');
        
        cancelBtn.addEventListener('click', handleCancel);
        confirmBtn.addEventListener('click', handleConfirm);
        modal.addEventListener('click', handleOutsideClick);
    });
}

// Просмотр вариаций товара
async function viewProductVariations(id) {
    try {
        console.log('👁️ Просмотр вариаций товара с ID:', id);
        console.log('🔍 Начинаем загрузку данных товара...');
        
        const response = await fetchData(`/api/v1/products/${id}`);
        console.log('📡 Ответ API для товара:', response);
        console.log('🔍 Тип ответа:', typeof response);
        console.log('🔍 Ключи ответа:', Object.keys(response));
        
        // Проверяем разные возможные форматы ответа
        let product;
        if (response.data) {
            product = response.data;
            console.log('🔍 Используем response.data');
        } else if (response.success && response.data) {
            product = response.data;
            console.log('🔍 Используем response.success.data');
        } else if (response.product) {
            product = response.product;
            console.log('🔍 Используем response.product');
        } else {
            product = response;
            console.log('🔍 Используем response напрямую');
        }
        
        console.log('📦 Данные товара:', product);
        console.log('🔍 Проверяем вариации:');
        console.log('  - product существует:', !!product);
        console.log('  - product.variations существует:', !!product?.variations);
        console.log('  - product.variations.length:', product?.variations?.length);
        console.log('  - product.variations тип:', typeof product?.variations);
        console.log('  - product.variations содержимое:', product?.variations);
        console.log('  - product.product существует:', !!product?.product);
        console.log('  - product.product?.variations существует:', !!product?.product?.variations);
        console.log('  - product.product?.variations.length:', product?.product?.variations?.length);
        
        // Проверяем вариации в правильном месте
        const variations = product?.variations || product?.product?.variations;
        console.log('🔍 Найденные вариации:', variations);
        console.log('🔍 Количество вариаций:', variations?.length);
        
        if (!variations || variations.length === 0) {
            console.log('❌ Вариации не найдены, показываем сообщение');
            showMessage('У этого товара нет вариаций', 'info');
            return;
        }
        
        // Используем правильный объект товара для отображения
        const productData = product?.product || product;
        
        // Создаем модальное окно для просмотра вариаций
        const modal = document.createElement('div');
        modal.className = 'modal';
        modal.style.display = 'block';
        modal.style.zIndex = '9999';
        
        modal.innerHTML = `
            <div class="modal-content" style="max-width: 800px; margin: 50px auto;">
                <div class="modal-header">
                    <h3 style="color:rgb(210, 217, 245);">
                        <i class="fas fa-layer-group"></i> Вариации товара: ${productData.name}
                    </h3>
                    <span class="close" onclick="this.closest('.modal').remove()">&times;</span>
                </div>
                <div style="padding: 20px; max-height: 70vh; overflow-y: auto;">
                    <div style="margin-bottom: 20px; padding: 15px; background: linear-gradient(135deg, #f8f9fa, #e9ecef); border-radius: 10px;">
                        <h4 style="margin: 0 0 10px 0; color: #2c3e50;">
                            <i class="fas fa-box"></i> Информация о товаре
                        </h4>
                        <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 10px; font-size: 14px;">
                            <div><strong>Название:</strong> ${productData.name}</div>
                            <div><strong>Бренд:</strong> ${productData.brand || 'Не указан'}</div>
                            <div><strong>Пол:</strong> ${getGenderText(productData.gender)}</div>
                            <div><strong>Категория:</strong> ${productData.category?.name || 'Не указана'}</div>
                        </div>
                    </div>
                    
                    <div style="margin-bottom: 20px;">
                        <h4 style="margin: 0 0 15px 0; color: #2c3e50;">
                            <i class="fas fa-list"></i> Вариации (${variations.length})
                        </h4>
                        <div style="display: grid; gap: 15px;">
                            ${variations.map((variation, index) => `
                                <div style="border: 2px solid #e9ecef; border-radius: 12px; padding: 20px; background: white; position: relative; overflow: hidden;">
                                    <div style="position: absolute; top: 0; left: 0; right: 0; height: 4px; background: linear-gradient(135deg, #667eea, #764ba2);"></div>
                                    
                                    <div style="display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 15px;">
                                        <h5 style="margin: 0; color: #2c3e50; font-size: 16px;">
                                            <i class="fas fa-tag"></i> Вариация ${index + 1}
                                        </h5>
                                        <span class="badge" style="background: linear-gradient(135deg, #667eea, #764ba2); color: white; font-size: 12px; padding: 6px 12px;">
                                            ID: ${variation.id?.substring(0, 8)}...
                                        </span>
                                    </div>
                                    
                                    <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); gap: 15px; margin-bottom: 15px;">
                                        <div>
                                            <strong style="color: #495057; font-size: 13px;">Размеры:</strong>
                                            <div style="margin-top: 5px;">
                                                ${variation.sizes && variation.sizes.length > 0 
                                                    ? variation.sizes.map(size => `<span class="badge" style="background: #e9ecef; color: #495057; margin: 2px; padding: 4px 8px; font-size: 11px;">${size}</span>`).join('')
                                                    : '<span style="color: #6c757d; font-style: italic;">Не указаны</span>'
                                                }
                                            </div>
                                        </div>
                                        
                                        <div>
                                            <strong style="color: #495057; font-size: 13px;">Цвета:</strong>
                                            <div style="margin-top: 5px;">
                                                ${variation.colors && variation.colors.length > 0 
                                                    ? variation.colors.map(color => `<span class="badge" style="background: #e9ecef; color: #495057; margin: 2px; padding: 4px 8px; font-size: 11px;">${color}</span>`).join('')
                                                    : '<span style="color: #6c757d; font-style: italic;">Не указаны</span>'
                                                }
                                            </div>
                                        </div>
                                        
                                        <div>
                                            <strong style="color: #495057; font-size: 13px;">Цена:</strong>
                                            <div style="margin-top: 5px; font-size: 18px; font-weight: bold; color: #28a745;">
                                                ₽${variation.price || 0}
                                            </div>
                                        </div>
                                        
                                        <div>
                                            <strong style="color: #495057; font-size: 13px;">Остаток:</strong>
                                            <div style="margin-top: 5px; font-size: 16px; font-weight: bold; color: ${variation.stockQuantity > 0 ? '#28a745' : '#dc3545'};">
                                                ${variation.stockQuantity || 0} шт.
                                            </div>
                                        </div>
                                    </div>
                                    
                                    ${(variation.imageUrls && variation.imageUrls.length > 0 ? `
                                        <div>
                                            <strong style="color: #495057; font-size: 13px;">Фотографии:</strong>
                                            <div class="variation-images-preview">
                                                ${variation.imageUrls.map((url, imgIndex) => {
                                                    // Единственная точка истины
                                                    const imageUrl = window.getImageUrl ? window.getImageUrl(url) : (() => {
                                                        let finalUrl = url;
                                                        if (!url.startsWith('http')) {
                                                            if (url.startsWith('/')) {
                                                                finalUrl = API_BASE_URL + url;
                                                            } else {
                                                                finalUrl = API_BASE_URL + '/' + url;
                                                            }
                                                        } else {
                                                            // Если URL содержит 0.0.0.0, заменяем на localhost
                                                            finalUrl = url.replace('0.0.0.0', 'localhost');
                                                        }
                                                        return finalUrl;
                                                    })();

                                                    console.log(`🖼️ Формируем URL для изображения: ${url} -> ${imageUrl}`);

                                                    return `
                                                        <div class="image-preview-item">
                                                            <img
                                                                src="${imageUrl}"
                                                                alt="Preview"
                                                                onclick="openImageModal('${imageUrl}', 'Фото вариации ${index + 1}')"
                                                                style="cursor: pointer;"
                                                                onerror="this.style.display='none'; console.error('❌ Ошибка загрузки изображения:', '${imageUrl}');"
                                                            >
                                                            <button type="button" class="remove-image" onclick="removeVariationImage(${index}, ${imgIndex})">×</button>
                                                        </div>
                                                    `;
                                                }).join('')}
                                            </div>
                                        </div>
                                    ` : '')}
                                    
                                    ${variation.sku ? `
                                        <div style="margin-top: 10px;">
                                            <strong style="color: #495057; font-size: 13px;">SKU:</strong>
                                            <span style="font-family: monospace; background: #f8f9fa; padding: 4px 8px; border-radius: 4px; font-size: 12px;">${variation.sku}</span>
                                        </div>
                                    ` : ''}
                                </div>
                            `).join('')}
                        </div>
                    </div>
                </div>
                <div style="padding: 20px; border-top: 1px solid #e9ecef; text-align: center;">
                    <button class="btn btn-primary" onclick="this.closest('.modal').remove()">
                        <i class="fas fa-times"></i> Закрыть
                    </button>
                </div>
            </div>
        `;
        
        document.body.appendChild(modal);
        
        // Закрытие по клику вне модального окна
        modal.addEventListener('click', (e) => {
            if (e.target === modal) {
                modal.remove();
            }
        });
        
    } catch (error) {
        console.error('❌ Ошибка загрузки вариаций товара:', error);
        showMessage('Ошибка загрузки вариаций товара: ' + error.message, 'error');
    }
}

// Редактирование товара
async function editProduct(id) {
    console.log('🔄 Редактирование товара с ID:', id);
    await loadProductData(id);
}

async function createCategory(data) {
    // Получаем роль пользователя
    const userRole = localStorage.getItem('userRole') || 'admin';
    
    // Создание категорий доступно только супер админу
    if (userRole !== 'super_admin') {
        throw new Error('Создание категорий доступно только супер администратору');
    }
    
    const endpoint = '/api/v1/admin/categories/';
    return await fetchData(endpoint, {
        method: 'POST',
        body: JSON.stringify(data)
    });
}

async function updateCategory(id, data) {
    // Получаем роль пользователя
    const userRole = localStorage.getItem('userRole') || 'admin';
    
    // Обновление категорий доступно супер админу и обычному админу
    if (userRole !== 'super_admin' && userRole !== 'admin') {
        throw new Error('Обновление категорий доступно только администраторам');
    }
    
    const endpoint = `/api/v1/admin/categories/${id}/`;
    return await fetchData(endpoint, {
        method: 'PUT',
        body: JSON.stringify(data)
    });
}

async function deleteCategory(id) {
    if (!confirm('Вы уверены, что хотите удалить эту категорию?')) return;
    
    try {
        // Получаем роль пользователя
        const userRole = localStorage.getItem('userRole') || 'admin';
        
        // Удаление категорий доступно супер админу и обычному админу
        if (userRole !== 'super_admin' && userRole !== 'admin') {
            throw new Error('Удаление категорий доступно только администраторам');
        }
        
        const endpoint = `/api/v1/admin/categories/${id}/`;
        await fetchData(endpoint, {
            method: 'DELETE'
        });
        loadCategories();
        showMessage('Категория успешно удалена', 'success');
    } catch (error) {
        console.error('Ошибка удаления категории:', error);
        showMessage('Ошибка удаления категории', 'error');
    }
}

// Редактирование категории
async function editCategory(id) {
    openCategoryModal(id);
}

// Просмотр заказа
async function viewOrder(id) {
    try {
        const response = await fetchData(`/api/v1/admin/orders/${id}`);
        const order = response.data;
        
        // Показываем информацию о заказе в модальном окне или alert
        alert(`Заказ #${order.id.substring(0, 8)}...\nСтатус: ${order.status}\nСумма: ₽${order.total_amount}\nДата: ${new Date(order.created_at).toLocaleDateString()}`);
    } catch (error) {
        console.error('Ошибка загрузки заказа:', error);
        showMessage('Ошибка загрузки заказа', 'error');
    }
}

// Загрузка данных для редактирования
async function loadProductData(id) {
    try {
        const response = await fetchData(`/api/v1/products/${id}`);
        console.log('📡 Ответ API для товара:', response);
        
        // Проверяем разные возможные форматы ответа
        let product;
        if (response.product) {
            product = response.product;
        } else if (response.data) {
            product = response.data;
        } else {
            product = response;
        }
        
        console.log('📦 Данные товара:', product);
        
        currentProductId = id;
        document.getElementById('product-modal-title').textContent = 'Редактировать товар';
        
        // Заполняем основные поля
        document.getElementById('product-name').value = product.name;
        document.getElementById('product-description').value = product.description;
        document.getElementById('product-gender').value = product.gender;
        document.getElementById('product-category').value = product.categoryId;
        document.getElementById('product-brand').value = product.brand;
        
        // Загружаем вариации
        if (product.variations && product.variations.length > 0) {
            variations = product.variations.map(v => ({
                id: v.id,
                sizes: v.sizes || [],
                colors: v.colors || [],
                price: v.price,
                originalPrice: v.originalPrice,
                stockQuantity: v.stockQuantity,
                sku: v.sku,
                imageUrls: v.imageUrls || [] // Обновляем загрузку множественных фото
            }));
        } else {
            variations = [];
        }
        renderVariations();
        
        document.getElementById('product-modal').style.display = 'block';
    } catch (error) {
        console.error('Ошибка загрузки данных товара:', error);
        alert('Ошибка загрузки данных товара');
    }
}

async function loadCategoryData(id) {
    try {
        const response = await fetchData(`/api/v1/categories/${id}`);
        console.log('📡 Ответ API для категории:', response);
        
        // Проверяем разные возможные форматы ответа
        let category;
        if (response.category) {
            category = response.category;
        } else if (response.data) {
            category = response.data;
        } else {
            category = response;
        }
        
        console.log('📦 Данные категории:', category);
        
        document.getElementById('category-name').value = category.name;
        document.getElementById('category-description').value = category.description || '';
        document.getElementById('category-parent').value = category.parent_id || '';
    } catch (error) {
        console.error('Ошибка загрузки данных категории:', error);
        showMessage('Ошибка загрузки данных категории', 'error');
    }
}

// Заполнение селектов категорий
function populateCategorySelects(categories) {
    console.log('🔄 populateCategorySelects вызвана с категориями:', categories);
    
    const selects = [
        document.getElementById('product-category'),
        document.getElementById('category-parent'),
        document.getElementById('category-filter')
    ];
    
    console.log('🔍 Найдено селектов:', selects.filter(s => s).length);
    console.log('🔍 product-category:', !!selects[0]);
    console.log('🔍 category-parent:', !!selects[1]);
    console.log('🔍 category-filter:', !!selects[2]);
    
    selects.forEach(select => {
        if (select) {
            const currentValue = select.value;
            select.innerHTML = select.id === 'category-filter' ? 
                '<option value="">Все категории</option>' : 
                '<option value="">Выберите категорию</option>';
            
            categories.forEach(category => {
                const option = document.createElement('option');
                option.value = category.id;
                option.textContent = category.name;
                select.appendChild(option);
            });
            
            select.value = currentValue;
        }
    });
}

// Настройки
function loadSettings() {
    const savedUrl = localStorage.getItem('api_url');
    if (savedUrl) {
        document.getElementById('api-url').value = savedUrl;
        API_BASE_URL = savedUrl;
    }
}

function saveSettings() {
    const apiUrl = document.getElementById('api-url').value;
    localStorage.setItem('api_url', apiUrl);
    API_BASE_URL = apiUrl;
    showMessage('Настройки сохранены', 'success');
}

// Фильтры
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

// Глобальная переменная для хранения всех товаров
let allProducts = [];

function filterProducts() {
    console.log('🔄 Фильтрация товаров...');
    
    const searchTerm = document.getElementById('product-search')?.value?.toLowerCase() || '';
    const categoryFilter = document.getElementById('category-filter')?.value || '';
    
    console.log('🔍 Поисковый запрос:', searchTerm);
    console.log('📂 Фильтр категории:', categoryFilter);
    
    // Фильтруем товары
    let filteredProducts = allProducts.filter(product => {
        // Поиск по названию и бренду
        const matchesSearch = !searchTerm || 
            product.name?.toLowerCase().includes(searchTerm) ||
            product.brand?.toLowerCase().includes(searchTerm) ||
            product.description?.toLowerCase().includes(searchTerm);
        
        // Фильтр по категории
        const matchesCategory = !categoryFilter || 
            product.categoryId === categoryFilter || 
            (product.category && product.category.id === categoryFilter);
        
        console.log(`🔍 Товар "${product.name}": categoryId=${product.categoryId}, category.id=${product.category?.id}, filter=${categoryFilter}, matches=${matchesCategory}`);
        
        return matchesSearch && matchesCategory;
    });
    
    console.log(`📊 Отфильтровано ${filteredProducts.length} из ${allProducts.length} товаров`);
    
    // Отображаем отфильтрованные товары
    displayProducts(filteredProducts);
}

// Утилиты
function showMessage(text, type = 'success') {
    // Удаляем предыдущие сообщения
    const existingMessages = document.querySelectorAll('.message');
    existingMessages.forEach(msg => msg.remove());
    
    const message = document.createElement('div');
    message.className = `message ${type}`;
    
    // Выбираем иконку в зависимости от типа сообщения
    let icon = 'ℹ️';
    if (type === 'success') icon = '✅';
    else if (type === 'error') icon = '❌';
    else if (type === 'warning') icon = '⚠️';
    
    message.innerHTML = `
        <div class="message-content">
            <span class="message-icon">${icon}</span>
            <span class="message-text">${text}</span>
            <button class="message-close" onclick="this.parentElement.parentElement.remove()">×</button>
        </div>
    `;
    
    // Добавляем в начало body для лучшей видимости
    document.body.insertBefore(message, document.body.firstChild);
    
    // Автоматически скрываем через 4 секунды
    setTimeout(() => {
        if (message.parentElement) {
            message.remove();
        }
    }, 4000);
}

// Инициализация загрузки изображений
function setupImageUpload() {
    const uploadArea = document.getElementById('image-upload-area');
    const fileInput = document.getElementById('image-upload');
    const preview = document.getElementById('image-preview');
    const urlsContainer = document.getElementById('image-urls');

    // Проверяем существование элементов перед добавлением обработчиков
    if (uploadArea) {
        // Клик по области загрузки
        uploadArea.addEventListener('click', () => {
            if (fileInput) fileInput.click();
        });

        // Drag and drop
        uploadArea.addEventListener('dragover', (e) => {
            e.preventDefault();
            uploadArea.classList.add('dragover');
        });

        uploadArea.addEventListener('dragleave', () => {
            uploadArea.classList.remove('dragover');
        });

        uploadArea.addEventListener('drop', (e) => {
            e.preventDefault();
            uploadArea.classList.remove('dragover');
            const files = e.dataTransfer.files;
            handleFiles(files);
        });
    }

    if (fileInput) {
        // Выбор файлов
        fileInput.addEventListener('change', (e) => {
            handleFiles(e.target.files);
        });
    }
}

// Обработка выбранных файлов
async function handleFiles(files) {
    for (let file of files) {
        if (file.type.startsWith('image/')) {
            await uploadImage(file);
        }
    }
}

// Загрузка изображения на сервер
async function uploadImage(file) {
    const formData = new FormData();
    formData.append('image', file);
    formData.append('folder', 'products');

    try {
        // Показываем индикатор загрузки с файлом
        showUploadStatus(`📤 Загружаем ${file.name}...`, 'loading');
        
        // Добавляем спиннер в превью
        const preview = document.getElementById('image-preview');
        if (preview) {
            const loadingItem = document.createElement('div');
            loadingItem.className = 'image-preview-item loading';
            loadingItem.innerHTML = `
                <div class="loading-spinner">
                    <div class="spinner"></div>
                    <div class="loading-text">Загрузка...</div>
                </div>
                <div class="file-info">${file.name}</div>
            `;
            preview.appendChild(loadingItem);
        }
        
        const response = await fetch(`${API_BASE_URL}/api/v1/upload/image?folder=products`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${adminToken}`
            },
            body: formData
        });

        const result = await response.json();

        if (result.success) {
            // Удаляем спиннер загрузки
            if (preview) {
                const loadingItem = preview.querySelector('.loading');
                if (loadingItem) loadingItem.remove();
            }
            
            uploadedImages.push({
                url: result.url,
                filename: result.filename
            });
            addImagePreview(result.url, result.filename);
            addImageUrl(result.url);
            showUploadStatus(`✅ ${file.name} загружен успешно!`, 'success');
        } else {
            // Удаляем спиннер загрузки при ошибке
            if (preview) {
                const loadingItem = preview.querySelector('.loading');
                if (loadingItem) loadingItem.remove();
            }
            showUploadStatus(`❌ Ошибка загрузки ${file.name}: ${result.error}`, 'error');
        }
    } catch (error) {
        console.error('Ошибка загрузки:', error);
        
        // Удаляем спиннер загрузки при ошибке
        const preview = document.getElementById('image-preview');
        if (preview) {
            const loadingItem = preview.querySelector('.loading');
            if (loadingItem) loadingItem.remove();
        }
        
        showUploadStatus(`❌ Ошибка загрузки ${file.name}`, 'error');
    }
}

// Добавление превью изображения
function addImagePreview(url, filename) {
    const preview = document.getElementById('image-preview');
    if (!preview) {
        console.warn('⚠️ Элемент image-preview не найден');
        return;
    }
    
    const item = document.createElement('div');
    item.className = 'image-preview-item';
    item.innerHTML = `
        <img src="${url}" alt="Preview">
        <button class="remove-image" onclick="removeImage('${filename}')">×</button>
    `;
    preview.appendChild(item);
}

// Добавление URL изображения
function addImageUrl(url) {
    const urlsContainer = document.getElementById('image-urls');
    if (!urlsContainer) {
        console.warn('⚠️ Элемент image-urls не найден');
        return;
    }
    
    const item = document.createElement('div');
    item.className = 'image-url-item';
    item.innerHTML = `
        <input type="text" value="${url}" readonly>
        <button class="remove-url" onclick="removeImageUrl('${url}')">Удалить</button>
    `;
    urlsContainer.appendChild(item);
    imageUrls.push(url);
}

// Удаление изображения
async function removeImage(filename) {
    try {
        const response = await fetch(`${API_BASE_URL}/api/v1/upload/image/${filename}?folder=products`, {
            method: 'DELETE',
            headers: {
                'Authorization': `Bearer ${adminToken}`
            }
        });

        const result = await response.json();
        if (result.success) {
            // Удаляем из массивов
            uploadedImages = uploadedImages.filter(img => img.filename !== filename);
            imageUrls = imageUrls.filter(url => !url.includes(filename));
            
            // Обновляем интерфейс
            updateImageInterface();
            showUploadStatus('Изображение удалено', 'success');
        }
    } catch (error) {
        console.error('Ошибка удаления:', error);
        showUploadStatus('Ошибка удаления изображения', 'error');
    }
}

// Удаление URL изображения
function removeImageUrl(url) {
    imageUrls = imageUrls.filter(u => u !== url);
    updateImageInterface();
}

// Обновление интерфейса изображений
function updateImageInterface() {
    const preview = document.getElementById('image-preview');
    const urlsContainer = document.getElementById('image-urls');
    
    // Проверяем существование элементов
    if (!preview || !urlsContainer) {
        console.warn('⚠️ Элементы image-preview или image-urls не найдены');
        return;
    }
    
    // Очищаем превью
    preview.innerHTML = '';
    
    // Добавляем превью для загруженных изображений
    uploadedImages.forEach(img => {
        addImagePreview(img.url, img.filename);
    });
    
    // Очищаем URL
    urlsContainer.innerHTML = '';
    
    // Добавляем URL
    imageUrls.forEach(url => {
        addImageUrl(url);
    });
}

// Добавление вариации
function addVariation() {
    const variation = {
        id: Date.now(), // Временный ID для фронтенда
        sizes: [],
        colors: [],
        price: 0,
        originalPrice: null,
        discount: 0,
        stockQuantity: 0,
        sku: '',
        imageUrls: [] // Множественные фото
    };
    
    variations.push(variation);
    console.log('➕ Добавлена новая вариация:', variation);
    console.log('📦 Всего вариаций:', variations.length);
    renderVariations();
}

// Удаление вариации
function removeVariation(index) {
    variations.splice(index, 1);
    renderVariations();
}

// Обновление вариации
function updateVariation(index, field, value) {
    if (variations[index]) {
        variations[index][field] = value;
        console.log(`🔄 Обновлена вариация ${index + 1}: ${field} = ${value}`);
    }
}

// Обновление множественных значений
function updateVariationMulti(index, field, value, checked) {
    if (variations[index]) {
        if (!variations[index][field]) {
            variations[index][field] = [];
        }
        
        if (checked) {
            if (!variations[index][field].includes(value)) {
                variations[index][field].push(value);
            }
        } else {
            variations[index][field] = variations[index][field].filter(item => item !== value);
        }
        
        console.log(`🔄 Обновлена вариация ${index + 1}: ${field} =`, variations[index][field]);
    }
}

// Обновление множественных фото
function updateVariationImages(index, imageUrls) {
    if (variations[index]) {
        variations[index].imageUrls = imageUrls;
    }
}

// Валидация файлов перед загрузкой
function validateFiles(files) {
    const allowedTypes = ['image/jpeg', 'image/jpg', 'image/png', 'image/gif', 'image/webp'];
    const maxSize = 10 * 1024 * 1024; // 10MB
    const errors = [];
    
    for (let i = 0; i < files.length; i++) {
        const file = files[i];
        
        // Проверяем тип файла
        if (!allowedTypes.includes(file.type)) {
            errors.push(`Файл "${file.name}" имеет неподдерживаемый тип: ${file.type}`);
        }
        
        // Проверяем размер файла
        if (file.size > maxSize) {
            errors.push(`Файл "${file.name}" слишком большой: ${(file.size / 1024 / 1024).toFixed(2)}MB (максимум 10MB)`);
        }
    }
    
    return errors;
}

// Загрузка множественных фото для вариации
async function uploadVariationImages(variationIndex, input) {
    const files = Array.from(input.files);
    if (files.length === 0) return;
    
    console.log(`📸 Начинаем загрузку ${files.length} файлов для вариации ${variationIndex}`);
    
    // Валидируем файлы перед загрузкой
    const validationErrors = validateFiles(files);
    if (validationErrors.length > 0) {
        alert('Ошибки валидации файлов:\n' + validationErrors.join('\n'));
        input.value = '';
        return;
    }
    
    // Показываем индикатор загрузки для вариации
    showVariationUploadStatus(variationIndex, `📤 Загружаем ${files.length} файлов...`, 'loading');
    
    const uploadedUrls = [];
    let successCount = 0;
    let errorCount = 0;
    
    for (let i = 0; i < files.length; i++) {
        const file = files[i];
        console.log(`📁 Загружаем файл ${i + 1}/${files.length}: ${file.name} (${file.size} байт)`);
        
        // Обновляем статус для каждого файла
        showVariationUploadStatus(variationIndex, `📤 Загружаем ${file.name} (${i + 1}/${files.length})...`, 'loading');
        
        try {
            const formData = new FormData();
            formData.append('image', file);
            
            console.log(`🚀 Отправляем запрос для файла ${file.name}`);
            const response = await fetch(`${API_BASE_URL}/api/v1/upload/image?folder=variations`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${adminToken}`
                },
                body: formData
            });
            
            console.log(`📡 Ответ сервера: ${response.status} ${response.statusText}`);
            
            if (response.ok) {
                const result = await response.json();
                console.log(`✅ Файл ${file.name} успешно загружен:`, result);
                uploadedUrls.push(result.url);
                successCount++;
                
                // Показываем прогресс
                showVariationUploadStatus(variationIndex, `✅ ${file.name} загружен (${successCount}/${files.length})`, 'success');
            } else {
                const errorData = await response.json().catch(() => ({}));
                console.error(`❌ Ошибка загрузки файла ${file.name}:`, errorData);
                showVariationUploadStatus(variationIndex, `❌ Ошибка загрузки ${file.name}`, 'error');
                errorCount++;
            }
        } catch (error) {
            console.error(`💥 Исключение при загрузке файла ${file.name}:`, error);
            showVariationUploadStatus(variationIndex, `❌ Ошибка загрузки ${file.name}`, 'error');
            errorCount++;
        }
    }
    
    console.log(`📊 Результат загрузки: ${successCount} успешно, ${errorCount} ошибок`);
    
    // Добавляем новые фото к существующим только если есть успешные загрузки
    if (uploadedUrls.length > 0 && variations[variationIndex]) {
        variations[variationIndex].imageUrls = [...(variations[variationIndex].imageUrls || []), ...uploadedUrls];
        console.log(`🔄 Обновляем вариацию ${variationIndex}, всего фото: ${variations[variationIndex].imageUrls.length}`);
        renderVariations();
    }
    
    // Очищаем input для возможности повторной загрузки
    input.value = '';
    
    // Финальный статус
    if (successCount > 0) {
        showVariationUploadStatus(variationIndex, `🎉 Загружено ${successCount} из ${files.length} файлов`, 'success');
        console.log(`🎉 Успешно загружено ${successCount} файлов`);
        
        // Скрываем статус через 3 секунды
        setTimeout(() => {
            hideVariationUploadStatus(variationIndex);
        }, 3000);
    } else {
        showVariationUploadStatus(variationIndex, `❌ Не удалось загрузить ни одного файла`, 'error');
    }
}

// Удаление фото из вариации
function removeVariationImage(variationIndex, imageIndex) {
    if (variations[variationIndex] && variations[variationIndex].imageUrls) {
        variations[variationIndex].imageUrls.splice(imageIndex, 1);
        renderVariations();
    }
}

// Отрисовка вариаций
function renderVariations() {
    const container = document.getElementById('variations-list');
    
    if (variations.length === 0) {
        container.innerHTML = '<p class="no-variations">Нет вариаций. Добавьте хотя бы одну вариацию.</p>';
        return;
    }
    
    container.innerHTML = variations.map((variation, index) => `
        <div class="variation-item" data-variation-index="${index}">
            <button type="button" class="remove-variation" onclick="removeVariation(${index})">×</button>
            <div class="variation-fields">
                <div class="variation-field variation-multi-select">
                    <label>Размеры</label>
                    <div class="checkbox-group">
                        <label><input type="checkbox" value="XS" ${variation.sizes?.includes('XS') ? 'checked' : ''} onchange="updateVariationMulti(${index}, 'sizes', 'XS', this.checked)"> XS</label>
                        <label><input type="checkbox" value="S" ${variation.sizes?.includes('S') ? 'checked' : ''} onchange="updateVariationMulti(${index}, 'sizes', 'S', this.checked)"> S</label>
                        <label><input type="checkbox" value="M" ${variation.sizes?.includes('M') ? 'checked' : ''} onchange="updateVariationMulti(${index}, 'sizes', 'M', this.checked)"> M</label>
                        <label><input type="checkbox" value="L" ${variation.sizes?.includes('L') ? 'checked' : ''} onchange="updateVariationMulti(${index}, 'sizes', 'L', this.checked)"> L</label>
                        <label><input type="checkbox" value="XL" ${variation.sizes?.includes('XL') ? 'checked' : ''} onchange="updateVariationMulti(${index}, 'sizes', 'XL', this.checked)"> XL</label>
                        <label><input type="checkbox" value="XXL" ${variation.sizes?.includes('XXL') ? 'checked' : ''} onchange="updateVariationMulti(${index}, 'sizes', 'XXL', this.checked)"> XXL</label>
                    </div>
                    <div class="checkbox-group">
                        <label><input type="checkbox" value="36" ${variation.sizes?.includes('36') ? 'checked' : ''} onchange="updateVariationMulti(${index}, 'sizes', '36', this.checked)"> 36</label>
                        <label><input type="checkbox" value="37" ${variation.sizes?.includes('37') ? 'checked' : ''} onchange="updateVariationMulti(${index}, 'sizes', '37', this.checked)"> 37</label>
                        <label><input type="checkbox" value="38" ${variation.sizes?.includes('38') ? 'checked' : ''} onchange="updateVariationMulti(${index}, 'sizes', '38', this.checked)"> 38</label>
                        <label><input type="checkbox" value="39" ${variation.sizes?.includes('39') ? 'checked' : ''} onchange="updateVariationMulti(${index}, 'sizes', '39', this.checked)"> 39</label>
                        <label><input type="checkbox" value="40" ${variation.sizes?.includes('40') ? 'checked' : ''} onchange="updateVariationMulti(${index}, 'sizes', '40', this.checked)"> 40</label>
                        <label><input type="checkbox" value="41" ${variation.sizes?.includes('41') ? 'checked' : ''} onchange="updateVariationMulti(${index}, 'sizes', '41', this.checked)"> 41</label>
                        <label><input type="checkbox" value="42" ${variation.sizes?.includes('42') ? 'checked' : ''} onchange="updateVariationMulti(${index}, 'sizes', '42', this.checked)"> 42</label>
                        <label><input type="checkbox" value="43" ${variation.sizes?.includes('43') ? 'checked' : ''} onchange="updateVariationMulti(${index}, 'sizes', '43', this.checked)"> 43</label>
                        <label><input type="checkbox" value="44" ${variation.sizes?.includes('44') ? 'checked' : ''} onchange="updateVariationMulti(${index}, 'sizes', '44', this.checked)"> 44</label>
                        <label><input type="checkbox" value="45" ${variation.sizes?.includes('45') ? 'checked' : ''} onchange="updateVariationMulti(${index}, 'sizes', '45', this.checked)"> 45</label>
                        <label><input type="checkbox" value="46" ${variation.sizes?.includes('46') ? 'checked' : ''} onchange="updateVariationMulti(${index}, 'sizes', '46', this.checked)"> 46</label>
                    </div>
                </div>
                <div class="variation-field variation-multi-select">
                    <label>Цвета</label>
                    <div class="checkbox-group">
                        <label><input type="checkbox" value="Красный" ${variation.colors?.includes('Красный') ? 'checked' : ''} onchange="updateVariationMulti(${index}, 'colors', 'Красный', this.checked)"> Красный</label>
                        <label><input type="checkbox" value="Синий" ${variation.colors?.includes('Синий') ? 'checked' : ''} onchange="updateVariationMulti(${index}, 'colors', 'Синий', this.checked)"> Синий</label>
                        <label><input type="checkbox" value="Зеленый" ${variation.colors?.includes('Зеленый') ? 'checked' : ''} onchange="updateVariationMulti(${index}, 'colors', 'Зеленый', this.checked)"> Зеленый</label>
                        <label><input type="checkbox" value="Желтый" ${variation.colors?.includes('Желтый') ? 'checked' : ''} onchange="updateVariationMulti(${index}, 'colors', 'Желтый', this.checked)"> Желтый</label>
                        <label><input type="checkbox" value="Черный" ${variation.colors?.includes('Черный') ? 'checked' : ''} onchange="updateVariationMulti(${index}, 'colors', 'Черный', this.checked)"> Черный</label>
                        <label><input type="checkbox" value="Белый" ${variation.colors?.includes('Белый') ? 'checked' : ''} onchange="updateVariationMulti(${index}, 'colors', 'Белый', this.checked)"> Белый</label>
                        <label><input type="checkbox" value="Серый" ${variation.colors?.includes('Серый') ? 'checked' : ''} onchange="updateVariationMulti(${index}, 'colors', 'Серый', this.checked)"> Серый</label>
                        <label><input type="checkbox" value="Коричневый" ${variation.colors?.includes('Коричневый') ? 'checked' : ''} onchange="updateVariationMulti(${index}, 'colors', 'Коричневый', this.checked)"> Коричневый</label>
                        <label><input type="checkbox" value="Розовый" ${variation.colors?.includes('Розовый') ? 'checked' : ''} onchange="updateVariationMulti(${index}, 'colors', 'Розовый', this.checked)"> Розовый</label>
                        <label><input type="checkbox" value="Фиолетовый" ${variation.colors?.includes('Фиолетовый') ? 'checked' : ''} onchange="updateVariationMulti(${index}, 'colors', 'Фиолетовый', this.checked)"> Фиолетовый</label>
                        <label><input type="checkbox" value="Оранжевый" ${variation.colors?.includes('Оранжевый') ? 'checked' : ''} onchange="updateVariationMulti(${index}, 'colors', 'Оранжевый', this.checked)"> Оранжевый</label>
                    </div>
                </div>
                <div class="variation-field">
                    <label>Цена (₽)</label>
                    <input type="number" 
                           value="${variation.price}" 
                           onchange="updateVariation(${index}, 'price', parseFloat(this.value) || 0)"
                           min="0" 
                           step="0.01" 
                           placeholder="0.00">
                </div>
                <div class="variation-field">
                    <label>Скидка (в процентах 0-100%)</label>
                    <input type="number" 
                           value="${variation.discount || 0}" 
                           onchange="updateVariation(${index}, 'discount', parseInt(this.value) || 0)"
                           min="0" 
                           max="100" 
                           placeholder="0"
                           title="Например: 15 = скидка 15%">
                </div>
                <div class="variation-field">
                    <label>Количество</label>
                    <input type="number" 
                           value="${variation.stockQuantity}" 
                           onchange="updateVariation(${index}, 'stockQuantity', parseInt(this.value) || 0)"
                           min="0" 
                           placeholder="0">
                </div>
                <div class="variation-field">
                    <label>SKU</label>
                    <input type="text" 
                           value="${variation.sku}" 
                           onchange="updateVariation(${index}, 'sku', this.value)"
                           placeholder="SKU">
                </div>
                <div class="variation-field variation-image-upload">
                    <label>Фото для этой вариации (можно выбрать несколько)</label>
                    <input type="file" 
                           accept="image/*" 
                           multiple
                           onchange="uploadVariationImages(${index}, this)"
                           style="padding: 8px;">
                           ${(variation.imageUrls && variation.imageUrls.length > 0 ? `
                            <div class="variation-images-preview">
                              ${variation.imageUrls.map((url, imgIndex) => {
                                // Единственная точка истины
                                const imageUrl = window.getImageUrl(url);
                          
                                console.log(`🖼️ Формируем URL для изображения: ${url} -> ${imageUrl}`);
                          
                                return `
                                  <div class="image-preview-item">
                                    <img
                                      src="${imageUrl}"
                                      alt="Preview"
                                      onclick="openImageModal('${imageUrl}', 'Фото вариации ${index + 1}')"
                                      style="cursor: pointer;"
                                      onerror="this.style.display='none'; console.error('❌ Ошибка загрузки изображения:', '${imageUrl}');"
                                    >
                                    <button type="button" class="remove-image" onclick="removeVariationImage(${index}, ${imgIndex})">×</button>
                                  </div>
                                `;
                              }).join('')}
                            </div>
                          ` : '')}
                </div>
            </div>
        </div>
    `).join('');
}

// Очистка формы вариаций
function clearVariationsForm() {
    variations = [];
    renderVariations();
}

// Показать статус загрузки
function showUploadStatus(message, type) {
    const container = document.querySelector('.image-upload-container');
    let status = container.querySelector('.upload-status');
    
    if (!status) {
        status = document.createElement('div');
        status.className = 'upload-status';
        container.appendChild(status);
    }
    
    status.textContent = message;
    status.className = `upload-status ${type}`;
    
    if (type === 'success') {
        setTimeout(() => {
            status.remove();
        }, 3000);
    }
}

// Показать статус загрузки для вариации
function showVariationUploadStatus(variationIndex, message, type) {
    const variationItem = document.querySelector(`[data-variation-index="${variationIndex}"]`);
    if (!variationItem) return;
    
    let status = variationItem.querySelector('.variation-upload-status');
    
    if (!status) {
        status = document.createElement('div');
        status.className = 'variation-upload-status';
        variationItem.appendChild(status);
    }
    
    status.innerHTML = `
        <div class="upload-status-content ${type}">
            <div class="status-icon">
                ${type === 'loading' ? '<div class="spinner"></div>' : 
                  type === 'success' ? '✅' : '❌'}
            </div>
            <div class="status-message">${message}</div>
        </div>
    `;
    status.className = `variation-upload-status ${type}`;
}

// Скрыть статус загрузки для вариации
function hideVariationUploadStatus(variationIndex) {
    const variationItem = document.querySelector(`[data-variation-index="${variationIndex}"]`);
    if (!variationItem) return;
    
    const status = variationItem.querySelector('.variation-upload-status');
    if (status) {
        status.remove();
    }
}

// Открыть модальное окно с изображением
function openImageModal(imageUrl, alt = 'Изображение') {
    const modal = document.createElement('div');
    modal.className = 'modal image-modal';
    modal.style.display = 'flex';
    modal.style.zIndex = '10000';
    
    modal.innerHTML = `
        <div class="modal-content image-modal-content">
            <div class="modal-header">
                <h3>Просмотр изображения</h3>
                <span class="close" onclick="this.closest('.modal').remove()">&times;</span>
            </div>
            <div class="image-modal-body">
                <img src="${imageUrl}" alt="${alt}" class="modal-image">
            </div>
        </div>
    `;
    
    document.body.appendChild(modal);
    
    // Закрытие по клику вне модального окна
    modal.addEventListener('click', (e) => {
        if (e.target === modal) {
            modal.remove();
        }
    });
    
    // Закрытие по Escape
    document.addEventListener('keydown', function closeOnEscape(e) {
        if (e.key === 'Escape') {
            modal.remove();
            document.removeEventListener('keydown', closeOnEscape);
        }
    });
}

// Очистка формы изображений
function clearImageForm() {
    uploadedImages = [];
    imageUrls = [];
    
    // Безопасная очистка элементов (проверяем существование)
    const imagePreviewElement = document.getElementById('image-preview');
    const imageUrlsElement = document.getElementById('image-urls');
    const imageUploadElement = document.getElementById('image-upload');
    
    if (imagePreviewElement) {
        imagePreviewElement.innerHTML = '';
    }
    
    if (imageUrlsElement) {
        imageUrlsElement.innerHTML = '';
    }
    
    if (imageUploadElement) {
        imageUploadElement.value = '';
    }
}

// Закрытие модальных окон при клике вне их
window.onclick = function(event) {
    const productModal = document.getElementById('product-modal');
    const categoryModal = document.getElementById('category-modal');
    
    if (event.target === productModal) {
        closeProductModal();
    }
    
    if (event.target === categoryModal) {
        closeCategoryModal();
    }
}

// Экспорт функций для использования в HTML
window.openProductModal = openProductModal;
window.closeProductModal = closeProductModal;
window.openCategoryModal = openCategoryModal;
window.closeCategoryModal = closeCategoryModal;
window.editProduct = loadProductData;
window.editCategory = loadCategoryData;
window.deleteProduct = deleteProduct;
window.deleteCategory = deleteCategory;
window.saveSettings = saveSettings;
window.openUserModal = openUserModal;
window.closeUserModal = closeUserModal;
window.openRoleModal = openRoleModal;
window.closeRoleModal = closeRoleModal;
window.showVariationUploadStatus = showVariationUploadStatus;
window.hideVariationUploadStatus = hideVariationUploadStatus;
window.openImageModal = openImageModal;

// ===== НОВАЯ ФУНКЦИЯ МОНИТОРИНГА СИСТЕМЫ =====

/**
 * 🚀 Системный мониторинг и диагностика
 * Показывает состояние системы, производительность и отладочную информацию
 */
function systemMonitor() {
    console.log('🔍 === СИСТЕМНЫЙ МОНИТОРИНГ ===');
    
    // Информация о браузере
    console.log('🌐 Браузер:', {
        userAgent: navigator.userAgent,
        language: navigator.language,
        cookieEnabled: navigator.cookieEnabled,
        onLine: navigator.onLine,
        platform: navigator.platform
    });
    
    // Производительность
    const perf = performance.getEntriesByType('navigation')[0];
    console.log('⚡ Производительность:', {
        loadTime: perf ? Math.round(perf.loadEventEnd - perf.loadEventStart) + 'ms' : 'N/A',
        domContentLoaded: perf ? Math.round(perf.domContentLoadedEventEnd - perf.domContentLoadedEventStart) + 'ms' : 'N/A',
        totalTime: perf ? Math.round(perf.loadEventEnd - perf.navigationStart) + 'ms' : 'N/A'
    });
    
    // Состояние localStorage
    console.log('💾 localStorage:', {
        adminToken: localStorage.getItem('adminToken') ? 'Присутствует' : 'Отсутствует',
        userRole: localStorage.getItem('userRole') || 'Не установлена',
        lastActivity: localStorage.getItem('lastActivity') || 'Не установлено',
        userData: localStorage.getItem('userData') ? 'Присутствует' : 'Отсутствует'
    });
    
    // Состояние DOM
    console.log('🏗️ DOM:', {
        readyState: document.readyState,
        title: document.title,
        url: window.location.href,
        viewport: {
            width: window.innerWidth,
            height: window.innerHeight
        }
    });
    
    // Состояние API
    console.log('🌐 API:', {
        baseUrl: API_BASE_URL,
        config: CONFIG ? 'Загружен' : 'Не загружен',
        endpoints: CONFIG ? Object.keys(CONFIG.API.ENDPOINTS) : 'N/A'
    });
    
    // Глобальные переменные
    console.log('🔧 Глобальные переменные:', {
        adminToken: adminToken ? 'Присутствует' : 'Отсутствует',
        userRole: userRole || 'Не установлена',
        currentProductId: currentProductId || 'Не установлен',
        currentCategoryId: currentCategoryId || 'Не установлен',
        variations: variations ? variations.length : 0,
        uploadedImages: uploadedImages ? uploadedImages.length : 0
    });
    
    // Проверка подключения к API
    testAPIConnection();
    
    console.log('🔍 === МОНИТОРИНГ ЗАВЕРШЕН ===');
}

/**
 * 🧪 Тестирование подключения к API
 */
async function testAPIConnection() {
    console.log('🧪 Тестируем подключение к API...');
    
    try {
        const startTime = performance.now();
        const response = await fetch(`${API_BASE_URL}/health`);
        const endTime = performance.now();
        const responseTime = Math.round(endTime - startTime);
        
        if (response.ok) {
            const data = await response.json();
            console.log('✅ API подключение:', {
                status: response.status,
                responseTime: responseTime + 'ms',
                data: data
            });
        } else {
            console.log('⚠️ API подключение:', {
                status: response.status,
                responseTime: responseTime + 'ms',
                statusText: response.statusText
            });
        }
    } catch (error) {
        console.error('❌ Ошибка подключения к API:', {
            message: error.message,
            type: error.name
        });
    }
}

/**
 * 🧹 Очистка системы и сброс состояния
 */
function systemReset() {
    console.log('🧹 Сброс системы...');
    
    if (confirm('Вы уверены, что хотите сбросить состояние системы? Это очистит все данные и перезагрузит страницу.')) {
        // Очищаем localStorage
        localStorage.clear();
        
        // Сбрасываем глобальные переменные
        adminToken = null;
        userRole = null;
        currentProductId = null;
        currentCategoryId = null;
        variations = [];
        uploadedImages = [];
        imageUrls = [];
        
        // Перезагружаем страницу
        window.location.reload();
    }
}

/**
 * 📊 Экспорт данных системы
 */
function exportSystemData() {
    console.log('📊 Экспорт данных системы...');
    
    const systemData = {
        timestamp: new Date().toISOString(),
        localStorage: {
            adminToken: localStorage.getItem('adminToken') ? 'Присутствует' : 'Отсутствует',
            userRole: localStorage.getItem('userRole'),
            lastActivity: localStorage.getItem('lastActivity'),
            userData: localStorage.getItem('userData')
        },
        globals: {
            adminToken: adminToken ? 'Присутствует' : 'Отсутствует',
            userRole: userRole,
            currentProductId: currentProductId,
            currentCategoryId: currentCategoryId,
            variationsCount: variations.length,
            imagesCount: uploadedImages.length
        },
        environment: {
            userAgent: navigator.userAgent,
            url: window.location.href,
            timestamp: Date.now()
        }
    };
    
    // Создаем файл для скачивания
    const dataStr = JSON.stringify(systemData, null, 2);
    const dataBlob = new Blob([dataStr], {type: 'application/json'});
    const url = URL.createObjectURL(dataBlob);
    
    const link = document.createElement('a');
    link.href = url;
    link.download = `system-data-${Date.now()}.json`;
    link.click();
    
    URL.revokeObjectURL(url);
    console.log('✅ Данные системы экспортированы');
}

// Экспортируем новые функции
window.systemMonitor = systemMonitor;
window.testAPIConnection = testAPIConnection;
window.systemReset = systemReset;
window.exportSystemData = exportSystemData; 

// ===== ФУНКЦИИ ДЛЯ РАБОТЫ С ПОЛЬЗОВАТЕЛЯМИ =====

function openUserModal() {
    document.getElementById('user-modal').style.display = 'block';
    loadRolesForSelect();
}

function closeUserModal() {
    document.getElementById('user-modal').style.display = 'none';
    document.getElementById('user-form').reset();
}

async function loadRolesForSelect() {
    try {
        console.log('🔄 Загружаем роли для селекта...');
        const response = await fetchData('/api/v1/admin/roles/');
        console.log('📡 Ответ API ролей для селекта:', response);
        
        if (response.success && response.data && response.data.roles) {
            const roleSelect = document.getElementById('modal-user-role');
            console.log('🔍 Найден селект ролей:', roleSelect);
            
            if (roleSelect) {
                roleSelect.innerHTML = '<option value="">Выберите роль</option>';
                
                response.data.roles.forEach(role => {
                    console.log('➕ Добавляем роль в селект:', role);
                    const option = document.createElement('option');
                    option.value = role.id;
                    option.textContent = role.displayName || role.name;
                    roleSelect.appendChild(option);
                });
                
                console.log(`✅ Загружено ${response.data.roles.length} ролей в селект`);
            } else {
                console.error('❌ Селект ролей не найден!');
            }
        } else {
            console.error('❌ Неверный формат ответа API ролей:', response);
        }
    } catch (error) {
        console.error('❌ Ошибка загрузки ролей для селекта:', error);
        showMessage('Ошибка загрузки ролей: ' + error.message, 'error');
    }
}

// ===== ФУНКЦИИ ДЛЯ РАБОТЫ С РОЛЯМИ =====

async function loadRoles() {
    console.log('🔄 Загружаем роли...');
    try {
        const response = await fetchData('/api/v1/admin/roles/');
        console.log('📡 Ответ API для ролей:', response);
        if (response.success) {
            console.log('✅ Роли загружены успешно:', response.data.roles);
            displayRoles(response.data.roles);
        } else {
            console.error('❌ Ошибка в ответе API для ролей:', response);
        }
    } catch (error) {
        console.error('❌ Ошибка загрузки ролей:', error);
        showMessage('Ошибка загрузки ролей', 'error');
    }
}

function displayRoles(roles) {
    console.log('🔄 displayRoles вызвана с данными:', roles);
    const tbody = document.getElementById('roles-table-body');
    console.log('🔍 Найден tbody:', tbody);
    if (!tbody) {
        console.error('❌ tbody для ролей не найден!');
        return;
    }

    tbody.innerHTML = '';
    
    if (!roles || roles.length === 0) {
        tbody.innerHTML = `
            <tr>
                <td colspan="7" class="text-center">
                    <div style="padding: 40px 20px;">
                        <i class="fas fa-user-shield" style="font-size: 48px; color: #ccc; margin-bottom: 15px;"></i>
                        <div style="font-size: 18px; color: #666; margin-bottom: 10px;">Ролей не найдено</div>
                        <div style="font-size: 14px; color: #999;">Создайте первую роль, нажав кнопку "Добавить роль"</div>
                    </div>
                </td>
            </tr>
        `;
        return;
    }
    
    roles.forEach((role, index) => {
        const row = document.createElement('tr');
        row.style.animationDelay = `${index * 0.1}s`;
        
        row.innerHTML = `
            <td>
                <div style="display: flex; align-items: center; gap: 12px;">
                    <div style="width: 45px; height: 45px; border-radius: 50%; background: linear-gradient(135deg, #f093fb, #f5576c); display: flex; align-items: center; justify-content: center; color: white; font-weight: bold; font-size: 18px; box-shadow: 0 4px 15px rgba(240, 147, 251, 0.3);">
                        <i class="fas fa-user-shield"></i>
                    </div>
                    <div>
                        <div style="font-weight: 700; color: #333; font-size: 15px;">${role.displayName}</div>
                        <div style="font-size: 11px; color: #888; font-family: monospace;">${role.name}</div>
                    </div>
                </div>
            </td>
            <td>
                <div style="display: flex; align-items: center; gap: 10px;">
                    <i class="fas fa-eye" style="color:rgb(206, 213, 245); font-size: 16px;"></i>
                    <span style="font-weight: 500;">${role.displayName}</span>
                </div>
            </td>
            <td>
                <div style="display: flex; align-items: center; gap: 10px;">
                    <i class="fas fa-info-circle" style="color: #4ecdc4; font-size: 16px;"></i>
                    <span style="font-weight: 500; max-width: 200px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;" title="${role.description || 'Нет описания'}">
                        ${role.description || 'Нет описания'}
                    </span>
                </div>
            </td>
            <td>
                <div style="display: flex; align-items: center; gap: 8px;">
                    <i class="fas fa-key" style="color: #f093fb; font-size: 14px;"></i>
                    <span style="font-size: 13px; color: #666; font-weight: 500;">
                        ${role.permissions ? role.permissions.split(',').length : 0} разрешений
                    </span>
                </div>
            </td>
            <td>
                <span class="badge ${role.isActive ? 'role-user' : 'role-admin'}" style="background: ${role.isActive ? 'linear-gradient(135deg, #4ecdc4, #44a08d)' : 'linear-gradient(135deg, #ff6b6b, #ee5a24)'};">
                    <i class="fas ${role.isActive ? 'fa-check-circle' : 'fa-times-circle'}"></i>
                    ${role.isActive ? 'Активна' : 'Неактивна'}
                </span>
            </td>
            <td>
                <div style="display: flex; align-items: center; gap: 8px;">
                    <i class="fas fa-users" style="color: #45b7d1; font-size: 14px;"></i>
                    <span class="badge" style="background: linear-gradient(135deg, #45b7d1, #96ceb4); color: white;">
                        ${role.userCount || 0}
                    </span>
                </div>
            </td>
            <td>
                <div style="display: flex; gap: 6px;">
                    <button class="btn-sm btn-primary" onclick="viewRole('${role.id}')" title="Просмотр">
                        <i class="fas fa-eye"></i>
                    </button>
                    <button class="btn-sm btn-secondary" onclick="editRole('${role.id}')" title="Редактировать" ${role.isSystem ? 'disabled' : ''}>
                        <i class="fas fa-edit"></i>
                    </button>
                    <button class="btn-sm btn-danger" onclick="deleteRole('${role.id}')" title="Удалить" ${role.isSystem ? 'disabled' : ''}>
                        <i class="fas fa-trash"></i>
                    </button>
                </div>
            </td>
        `;
        tbody.appendChild(row);
    });
}

function openRoleModal() {
    document.getElementById('role-modal').style.display = 'block';
}

function closeRoleModal() {
    document.getElementById('role-modal').style.display = 'none';
    document.getElementById('role-form').reset();
}

// Инициализация форм при загрузке страницы
document.addEventListener('DOMContentLoaded', function() {
    const userForm = document.getElementById('user-form');
    if (userForm) {
        userForm.addEventListener('submit', handleUserSubmit);
    }
    
    const roleForm = document.getElementById('role-form');
    
    if (roleForm) {
        roleForm.addEventListener('submit', handleRoleSubmit);
    }
    
    // Мобильная навигация
    setupMobileNavigation();
    
    // Адаптивность
    
    
    // Настройка мобильного тапбара
    // setupMobileTabbar(); // Отключено - убираем нижний таббар
});

// Функция для управления мобильной навигацией
function setupMobileNavigation() {
    console.log('🔧 Настройка мобильной навигации...');
    
    // Функция инициализации с повторными попытками
    function initMobileNav() {
        const mobileNavToggle = document.getElementById('mobile-nav-toggle');
        const sidebar = document.querySelector('.sidebar');
        
        console.log('🔍 Поиск элементов мобильной навигации:');
        console.log('  - mobileNavToggle:', !!mobileNavToggle);
        console.log('  - sidebar:', !!sidebar);
        
        if (!mobileNavToggle || !sidebar) {
            console.log('⚠️ Мобильные элементы навигации не найдены, повторяем через 100ms...');
            setTimeout(initMobileNav, 100);
            return;
        }
        
        console.log('✅ Элементы мобильной навигации найдены, настраиваем...');
        
        // Удаляем старые обработчики (если есть)
        const newToggle = mobileNavToggle.cloneNode(true);
        mobileNavToggle.parentNode.replaceChild(newToggle, mobileNavToggle);
        
        // Показывать/скрывать сайдбар на мобильных
        newToggle.addEventListener('click', (e) => {
            e.preventDefault();
            e.stopPropagation();
            console.log('📱 Клик по мобильной кнопке навигации');
            sidebar.classList.toggle('show');
            console.log('🔍 Сайдбар видимый:', sidebar.classList.contains('show'));
        });
        
        // Закрыть сайдбар при клике вне его
        document.addEventListener('click', (e) => {
            if (!sidebar.contains(e.target) && !newToggle.contains(e.target)) {
                sidebar.classList.remove('show');
            }
        });
        
        // Автозакрытие мобильного меню при клике на пункты навигации
        const navItems = sidebar.querySelectorAll('.nav-item a, .nav-item button');
        navItems.forEach(item => {
            item.addEventListener('click', () => {
                console.log('📱 Клик по пункту навигации, закрываем мобильное меню');
                sidebar.classList.remove('show');
            });
        });
        
        // Автозакрытие при клике на логотип или заголовок
        const logo = sidebar.querySelector('.logo, .sidebar-header');
        if (logo) {
            logo.addEventListener('click', () => {
                console.log('📱 Клик по логотипу, закрываем мобильное меню');
                sidebar.classList.remove('show');
            });
        }
        
        // Автозакрытие при переключении вкладок (для всех функций showTab)
        const originalShowTab = window.showTab;
        if (originalShowTab) {
            window.showTab = function(tabName) {
                // Закрываем мобильное меню перед переключением
                if (window.innerWidth <= 768) {
                    console.log('📱 Переключение вкладки, закрываем мобильное меню');
                    sidebar.classList.remove('show');
                }
                // Вызываем оригинальную функцию
                return originalShowTab.call(this, tabName);
            };
        }
        
        // Проверяем размер экрана и показываем/скрываем элементы
        function checkMobile() {
            const isMobile = window.innerWidth <= 768;
            
            if (isMobile) {
                newToggle.style.display = 'block';
                sidebar.classList.remove('show');
                console.log('📱 Мобильный режим: кнопка показана');
            } else {
                newToggle.style.display = 'none';
                sidebar.classList.remove('show');
                console.log('💻 Десктопный режим: кнопка скрыта');
            }
        }
        
        // Проверяем при загрузке и изменении размера окна
        checkMobile();
        window.addEventListener('resize', checkMobile);
        
        console.log('✅ Мобильная навигация настроена успешно');
    }
    
    // Запускаем инициализацию
    initMobileNav();
}

// Функция для адаптивного отображения таблиц
function setupResponsiveTables() {
    const tables = document.querySelectorAll('.data-table');
    
    tables.forEach(table => {
        const rows = table.querySelectorAll('tbody tr');
        
        rows.forEach(row => {
            const cells = row.querySelectorAll('td');
            
            cells.forEach((cell, index) => {
                const header = table.querySelector(`thead th:nth-child(${index + 1})`);
                if (header) {
                    cell.setAttribute('data-label', header.textContent.trim());
                }
            });
        });
    });
}

// Функция для оптимизации мобильного интерфейса
function optimizeForMobile() {
    // Увеличиваем размеры кнопок для удобства на мобильных
    const buttons = document.querySelectorAll('.btn, .btn-sm');
    buttons.forEach(btn => {
        btn.style.minHeight = '44px';
        btn.style.minWidth = '44px';
    });
    
    // Оптимизируем формы для мобильных
    const inputs = document.querySelectorAll('.form-input');
    inputs.forEach(input => {
        input.style.fontSize = '16px'; // Предотвращает зум на iOS
    });
    
    // Добавляем поддержку свайпов для мобильной навигации
    let startX = 0;
    let startY = 0;
    
    document.addEventListener('touchstart', (e) => {
        startX = e.touches[0].clientX;
        startY = e.touches[0].clientY;
    });
    
    document.addEventListener('touchend', (e) => {
        const endX = e.changedTouches[0].clientX;
        const endY = e.changedTouches[0].clientY;
        const diffX = startX - endX;
        const diffY = startY - endY;
        
        // Свайп влево для закрытия сайдбара
        if (diffX > 50 && Math.abs(diffY) < 50) {
            document.querySelector('.sidebar').classList.remove('show');
        }
        
        // Свайп вправо для открытия сайдбара
        if (diffX < -50 && Math.abs(diffY) < 50) {
            document.querySelector('.sidebar').classList.add('show');
        }
    });
}

// Функция для управления мобильным тапбаром
function setupMobileTabbar() {
    const mobileTabbar = document.getElementById('mobile-tabbar');
    
    function checkMobileAndShowTabbar() {
        if (window.innerWidth <= 768) {
            mobileTabbar.style.display = 'flex';
            mobileTabbar.classList.add('show');
        } else {
            mobileTabbar.style.display = 'none';
        }
    }
    
    // Показываем тапбар при загрузке на мобильных
    checkMobileAndShowTabbar();
    
    // Обновляем при изменении размера окна
    window.addEventListener('resize', checkMobileAndShowTabbar);
    
    // Обработка нажатий на элементы тапбара
    const tabbarItems = document.querySelectorAll('.tabbar-item');
    tabbarItems.forEach(item => {
        item.addEventListener('click', () => {
            // Убираем активный класс у всех элементов
            tabbarItems.forEach(tab => tab.classList.remove('active'));
            
            // Добавляем активный класс к выбранному элементу
            item.classList.add('active');
            
            // Показываем соответствующую вкладку
            const tabName = item.dataset.tab;
            switchTab(tabName);
        });
    });
}

// Функция для переключения вкладок в мобильном режиме
function switchTab(tabName) {
    // Скрываем все вкладки
    const tabContents = document.querySelectorAll('.tab-content');
    tabContents.forEach(content => {
        content.classList.remove('active');
    });
    
    // Показываем выбранную вкладку
    const selectedTab = document.getElementById(`${tabName}-tab`);
    if (selectedTab) {
        selectedTab.classList.add('active');
    }
    
    // Обновляем активный элемент в тапбаре
    const tabbarItems = document.querySelectorAll('.tabbar-item');
    tabbarItems.forEach(item => {
        item.classList.remove('active');
        if (item.dataset.tab === tabName) {
            item.classList.add('active');
        }
    });
    
    // Загружаем данные для вкладки
    switch (tabName) {
        case 'dashboard':
            loadDashboard();
            break;
        case 'products':
            loadProducts();
            break;
        case 'categories':
            loadCategories();
            break;
        case 'users':
            loadUsers();
            break;
        case 'roles':
            loadRoles();
            break;
        case 'orders':
            loadOrders();
            break;
        case 'settings':
            loadSettings();
            break;
    }
}

// Функция для загрузки клиентов (для владельца магазина)
async function loadCustomers() {
    try {
        showMessage('Загрузка клиентов...', 'info');
        
        const response = await fetchData('/api/v1/shop/customers/');
        
        if (response.customers) {
            displayCustomers(response.customers);
            showMessage('✅ Клиенты загружены', 'success');
        } else {
            showMessage('❌ Ошибка загрузки клиентов', 'error');
        }
    } catch (error) {
        console.error('Ошибка загрузки клиентов:', error);
        showMessage('❌ Ошибка загрузки клиентов', 'error');
    }
}

// Функция для отображения клиентов
function displayCustomers(customers) {
    const container = document.getElementById('customers-tab');
    if (!container) return;
    
    if (!customers || customers.length === 0) {
        container.innerHTML = `
            <div class="empty-state">
                <i class="fas fa-users"></i>
                <p>Клиенты не найдены</p>
                <small>Клиенты появятся здесь после регистрации</small>
            </div>
        `;
        return;
    }
    
    let html = `
        <div class="content-header">
            <h2><i class="fas fa-users"></i> Клиенты магазина</h2>
            <div class="filters">
                <div class="filter-group">
                    <input type="text" class="filter-input" placeholder="Поиск клиентов..." onkeyup="filterCustomers(this.value)">
                </div>
            </div>
        </div>
        <div class="table-container">
            <h3>Список клиентов (${customers.length})</h3>
            <div class="table-responsive">
                <table class="data-table">
                    <thead>
                        <tr>
                            <th>👤 Клиент</th>
                            <th>📧 Email</th>
                            <th>📱 Телефон</th>
                            <th>📅 Дата регистрации</th>
                            <th>🛒 Заказов</th>
                            <th>💰 Сумма заказов</th>
                            <th>⚙️ Действия</th>
                        </tr>
                    </thead>
                    <tbody>
    `;
    
    customers.forEach(customer => {
        const orderCount = customer.orderCount || 0;
        const totalSpent = customer.totalSpent || 0;
        
        html += `
            <tr>
                <td data-label="Клиент">
                    <div class="user-info">
                        <div class="user-avatar">
                            <i class="fas fa-user"></i>
                        </div>
                        <div>
                            <div class="user-name">${customer.name || 'Не указано'}</div>
                            <div class="user-phone">${customer.phone || 'Не указано'}</div>
                        </div>
                    </div>
                </td>
                <td data-label="Email">
                    <div class="email-info">
                        <span class="email">${customer.email}</span>
                        ${customer.emailVerified ? '<span class="verified-badge">✓</span>' : '<span class="unverified-badge">✗</span>'}
                    </div>
                </td>
                <td data-label="Телефон">${customer.phone || 'Не указано'}</td>
                <td data-label="Дата регистрации">
                    <div class="date-info">
                        <div class="date">${new Date(customer.createdAt).toLocaleDateString()}</div>
                        <div class="time">${new Date(customer.createdAt).toLocaleTimeString()}</div>
                    </div>
                </td>
                <td data-label="Заказов">
                    <span class="badge badge-info">${orderCount}</span>
                </td>
                <td data-label="Сумма заказов">
                    <span class="amount">${totalSpent.toLocaleString()} ₽</span>
                </td>
                <td data-label="Действия">
                    <div class="action-buttons">
                        <button class="btn btn-sm btn-primary" onclick="viewCustomer('${customer.id}')">
                            <i class="fas fa-eye"></i>
                        </button>
                        <button class="btn btn-sm btn-info" onclick="viewCustomerOrders('${customer.id}')">
                            <i class="fas fa-shopping-cart"></i>
                        </button>
                    </div>
                </td>
            </tr>
        `;
    });
    
    html += `
                    </tbody>
                </table>
            </div>
        </div>
    `;
    
    container.innerHTML = html;
}

// Функция для фильтрации клиентов
function filterCustomers(searchTerm) {
    const rows = document.querySelectorAll('#customers-tab .data-table tbody tr');
    
    rows.forEach(row => {
        const text = row.textContent.toLowerCase();
        const matches = text.includes(searchTerm.toLowerCase());
        row.style.display = matches ? '' : 'none';
    });
}

// Функция для просмотра клиента
async function viewCustomer(customerId) {
    try {
        const response = await fetchData(`/api/v1/shop/customers/${customerId}/`);
        
        if (response.customer) {
            const customer = response.customer;
            
            const modal = document.createElement('div');
            modal.className = 'modal';
            modal.style.display = 'block';
            
            modal.innerHTML = `
                <div class="modal-content">
                    <div class="modal-header">
                        <h3><i class="fas fa-user"></i> Информация о клиенте</h3>
                        <span class="close" onclick="this.closest('.modal').remove()">&times;</span>
                    </div>
                    <div style="padding: 20px;">
                        <div class="user-avatar-large">
                            <i class="fas fa-user"></i>
                        </div>
                        <div class="user-info-grid">
                            <div class="info-item">
                                <label>Имя:</label>
                                <span>${customer.name || 'Не указано'}</span>
                            </div>
                            <div class="info-item">
                                <label>Email:</label>
                                <span>${customer.email}</span>
                            </div>
                            <div class="info-item">
                                <label>Телефон:</label>
                                <span>${customer.phone || 'Не указано'}</span>
                            </div>
                            <div class="info-item">
                                <label>Дата регистрации:</label>
                                <span>${new Date(customer.createdAt).toLocaleString()}</span>
                            </div>
                            <div class="info-item">
                                <label>Количество заказов:</label>
                                <span>${customer.orderCount || 0}</span>
                            </div>
                            <div class="info-item">
                                <label>Общая сумма заказов:</label>
                                <span>${(customer.totalSpent || 0).toLocaleString()} ₽</span>
                            </div>
                        </div>
                    </div>
                </div>
            `;
            
            document.body.appendChild(modal);
            
            // Закрытие по клику вне модального окна
            modal.addEventListener('click', (e) => {
                if (e.target === modal) {
                    modal.remove();
                }
            });
        }
    } catch (error) {
        console.error('Ошибка загрузки информации о клиенте:', error);
        showMessage('❌ Ошибка загрузки информации о клиенте', 'error');
    }
}

// Функция для просмотра заказов клиента
async function viewCustomerOrders(customerId) {
    try {
        const response = await fetchData(`/api/v1/shop/customers/${customerId}/orders/`);
        
        if (response.orders) {
            // Здесь можно показать модальное окно с заказами клиента
            showMessage(`Загружено ${response.orders.length} заказов клиента`, 'info');
        }
    } catch (error) {
        console.error('Ошибка загрузки заказов клиента:', error);
        showMessage('❌ Ошибка загрузки заказов клиента', 'error');
    }
}

async function handleUserSubmit(e) {
    e.preventDefault();
    
    const formData = {
        name: document.getElementById('modal-user-name').value,
        email: document.getElementById('modal-user-email').value,
        password: document.getElementById('modal-user-password').value,
        phone: document.getElementById('modal-user-phone').value,
        isActive: document.getElementById('user-active').checked
    };
    
    const roleId = document.getElementById('modal-user-role').value;
    if (roleId) {
        formData.roleId = roleId;
    }
    
    console.log('🔍 Отправляемые данные:', formData);
    
    try {
        const response = await fetch(`${API_BASE_URL}/api/v1/admin/users/`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${adminToken}`
            },
            body: JSON.stringify(formData)
        });
        
        const data = await response.json();
        console.log('🔍 Ответ сервера:', data);
        console.log('🔍 Статус ответа:', response.status);
        
        if (response.ok && data.success) {
            showMessage('Пользователь создан успешно!', 'success');
            closeUserModal();
            loadUsers();
        } else {
            showMessage(data.message || 'Ошибка создания пользователя', 'error');
        }
    } catch (error) {
        console.error('Ошибка создания пользователя:', error);
        showMessage('Ошибка создания пользователя', 'error');
    }
}

async function handleRoleSubmit(e) {
    e.preventDefault();
    
    const formData = {
        name: document.getElementById('role-name').value,
        displayName: document.getElementById('role-display-name').value,
        description: document.getElementById('role-description').value,
        permissions: document.getElementById('role-permissions').value || '{}'
    };
    
    try {
        const response = await fetch(`${API_BASE_URL}/api/v1/admin/roles/`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${adminToken}`
            },
            body: JSON.stringify(formData)
        });
        
        const data = await response.json();
        
        if (response.ok && data.success) {
            showMessage('Роль создана успешно!', 'success');
            closeRoleModal();
            loadRoles();
        } else {
            showMessage(data.message || 'Ошибка создания роли', 'error');
        }
    } catch (error) {
        console.error('Ошибка создания роли:', error);
        showMessage('Ошибка создания роли', 'error');
    }
}

