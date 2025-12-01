// ===== USERS.JS - Управление пользователями =====

// Загрузка пользователей
window.loadUsers = async function loadUsers() {
    console.log('🔄 Загружаем пользователей...');
    try {
        const response = await window.api.fetchData('/api/v1/admin/users/');
        console.log('📡 Ответ API для пользователей:', response);
        if (response.success && response.data) {
            const users = response.data.users || [];
            console.log('✅ Пользователи загружены успешно:', users.length, 'пользователей');
            displayUsers(users);
        } else {
            console.error('❌ Ошибка в ответе API для пользователей:', response);
            if (window.ui && window.ui.showMessage) {
                window.ui.showMessage('Ошибка загрузки пользователей: ' + (response.message || 'Неизвестная ошибка'), 'error');
            }
        }
    } catch (error) {
        console.error('❌ Ошибка загрузки пользователей:', error);
        if (window.ui && window.ui.showMessage) {
            window.ui.showMessage('Ошибка загрузки пользователей: ' + error.message, 'error');
        }
    }
};

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
    
    if (!users || users.length === 0) {
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
        const row = document.createElement('tr');
        row.style.animationDelay = `${index * 0.1}s`;
        
        const rowHtml = `
            <td data-label="Имя">
                <div style="display: flex; align-items: center; gap: 12px;">
                    <div style="width: 45px; height: 45px; border-radius: 50%; background: linear-gradient(135deg, #667eea, #764ba2); display: flex; align-items: center; justify-content: center; color: white; font-weight: bold; font-size: 18px;">
                        ${user.name ? user.name.charAt(0).toUpperCase() : 'U'}
                    </div>
                    <div>
                        <div style="font-weight: 700; color: #333; font-size: 15px;">${user.name || 'Не указано'}</div>
                        <div style="font-size: 11px; color: #888; font-family: monospace;">${user.id ? user.id.substring(0, 8) + '...' : 'N/A'}</div>
                    </div>
                </div>
            </td>
            <td data-label="Email">
                <div style="display: flex; align-items: center; gap: 10px;">
                    <i class="fas fa-envelope" style="color: #667eea;"></i>
                    <span>${user.email || 'N/A'}</span>
                </div>
            </td>
            <td data-label="Телефон">
                <div style="display: flex; align-items: center; gap: 10px;">
                    <i class="fas fa-phone" style="color: #4ecdc4;"></i>
                    <span>${user.phone || 'Не указан'}</span>
                </div>
            </td>
            <td data-label="Роль">
                <span class="badge role-${user.role?.name || 'user'}">
                    <i class="fas fa-user-shield"></i>
                    ${user.role?.displayName || 'Пользователь'}
                </span>
            </td>
            <td data-label="Статус">
                <span class="badge ${user.isActive ? 'role-user' : 'role-admin'}" style="background: ${user.isActive ? 'linear-gradient(135deg, #4ecdc4, #44a08d)' : 'linear-gradient(135deg, #ff6b6b, #ee5a24)'};">
                    <i class="fas ${user.isActive ? 'fa-check-circle' : 'fa-times-circle'}"></i>
                    ${user.isActive ? 'Активен' : 'Неактивен'}
                </span>
            </td>
            <td data-label="Дата регистрации">
                <div style="display: flex; align-items: center; gap: 8px;">
                    <i class="fas fa-calendar" style="color: #f093fb;"></i>
                    <span style="font-size: 13px; color: #666;">
                        ${user.createdAt ? new Date(user.createdAt).toLocaleDateString('ru-RU', { day: '2-digit', month: '2-digit', year: 'numeric' }) : 'N/A'}
                    </span>
                </div>
            </td>
            <td data-label="Действия">
                <div style="display: flex; gap: 6px;">
                    <button class="btn-sm btn-primary" onclick="window.users.viewUser('${user.id}')" title="Просмотр">
                        <i class="fas fa-eye"></i>
                    </button>
                    <button class="btn-sm btn-secondary" onclick="window.users.editUser('${user.id}')" title="Редактировать">
                        <i class="fas fa-edit"></i>
                    </button>
                    <button class="btn-sm btn-danger" onclick="window.users.deleteUser('${user.id}')" title="Удалить">
                        <i class="fas fa-trash"></i>
                    </button>
                </div>
            </td>
        `;
        
        row.setAttribute('data-user-id', user.id);
        row.innerHTML = rowHtml;
        tbody.appendChild(row);
    });
    
    console.log('✅ displayUsers завершена');
}

// Просмотр пользователя
async function viewUser(id) {
    try {
        const response = await window.api.fetchData(`/api/v1/admin/users/${id}`);
        const user = response.data;
        
        // Показываем информацию о пользователе
        if (window.ui && window.ui.showModal) {
            window.ui.showModal('Информация о пользователе', `
                <div class="user-details-modal">
                    <div class="user-avatar-large">
                        ${user.avatar ? `<img src="${user.avatar}" alt="${user.name}">` : `<i class="fas fa-user-circle"></i>`}
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
                            <span>${user.role?.displayName || 'Пользователь'}</span>
                        </div>
                        <div class="info-item">
                            <label>Статус:</label>
                            <span>${user.isActive ? 'Активен' : 'Неактивен'}</span>
                        </div>
                        <div class="info-item">
                            <label>Дата регистрации:</label>
                            <span>${new Date(user.created_at).toLocaleString()}</span>
                        </div>
                    </div>
                </div>
            `);
        }
    } catch (error) {
        console.error('Ошибка загрузки данных пользователя:', error);
        if (window.ui && window.ui.showMessage) {
            window.ui.showMessage('Ошибка загрузки данных пользователя', 'error');
        }
    }
}

// Редактирование пользователя
async function editUser(id) {
    try {
        const response = await window.api.fetchData(`/api/v1/admin/users/${id}`);
        const user = response.data;
        
        // Показываем форму редактирования
        const modal = document.createElement('div');
        modal.className = 'modal';
        modal.style.display = 'block';
        
        modal.innerHTML = `
            <div class="modal-content" style="max-width: 500px;">
                <div class="modal-header">
                    <h3><i class="fas fa-edit"></i> Редактировать пользователя</h3>
                    <span class="close" onclick="this.closest('.modal').remove()">&times;</span>
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
                        <button type="button" class="btn btn-secondary" onclick="this.closest('.modal').remove()">Отмена</button>
                    </div>
                </form>
            </div>
        `;
        
        document.body.appendChild(modal);
        
        document.getElementById('edit-user-form').addEventListener('submit', async (e) => {
            e.preventDefault();
            
            const formData = {
                name: document.getElementById('edit-user-name').value,
                phone: document.getElementById('edit-user-phone').value,
                role: document.getElementById('edit-user-role').value,
                isActive: document.getElementById('edit-user-active').checked
            };
            
            try {
                await window.api.fetchData(`/api/v1/admin/users/${id}`, {
                    method: 'PUT',
                    body: JSON.stringify(formData)
                });
                
                if (window.ui && window.ui.showMessage) {
                    window.ui.showMessage('Пользователь успешно обновлен', 'success');
                }
                document.querySelector('.modal').remove();
                loadUsers();
                
            } catch (error) {
                console.error('Ошибка обновления пользователя:', error);
                if (window.ui && window.ui.showMessage) {
                    window.ui.showMessage('Ошибка обновления пользователя', 'error');
                }
            }
        });
        
    } catch (error) {
        console.error('Ошибка загрузки данных пользователя:', error);
        if (window.ui && window.ui.showMessage) {
            window.ui.showMessage('Ошибка загрузки данных пользователя', 'error');
        }
    }
}

// Удаление пользователя
async function deleteUser(id) {
    if (!confirm('Вы уверены, что хотите удалить этого пользователя? Это действие нельзя отменить.')) {
        return;
    }
    
    try {
        await window.api.fetchData(`/api/v1/admin/users/${id}`, {
            method: 'DELETE'
        });
        
        if (window.ui && window.ui.showMessage) {
            window.ui.showMessage('Пользователь успешно удален', 'success');
        }
        loadUsers();
        
    } catch (error) {
        console.error('Ошибка удаления пользователя:', error);
        if (window.ui && window.ui.showMessage) {
            window.ui.showMessage('Ошибка удаления пользователя', 'error');
        }
    }
}

// Открытие модального окна пользователя
function openUserModal() {
    document.getElementById('user-modal').style.display = 'block';
    loadRolesForSelect();
}

// Закрытие модального окна пользователя
function closeUserModal() {
    document.getElementById('user-modal').style.display = 'none';
    document.getElementById('user-form').reset();
    // Скрываем поля магазина при закрытии
    const shopFields = document.getElementById('shop-fields');
    if (shopFields) {
        shopFields.style.display = 'none';
    }
}

// Загрузка ролей для селекта
async function loadRolesForSelect() {
    try {
        console.log('🔄 Загружаем роли для селекта...');
        const response = await window.api.fetchData('/api/v1/admin/roles/');
        
        if (response.success && response.data && response.data.roles) {
            const roleSelect = document.getElementById('modal-user-role');
            
            if (roleSelect) {
                roleSelect.innerHTML = '<option value="">Выберите роль</option>';
                
                // Сохраняем роли для проверки shop_owner
                window.availableRoles = response.data.roles;
                
                response.data.roles.forEach(role => {
                    const option = document.createElement('option');
                    option.value = role.id;
                    option.textContent = role.displayName || role.name;
                    option.dataset.roleName = role.name; // Сохраняем имя роли для проверки
                    roleSelect.appendChild(option);
                });
                
                console.log(`✅ Загружено ${response.data.roles.length} ролей в селект`);
            }
        }
    } catch (error) {
        console.error('❌ Ошибка загрузки ролей для селекта:', error);
    }
}

// Обработка изменения роли - показываем/скрываем поля магазина
function handleRoleChange() {
    const roleSelect = document.getElementById('modal-user-role');
    const shopFields = document.getElementById('shop-fields');
    const selectedOption = roleSelect.options[roleSelect.selectedIndex];
    const roleName = selectedOption ? selectedOption.dataset.roleName : '';
    
    if (roleName === 'shop_owner') {
        shopFields.style.display = 'block';
        // Делаем ИНН обязательным
        document.getElementById('modal-shop-inn').required = true;
        // Автозаполняем название магазина из имени пользователя, если пусто
        const shopNameInput = document.getElementById('modal-shop-name');
        const userNameInput = document.getElementById('modal-user-name');
        if (!shopNameInput.value && userNameInput.value) {
            shopNameInput.value = userNameInput.value;
        }
        // Название магазина не обязательное - будет использовано имя пользователя, если не указано
        shopNameInput.required = false;
        // Загружаем города для выбора
        loadCitiesForSelect();
    } else {
        shopFields.style.display = 'none';
        // Убираем обязательность полей
        document.getElementById('modal-shop-name').required = false;
        document.getElementById('modal-shop-inn').required = false;
        // Очищаем поля
        document.getElementById('modal-shop-name').value = '';
        document.getElementById('modal-shop-inn').value = '';
        document.getElementById('modal-shop-description').value = '';
        document.getElementById('modal-shop-address').value = '';
    }
}

// Обработка создания пользователя
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
    const roleSelect = document.getElementById('modal-user-role');
    const selectedOption = roleSelect.options[roleSelect.selectedIndex];
    const roleName = selectedOption ? selectedOption.dataset.roleName : '';
    
    if (roleId) {
        formData.roleId = roleId;
    }
    
    // Если выбрана роль shop_owner, добавляем данные магазина
    if (roleName === 'shop_owner') {
        const citySelect = document.getElementById('modal-shop-city');
        const cityId = citySelect ? citySelect.value : null;
        
        formData.shop = {
            name: document.getElementById('modal-shop-name').value,
            inn: document.getElementById('modal-shop-inn').value,
            description: document.getElementById('modal-shop-description').value,
            address: document.getElementById('modal-shop-address').value,
            email: document.getElementById('modal-user-email').value, // Используем email пользователя
            phone: document.getElementById('modal-user-phone').value, // Используем телефон пользователя
            cityId: cityId || null // ID города
        };
    }
    
    try {
        const adminToken = window.storage ? window.storage.getAdminToken() : null;
        
        if (!adminToken) {
            console.error('❌ Токен админа не найден!');
            if (window.ui && window.ui.showMessage) {
                window.ui.showMessage('Токен авторизации не найден. Пожалуйста, войдите заново.', 'error');
            }
            return;
        }
        
        console.log('📤 Отправляем данные для создания пользователя:', formData);
        
        const API_BASE_URL = window.getApiUrl ? window.getApiUrl('') : (CONFIG && CONFIG.API && CONFIG.API.BASE_URL ? CONFIG.API.BASE_URL : '');
        const response = await fetch(`${API_BASE_URL}/api/v1/admin/users/`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${adminToken}`
            },
            body: JSON.stringify(formData)
        });
        
        console.log('📡 Ответ сервера:', response.status, response.statusText);
        
        const data = await response.json();
        console.log('📦 Данные ответа:', data);
        
        if (response.ok && data.success) {
            if (window.ui && window.ui.showMessage) {
                window.ui.showMessage('Пользователь создан успешно!', 'success');
            }
            closeUserModal();
            if (window.loadUsers) {
                window.loadUsers();
            } else if (window.users && window.users.loadUsers) {
                window.users.loadUsers();
            }
        } else {
            const errorMessage = data.message || data.error || 'Ошибка создания пользователя';
            console.error('❌ Ошибка создания пользователя:', errorMessage, data);
            if (window.ui && window.ui.showMessage) {
                window.ui.showMessage(errorMessage, 'error');
            }
        }
    } catch (error) {
        console.error('❌ Ошибка создания пользователя:', error);
        if (window.ui && window.ui.showMessage) {
            window.ui.showMessage('Ошибка создания пользователя: ' + error.message, 'error');
        }
    }
}

// Экспорт
window.users = {
    loadUsers,
    displayUsers,
    viewUser,
    editUser,
    deleteUser,
    openUserModal,
    closeUserModal,
    loadRolesForSelect,
    handleUserSubmit,
    handleRoleChange,
    loadCitiesForSelect
};

// Экспортируем handleRoleChange глобально для использования в HTML
window.handleRoleChange = handleRoleChange;

// Глобальные функции для обратной совместимости
window.openUserModal = openUserModal;
window.closeUserModal = closeUserModal;

