// ===== PRODUCTS.JS - Управление товарами =====

// Получаем переменные из storage
function getVariations() {
    return window.storage ? window.storage.getVariations() : (typeof variations !== 'undefined' ? variations : []);
}

function setVariations(vars) {
    if (window.storage) {
        window.storage.setVariations(vars);
    } else if (typeof variations !== 'undefined') {
        variations = vars;
    }
}

// Загрузка товаров
async function loadProducts() {
    console.log('🔄 Начинаем загрузку товаров...');
    
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
            endpoint = CONFIG.API.ENDPOINTS.PRODUCTS.LIST;
            title = 'Список всех товаров (Админ)';
            console.log('👑 Админ загружает все товары');
        } else {
            endpoint = CONFIG.API.ENDPOINTS.PRODUCTS.LIST;
            title = 'Мои товары';
            console.log('🏪 Владелец магазина загружает все товары, фильтруем по ownerId');
        }
        
        console.log(`🔗 Используем эндпоинт: ${endpoint}`);
        
        const response = await window.api.fetchData(endpoint);
        console.log('📡 Ответ API товаров:', response);
        
        // Проверяем разные возможные форматы ответа
        let products = [];
        
        if (response.data && response.data.products && Array.isArray(response.data.products)) {
            products = response.data.products;
        } else if (response.products && Array.isArray(response.products)) {
            products = response.products;
        } else if (response.data && Array.isArray(response.data)) {
            products = response.data;
        } else if (Array.isArray(response)) {
            products = response;
        } else {
            products = [];
        }
        
        console.log(`📦 Получено ${products.length} товаров из API`);
        
        // Фильтруем товары по ownerId для владельцев магазинов
        if (userRole === 'shop_owner' || userRole === 'user') {
            const userData = JSON.parse(localStorage.getItem('userData'));
            const userId = userData?.id;
            
            if (userId) {
                const originalCount = products.length;
                products = products.filter(product => product.ownerId === userId);
                console.log(`✅ Отфильтровано: ${originalCount} → ${products.length} товаров`);
            }
        }
        
        // Сохраняем товары в глобальную переменную для фильтрации
        if (window.storage && window.storage.setAllProducts) {
            window.storage.setAllProducts(products);
        } else {
            allProducts = products;
        }
        
        if (!container) {
            console.error('❌ Контейнер products-table не найден!');
            return;
        }
        
        // Обновляем заголовок
        if (container) {
            const titleElement = container.querySelector('h3');
            if (titleElement) {
                titleElement.innerHTML = `<i class="fas fa-box"></i> ${title}`;
            }
        }
        
        displayProducts(products);
        
        if (products.length > 0) {
            const roleText = userRole === 'shop_owner' ? 'ваших' : '';
            if (window.ui && window.ui.showMessage) {
                window.ui.showMessage(`Успешно загружено ${products.length} ${roleText} товаров`, 'success');
            }
        }
        
    } catch (error) {
        console.error('❌ Ошибка загрузки товаров:', error);
        
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
                        <button class="btn btn-primary" onclick="window.products.loadProducts()">
                            <i class="fas fa-redo"></i> Попробовать снова
                        </button>
                    </div>
                </div>
            `;
        }
        
        if (window.ui && window.ui.showMessage) {
            window.ui.showMessage('Ошибка загрузки товаров: ' + error.message, 'error');
        }
    }
}

// Обновление списка товаров
async function refreshProductsList() {
    console.log('🔄 Принудительное обновление списка товаров...');
    
    const productsTab = document.getElementById('products');
    if (!productsTab || !productsTab.classList.contains('active')) {
        showTab('products');
        await new Promise(resolve => setTimeout(resolve, 150));
    }
    
    try {
        await loadProducts();
        console.log('✅ Список товаров успешно обновлен');
        return true;
    } catch (error) {
        console.error('❌ Ошибка обновления списка товаров:', error);
        if (window.ui && window.ui.showMessage) {
            window.ui.showMessage('Не удалось обновить список товаров: ' + error.message, 'error');
        }
        return false;
    }
}

// Отображение товаров
function displayProducts(products) {
    console.log('🔄 Отображение товаров:', products);
    
    const container = document.getElementById('products-table');
    
    if (!container) {
        console.error('❌ Контейнер products-table не найден!');
        return;
    }
    
    if (!Array.isArray(products)) {
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
    
    // Вспомогательные функции для пола
    function getGenderColor(gender) {
        switch (gender?.toLowerCase()) {
            case 'male': return 'linear-gradient(135deg, #4ecdc4, #44a08d)';
            case 'female': return 'linear-gradient(135deg, #f093fb, #f5576c)';
            default: return 'linear-gradient(135deg, #45b7d1, #96ceb4)';
        }
    }
    
    function getGenderIcon(gender) {
        switch (gender?.toLowerCase()) {
            case 'male': return 'fa-mars';
            case 'female': return 'fa-venus';
            default: return 'fa-venus-mars';
        }
    }
    
    function getGenderText(gender) {
        switch (gender?.toLowerCase()) {
            case 'male': return 'Мужской';
            case 'female': return 'Женский';
            default: return 'Унисекс';
        }
    }
    
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
                                        <div style="width: 50px; height: 50px; border-radius: 12px; background: linear-gradient(135deg, #667eea, #764ba2); display: flex; align-items: center; justify-content: center; color: white; font-weight: bold; font-size: 20px; box-shadow: 0 4px 15px rgba(102, 126, 234, 0.3);">
                                            <i class="fas fa-box"></i>
                                        </div>
                                        <div>
                                            <div style="font-weight: 700; color: #333; font-size: 16px; margin-bottom: 4px;">${product.name}</div>
                                            <div style="font-size: 12px; color: #888; font-family: monospace; background: #f8f9fa; padding: 2px 6px; border-radius: 4px;">${product.id?.substring(0, 8)}...</div>
                                        </div>
                                    </div>
                                </td>
                                <td data-label="Бренд">${product.brand || 'Не указан'}</td>
                                <td data-label="Пол">
                                    <span class="badge" style="background: ${getGenderColor(product.gender)}; font-size: 12px; padding: 8px 12px;">
                                        <i class="fas ${getGenderIcon(product.gender)}"></i>
                                        ${getGenderText(product.gender)}
                                    </span>
                                </td>
                                <td data-label="Категория">${product.category?.name || 'Без категории'}</td>
                                <td data-label="Вариации">
                                    <div style="display: flex; flex-direction: column; gap: 4px;">
                                        <span class="badge" style="background: linear-gradient(135deg, #45b7d1, #96ceb4); color: white; font-size: 12px; padding: 8px 12px;">
                                            ${product.variations?.length || 0} вариаций
                                        </span>
                                        ${product.variations?.some(v => v.barcode) ? `
                                            <span style="font-size: 11px; color: #666;">
                                                <i class="fas fa-barcode"></i> Есть штрих-коды
                                            </span>
                                        ` : ''}
                                    </div>
                                </td>
                                <td data-label="Дата создания">${product.createdAt ? new Date(product.createdAt).toLocaleDateString('ru-RU') : 'N/A'}</td>
                                <td data-label="Действия">
                                    <div style="display: flex; gap: 8px; justify-content: center;">
                                        <button class="btn-sm btn-info" onclick="viewProductVariations('${product.id}')" title="Просмотр вариаций">
                                            <i class="fas fa-eye"></i>
                                        </button>
                                        <button class="btn-sm btn-primary" onclick="window.products.editProduct('${product.id}')" title="Редактировать">
                                            <i class="fas fa-edit"></i>
                                        </button>
                                        <button class="btn-sm btn-danger" onclick="window.products.deleteProduct('${product.id}')" title="Удалить">
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
}

// Редактирование товара
async function editProduct(id) {
    console.log('🔄 Редактирование товара с ID:', id);
    await loadProductData(id);
}

// Загрузка данных товара
async function loadProductData(id) {
    try {
        const response = await window.api.fetchData(`/api/v1/products/${id}`);
        
        let product;
        if (response.product) {
            product = response.product;
        } else if (response.data) {
            product = response.data;
        } else {
            product = response;
        }
        
        if (window.storage && window.storage.setCurrentProductId) {
            window.storage.setCurrentProductId(id);
        }
        document.getElementById('product-modal-title').textContent = 'Редактировать товар';
        
        document.getElementById('product-name').value = product.name;
        document.getElementById('product-description').value = product.description;
        document.getElementById('product-gender').value = product.gender;
        document.getElementById('product-category').value = product.categoryId;
        document.getElementById('product-brand').value = product.brand;
        
        // Загружаем вариации (убеждаемся, что barcode есть)
        const variations = (product.variations || []).map(v => ({
            ...v,
            barcode: v.barcode || ''
        }));
        setVariations(variations);
        renderVariations();
        document.getElementById('product-modal').style.display = 'block';
    } catch (error) {
        console.error('Ошибка загрузки данных товара:', error);
        alert('Ошибка загрузки данных товара');
    }
}

// Удаление товара
async function deleteProduct(id) {
    if (!window.ui || !window.ui.showConfirmDialog) {
        if (!confirm('Вы уверены, что хотите удалить этот товар?')) {
            return;
        }
    } else {
        const confirmed = await window.ui.showConfirmDialog(
            'Удаление товара',
            'Вы уверены, что хотите удалить этот товар?',
            'Это действие нельзя отменить.',
            'Удалить',
            'Отмена'
        );
        
        if (!confirmed) return;
    }
    
    try {
        if (window.ui && window.ui.showMessage) {
            window.ui.showMessage('Удаление товара...', 'info');
        }
        
        // Определяем endpoint в зависимости от роли
        const userRole = localStorage.getItem('userRole') || 'admin';
        let endpoint;
        
        if (userRole === 'super_admin' || userRole === 'admin') {
            endpoint = `/api/v1/admin/products/${id}`;
        } else if (userRole === 'shop_owner') {
            endpoint = `/api/v1/shop/products/${id}`;
        } else {
            endpoint = `/api/v1/shop/products/${id}`;
        }
        
        await window.api.fetchData(endpoint, { method: 'DELETE' });
        
        if (window.ui && window.ui.showMessage) {
            window.ui.showMessage('✅ Товар успешно удален', 'success');
        }
        
        await refreshProductsList();
        
    } catch (error) {
        console.error('Ошибка удаления товара:', error);
        
        let errorMessage = 'Неизвестная ошибка';
        if (error.error && error.error.message) {
            errorMessage = error.error.message;
        } else if (error.message) {
            errorMessage = error.message;
        }
        
        if (window.ui && window.ui.showMessage) {
            window.ui.showMessage(`❌ Ошибка удаления товара: ${errorMessage}`, 'error');
        }
    }
}

// Просмотр вариаций товара
async function viewProductVariations(id) {
    try {
        console.log('👁️ Просмотр вариаций товара с ID:', id);
        
        const response = await window.api.fetchData(`/api/v1/products/${id}`);
        console.log('📡 Ответ API для товара:', response);
        
        let product;
        if (response.data) {
            product = response.data;
        } else if (response.success && response.data) {
            product = response.data;
        } else if (response.product) {
            product = response.product;
        } else {
            product = response;
        }
        
        const variations = product?.variations || product?.product?.variations || [];
        const productData = product?.product || product;
        
        console.log('📦 Данные товара:', productData);
        console.log('🎨 Вариации:', variations);
        
        if (!variations || variations.length === 0) {
            if (window.ui && window.ui.showMessage) {
                window.ui.showMessage('У этого товара нет вариаций', 'info');
            }
            return;
        }
        
        // Логируем первую вариацию для отладки
        if (variations[0]) {
            console.log('🔍 Первая вариация (структура):', {
                sizes: variations[0].sizes,
                colors: variations[0].colors,
                price: variations[0].price,
                discount: variations[0].discount,
                stock: variations[0].stockQuantity,
                sku: variations[0].sku,
                images: variations[0].imageUrls
            });
        }
        
        // Вспомогательные функции для текста пола
        function getGenderText(gender) {
            switch (gender?.toLowerCase()) {
                case 'male': return 'Мужской';
                case 'female': return 'Женский';
                default: return 'Унисекс';
            }
        }
        
        const modal = document.createElement('div');
        modal.className = 'modal';
        modal.style.display = 'block';
        
        modal.innerHTML = `
            <div class="modal-content" style="max-width: 900px; margin: 50px auto;">
                <div class="modal-header">
                    <h3><i class="fas fa-layer-group"></i> Вариации товара: ${productData.name}</h3>
                    <span class="close" onclick="this.closest('.modal').remove()">&times;</span>
                </div>
                <div style="padding: 20px; max-height: 70vh; overflow-y: auto;">
                    <div style="margin-bottom: 20px; padding: 15px; background: linear-gradient(135deg, #f8f9fa, #e9ecef); border-radius: 10px;">
                        <h4 style="margin: 0 0 10px 0; color: #2c3e50;"><i class="fas fa-box"></i> Информация о товаре</h4>
                        <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 10px; font-size: 14px;">
                            <div><strong>Название:</strong> ${productData.name}</div>
                            <div><strong>Бренд:</strong> ${productData.brand || 'Не указан'}</div>
                            <div><strong>Пол:</strong> ${getGenderText(productData.gender)}</div>
                            <div><strong>Категория:</strong> ${productData.category?.name || 'Не указана'}</div>
                        </div>
                    </div>
                    
                    <h4 style="margin: 0 0 15px 0; color: #2c3e50;"><i class="fas fa-list"></i> Вариации (${variations.length})</h4>
                    <div style="display: grid; gap: 15px;">
                        ${variations.map((variation, index) => `
                            <div style="border: 2px solid #e9ecef; border-radius: 12px; padding: 20px; background: white; position: relative; overflow: hidden;">
                                <div style="position: absolute; top: 0; left: 0; right: 0; height: 4px; background: linear-gradient(135deg, #667eea, #764ba2);"></div>
                                
                                <div style="display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 15px;">
                                    <h5 style="margin: 0; color: #2c3e50; font-size: 16px;"><i class="fas fa-tag"></i> Вариация ${index + 1}</h5>
                                    ${variation.id ? `<span class="badge" style="background: linear-gradient(135deg, #667eea, #764ba2); color: white; font-size: 12px; padding: 6px 12px;">ID: ${variation.id?.substring(0, 8)}...</span>` : ''}
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
                                        <strong style="color: #495057; font-size: 13px;">Скидка:</strong>
                                        <div style="margin-top: 5px; font-size: 16px; font-weight: bold; color: ${variation.discount > 0 ? '#e67e22' : '#6c757d'};">
                                            ${variation.discount || 0}%
                                        </div>
                                    </div>
                                    
                                    <div>
                                        <strong style="color: #495057; font-size: 13px;">Остаток:</strong>
                                        <div style="margin-top: 5px; font-size: 16px; font-weight: bold; color: ${variation.stockQuantity > 0 ? '#28a745' : '#dc3545'};">
                                            ${variation.stockQuantity || 0} шт.
                                        </div>
                                    </div>
                                </div>
                                
                                ${variation.sku ? `
                                    <div style="margin-bottom: 15px;">
                                        <strong style="color: #495057; font-size: 13px;">SKU:</strong>
                                        <span style="font-family: monospace; background: #f8f9fa; padding: 4px 8px; border-radius: 4px; font-size: 12px;">${variation.sku}</span>
                                    </div>
                                ` : ''}
                                
                                ${variation.barcode ? `
                                    <div style="margin-bottom: 15px;">
                                        <strong style="color: #495057; font-size: 13px;"><i class="fas fa-barcode"></i> Штрих-код:</strong>
                                        <span style="font-family: monospace; background: #e3f2fd; padding: 4px 8px; border-radius: 4px; font-size: 12px; font-weight: 600; color: #1976d2;">${variation.barcode}</span>
                                    </div>
                                ` : ''}
                                
                                ${(variation.imageUrls && variation.imageUrls.length > 0) ? `
                                    <div style="margin-bottom: 15px;">
                                        <strong style="color: #495057; font-size: 13px;"><i class="fas fa-images"></i> Фотографии (${variation.imageUrls.length}):</strong>
                                        <div class="variation-images-preview" style="display: flex; gap: 10px; margin-top: 10px; flex-wrap: wrap;">
                                            ${variation.imageUrls.map((url, imgIndex) => {
                                                // Получаем полный URL изображения
                                                let imageUrl = url;
                                                if (typeof window.getImageUrl === 'function') {
                                                    imageUrl = window.getImageUrl(url);
                                                } else if (typeof getImageUrl === 'function') {
                                                    imageUrl = getImageUrl(url);
                                                } else {
                                                    // Формируем URL вручную
                                                    const API_BASE_URL = window.getApiUrl ? window.getApiUrl('') : (CONFIG && CONFIG.API && CONFIG.API.BASE_URL ? CONFIG.API.BASE_URL : '');
                                                    imageUrl = url.startsWith('http') ? url : `${API_BASE_URL}${url.startsWith('/') ? url : '/' + url}`;
                                                }
                                                return `
                                                    <div class="image-preview-item" style="position: relative; border: 2px solid #e9ecef; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.1);">
                                                        <img 
                                                            src="${imageUrl}" 
                                                            alt="Photo ${imgIndex + 1}" 
                                                            onclick="window.openImageModal('${imageUrl}', 'Фото вариации ${index + 1}')"
                                                            style="width: 120px; height: 120px; object-fit: cover; cursor: pointer; transition: transform 0.2s;"
                                                            onerror="this.src='data:image/svg+xml,%3Csvg xmlns=%22http://www.w3.org/2000/svg%22 width=%22120%22 height=%22120%22%3E%3Crect fill=%22%23f8f9fa%22 width=%22120%22 height=%22120%22/%3E%3Ctext x=%2250%25%22 y=%2250%25%22 fill=%22%236c757d%22 text-anchor=%22middle%22 dy=%22.3em%22 font-size=%2214%22%3EНет фото%3C/text%3E%3C/svg%3E'; console.error('Ошибка загрузки:', '${imageUrl}');"
                                                            onmouseover="this.style.transform='scale(1.05)'"
                                                            onmouseout="this.style.transform='scale(1)'"
                                                        >
                                                        <div style="position: absolute; bottom: 0; left: 0; right: 0; background: rgba(0,0,0,0.7); color: white; font-size: 10px; padding: 2px 6px; text-align: center;">
                                                            #${imgIndex + 1}
                                                        </div>
                                                    </div>
                                                `;
                                            }).join('')}
                                        </div>
                                    </div>
                                ` : '<div style="margin-bottom: 15px; color: #6c757d; font-style: italic;"><i class="fas fa-ban"></i> Нет фотографий</div>'}
                            </div>
                        `).join('')}
                    </div>
                </div>
                <div style="padding: 20px; border-top: 1px solid #e9ecef; text-align: center;">
                    <button class="btn btn-primary" onclick="this.closest('.modal').remove()"><i class="fas fa-times"></i> Закрыть</button>
                </div>
            </div>
        `;
        
        document.body.appendChild(modal);
        
        modal.addEventListener('click', (e) => {
            if (e.target === modal) {
                modal.remove();
            }
        });
        
    } catch (error) {
        console.error('❌ Ошибка загрузки вариаций товара:', error);
        if (window.ui && window.ui.showMessage) {
            window.ui.showMessage('Ошибка загрузки вариаций товара: ' + error.message, 'error');
        }
    }
}

// Утилиты для вариаций
function addVariation() {
    const variation = {
        id: Date.now(),
        sizes: [],
        colors: [],
        price: 0,
        originalPrice: null,
        discount: 0,
        stockQuantity: 0,
        sku: '',
        barcode: '',
        imageUrls: []
    };
    
    if (window.storage && window.storage.addVariation) {
        window.storage.addVariation(variation);
    } else {
        variations.push(variation);
    }
    
    renderVariations();
}

function removeVariation(index) {
    if (window.storage && window.storage.removeVariation) {
        window.storage.removeVariation(index);
    } else {
        variations.splice(index, 1);
    }
    renderVariations();
}

function renderVariations() {
    const container = document.getElementById('variations-list');
    if (!container) return;
    
    const vars = getVariations();
    
    if (vars.length === 0) {
        container.innerHTML = '<p class="no-variations">Нет вариаций. Добавьте хотя бы одну вариацию.</p>';
        return;
    }
    
    container.innerHTML = vars.map((variation, index) => `
        <div class="variation-item" data-variation-index="${index}">
            <button type="button" class="remove-variation" onclick="window.products.removeVariation(${index})">×</button>
            <div class="variation-fields">
                <div class="variation-field">
                    <label>Тип размера</label>
                    <select onchange="window.products.changeSizeType(${index}, this.value)" class="form-input" style="margin-bottom: 10px;">
                        <option value="clothing">Одежда (XS-XXL)</option>
                        <option value="shoes">Обувь (36-46)</option>
                        <option value="pants">Штаны/Джинсы (28-40)</option>
                    </select>
                </div>
                <div class="variation-field">
                    <label>Размеры</label>
                    <div class="checkbox-group" id="sizes-container-${index}">
                        <!-- Размеры одежды (по умолчанию) -->
                        <label><input type="checkbox" value="XS" ${variation.sizes?.includes('XS') ? 'checked' : ''}> XS</label>
                        <label><input type="checkbox" value="S" ${variation.sizes?.includes('S') ? 'checked' : ''}> S</label>
                        <label><input type="checkbox" value="M" ${variation.sizes?.includes('M') ? 'checked' : ''}> M</label>
                        <label><input type="checkbox" value="L" ${variation.sizes?.includes('L') ? 'checked' : ''}> L</label>
                        <label><input type="checkbox" value="XL" ${variation.sizes?.includes('XL') ? 'checked' : ''}> XL</label>
                        <label><input type="checkbox" value="XXL" ${variation.sizes?.includes('XXL') ? 'checked' : ''}> XXL</label>
                    </div>
                </div>
                <div class="variation-field">
                    <label>Цвета</label>
                    <div class="checkbox-group" style="max-height: 200px; overflow-y: auto; padding: 5px;">
                        <label><input type="checkbox" value="Красный" ${variation.colors?.includes('Красный') ? 'checked' : ''}> 🔴 Красный</label>
                        <label><input type="checkbox" value="Синий" ${variation.colors?.includes('Синий') ? 'checked' : ''}> 🔵 Синий</label>
                        <label><input type="checkbox" value="Зеленый" ${variation.colors?.includes('Зеленый') ? 'checked' : ''}> 🟢 Зеленый</label>
                        <label><input type="checkbox" value="Черный" ${variation.colors?.includes('Черный') ? 'checked' : ''}> ⚫ Черный</label>
                        <label><input type="checkbox" value="Белый" ${variation.colors?.includes('Белый') ? 'checked' : ''}> ⚪ Белый</label>
                        <label><input type="checkbox" value="Серый" ${variation.colors?.includes('Серый') ? 'checked' : ''}> ⚫ Серый</label>
                        <label><input type="checkbox" value="Желтый" ${variation.colors?.includes('Желтый') ? 'checked' : ''}> 🟡 Желтый</label>
                        <label><input type="checkbox" value="Оранжевый" ${variation.colors?.includes('Оранжевый') ? 'checked' : ''}> 🟠 Оранжевый</label>
                        <label><input type="checkbox" value="Розовый" ${variation.colors?.includes('Розовый') ? 'checked' : ''}> 🌸 Розовый</label>
                        <label><input type="checkbox" value="Фиолетовый" ${variation.colors?.includes('Фиолетовый') ? 'checked' : ''}> 🟣 Фиолетовый</label>
                        <label><input type="checkbox" value="Коричневый" ${variation.colors?.includes('Коричневый') ? 'checked' : ''}> 🟤 Коричневый</label>
                        <label><input type="checkbox" value="Бежевый" ${variation.colors?.includes('Бежевый') ? 'checked' : ''}> 🟫 Бежевый</label>
                        <label><input type="checkbox" value="Голубой" ${variation.colors?.includes('Голубой') ? 'checked' : ''}> 🔵 Голубой</label>
                        <label><input type="checkbox" value="Салатовый" ${variation.colors?.includes('Салатовый') ? 'checked' : ''}> 🟢 Салатовый</label>
                        <label><input type="checkbox" value="Бордовый" ${variation.colors?.includes('Бордовый') ? 'checked' : ''}> 🔴 Бордовый</label>
                        <label><input type="checkbox" value="Темно-синий" ${variation.colors?.includes('Темно-синий') ? 'checked' : ''}> 🔵 Темно-синий</label>
                    </div>
                </div>
                <div class="variation-field">
                    <label>Цена (₽)</label>
                    <input type="number" value="${variation.price}" min="0" step="0.01" placeholder="0.00" oninput="window.products.updateVariation(${index}, 'price', parseFloat(this.value)||0)">
                </div>
                <div class="variation-field">
                    <label>Скидка (%)</label>
                    <input type="number" value="${variation.discount || 0}" min="0" max="100" step="1" placeholder="0" oninput="window.products.updateVariation(${index}, 'discount', Math.min(100, Math.max(0, parseInt(this.value)||0)))">
                </div>
                <div class="variation-field">
                    <label>Количество</label>
                    <input type="number" value="${variation.stockQuantity}" min="0" placeholder="0" oninput="window.products.updateVariation(${index}, 'stockQuantity', Math.max(0, parseInt(this.value)||0))">
                </div>
                <div class="variation-field">
                    <label>SKU</label>
                    <input type="text" value="${variation.sku||''}" placeholder="SKU" oninput="window.products.updateVariation(${index}, 'sku', this.value)">
                </div>
                <div class="variation-field">
                    <label><i class="fas fa-barcode"></i> Штрих-код</label>
                    <input type="text" value="${variation.barcode||''}" placeholder="EAN-13, UPC, Code128" oninput="window.products.updateVariation(${index}, 'barcode', this.value)" maxlength="50">
                    <small style="color: #666; font-size: 11px; display: block; margin-top: 4px;">Введите штрих-код (EAN-13, UPC, Code128 и т.д.)</small>
                </div>
                <div class="variation-field">
                    <label>Фото вариации (несколько)</label>
                    <input type="file" accept="image/*" multiple onchange="window.products.uploadVariationImages(${index}, this)">
                    ${variation.imageUrls && variation.imageUrls.length ? `
                    <div class="variation-images-preview">
                        ${variation.imageUrls.map((url, imgIndex) => {
                            const imageUrl = window.getImageUrl ? window.getImageUrl(url) : url;
                            return `
                            <div class="image-preview-item">
                                <img src="${imageUrl}" alt="Preview" style="max-width: 70px; max-height: 70px; object-fit: cover; border-radius: 6px;">
                                <button type="button" class="remove-image" onclick="window.products.removeVariationImage(${index}, ${imgIndex})">×</button>
                            </div>`;
                        }).join('')}
                    </div>` : ''}
                </div>
            </div>
        </div>
    `).join('');

    // Установим обработчики чекбоксов после рендера
    container.querySelectorAll('.variation-item').forEach(item => {
        const idx = parseInt(item.getAttribute('data-variation-index'));
        // sizes
        item.querySelectorAll('.checkbox-group input[type="checkbox"]').forEach(input => {
            input.addEventListener('change', (e) => {
                const field = item.querySelector('label').textContent.includes('Размеры') ? 'sizes' : undefined;
                // Определяем по ближайшему родителю
                const parentLabel = e.target.closest('.variation-field').querySelector('label').textContent;
                const targetField = parentLabel.includes('Размер') ? 'sizes' : (parentLabel.includes('Цвет') ? 'colors' : null);
                if (!targetField) return;
                window.products.updateVariationMulti(idx, targetField, e.target.value, e.target.checked);
            });
        });
    });
}

// Обновление полей вариации
function updateVariation(index, field, value) {
    const vars = getVariations();
    if (!vars[index]) return;
    vars[index][field] = value;
    setVariations(vars);
}

function updateVariationMulti(index, field, value, checked) {
    const vars = getVariations();
    if (!vars[index]) return;
    if (!Array.isArray(vars[index][field])) vars[index][field] = [];
    if (checked) {
        if (!vars[index][field].includes(value)) vars[index][field].push(value);
    } else {
        vars[index][field] = vars[index][field].filter(v => v !== value);
    }
    setVariations(vars);
}

// Загрузка изображений вариации
async function uploadVariationImages(variationIndex, inputEl) {
    const files = Array.from(inputEl.files || []);
    if (files.length === 0) return;

    const adminToken = window.storage && window.storage.getAdminToken ? window.storage.getAdminToken() : null;
    const folder = 'variations';

    // Показываем индикатор загрузки
    if (window.ui && window.ui.showMessage) {
        window.ui.showMessage(`⏳ Загрузка ${files.length} фото...`, 'info');
    }

    // Создаем временный контейнер для превью загрузки
    let loadingContainer = document.getElementById(`loading-preview-${variationIndex}`);
    
    if (!loadingContainer) {
        loadingContainer = document.createElement('div');
        loadingContainer.id = `loading-preview-${variationIndex}`;
        loadingContainer.className = 'loading-preview';
        loadingContainer.style.cssText = 'display: flex; gap: 10px; margin-top: 10px; flex-wrap: wrap; position: fixed; top: 20px; right: 20px; z-index: 9999; background: white; padding: 15px; border-radius: 10px; box-shadow: 0 4px 15px rgba(0,0,0,0.2); max-width: 400px; max-height: 80vh; overflow-y: auto;';
        document.body.appendChild(loadingContainer);
    }

    const uploadedUrls = [];
    let successCount = 0;
    let failCount = 0;

    for (let i = 0; i < files.length; i++) {
        const file = files[i];
        
        // Создаем превью файла перед загрузкой
        const filePreview = URL.createObjectURL(file);
        
        // Добавляем превью загружаемого файла с большим размером
        const loadingItem = document.createElement('div');
        loadingItem.className = 'loading-item';
        loadingItem.style.cssText = 'position: relative; width: 120px; height: 120px; border: 2px solid #667eea; border-radius: 8px; overflow: hidden; background: #f8f9fa;';
        loadingItem.innerHTML = `
            <div style="position: absolute; inset: 0; display: flex; align-items: center; justify-content: center; background: rgba(255,255,255,0.9); z-index: 2;">
                <div style="text-align: center;">
                    <i class="fas fa-spinner fa-spin" style="font-size: 32px; color: #667eea;"></i>
                    <div style="margin-top: 8px; font-size: 12px; color: #666; font-weight: bold;">${i + 1}/${files.length}</div>
                </div>
            </div>
            <img src="${filePreview}" style="width: 100%; height: 100%; object-fit: cover;" alt="Preview">
        `;
        loadingContainer?.appendChild(loadingItem);
        
        const formData = new FormData();
        formData.append('image', file);
        formData.append('folder', folder);
        
        try {
            const resp = await fetch(getApiUrl(CONFIG.API.ENDPOINTS.UPLOAD.IMAGE) + `?folder=${folder}`, {
                method: 'POST',
                headers: adminToken ? { 'Authorization': `Bearer ${adminToken}` } : {},
                body: formData
            });
            const data = await resp.json();
            
            if (resp.ok && (data.url || data.data?.url)) {
                const url = data.url || data.data.url;
                uploadedUrls.push(url);
                successCount++;
                
                // Меняем индикатор на успешный с превью
                const imageUrl = window.getImageUrl ? window.getImageUrl(url) : url;
                loadingItem.innerHTML = `
                    <div style="position: absolute; top: 5px; right: 5px; background: #28a745; border-radius: 50%; width: 24px; height: 24px; display: flex; align-items: center; justify-content: center; z-index: 3; box-shadow: 0 2px 4px rgba(0,0,0,0.2);">
                        <i class="fas fa-check" style="font-size: 12px; color: white;"></i>
                    </div>
                    <img src="${imageUrl}" style="width: 100%; height: 100%; object-fit: cover;" alt="Uploaded" onerror="this.src='data:image/svg+xml,%3Csvg xmlns=%22http://www.w3.org/2000/svg%22%3E%3Crect fill=%22%23f8f9fa%22 width=%22120%22 height=%22120%22/%3E%3Ctext x=%2250%25%22 y=%2250%25%22 fill=%22%236c757d%22 text-anchor=%22middle%22 dy=%22.3em%22%3EОшибка%3C/text%3E%3C/svg%3E';">
                `;
                
                // Добавляем в вариацию, но не перерисовываем сразу
                const vars = getVariations();
                if (!Array.isArray(vars[variationIndex].imageUrls)) vars[variationIndex].imageUrls = [];
                vars[variationIndex].imageUrls.push(url);
                setVariations(vars);
                
                // Освобождаем память от превью
                URL.revokeObjectURL(filePreview);
                
            } else {
                failCount++;
                loadingItem.innerHTML = `
                    <div style="position: absolute; inset: 0; display: flex; align-items: center; justify-content: center; background: rgba(220,53,69,0.1);">
                        <i class="fas fa-times-circle" style="font-size: 32px; color: #dc3545;"></i>
                    </div>
                `;
                URL.revokeObjectURL(filePreview);
            }
        } catch (e) {
            console.error('upload error', e);
            failCount++;
            loadingItem.innerHTML = `
                <div style="position: absolute; inset: 0; display: flex; align-items: center; justify-content: center; background: rgba(220,53,69,0.1);">
                    <i class="fas fa-times-circle" style="font-size: 32px; color: #dc3545;"></i>
                </div>
            `;
            URL.revokeObjectURL(filePreview);
        }
    }

    // Перерисовываем вариации один раз после всех загрузок
    renderVariations();

    // Убираем контейнер загрузки через 3 секунды (увеличено для лучшей видимости)
    setTimeout(() => {
        if (loadingContainer) {
            loadingContainer.remove();
        }
    }, 3000);

    // Показываем финальное сообщение
    if (window.ui && window.ui.showMessage) {
        if (failCount === 0) {
            window.ui.showMessage(`✅ Загружено ${successCount} фото`, 'success');
        } else {
            window.ui.showMessage(`⚠️ Загружено ${successCount}, ошибок ${failCount}`, 'warning');
        }
    }

    inputEl.value = '';
}

function removeVariationImage(variationIndex, imgIndex) {
    const vars = getVariations();
    if (!vars[variationIndex] || !Array.isArray(vars[variationIndex].imageUrls)) return;
    vars[variationIndex].imageUrls.splice(imgIndex, 1);
    setVariations(vars);
    renderVariations();
}

// Открытие модального окна товара
function openProductModal() {
    const modal = document.getElementById('product-modal');
    const title = document.getElementById('product-modal-title');
    
    if (modal && title) {
        title.textContent = 'Добавить товар';
        document.getElementById('product-form')?.reset();
        
        // Сбрасываем вариации
        setVariations([]);
        renderVariations();
        
        // Сбрасываем ID
        if (window.storage) {
            if (window.storage.setCurrentProductId) {
                window.storage.setCurrentProductId(null);
            }
        }
        modal.style.display = 'block';
    }
}

// Закрытие модального окна товара
function closeProductModal() {
    const modal = document.getElementById('product-modal');
    if (modal) {
        modal.style.display = 'none';
        currentProductId = null;
    }
}

// Обработка отправки формы товара
async function handleProductSubmit(e) {
    e.preventDefault();
    
    console.log('📝 Отправка формы товара...');
    
    // Синхронизируем размеры из DOM перед сохранением
    syncVariationsFromDOM();
    
    try {
        const formData = {
            name: document.getElementById('product-name').value,
            description: document.getElementById('product-description').value,
            gender: document.getElementById('product-gender').value,
            categoryId: document.getElementById('product-category').value,
            brand: document.getElementById('product-brand').value,
            variations: getVariations()
        };
        
        console.log('📦 Данные формы:', formData);
        console.log('📦 Вариации с размерами:', formData.variations.map(v => ({ sizes: v.sizes, colors: v.colors })));
        
        // Проверка: обязательно наличие хотя бы одной вариации
        if (!formData.variations || formData.variations.length === 0) {
            if (window.ui && window.ui.showMessage) {
                window.ui.showMessage('❌ Добавьте хотя бы одну вариацию товара!', 'error');
            } else {
                alert('Добавьте хотя бы одну вариацию товара!');
            }
            return;
        }
        
        // Определяем роль и endpoint
        const userRole = localStorage.getItem('userRole') || 'admin';
        let endpoint;
        
        if (userRole === 'super_admin' || userRole === 'admin') {
            endpoint = '/api/v1/admin/products/';
        } else if (userRole === 'shop_owner') {
            endpoint = '/api/v1/shop/products/';
        } else {
            endpoint = '/api/v1/shop/products/';
        }
        
        console.log(`🔗 Используем endpoint: ${endpoint}`);
        
        const currentProductId = window.storage ? window.storage.getCurrentProductId() : null;
        
        let response;
        if (currentProductId) {
            // Обновление (без слэша в конце)
            response = await window.api.fetchData(`${endpoint}${currentProductId}`, {
                method: 'PUT',
                body: JSON.stringify(formData)
            });
        } else {
            // Создание
            response = await window.api.fetchData(endpoint, {
                method: 'POST',
                body: JSON.stringify(formData)
            });
        }
        
        console.log('✅ Ответ сервера:', response);
        
        if (window.ui && window.ui.showMessage) {
            window.ui.showMessage('Товар успешно сохранен!', 'success');
        }
        
        closeProductModal();
        loadProducts();
        
    } catch (error) {
        console.error('❌ Ошибка сохранения товара:', error);
        if (window.ui && window.ui.showMessage) {
            window.ui.showMessage('Ошибка сохранения товара: ' + error.message, 'error');
        }
    }
}

// Экспорт
// Функция для смены типа размеров
function changeSizeType(variationIndex, sizeType) {
    const container = document.getElementById(`sizes-container-${variationIndex}`);
    if (!container) {
        console.error(`❌ Контейнер sizes-container-${variationIndex} не найден`);
        return;
    }
    
    // Получаем текущие выбранные размеры из чекбоксов
    const currentCheckboxes = container.querySelectorAll('input[type="checkbox"]:checked');
    const currentSizes = Array.from(currentCheckboxes).map(cb => cb.value);
    
    console.log('📦 Текущие выбранные размеры:', currentSizes);
    
    let sizesHTML = '';
    
    switch(sizeType) {
        case 'shoes':
            // Размеры обуви (36-46)
            const shoeSizes = ['36', '37', '38', '39', '40', '41', '42', '43', '44', '45', '46'];
            sizesHTML = shoeSizes.map(size => 
                `<label><input type="checkbox" value="${size}" onchange="window.products.updateVariationSize(${variationIndex})" ${currentSizes.includes(size) ? 'checked' : ''}> ${size}</label>`
            ).join('');
            break;
            
        case 'pants':
            // Размеры штанов/джинсов (28-40)
            const pantsSizes = ['28', '29', '30', '31', '32', '33', '34', '36', '38', '40'];
            sizesHTML = pantsSizes.map(size => 
                `<label><input type="checkbox" value="${size}" onchange="window.products.updateVariationSize(${variationIndex})" ${currentSizes.includes(size) ? 'checked' : ''}> ${size}</label>`
            ).join('');
            break;
            
        case 'clothing':
        default:
            // Размеры одежды (XS-XXL)
            const clothingSizes = ['XS', 'S', 'M', 'L', 'XL', 'XXL'];
            sizesHTML = clothingSizes.map(size => 
                `<label><input type="checkbox" value="${size}" onchange="window.products.updateVariationSize(${variationIndex})" ${currentSizes.includes(size) ? 'checked' : ''}> ${size}</label>`
            ).join('');
            break;
    }
    
    container.innerHTML = sizesHTML;
    
    // Обновляем размеры в хранилище после смены типа
    // Извлекаем размеры из новых чекбоксов (которые могут быть отмечены, если были совпадения)
    setTimeout(() => {
        updateVariationSize(variationIndex);
    }, 0);
    
    console.log(`✅ Тип размеров изменен на: ${sizeType} для вариации ${variationIndex}`);
    console.log(`📊 Новые размеры:`, sizeType === 'shoes' ? '36-46' : sizeType === 'pants' ? '28-40' : 'XS-XXL');
}

// Синхронизация всех вариаций из DOM в хранилище
function syncVariationsFromDOM() {
    const vars = getVariations();
    let updated = false;
    
    for (let i = 0; i < vars.length; i++) {
        const container = document.getElementById(`sizes-container-${i}`);
        if (container) {
            const checkboxes = container.querySelectorAll('input[type="checkbox"]:checked');
            const selectedSizes = Array.from(checkboxes).map(cb => cb.value);
            
            if (JSON.stringify(vars[i].sizes) !== JSON.stringify(selectedSizes)) {
                vars[i].sizes = selectedSizes;
                updated = true;
                console.log(`🔄 Синхронизированы размеры для вариации ${i}:`, selectedSizes);
            }
        }
    }
    
    if (updated) {
        setVariations(vars);
        console.log('✅ Все вариации синхронизированы из DOM');
    }
}

// Функция для обновления размеров вариации
function updateVariationSize(variationIndex) {
    const container = document.getElementById(`sizes-container-${variationIndex}`);
    if (!container) return;
    
    const checkboxes = container.querySelectorAll('input[type="checkbox"]:checked');
    const selectedSizes = Array.from(checkboxes).map(cb => cb.value);
    
    // Обновляем в хранилище (важно!)
    const vars = getVariations();
    if (vars[variationIndex]) {
        vars[variationIndex].sizes = selectedSizes;
        setVariations(vars);
        console.log(`✅ Размеры обновлены для вариации ${variationIndex}:`, selectedSizes);
    }
    
    // Также обновляем в глобальном объекте для совместимости
    if (currentProduct && currentProduct.variations && currentProduct.variations[variationIndex]) {
        currentProduct.variations[variationIndex].sizes = selectedSizes;
    }
}

window.products = {
    loadProducts,
    displayProducts,
    refreshProductsList,
    editProduct,
    loadProductData,
    deleteProduct,
    viewProductVariations,
    addVariation,
    removeVariation,
    renderVariations,
    updateVariation,
    updateVariationMulti,
    uploadVariationImages,
    removeVariationImage,
    openProductModal,
    closeProductModal,
    handleProductSubmit,
    changeSizeType,
    updateVariationSize
};

// Глобальные функции для обратной совместимости
window.viewProductVariations = viewProductVariations;
window.openProductModal = openProductModal;
window.closeProductModal = closeProductModal;

// Модальное окно для просмотра изображения
window.openImageModal = function(imageUrl, title = 'Изображение') {
    const modal = document.createElement('div');
    modal.className = 'modal';
    modal.style.display = 'block';
    modal.style.zIndex = '10000';
    modal.style.backgroundColor = 'rgba(0, 0, 0, 0.9)';
    
    modal.innerHTML = `
        <div class="modal-content" style="max-width: 90vw; max-height: 90vh; margin: 5vh auto; background: transparent; box-shadow: none;">
            <div style="position: relative; display: flex; align-items: center; justify-content: center;">
                <span class="close" onclick="this.closest('.modal').remove()" style="position: absolute; top: -40px; right: 0; color: white; font-size: 40px; cursor: pointer; z-index: 10001;">&times;</span>
                <img src="${imageUrl}" alt="${title}" style="max-width: 100%; max-height: 90vh; object-fit: contain; border-radius: 10px; box-shadow: 0 10px 50px rgba(0,0,0,0.5);">
            </div>
        </div>
    `;
    
    document.body.appendChild(modal);
    
    // Закрытие по клику вне изображения
    modal.addEventListener('click', (e) => {
        if (e.target === modal) {
            modal.remove();
        }
    });
    
    // Закрытие по Escape
    const handleEscape = (e) => {
        if (e.key === 'Escape') {
            modal.remove();
            document.removeEventListener('keydown', handleEscape);
        }
    };
    document.addEventListener('keydown', handleEscape);
};

