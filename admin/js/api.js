// ===== API.JS - Работа с API =====

// Основная функция для API запросов
async function fetchData(endpoint, options = {}) {
    const API_BASE_URL = window.getApiUrl ? window.getApiUrl('') : CONFIG.API.BASE_URL;
    const adminToken = (window.storage && window.storage.getAdminToken) ? window.storage.getAdminToken() : (window.getAdminToken ? window.getAdminToken() : null);
    
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

// CRUD операции для продуктов
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
    return await fetchData(`/api/v1/shop/products/${id}/`, {
        method: 'DELETE'
    });
}

// CRUD операции для категорий
async function createCategory(data) {
    return await fetchData('/api/v1/admin/categories/', {
        method: 'POST',
        body: JSON.stringify(data)
    });
}

async function updateCategory(id, data) {
    return await fetchData(`/api/v1/admin/categories/${id}/`, {
        method: 'PUT',
        body: JSON.stringify(data)
    });
}

async function deleteCategory(id) {
    return await fetchData(`/api/v1/admin/categories/${id}/`, {
        method: 'DELETE'
    });
}

// CRUD операции для пользователей
async function createUser(data) {
    return await fetchData('/api/v1/admin/users/', {
        method: 'POST',
        body: JSON.stringify(data)
    });
}

async function updateUser(id, data) {
    return await fetchData(`/api/v1/admin/users/${id}/`, {
        method: 'PUT',
        body: JSON.stringify(data)
    });
}

async function deleteUser(id) {
    return await fetchData(`/api/v1/admin/users/${id}/`, {
        method: 'DELETE'
    });
}

// Тестирование подключения к API
async function testConnection() {
    try {
        const response = await fetch(getApiUrl(CONFIG.API.ENDPOINTS.HEALTH));
        const data = await response.json();
        
        if (response.ok) {
            return { success: true, data };
        } else {
            return { success: false, error: 'API сервер не отвечает' };
        }
    } catch (error) {
        console.error('Ошибка подключения:', error);
        return { success: false, error: error.message };
    }
}

// Экспорт функций
window.api = {
    fetchData,
    createProduct,
    updateProduct,
    deleteProduct,
    createCategory,
    updateCategory,
    deleteCategory,
    createUser,
    updateUser,
    deleteUser,
    testConnection
};

// Делаем testConnection глобальной для обратной совместимости
window.testConnection = async function() {
    const result = await window.api.testConnection();
    if (result.success) {
        if (window.ui && window.ui.showMessage) {
            window.ui.showMessage('✅ API подключение работает!', 'success');
        } else {
            alert('✅ API подключение работает!');
        }
    } else {
        if (window.ui && window.ui.showMessage) {
            window.ui.showMessage(`❌ Ошибка подключения: ${result.error}`, 'error');
        } else {
            alert(`❌ Ошибка подключения: ${result.error}`);
        }
    }
    return result;
};

