// ===== AUTH.JS - Авторизация и управление пользователями =====

// Обработка входа
async function handleLogin(e) {
    if (e) e.preventDefault();
    
    const phone = document.getElementById('login-phone').value;
    const password = document.getElementById('login-password').value;
    
    // Показываем сообщение о загрузке
    if (window.ui && window.ui.showMessage) {
        window.ui.showMessage('Проверяем данные...', 'info');
    }
    
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
            const token = data.data.token;
            const role = data.data.user?.role?.name || 'user';
            const userData = data.data.user;
            
            // Сохраняем токен и роль
            if (window.storage && window.storage.setAdminToken && window.storage.setUserRole) {
                window.storage.setAdminToken(token);
                window.storage.setUserRole(role);
            } else {
                // Fallback
                window.setAdminToken(token);
                window.setUserRole(role);
            }
            localStorage.setItem('lastActivity', Date.now().toString());
            localStorage.setItem('userData', JSON.stringify(userData));
            
            console.log('✅ Успешный вход, токен получен, роль:', role);
            
            // Скрываем форму входа и показываем админ панель
            document.getElementById('login-modal').style.display = 'none';
            document.getElementById('admin-content').style.display = 'flex';
            
            // Обновляем информацию о пользователе
            updateUserInfo(userData, role);
            
            // Настройка навигации
            if (window.navigation && window.navigation.setupNavigation) {
                window.navigation.setupNavigation(role);
            }
            
            // Загружаем данные
            setTimeout(() => {
                if (window.dashboard && window.dashboard.loadDashboard) {
                    window.dashboard.loadDashboard(role);
                }
                if (window.app && window.app.loadInitialData) {
                    window.app.loadInitialData(role);
                }
                // Регистрируем FCM токен для push-уведомлений
                if (window.fcm && window.fcm.checkAndRegisterFCMToken) {
                    window.fcm.checkAndRegisterFCMToken();
                }
            }, 100);
            
            const roleText = (role === 'super_admin' || role === 'admin') ? 'админ панель' : 'панель управления магазином';
            if (window.ui && window.ui.showMessage) {
                window.ui.showMessage(`Успешный вход в ${roleText}!`, 'success');
            }
        } else {
            let errorMessage = 'Неверный email или пароль';
            if (response.status === 401) {
                errorMessage = '❌ Неверный email или пароль. Проверьте данные и попробуйте снова.';
            }
            
            if (window.ui && window.ui.showMessage) {
                window.ui.showMessage(errorMessage, 'error');
            }
        }
    } catch (error) {
        console.error('Ошибка входа:', error);
        const errorMessage = 'Ошибка подключения к серверу.';
        if (window.ui && window.ui.showMessage) {
            window.ui.showMessage(errorMessage, 'error');
        }
    }
}

// Обновление информации о пользователе
function updateUserInfo(user, userRole) {
    console.log('🔄 updateUserInfo вызвана с данными:', { user, userRole });
    
    let userName = 'Пользователь';
    let userEmail = '';
    
    if (user) {
        if (user.name && user.name.trim() !== '') {
            userName = user.name;
        } else if (user.email && user.email.trim() !== '') {
            userName = user.email.split('@')[0];
        }
        userEmail = user.email || '';
    } else {
        // Если нет данных пользователя, пробуем взять из localStorage
        const savedUserData = localStorage.getItem('userData');
        if (savedUserData) {
            try {
                const userData = JSON.parse(savedUserData);
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
    
    // Определяем роль для отображения
    let roleDisplay = '';
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
    
    // Обновляем элементы header
    const headerUserName = document.getElementById('header-user-name');
    const headerUserEmail = document.getElementById('header-user-email');
    const headerUserRole = document.getElementById('header-user-role');
    
    if (headerUserName) headerUserName.textContent = userName;
    if (headerUserEmail) headerUserEmail.textContent = userEmail;
    if (headerUserRole) headerUserRole.textContent = roleDisplay;
}

// Переключение выпадающего меню пользователя
function toggleUserDropdown() {
    if (window.storage && window.storage.updateLastActivity) {
        window.storage.updateLastActivity();
    } else if (typeof updateLastActivity === 'function') {
        updateLastActivity();
    }
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
    if (window.storage && window.storage.clearAllStorage) {
        window.storage.clearAllStorage();
    } else if (typeof clearAllStorage === 'function') {
        clearAllStorage();
    } else {
        localStorage.clear();
    }
    
    // Скрываем основной контент
    document.getElementById('admin-content').style.display = 'none';
    document.getElementById('login-modal').style.display = 'block';
    document.getElementById('login-form').reset();

    // Очищаем контент вкладок, чтобы избежать "залипания" данных предыдущей роли
    const ordersContainer = document.getElementById('orders-table');
    if (ordersContainer) ordersContainer.innerHTML = '';
    const productsContainer = document.getElementById('products-table');
    if (productsContainer) productsContainer.innerHTML = '';
    const usersTbody = document.getElementById('users-table-body');
    if (usersTbody) usersTbody.innerHTML = '';
    const categoriesContainer = document.getElementById('categories-table');
    if (categoriesContainer) categoriesContainer.innerHTML = '';
    
    // Скрываем выпадающее меню
    const dropdown = document.getElementById('user-dropdown');
    if (dropdown) {
        dropdown.classList.remove('show');
    }
    
    if (window.ui && window.ui.showMessage) {
        window.ui.showMessage('Вы успешно вышли из системы', 'success');
    }
}

// Экспорт функций
window.auth = {
    handleLogin,
    updateUserInfo,
    toggleUserDropdown,
    logout
};

