// ===== CATEGORIES.JS - Управление категориями =====

// Загрузка категорий
async function loadCategories() {
    try {
        const response = await window.api.fetchData(CONFIG.API.ENDPOINTS.CATEGORIES.LIST);
        console.log('📡 Ответ API категорий:', response);
        
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
            categories = [];
        }
        
        console.log(`📦 Получено ${categories.length} категорий:`, categories);
        
        displayCategories(categories);
        populateCategorySelects(categories);
        
        console.log('✅ Категории загружены и селекты обновлены');
    } catch (error) {
        console.error('Ошибка загрузки категорий:', error);
        if (window.ui && window.ui.showMessage) {
            window.ui.showMessage('Ошибка загрузки категорий', 'error');
        }
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
    
    // Функция для рендера категории с иконкой
    const renderCategoryIcon = (category) => {
        if (category.iconUrl) {
            return `<img src="${category.iconUrl}" alt="${category.name}" style="width: 50px; height: 50px; border-radius: 12px; object-fit: cover; border: 2px solid #e0e0e0;">`;
        }
        return `<div style="width: 50px; height: 50px; border-radius: 12px; background: linear-gradient(135deg, #f093fb, #f5576c); display: flex; align-items: center; justify-content: center; color: white; font-size: 20px;">
                    <i class="fas fa-tag"></i>
                </div>`;
    };
    
    // Функция для рендера иерархии категорий
    const renderCategoryTree = (categories, level = 0) => {
        return categories.map((category, index) => {
            const indent = level * 20;
            const hasChildren = category.children && category.children.length > 0;
            
            let html = `
                <tr style="animation-delay: ${index * 0.1}s;" data-category-id="${category.id}">
                    <td>
                        <div style="display: flex; align-items: center; gap: 12px; padding-left: ${indent}px;">
                            ${renderCategoryIcon(category)}
                            <div>
                                <div style="font-weight: 700; color: #333; font-size: 16px;">
                                    ${level > 0 ? '<i class="fas fa-level-up-alt" style="transform: rotate(90deg); margin-right: 5px; color: #999;"></i>' : ''}
                                    ${category.name}
                                </div>
                                <div style="font-size: 12px; color: #888; font-family: monospace;">${category.id?.substring(0, 8)}...</div>
                            </div>
                        </div>
                    </td>
                    <td>${category.description || 'Нет описания'}</td>
                    <td>
                        ${category.parent?.name ? 
                            `<span class="badge" style="background: #6c757d; color: white;">${category.parent.name}</span>` : 
                            '<span style="color: #999;">Корневая</span>'}
                    </td>
                    <td>
                        <span class="badge" style="background: linear-gradient(135deg, #45b7d1, #96ceb4); color: white;">
                            ${category.sortOrder || 0}
                        </span>
                    </td>
                    <td>
                        <div>
                            <div>${category.createdAt ? new Date(category.createdAt).toLocaleDateString('ru-RU') : 'N/A'}</div>
                            <div style="font-size: 11px; color: #7f8c8d;">
                                ${category.createdAt ? new Date(category.createdAt).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' }) : ''}
                            </div>
                        </div>
                    </td>
                    <td>
                        <div style="display: flex; gap: 8px;">
                            <button class="btn-sm btn-primary" onclick="window.categories.editCategory('${category.id}')" title="Редактировать">
                                <i class="fas fa-edit"></i>
                            </button>
                            <button class="btn-sm btn-danger" onclick="window.categories.deleteCategory('${category.id}')" title="Удалить">
                                <i class="fas fa-trash"></i>
                            </button>
                        </div>
                    </td>
                </tr>
            `;
            
            // Рекурсивно добавляем подкатегории
            if (hasChildren) {
                html += renderCategoryTree(category.children, level + 1);
            }
            
            return html;
        }).join('');
    };
    
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
                        ${renderCategoryTree(categories)}
                    </tbody>
                </table>
            </div>
        </div>
    `;
    
    container.innerHTML = table;
}

// Заполнение селектов категорий
function populateCategorySelects(categories) {
    console.log('🔄 populateCategorySelects вызвана с категориями:', categories);
    
    // Функция для рекурсивного добавления категорий с отступами
    const addCategoryOptions = (select, categories, level = 0, excludeId = null) => {
        categories.forEach(category => {
            // Исключаем текущую категорию и её потомков при редактировании (для parent select)
            if (category.id !== excludeId) {
                const option = document.createElement('option');
                option.value = category.id;
                const indent = '  '.repeat(level); // Отступы для визуализации уровня
                const prefix = level > 0 ? '└─ ' : '';
                option.textContent = indent + prefix + category.name;
                select.appendChild(option);
                
                // Рекурсивно добавляем подкатегории
                if (category.children && category.children.length > 0) {
                    addCategoryOptions(select, category.children, level + 1, excludeId);
                } else if (category.subcategories && category.subcategories.length > 0) {
                    addCategoryOptions(select, category.subcategories, level + 1, excludeId);
                }
            }
        });
    };
    
    const selects = [
        { element: document.getElementById('product-category'), default: 'Выберите категорию' },
        { element: document.getElementById('category-parent'), default: 'Нет родительской категории (корневая)' },
        { element: document.getElementById('category-filter'), default: 'Все категории' }
    ];
    
    selects.forEach(({ element, default: defaultText }) => {
        if (element) {
            const currentValue = element.value;
            element.innerHTML = `<option value="">${defaultText}</option>`;
            
            // Для category-parent исключаем текущую редактируемую категорию
            const excludeId = element.id === 'category-parent' && window.currentCategoryId ? 
                window.currentCategoryId : null;
            
            addCategoryOptions(element, categories, 0, excludeId);
            
            // Восстанавливаем выбранное значение
            if (currentValue) {
                element.value = currentValue;
            }
        }
    });
}

// Создание категории
async function createCategory(data) {
    const userRole = localStorage.getItem('userRole') || 'admin';
    
    if (userRole !== 'super_admin' && userRole !== 'admin') {
        throw new Error('Создание категорий доступно только администраторам');
    }
    
    const endpoint = '/api/v1/admin/categories/';
    return await window.api.fetchData(endpoint, {
        method: 'POST',
        body: JSON.stringify(data)
    });
}

// Обновление категории
async function updateCategory(id, data) {
    const userRole = localStorage.getItem('userRole') || 'admin';
    
    if (userRole !== 'super_admin' && userRole !== 'admin') {
        throw new Error('Обновление категорий доступно только администраторам');
    }
    
    const endpoint = `/api/v1/admin/categories/${id}/`;
    return await window.api.fetchData(endpoint, {
        method: 'PUT',
        body: JSON.stringify(data)
    });
}

// Удаление категории
async function deleteCategory(id) {
    if (!confirm('Вы уверены, что хотите удалить эту категорию?')) return;
    
    try {
        const userRole = localStorage.getItem('userRole') || 'admin';
        
        if (userRole !== 'super_admin' && userRole !== 'admin') {
            throw new Error('Удаление категорий доступно только администраторам');
        }
        
        const endpoint = `/api/v1/admin/categories/${id}/`;
        await window.api.fetchData(endpoint, { method: 'DELETE' });
        
        loadCategories();
        
        if (window.ui && window.ui.showMessage) {
            window.ui.showMessage('Категория успешно удалена', 'success');
        }
    } catch (error) {
        console.error('Ошибка удаления категории:', error);
        if (window.ui && window.ui.showMessage) {
            window.ui.showMessage('Ошибка удаления категории', 'error');
        }
    }
}

// Редактирование категории
async function editCategory(id) {
    openCategoryModal(id);
}

// Загрузка данных категории
async function loadCategoryData(id) {
    try {
        const response = await window.api.fetchData(`/api/v1/categories/${id}`);
        
        let category;
        if (response.category) {
            category = response.category;
        } else if (response.data) {
            category = response.data;
        } else {
            category = response;
        }
        
        document.getElementById('category-name').value = category.name;
        document.getElementById('category-description').value = category.description || '';
        document.getElementById('category-parent').value = category.parentId || category.parent_id || '';
        document.getElementById('category-sort-order').value = category.sortOrder || category.sort_order || 0;
        document.getElementById('category-is-active').checked = category.isActive !== undefined ? category.isActive : (category.is_active !== undefined ? category.is_active : true);
        
        // Загрузка иконки
        if (category.iconUrl || category.icon_url) {
            const iconUrl = category.iconUrl || category.icon_url;
            document.getElementById('category-icon-url').value = iconUrl;
            document.getElementById('category-icon-preview-img').src = iconUrl;
            document.getElementById('category-icon-preview').style.display = 'block';
        } else {
            document.getElementById('category-icon-url').value = '';
            document.getElementById('category-icon-preview').style.display = 'none';
        }
    } catch (error) {
        console.error('Ошибка загрузки данных категории:', error);
        if (window.ui && window.ui.showMessage) {
            window.ui.showMessage('Ошибка загрузки данных категории', 'error');
        }
    }
}

// Открытие модального окна категории
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
        document.getElementById('category-icon-url').value = '';
        document.getElementById('category-icon-preview').style.display = 'none';
        document.getElementById('category-sort-order').value = 0;
        document.getElementById('category-is-active').checked = true;
    }
    
    modal.style.display = 'block';
    
    // Добавляем обработчик для предпросмотра иконки
    const iconInput = document.getElementById('category-icon');
    if (iconInput && !iconInput.dataset.listenerAttached) {
        iconInput.addEventListener('change', handleIconPreview);
        iconInput.dataset.listenerAttached = 'true';
    }
}

// Предпросмотр иконки
function handleIconPreview(e) {
    const file = e.target.files[0];
    if (file) {
        const reader = new FileReader();
        reader.onload = function(e) {
            document.getElementById('category-icon-preview-img').src = e.target.result;
            document.getElementById('category-icon-preview').style.display = 'block';
        };
        reader.readAsDataURL(file);
    }
}

// Очистка иконки
function clearCategoryIcon() {
    document.getElementById('category-icon').value = '';
    document.getElementById('category-icon-url').value = '';
    document.getElementById('category-icon-preview').style.display = 'none';
}

// Глобальная функция для очистки иконки
window.clearCategoryIcon = clearCategoryIcon;

// Закрытие модального окна категории
function closeCategoryModal() {
    document.getElementById('category-modal').style.display = 'none';
    currentCategoryId = null;
}

// Загрузка иконки на сервер
async function uploadCategoryIcon(file) {
    const formData = new FormData();
    formData.append('image', file);
    
    try {
        const token = localStorage.getItem('token');
        const response = await fetch(`${CONFIG.API.BASE_URL}/api/v1/upload/image`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`
            },
            body: formData
        });
        
        if (!response.ok) {
            throw new Error('Failed to upload icon');
        }
        
        const result = await response.json();
        
        // API может вернуть URL в разных форматах
        if (result.success && result.data) {
            return result.data.url || result.data.imageUrl || result.data.path;
        } else if (result.url) {
            return result.url;
        } else if (result.imageUrl) {
            return result.imageUrl;
        }
        
        throw new Error('Invalid response format');
    } catch (error) {
        console.error('Ошибка загрузки иконки:', error);
        throw error;
    }
}

// Обработка отправки формы категории
async function handleCategorySubmit(e) {
    e.preventDefault();
    
    try {
        let iconUrl = document.getElementById('category-icon-url').value;
        
        // Если выбран новый файл иконки, загружаем его
        const iconFile = document.getElementById('category-icon').files[0];
        if (iconFile) {
            if (window.ui && window.ui.showMessage) {
                window.ui.showMessage('Загрузка иконки...', 'info');
            }
            iconUrl = await uploadCategoryIcon(iconFile);
        }
        
        const parentValue = document.getElementById('category-parent').value;
        
        const formData = {
            name: document.getElementById('category-name').value,
            description: document.getElementById('category-description').value,
            iconUrl: iconUrl || '',
            parentId: parentValue ? parentValue : null,
            sortOrder: parseInt(document.getElementById('category-sort-order').value) || 0,
            isActive: document.getElementById('category-is-active').checked
        };
        
        console.log('Отправка данных категории:', formData);
        
        if (currentCategoryId) {
            await updateCategory(currentCategoryId, formData);
        } else {
            await createCategory(formData);
        }
        
        closeCategoryModal();
        loadCategories();
        
        if (window.ui && window.ui.showMessage) {
            window.ui.showMessage('Категория успешно сохранена', 'success');
        }
    } catch (error) {
        console.error('Ошибка сохранения категории:', error);
        if (window.ui && window.ui.showMessage) {
            window.ui.showMessage('Ошибка сохранения категории: ' + (error.message || 'Неизвестная ошибка'), 'error');
        }
    }
}

// Экспорт
window.categories = {
    loadCategories,
    displayCategories,
    populateCategorySelects,
    createCategory,
    updateCategory,
    deleteCategory,
    editCategory,
    loadCategoryData,
    openCategoryModal,
    closeCategoryModal,
    handleCategorySubmit,
    uploadCategoryIcon,
    handleIconPreview,
    clearCategoryIcon
};

// Глобальные функции для обратной совместимости
window.openCategoryModal = openCategoryModal;
window.closeCategoryModal = closeCategoryModal;
window.clearCategoryIcon = clearCategoryIcon;

