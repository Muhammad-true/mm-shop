// ===== ROLES.JS - Управление ролями =====

// Загрузка ролей
window.loadRoles = async function loadRoles() {
    console.log('🔄 Загружаем роли...');
    try {
        const response = await window.api.fetchData('/api/v1/admin/roles/');
        console.log('📡 Ответ API для ролей:', response);
        if (response.success && response.data) {
            const roles = response.data.roles || [];
            console.log('✅ Роли загружены успешно:', roles.length, 'ролей');
            displayRoles(roles);
        } else {
            console.error('❌ Ошибка в ответе API для ролей:', response);
            if (window.ui && window.ui.showMessage) {
                window.ui.showMessage('Ошибка загрузки ролей: ' + (response.message || 'Неизвестная ошибка'), 'error');
            }
        }
    } catch (error) {
        console.error('❌ Ошибка загрузки ролей:', error);
        if (window.ui && window.ui.showMessage) {
            window.ui.showMessage('Ошибка загрузки ролей: ' + error.message, 'error');
        }
    }
};

// Отображение ролей
function displayRoles(roles) {
    console.log('🔄 displayRoles вызвана с данными:', roles);
    const tbody = document.getElementById('roles-table-body');
    
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
                    <div style="width: 45px; height: 45px; border-radius: 50%; background: linear-gradient(135deg, #f093fb, #f5576c); display: flex; align-items: center; justify-content: center; color: white; font-size: 18px;">
                        <i class="fas fa-user-shield"></i>
                    </div>
                    <div>
                        <div style="font-weight: 700; color: #333; font-size: 15px;">${role.displayName}</div>
                        <div style="font-size: 11px; color: #888; font-family: monospace;">${role.name}</div>
                    </div>
                </div>
            </td>
            <td>${role.displayName}</td>
            <td>
                <div style="display: flex; align-items: center; gap: 10px;">
                    <i class="fas fa-info-circle" style="color: #4ecdc4;"></i>
                    <span style="font-weight: 500; max-width: 200px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;" title="${role.description || 'Нет описания'}">
                        ${role.description || 'Нет описания'}
                    </span>
                </div>
            </td>
            <td>
                <div style="display: flex; align-items: center; gap: 8px;">
                    <i class="fas fa-key" style="color: #f093fb;"></i>
                    <span style="font-size: 13px; color: #666;">
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
                    <i class="fas fa-users" style="color: #45b7d1;"></i>
                    <span class="badge" style="background: linear-gradient(135deg, #45b7d1, #96ceb4); color: white;">
                        ${role.userCount || 0}
                    </span>
                </div>
            </td>
            <td>
                <div style="display: flex; gap: 6px;">
                    <button class="btn-sm btn-primary" onclick="window.roles.viewRole('${role.id}')" title="Просмотр">
                        <i class="fas fa-eye"></i>
                    </button>
                    <button class="btn-sm btn-secondary" onclick="window.roles.editRole('${role.id}')" title="Редактировать" ${role.isSystem ? 'disabled' : ''}>
                        <i class="fas fa-edit"></i>
                    </button>
                    <button class="btn-sm btn-danger" onclick="window.roles.deleteRole('${role.id}')" title="Удалить" ${role.isSystem ? 'disabled' : ''}>
                        <i class="fas fa-trash"></i>
                    </button>
                </div>
            </td>
        `;
        tbody.appendChild(row);
    });
}

// Открытие модального окна роли
function openRoleModal() {
    document.getElementById('role-modal').style.display = 'block';
}

// Закрытие модального окна роли
function closeRoleModal() {
    document.getElementById('role-modal').style.display = 'none';
    document.getElementById('role-form').reset();
}

// Обработка отправки формы роли
async function handleRoleSubmit(e) {
    e.preventDefault();
    
    const formData = {
        name: document.getElementById('role-name').value,
        displayName: document.getElementById('role-display-name').value,
        description: document.getElementById('role-description').value,
        permissions: document.getElementById('role-permissions').value || '{}'
    };
    
    try {
        const adminToken = window.storage ? window.storage.getAdminToken() : adminToken;
        
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
            if (window.ui && window.ui.showMessage) {
                window.ui.showMessage('Роль создана успешно!', 'success');
            }
            closeRoleModal();
            loadRoles();
        } else {
            if (window.ui && window.ui.showMessage) {
                window.ui.showMessage(data.message || 'Ошибка создания роли', 'error');
            }
        }
    } catch (error) {
        console.error('Ошибка создания роли:', error);
        if (window.ui && window.ui.showMessage) {
            window.ui.showMessage('Ошибка создания роли', 'error');
        }
    }
}

// Просмотр роли
async function viewRole(id) {
    try {
        const response = await window.api.fetchData(`/api/v1/admin/roles/${id}`);
        const role = response.data;
        
        if (window.ui && window.ui.showModal) {
            window.ui.showModal('Информация о роли', `
                <div style="padding: 20px;">
                    <h4>${role.displayName}</h4>
                    <p><strong>Название:</strong> ${role.name}</p>
                    <p><strong>Описание:</strong> ${role.description || 'Нет описания'}</p>
                    <p><strong>Разрешения:</strong> ${role.permissions || 'Нет разрешений'}</p>
                    <p><strong>Статус:</strong> ${role.isActive ? 'Активна' : 'Неактивна'}</p>
                    <p><strong>Пользователей с этой ролью:</strong> ${role.userCount || 0}</p>
                </div>
            `);
        }
    } catch (error) {
        console.error('Ошибка загрузки роли:', error);
        if (window.ui && window.ui.showMessage) {
            window.ui.showMessage('Ошибка загрузки роли', 'error');
        }
    }
}

// Редактирование роли
async function editRole(id) {
    console.log('Редактирование роли:', id);
    if (window.ui && window.ui.showMessage) {
        window.ui.showMessage('Функция редактирования роли пока не реализована', 'info');
    }
}

// Удаление роли
async function deleteRole(id) {
    if (!confirm('Вы уверены, что хотите удалить эту роль?')) {
        return;
    }
    
    try {
        await window.api.fetchData(`/api/v1/admin/roles/${id}`, {
            method: 'DELETE'
        });
        
        if (window.ui && window.ui.showMessage) {
            window.ui.showMessage('Роль успешно удалена', 'success');
        }
        loadRoles();
    } catch (error) {
        console.error('Ошибка удаления роли:', error);
        if (window.ui && window.ui.showMessage) {
            window.ui.showMessage('Ошибка удаления роли', 'error');
        }
    }
}

// Экспорт
window.roles = {
    loadRoles,
    displayRoles,
    openRoleModal,
    closeRoleModal,
    handleRoleSubmit,
    viewRole,
    editRole,
    deleteRole
};


