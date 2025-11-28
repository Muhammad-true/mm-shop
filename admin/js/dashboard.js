// ===== DASHBOARD.JS - Дашборд =====

// Загрузка дашборда
async function loadDashboard(userRole = null) {
    console.log('🔄 Загружаем данные дашборда...');
    
    try {
        const adminToken = window.storage ? window.storage.getAdminToken() : null;
        console.log('🔑 Токен админа:', adminToken ? 'Присутствует' : 'Отсутствует');
        
        if (!adminToken) {
            console.warn('⚠️ Токен админа отсутствует, откладываем загрузку дашборда...');
            setTimeout(() => loadDashboard(userRole), 500);
            return;
        }
        
        // Получаем роль из localStorage, если не передана
        if (!userRole) {
            userRole = window.storage ? window.storage.getUserRole() : localStorage.getItem('userRole');
            console.log('📋 Роль из localStorage:', userRole);
        }
        
        // Нормализуем роль сразу
        const roleName = typeof userRole === 'object' ? userRole.name : userRole;
        console.log('🎭 Нормализованная роль:', roleName);
        
        // Проверяем актуальность роли через API профиль
        try {
            const profile = await window.api.fetchData('/api/v1/users/profile');
            const actualRole = profile?.data?.role?.name || profile?.role?.name;
            if (actualRole && actualRole !== roleName) {
                console.warn(`⚠️ Роль в localStorage (${roleName}) не совпадает с ролью на сервере (${actualRole}). Обновляем...`);
                if (window.storage && window.storage.setUserRole) {
                    window.storage.setUserRole(actualRole);
                } else {
                    localStorage.setItem('userRole', actualRole);
                }
                // Перезагружаем дашборд с актуальной ролью
                return loadDashboard(actualRole);
            }
        } catch (error) {
            console.warn('⚠️ Не удалось проверить актуальность роли:', error.message);
        }
        
        let products = { data: [] };
        let users = { data: { users: [] } };
        let orders = { data: { orders: [] } };
        let subscribers = { data: { subscribers: [] } };
        let shopId = null;
        
        // Получаем ID магазина для shop_owner
        if (roleName === 'shop_owner') {
            try {
                const profile = await window.api.fetchData('/api/v1/users/profile');
                shopId = profile?.data?.id || profile?.id;
                console.log('🏪 ID магазина:', shopId);
            } catch (error) {
                console.warn('⚠️ Ошибка получения профиля:', error.message);
            }
        }
        
        try {
            let productsEndpoint = CONFIG.API.ENDPOINTS.PRODUCTS.LIST;
            if (roleName === 'super_admin' || roleName === 'admin') {
                console.log('👑 Админ загружает все товары для дашборда');
            } else if (roleName === 'shop_owner') {
                console.log('🏪 Владелец магазина загружает свои товары для дашборда');
            } else {
                console.warn('⚠️ Неизвестная роль:', roleName);
            }
            
            products = await window.api.fetchData(productsEndpoint);
            console.log('✅ Товары загружены:', products.products?.length || 0);
        } catch (error) {
            console.warn('⚠️ Ошибка загрузки товаров:', error.message);
        }
        
        try {
            if (roleName === 'super_admin' || roleName === 'admin') {
                users = await window.api.fetchData(CONFIG.API.ENDPOINTS.USERS.LIST);
                console.log('✅ Пользователи загружены:', users.data?.users?.length || 0);
            } else if (roleName === 'shop_owner' && shopId) {
                // Загружаем подписчиков для магазина
                subscribers = await window.api.fetchData(`/api/v1/shops/${shopId}/subscribers`);
                console.log('✅ Подписчики загружены:', subscribers.data?.subscribers?.length || 0);
            }
        } catch (error) {
            console.warn('⚠️ Ошибка загрузки пользователей/подписчиков:', error.message);
        }
        
        try {
            if (roleName === 'super_admin' || roleName === 'admin') {
                orders = await window.api.fetchData(CONFIG.API.ENDPOINTS.ORDERS.LIST);
                console.log('📡 Ответ API заказов (админ):', orders);
            } else if (roleName === 'shop_owner') {
                orders = await window.api.fetchData('/api/v1/shop/orders/');
                console.log('📡 Ответ API заказов (владелец):', orders);
            }
        } catch (error) {
            console.warn('⚠️ Ошибка загрузки заказов:', error.message);
        }
        
        // Извлекаем данные с учетом разных структур ответа
        const productsList = Array.isArray(products?.products) ? products.products 
            : Array.isArray(products?.data?.products) ? products.data.products : [];
        const usersList = (roleName === 'super_admin' || roleName === 'admin') 
            ? (Array.isArray(users?.data?.users) ? users.data.users 
                : Array.isArray(users?.users) ? users.users : [])
            : [];
        const subscribersList = (roleName === 'shop_owner')
            ? (Array.isArray(subscribers?.data?.subscribers) ? subscribers.data.subscribers : [])
            : [];
        const ordersList = Array.isArray(orders?.data?.orders) ? orders.data.orders 
            : Array.isArray(orders?.orders) ? orders.orders 
            : Array.isArray(orders?.data) ? orders.data : [];
        
        const totalProducts = productsList.length;
        const totalUsers = roleName === 'shop_owner' ? subscribersList.length : usersList.length;
        const totalOrders = ordersList.length;
        
        // Доход считаем только из завершенных заказов
        const completedOrders = ordersList.filter(order => {
            const status = (order.status || '').toLowerCase();
            return status === 'completed' || status === 'завершен';
        });
        const revenue = completedOrders.reduce((sum, order) => sum + (order.total_amount || 0), 0);
        
        console.log('📊 Итоговые данные:', { 
            products: totalProducts, 
            users: totalUsers, 
            orders: totalOrders, 
            revenue,
            completedOrders: completedOrders.length
        });
        
        console.log('🎯 Обновляем счетчики:', { products: totalProducts, users: totalUsers, orders: totalOrders, revenue });
        
        // Обновляем UI для shop_owner
        if (roleName === 'shop_owner') {
            const usersLabel = document.querySelector('#total-users').parentElement.querySelector('p');
            if (usersLabel) {
                usersLabel.textContent = 'Подписчики';
            }
        }
        
        animateCounter('total-products', totalProducts);
        animateCounter('total-users', totalUsers);
        animateCounter('total-orders', totalOrders);
        animateRevenue('total-revenue', revenue);
        
        // Скрываем список последних заказов для shop_owner
        const recentSection = document.querySelector('.recent-section');
        if (roleName === 'shop_owner' && recentSection) {
            recentSection.style.display = 'none';
        } else if (recentSection) {
            recentSection.style.display = 'block';
            displayRecentOrders(ordersList.slice(0, 5));
        }
        
        // Делаем карточки кликабельными
        setupDashboardCards(roleName);
        
        // Загружаем уведомления
        await loadNotifications();
        
        console.log('✅ Дашборд загружен успешно');
        
    } catch (error) {
        console.error('❌ Ошибка загрузки дашборда:', error);
        if (window.ui && window.ui.showMessage) {
            window.ui.showMessage('Ошибка загрузки данных дашборда: ' + error.message, 'error');
        }
    }
}

// Анимация счетчиков
function animateCounter(elementId, targetValue) {
    const element = document.getElementById(elementId);
    if (!element) {
        console.warn(`⚠️ Элемент ${elementId} не найден для анимации счетчика`);
        return;
    }
    
    const startValue = 0;
    const duration = 1000;
    const startTime = performance.now();
    
    function updateCounter(currentTime) {
        const elapsed = currentTime - startTime;
        const progress = Math.min(elapsed / duration, 1);
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
    const duration = 1000;
    const startTime = performance.now();
    
    function updateRevenue(currentTime) {
        const elapsed = currentTime - startTime;
        const progress = Math.min(elapsed / duration, 1);
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
    
    // Функции для статусов
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
    
    const table = `
        <table class="data-table">
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
                    <tr data-order-id="${order.id}">
                        <td data-label="ID"><code>${order.id?.substring(0, 8)}...</code></td>
                        <td data-label="Пользователь">
                            <div class="user-info">
                                <i class="fas fa-user-circle"></i>
                                <span>${order.user_id?.substring(0, 8)}...</span>
                            </div>
                        </td>
                        <td data-label="Статус">
                            <span class="status-badge ${getStatusClass(order.status)}">
                                <i class="fas ${getStatusIcon(order.status)}"></i>
                                ${order.status || 'Новый'}
                            </span>
                        </td>
                        <td data-label="Сумма" class="amount">
                            <strong>₽${(order.total_amount || 0).toLocaleString()}</strong>
                        </td>
                        <td data-label="Дата">
                            <div class="date-info">
                                <div class="date">${new Date(order.created_at).toLocaleDateString()}</div>
                                <div class="time">${new Date(order.created_at).toLocaleTimeString()}</div>
                            </div>
                        </td>
                        <td data-label="Действия">
                            <button class="action-btn view" onclick="window.orders && window.orders.viewOrderDetails ? window.orders.viewOrderDetails('${order.id}') : (typeof viewOrderDetails === 'function' ? viewOrderDetails('${order.id}') : alert('Функция просмотра недоступна'))" title="Просмотр">
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

// Настройка кликабельных карточек дашборда
function setupDashboardCards(roleName) {
    const productsCard = document.querySelector('.stat-card:nth-child(1)');
    const usersCard = document.querySelector('.stat-card:nth-child(2)');
    const ordersCard = document.querySelector('.stat-card:nth-child(3)');
    const revenueCard = document.querySelector('.stat-card:nth-child(4)');
    
    if (productsCard) {
        productsCard.style.cursor = 'pointer';
        productsCard.addEventListener('click', () => {
            const productsTab = document.querySelector('[onclick*="showTab"]');
            if (productsTab) {
                const event = new Event('click');
                document.querySelector('[onclick*="products"]')?.dispatchEvent(event);
            } else {
                // Альтернативный способ перехода
                window.location.hash = '#products';
                if (window.products && window.products.loadProducts) {
                    window.products.loadProducts();
                }
            }
        });
    }
    
    if (usersCard && roleName === 'shop_owner') {
        usersCard.style.cursor = 'pointer';
        usersCard.addEventListener('click', () => {
            // Для shop_owner можно перейти на страницу подписчиков или клиентов
            console.log('Переход к подписчикам');
        });
    }
    
    if (ordersCard) {
        ordersCard.style.cursor = 'pointer';
        ordersCard.addEventListener('click', () => {
            const ordersTab = document.querySelector('[onclick*="orders"]');
            if (ordersTab) {
                ordersTab.click();
            } else {
                window.location.hash = '#orders';
                if (window.orders && window.orders.loadOrders) {
                    window.orders.loadOrders();
                }
            }
        });
    }
    
    if (revenueCard) {
        revenueCard.style.cursor = 'pointer';
        revenueCard.addEventListener('click', () => {
            const ordersTab = document.querySelector('[onclick*="orders"]');
            if (ordersTab) {
                ordersTab.click();
            } else {
                window.location.hash = '#orders';
                if (window.orders && window.orders.loadOrders) {
                    window.orders.loadOrders();
                }
            }
        });
    }
}

// Загрузка уведомлений
async function loadNotifications() {
    try {
        const notifications = await window.api.fetchData('/api/v1/notifications/?limit=10&isRead=false');
        const unreadCount = await window.api.fetchData('/api/v1/notifications/unread-count');
        
        const notificationsList = Array.isArray(notifications?.data) ? notifications.data 
            : Array.isArray(notifications?.notifications) ? notifications.notifications : [];
        const count = unreadCount?.data?.unreadCount || unreadCount?.unreadCount || 0;
        
        displayNotifications(notificationsList);
        updateUnreadCountBadge(count);
    } catch (error) {
        console.warn('⚠️ Ошибка загрузки уведомлений:', error.message);
        const container = document.getElementById('notifications-list');
        if (container) {
            container.innerHTML = '<p style="text-align:center;padding:20px;color:#777;">Ошибка загрузки уведомлений</p>';
        }
    }
}

// Отображение уведомлений
function displayNotifications(notifications) {
    const container = document.getElementById('notifications-list');
    if (!container) return;
    
    if (notifications.length === 0) {
        container.innerHTML = `
            <div class="empty-state">
                <i class="fas fa-bell-slash"></i>
                <p>Нет непрочитанных уведомлений</p>
            </div>
        `;
        return;
    }
    
    const notificationsHTML = notifications.map(notif => {
        const date = new Date(notif.timestamp || notif.createdAt);
        const typeIcon = {
            'order': 'fa-shopping-cart',
            'promotion': 'fa-tag',
            'system': 'fa-info-circle',
            'reminder': 'fa-clock'
        }[notif.type] || 'fa-bell';
        
        const typeColor = {
            'order': '#667eea',
            'promotion': '#f5576c',
            'system': '#4ecdc4',
            'reminder': '#ffa726'
        }[notif.type] || '#666';
        
        return `
            <div class="notification-item" style="
                background: white;
                border-radius: 10px;
                padding: 15px;
                margin-bottom: 10px;
                box-shadow: 0 2px 8px rgba(0,0,0,0.1);
                cursor: pointer;
                transition: all 0.3s ease;
                border-left: 4px solid ${typeColor};
            " onclick="window.dashboard && window.dashboard.openNotification('${notif.id}', '${notif.actionUrl || ''}')">
                <div style="display: flex; align-items: start; gap: 15px;">
                    <div style="
                        width: 40px;
                        height: 40px;
                        border-radius: 50%;
                        background: ${typeColor};
                        display: flex;
                        align-items: center;
                        justify-content: center;
                        color: white;
                        font-size: 18px;
                        flex-shrink: 0;
                    ">
                        <i class="fas ${typeIcon}"></i>
                    </div>
                    <div style="flex: 1;">
                        <div style="font-weight: 600; color: #333; margin-bottom: 5px; font-size: 14px;">
                            ${notif.title || 'Уведомление'}
                        </div>
                        <div style="color: #666; font-size: 13px; margin-bottom: 8px;">
                            ${notif.body || ''}
                        </div>
                        <div style="color: #999; font-size: 11px;">
                            ${date.toLocaleString('ru-RU')}
                        </div>
                    </div>
                    ${!notif.isRead ? `
                        <div style="
                            width: 10px;
                            height: 10px;
                            border-radius: 50%;
                            background: #ff6b6b;
                            flex-shrink: 0;
                            margin-top: 5px;
                        "></div>
                    ` : ''}
                </div>
            </div>
        `;
    }).join('');
    
    container.innerHTML = notificationsHTML;
    
    // Добавляем hover эффект
    const items = container.querySelectorAll('.notification-item');
    items.forEach(item => {
        item.addEventListener('mouseenter', function() {
            this.style.transform = 'translateX(5px)';
            this.style.boxShadow = '0 4px 12px rgba(0,0,0,0.15)';
        });
        item.addEventListener('mouseleave', function() {
            this.style.transform = 'translateX(0)';
            this.style.boxShadow = '0 2px 8px rgba(0,0,0,0.1)';
        });
    });
}

// Обновление бейджа непрочитанных
function updateUnreadCountBadge(count) {
    const badge = document.getElementById('unread-count-badge');
    if (badge) {
        if (count > 0) {
            badge.textContent = count;
            badge.style.display = 'inline-block';
        } else {
            badge.style.display = 'none';
        }
    }
}

// Открытие уведомления
async function openNotification(notificationId, actionUrl) {
    try {
        // Отмечаем как прочитанное
        await window.api.fetchData(`/api/v1/notifications/${notificationId}/read`, {
            method: 'PUT'
        });
        
        // Перезагружаем уведомления
        await loadNotifications();
        
        // Если есть actionUrl, переходим по нему
        if (actionUrl) {
            // Можно открыть в новой вкладке или перейти на страницу
            window.location.href = actionUrl;
        }
    } catch (error) {
        console.error('❌ Ошибка открытия уведомления:', error);
    }
}

// Отметить все уведомления как прочитанные
async function markAllNotificationsRead() {
    try {
        await window.api.fetchData('/api/v1/notifications/read-all', {
            method: 'PUT'
        });
        
        // Перезагружаем уведомления
        await loadNotifications();
        
        if (window.ui && window.ui.showMessage) {
            window.ui.showMessage('Все уведомления отмечены как прочитанные', 'success');
        }
    } catch (error) {
        console.error('❌ Ошибка отметки уведомлений:', error);
        if (window.ui && window.ui.showMessage) {
            window.ui.showMessage('Ошибка при отметке уведомлений', 'error');
        }
    }
}

// Экспорт
window.dashboard = {
    loadDashboard,
    displayRecentOrders,
    animateCounter,
    animateRevenue,
    setupDashboardCards,
    loadNotifications,
    displayNotifications,
    updateUnreadCountBadge,
    openNotification,
    markAllNotificationsRead
};


