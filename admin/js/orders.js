// ===== ORDERS.JS - Управление заказами =====

// Глобальные переменные для заказов
let currentOrdersPage = 1;
let currentOrdersFilters = {};
let ordersStats = {};

// Загрузка заказов
async function loadOrders(page = 1, filters = {}) {
    try {
        currentOrdersPage = page;
        currentOrdersFilters = filters;
        // Сброс кэша владельцев магазинов и очистка контейнера перед новой загрузкой
        window.shopOwners = undefined;
        const container = document.getElementById('orders-table');
        if (container) {
            container.innerHTML = '<p style="text-align:center;padding:20px;color:#777;">Загрузка заказов...</p>';
        }

        const userRole = localStorage.getItem('userRole') || 'admin';
        console.log('🔍 Роль пользователя для загрузки заказов:', userRole);
        
        let endpoint;
        if (userRole === 'super_admin' || userRole === 'admin') {
            endpoint = '/api/v1/admin/orders';
            console.log('👑 Используем админский эндпоинт для заказов');
        } else if (userRole === 'shop_owner') {
            endpoint = '/api/v1/shop/orders/';
            console.log('🏪 Используем эндпоинт владельца магазина для заказов');
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
        
        const response = await window.api.fetchData(fullEndpoint);
        console.log('📦 Ответ от сервера:', response);
        console.log('📊 Структура ответа:', {
            success: response.success,
            hasData: !!response.data,
            ordersCount: response.data?.orders?.length || 0,
            pagination: response.data?.pagination
        });
        
        if (response.data) {
            if (response.data.shop_owners) {
                window.shopOwners = response.data.shop_owners;
            }
            const orders = response.data.orders || [];
            console.log('✅ Заказов к отображению:', orders.length);
            
            // Логируем первый заказ для проверки
            if (orders.length > 0) {
                console.log('📋 Пример первого заказа:', {
                    id: orders[0].id,
                    user_id: orders[0].user_id,
                    items_count: orders[0].order_items?.length || 0,
                    total: orders[0].total_amount
                });
            }
            
            displayOrders(orders, response.data.pagination, response.data.stats);
        } else {
            console.warn('⚠️ response.data отсутствует!');
            displayOrders([], {}, {});
        }
    } catch (error) {
        console.error('❌ Ошибка загрузки заказов:', error);
        if (window.ui && window.ui.showMessage) {
            window.ui.showMessage('Ошибка загрузки заказов: ' + error.message, 'error');
        }
    }
}

// Отображение заказов
function displayOrders(orders, pagination = {}, stats = {}) {
    console.log('📊 Отображение заказов:', {
        ordersCount: orders.length,
        pagination: pagination,
        stats: stats
    });
    
    const container = document.getElementById('orders-table');
    
    if (!container) {
        console.warn('❌ Контейнер orders-table не найден');
        return;
    }
    
    ordersStats = stats;
    
    if (orders.length === 0) {
        container.innerHTML = `
            <p style="text-align: center; padding: 40px;">Заказов не найдено</p>
        `;
        return;
    }
    
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
        <div class="table-container">
            <h3><i class="fas fa-shopping-cart"></i> Список заказов</h3>
            <div class="table-responsive">
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
                        <th>Действия</th>
                    </tr>
                </thead>
                <tbody>
                    ${orders.map(order => `
                        <tr data-order-id="${order.id}">
                            <td data-label="№ Заказа"><strong>${order.order_number || order.id?.substring(0, 8)}</strong></td>
                            <td data-label="Клиент">
                                <div>
                                    <div><strong>${order.recipient_name || order.user?.name || 'N/A'}</strong></div>
                                    ${order.user?.is_guest ? '<small style="color: #999;">🎭 Гость</small>' : ''}
                                </div>
                            </td>
                            <td data-label="Телефон"><a href="tel:${order.phone}" style="color: #667eea; text-decoration: none;">${order.phone}</a></td>
                            <td data-label="Магазин">
                                <div>
                                    <div><strong>${order.shop_owner?.name || 'N/A'}</strong></div>
                                    <small style="color: #999;">${order.shop_owner?.phone || ''}</small>
                                </div>
                            </td>
                            <td data-label="Товары">${order.order_items?.length || 0} шт.</td>
                            <td data-label="Сумма"><strong>${order.total_amount || 0} ${order.currency || 'TJS'}</strong></td>
                            <td data-label="Статус">
                                <span class="status-badge ${order.status}" style="padding: 5px 10px; border-radius: 12px; font-size: 12px; font-weight: 600;">
                                    ${statusLabels[order.status] || order.status}
                                </span>
                            </td>
                            <td data-label="Дата">${new Date(order.created_at).toLocaleString('ru-RU', {day: '2-digit', month: '2-digit', year: 'numeric', hour: '2-digit', minute: '2-digit'})}</td>
                            <td data-label="Действия">
                                <div class="action-buttons" style="display: flex; gap: 5px;">
                                    <button class="action-btn view" onclick="viewOrderDetails('${order.id}')" title="Просмотр">
                                        <i class="fas fa-eye"></i>
                                    </button>
                                    <button class="action-btn edit" onclick="changeOrderStatus('${order.id}')" title="Изменить статус">
                                        <i class="fas fa-edit"></i>
                                    </button>
                                    ${order.status === 'pending' ? `
                                        <button class="action-btn success" onclick="confirmOrder('${order.id}')" title="Подтвердить">
                                            <i class="fas fa-check"></i>
                                        </button>
                                        <button class="action-btn danger" onclick="rejectOrder('${order.id}')" title="Отклонить">
                                            <i class="fas fa-times"></i>
                                        </button>
                                    ` : ''}
                                </div>
                            </td>
                        </tr>
                    `).join('')}
                </tbody>
                </table>
            </div>
        </div>
        ${pagination.totalPages > 1 ? createPagination(pagination) : ''}
    `;
    
    container.innerHTML = table;
}

// Создание пагинации
function createPagination(pagination) {
    const { page, totalPages } = pagination;
    let pages = '';
    
    for (let i = 1; i <= totalPages; i++) {
        if (i === page) {
            pages += `<button class="pagination-btn active">${i}</button>`;
        } else if (i === 1 || i === totalPages || (i >= page - 2 && i <= page + 2)) {
            pages += `<button class="pagination-btn" onclick="window.orders.loadOrders(${i}, currentOrdersFilters)">${i}</button>`;
        } else if (i === page - 3 || i === page + 3) {
            pages += `<span>...</span>`;
        }
    }
    
    return `
        <div class="pagination" style="display: flex; justify-content: center; gap: 5px; margin-top: 20px; padding: 20px;">
            ${page > 1 ? `<button class="pagination-btn" onclick="window.orders.loadOrders(${page - 1}, currentOrdersFilters)"><i class="fas fa-chevron-left"></i></button>` : ''}
            ${pages}
            ${page < totalPages ? `<button class="pagination-btn" onclick="window.orders.loadOrders(${page + 1}, currentOrdersFilters)"><i class="fas fa-chevron-right"></i></button>` : ''}
        </div>
    `;
}

// Подтверждение заказа
async function confirmOrder(orderId) {
    if (!confirm('Подтвердить этот заказ?')) return;
    
    try {
        const userRole = localStorage.getItem('userRole') || 'admin';
        let endpoint, method, body;
        
        if (userRole === 'shop_owner') {
            // Владелец магазина использует PUT /shop/orders/:id/status
            endpoint = `/api/v1/shop/orders/${orderId}/status`;
            method = 'PUT';
            body = JSON.stringify({ status: 'confirmed' });
        } else {
            // Админ использует POST /admin/orders/:id/confirm
            endpoint = `/api/v1/admin/orders/${orderId}/confirm`;
            method = 'POST';
            body = null;
        }
        
        const response = await window.api.fetchData(endpoint, {
            method: method,
            body: body
        });
        
        if (response.success) {
            if (window.ui && window.ui.showMessage) {
                window.ui.showMessage('Заказ подтвержден!', 'success');
            }
            loadOrders(currentOrdersPage, currentOrdersFilters);
        } else {
            if (window.ui && window.ui.showMessage) {
                window.ui.showMessage(response.message || 'Ошибка при подтверждении заказа', 'error');
            }
        }
    } catch (error) {
        console.error('Ошибка:', error);
        if (window.ui && window.ui.showMessage) {
            window.ui.showMessage('Ошибка при подтверждении заказа', 'error');
        }
    }
}

// Отклонение заказа
async function rejectOrder(orderId) {
    if (!confirm('Отклонить этот заказ?')) return;
    
    try {
        const userRole = localStorage.getItem('userRole') || 'admin';
        let endpoint, method, body;
        
        if (userRole === 'shop_owner') {
            // Владелец магазина использует PUT /shop/orders/:id/status
            endpoint = `/api/v1/shop/orders/${orderId}/status`;
            method = 'PUT';
            body = JSON.stringify({ status: 'cancelled' });
        } else {
            // Админ использует POST /admin/orders/:id/reject
            endpoint = `/api/v1/admin/orders/${orderId}/reject`;
            method = 'POST';
            body = null;
        }
        
        const response = await window.api.fetchData(endpoint, {
            method: method,
            body: body
        });
        
        if (response.success) {
            if (window.ui && window.ui.showMessage) {
                window.ui.showMessage('Заказ отклонен', 'success');
            }
            loadOrders(currentOrdersPage, currentOrdersFilters);
        } else {
            if (window.ui && window.ui.showMessage) {
                window.ui.showMessage(response.message || 'Ошибка при отклонении заказа', 'error');
            }
        }
    } catch (error) {
        console.error('Ошибка:', error);
        if (window.ui && window.ui.showMessage) {
            window.ui.showMessage('Ошибка при отклонении заказа', 'error');
        }
    }
}

// Просмотр деталей заказа
async function viewOrderDetails(orderId) {
    try {
        const response = await window.api.fetchData(`/api/v1/admin/orders/${orderId}`);
        
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
            
            if (window.ui && window.ui.showModal) {
                window.ui.showModal('Детали заказа', `
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
                                <p><strong>Дата создания:</strong> ${new Date(order.created_at).toLocaleString('ru-RU')}</p>
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
                `);
            }
        }
    } catch (error) {
        console.error('Ошибка загрузки деталей заказа:', error);
        if (window.ui && window.ui.showMessage) {
            window.ui.showMessage('Ошибка загрузки деталей заказа', 'error');
        }
    }
}

// Изменение статуса заказа
async function changeOrderStatus(orderId) {
    try {
        const userRole = localStorage.getItem('userRole') || 'admin';
        
        // Определяем endpoint в зависимости от роли
        let getOrderEndpoint;
        if (userRole === 'shop_owner') {
            getOrderEndpoint = `/api/v1/shop/orders/${orderId}`;
        } else {
            getOrderEndpoint = `/api/v1/admin/orders/${orderId}`;
        }
        
        const response = await window.api.fetchData(getOrderEndpoint);
        const order = response.data;
        
        const statusOptions = {
            pending: 'Ожидает',
            confirmed: 'Подтвержден',
            preparing: 'Готовится',
            inDelivery: 'В доставке',
            delivered: 'Доставлен',
            completed: 'Завершен',
            cancelled: 'Отменен'
        };
        
        const statusKeys = Object.keys(statusOptions);
        
        const modal = document.createElement('div');
        modal.className = 'modal';
        modal.style.display = 'block';
        
        modal.innerHTML = `
            <div class="modal-content" style="max-width: 500px;">
                <div class="modal-header">
                    <h3><i class="fas fa-edit"></i> Изменить статус заказа</h3>
                    <span class="close" onclick="this.closest('.modal').remove()">&times;</span>
                </div>
                <div style="padding: 20px;">
                    <p><strong>Текущий статус:</strong> ${statusOptions[order.status] || order.status}</p>
                    <label for="new-status"><strong>Новый статус:</strong></label>
                    <select id="new-status" class="form-input" style="width: 100%; margin: 10px 0;">
                        ${statusKeys.map(key => 
                            `<option value="${key}" ${key === order.status ? 'selected' : ''}>${statusOptions[key]}</option>`
                        ).join('')}
                    </select>
                </div>
                <div class="modal-actions">
                    <button type="button" class="btn btn-primary" onclick="saveOrderStatus('${orderId}')">Сохранить</button>
                    <button type="button" class="btn btn-secondary" onclick="this.closest('.modal').remove()">Отмена</button>
                </div>
            </div>
        `;
        
        document.body.appendChild(modal);
        
        // Сохранение статуса
        window.saveOrderStatus = async function(id) {
            const newStatus = document.getElementById('new-status').value;
            
            try {
                // Определяем endpoint в зависимости от роли
                let updateEndpoint;
                if (userRole === 'shop_owner') {
                    updateEndpoint = `/api/v1/shop/orders/${id}/status`;
                } else {
                    updateEndpoint = `/api/v1/admin/orders/${id}/status`;
                }
                
                const response = await window.api.fetchData(updateEndpoint, {
                    method: 'PUT',
                    body: JSON.stringify({ status: newStatus })
                });
                
                if (window.ui && window.ui.showMessage) {
                    window.ui.showMessage('Статус заказа успешно изменен', 'success');
                }
                
                document.querySelector('.modal').remove();
                loadOrders(currentOrdersPage, currentOrdersFilters);
            } catch (error) {
                console.error('Ошибка изменения статуса:', error);
                if (window.ui && window.ui.showMessage) {
                    window.ui.showMessage('Ошибка изменения статуса заказа', 'error');
                }
            }
        };
        
    } catch (error) {
        console.error('Ошибка загрузки данных заказа:', error);
        if (window.ui && window.ui.showMessage) {
            window.ui.showMessage('Ошибка загрузки данных заказа', 'error');
        }
    }
}

// Настройка фильтров для заказов
function setupOrderFilters() {
    const searchInput = document.getElementById('order-search');
    const statusFilter = document.getElementById('order-status-filter');
    const dateFromFilter = document.getElementById('order-date-from');
    const dateToFilter = document.getElementById('order-date-to');
    
    if (searchInput) {
        searchInput.addEventListener('input', () => applyOrderFilters());
    }
    
    if (statusFilter) {
        statusFilter.addEventListener('change', () => applyOrderFilters());
    }
    
    if (dateFromFilter) {
        dateFromFilter.addEventListener('change', () => applyOrderFilters());
    }
    
    if (dateToFilter) {
        dateToFilter.addEventListener('change', () => applyOrderFilters());
    }
    
    console.log('✅ Фильтры заказов настроены');
}

// Применение фильтров для заказов
function applyOrderFilters() {
    const filters = {
        search: document.getElementById('order-search')?.value || '',
        status: document.getElementById('order-status-filter')?.value || '',
        dateFrom: document.getElementById('order-date-from')?.value || '',
        dateTo: document.getElementById('order-date-to')?.value || ''
    };
    
    console.log('🔍 Применение фильтров заказов:', filters);
    
    // Перезагружаем заказы с новыми фильтрами
    loadOrders(1, filters);
}

// Глобальная функция для обратной совместимости
window.loadOrders = loadOrders;

// Экспорт
window.orders = {
    loadOrders,
    displayOrders,
    confirmOrder,
    rejectOrder,
    viewOrderDetails,
    changeOrderStatus,
    createPagination,
    setupOrderFilters,
    applyOrderFilters
};

// Глобальные функции для обратной совместимости
window.confirmOrder = confirmOrder;
window.rejectOrder = rejectOrder;
window.viewOrderDetails = viewOrderDetails;
window.changeOrderStatus = changeOrderStatus;

