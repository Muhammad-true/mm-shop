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
            } else if (roleName === 'shop_owner') {
                users = await window.api.fetchData('/api/v1/shop/customers/');
                console.log('✅ Клиенты загружены:', users.data?.customers?.length || 0);
            }
        } catch (error) {
            console.warn('⚠️ Ошибка загрузки пользователей/клиентов:', error.message);
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
        const productsList = products?.products || products?.data?.products || [];
        const usersList = (roleName === 'super_admin' || roleName === 'admin') 
            ? (users?.data?.users || users?.users || [])
            : (users?.data?.customers || users?.customers || []);
        const ordersList = orders?.data?.orders || orders?.orders || orders?.data || [];
        
        const totalProducts = productsList.length;
        const totalUsers = usersList.length;
        const totalOrders = ordersList.length;
        const revenue = ordersList.reduce((sum, order) => sum + (order.total_amount || 0), 0) || 0;
        
        console.log('📊 Итоговые данные:', { 
            products: totalProducts, 
            users: totalUsers, 
            orders: totalOrders, 
            revenue,
            ordersList: ordersList.slice(0, 5)
        });
        
        console.log('🎯 Обновляем счетчики:', { products: totalProducts, users: totalUsers, orders: totalOrders, revenue });
        
        animateCounter('total-products', totalProducts);
        animateCounter('total-users', totalUsers);
        animateCounter('total-orders', totalOrders);
        animateRevenue('total-revenue', revenue);
        
        displayRecentOrders(ordersList.slice(0, 5));
        
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

// Экспорт
window.dashboard = {
    loadDashboard,
    displayRecentOrders,
    animateCounter,
    animateRevenue
};


