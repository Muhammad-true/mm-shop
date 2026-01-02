// ===== NAVIGATION.JS - Навигация по вкладкам =====

// Настройка навигации
function setupNavigation(userRole = 'admin') {
    console.log('🔧 Настройка навигации для роли:', userRole);
    
    const navItems = document.querySelectorAll('.nav-item');
    
    // Скрываем все элементы навигации
    navItems.forEach(item => {
        item.style.display = 'none';
    });
    
    // Показываем элементы в зависимости от роли
    if (userRole === 'super_admin') {
        navItems.forEach(item => {
            item.style.display = 'flex';
        });
    } else if (userRole === 'admin') {
        const allowedTabs = ['dashboard', 'products', 'categories', 'users', 'orders', 'shops', 'updates', 'settings'];
        
        navItems.forEach(item => {
            const tabName = item.dataset.tab;
            if (allowedTabs.includes(tabName)) {
                item.style.display = 'flex';
            } else {
                item.style.display = 'none';
            }
        });
    } else if (userRole === 'shop_owner') {
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
            navItems.forEach(nav => nav.classList.remove('active'));
            this.classList.add('active');
            
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

// Переключение вкладок
function showTab(tabName, userRole = 'admin') {
    console.log(`🔄 Переключаемся на вкладку: ${tabName}, роль: ${userRole}`);
    
    const roleName = typeof userRole === 'object' ? userRole.name : userRole;
    const navItems = document.querySelectorAll('.nav-item');
    const tabContents = document.querySelectorAll('.tab-content');
    const currentPage = document.getElementById('current-page');
    
    // Убираем активный класс со всех элементов
    navItems.forEach(nav => nav.classList.remove('active'));
    tabContents.forEach(tab => tab.classList.remove('active'));
    
    // Находим нужную вкладку и активируем её
    const targetNav = document.querySelector(`[data-tab="${tabName}"]`);
    const targetTab = document.getElementById(tabName);
    
    if (targetNav && targetTab) {
        targetNav.classList.add('active');
        targetTab.classList.add('active');
        
        // Обновляем заголовок
        const navText = targetNav.querySelector('span').textContent;
        currentPage.textContent = navText;
        
        console.log(`✅ Вкладка ${tabName} активирована`);
        
        // Загружаем данные для активной вкладки
        setTimeout(() => {
            console.log(`🔄 Загружаем данные для вкладки: ${tabName}`);
            switch (tabName) {
                case 'dashboard':
                    if (window.dashboard && window.dashboard.loadDashboard) {
                        window.dashboard.loadDashboard(userRole);
                    } else if (typeof loadDashboard === 'function') {
                        loadDashboard(userRole);
                    }
                    break;
                case 'products':
                    setTimeout(() => {
                        if (window.products && window.products.loadProducts) window.products.loadProducts();
                        if (window.categories && window.categories.loadCategories) window.categories.loadCategories();
                    }, 200);
                    break;
                case 'categories':
                    if (window.categories && window.categories.loadCategories) window.categories.loadCategories();
                    break;
                case 'users':
                    console.log('👥 Загрузка пользователей, роль:', roleName);
                    // Пробуем разные способы вызова
                    if (window.loadUsers && typeof window.loadUsers === 'function') {
                        console.log('✅ Вызываем window.loadUsers()');
                        window.loadUsers();
                    } else if (window.users && window.users.loadUsers) {
                        console.log('✅ Вызываем window.users.loadUsers()');
                        window.users.loadUsers();
                    } else {
                        console.error('❌ loadUsers не найден!', {
                            hasLoadUsers: !!window.loadUsers,
                            hasUsers: !!window.users,
                            hasUsersLoadUsers: !!(window.users && window.users.loadUsers)
                        });
                    }
                    break;
                case 'roles':
                    console.log('🛡️ Загрузка ролей, роль:', roleName);
                    // Пробуем разные способы вызова
                    if (window.loadRoles && typeof window.loadRoles === 'function') {
                        console.log('✅ Вызываем window.loadRoles()');
                        window.loadRoles();
                    } else if (window.roles && window.roles.loadRoles) {
                        console.log('✅ Вызываем window.roles.loadRoles()');
                        window.roles.loadRoles();
                    } else {
                        console.error('❌ loadRoles не найден!', {
                            hasLoadRoles: !!window.loadRoles,
                            hasRoles: !!window.roles,
                            hasRolesLoadRoles: !!(window.roles && window.roles.loadRoles)
                        });
                    }
                    break;
                case 'orders':
                    if (typeof loadOrders === 'function') {
                        // Проверяем наличие параметра orderId в hash (deep link)
                        const hash = window.location.hash;
                        if (hash.includes('?')) {
                            const params = new URLSearchParams(hash.split('?')[1]);
                            const orderId = params.get('orderId');
                            loadOrders(1, {});
                            // Если есть orderId, открываем детали заказа
                            if (orderId && window.orders && window.orders.viewOrderDetails) {
                                setTimeout(() => {
                                    window.orders.viewOrderDetails(orderId);
                                }, 1000);
                            }
                        } else {
                            loadOrders(1, {});
                        }
                    }
                    break;
                case 'shops':
                    console.log('🛍️ Загрузка магазинов, роль:', roleName);
                    if (window.shops && window.shops.loadShops) {
                        window.shops.loadShops(1, {});
                    }
                    break;
                case 'updates':
                    if (window.updates && window.updates.init) {
                        window.updates.init();
                    }
                    break;
                case 'settings':
                    if (typeof loadSettings === 'function') loadSettings();
                    break;
            }
        }, 150);
    } else {
        console.error(`❌ Не удалось найти элементы для вкладки ${tabName}`);
    }
}

// Настройка мобильной навигации
function setupMobileNavigation() {
    console.log('🔧 Настройка мобильной навигации...');
    
    function initMobileNav() {
        const mobileNavToggle = document.getElementById('mobile-nav-toggle');
        const sidebar = document.querySelector('.sidebar');
        
        if (!mobileNavToggle || !sidebar) {
            setTimeout(initMobileNav, 100);
            return;
        }
        
        const newToggle = mobileNavToggle.cloneNode(true);
        mobileNavToggle.parentNode.replaceChild(newToggle, mobileNavToggle);
        
        newToggle.addEventListener('click', (e) => {
            e.preventDefault();
            e.stopPropagation();
            console.log('📱 Клик по мобильной кнопке навигации');
            sidebar.classList.toggle('show');
        });
        
        document.addEventListener('click', (e) => {
            if (!sidebar.contains(e.target) && !newToggle.contains(e.target)) {
                sidebar.classList.remove('show');
            }
        });
        
        function checkMobile() {
            const isMobile = window.innerWidth <= 768;
            
            if (isMobile) {
                newToggle.style.display = 'block';
                sidebar.classList.remove('show');
            } else {
                newToggle.style.display = 'none';
                sidebar.classList.remove('show');
            }
        }
        
        checkMobile();
        window.addEventListener('resize', checkMobile);
        
        console.log('✅ Мобильная навигация настроена успешно');
    }
    
    initMobileNav();
}

// Адаптивное отображение таблиц
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

// Оптимизация для мобильных
function optimizeForMobile() {
    const buttons = document.querySelectorAll('.btn, .btn-sm');
    buttons.forEach(btn => {
        btn.style.minHeight = '44px';
        btn.style.minWidth = '44px';
    });
    
    const inputs = document.querySelectorAll('.form-input');
    inputs.forEach(input => {
        input.style.fontSize = '16px';
    });
    
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
        
        if (diffX > 50 && Math.abs(diffY) < 50) {
            document.querySelector('.sidebar').classList.remove('show');
        }
        
        if (diffX < -50 && Math.abs(diffY) < 50) {
            document.querySelector('.sidebar').classList.add('show');
        }
    });
}

// Экспорт
window.navigation = {
    setupNavigation,
    showTab,
    setupMobileNavigation,
    setupResponsiveTables,
    optimizeForMobile
};


