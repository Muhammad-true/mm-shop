// ===== SHOPS.JS - Управление магазинами и лицензиями =====

window.shops = {
    currentPage: 1,
    currentLimit: 50,
    currentFilter: {},

    // Загрузка списка магазинов с лицензиями
    async loadShops(page = 1, filters = {}) {
        console.log('🛍️ Загрузка магазинов, страница:', page, 'фильтры:', filters);
        
        this.currentPage = page;
        this.currentFilter = filters;
        
        const shopsTable = document.getElementById('shops-table-body');
        if (!shopsTable) {
            console.error('❌ Таблица магазинов не найдена');
            return;
        }

        shopsTable.innerHTML = '<tr><td colspan="8" class="text-center loading"><i class="fas fa-spinner fa-spin"></i> Загрузка магазинов...</td></tr>';

        try {
            const token = window.storage?.getAdminToken() || localStorage.getItem('adminToken');
            if (!token) {
                throw new Error('Токен не найден');
            }

            // Формируем URL с параметрами
            const params = new URLSearchParams({
                page: page.toString(),
                limit: this.currentLimit.toString(),
                ...filters
            });

            const response = await fetch(`${getApiUrl('/api/v1/admin/shops')}?${params}`, {
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                }
            });

            if (!response.ok) {
                throw new Error(`Ошибка: ${response.status}`);
            }

            const result = await response.json();
            console.log('✅ Магазины загружены:', result);

            if (result.success && result.data) {
                this.renderShops(result.data.shops || [], result.data.pagination || {});
            } else {
                throw new Error('Неверный формат ответа');
            }
        } catch (error) {
            console.error('❌ Ошибка загрузки магазинов:', error);
            shopsTable.innerHTML = `<tr><td colspan="8" class="text-center error">Ошибка загрузки: ${error.message}</td></tr>`;
        }
    },

    // Отображение магазинов в таблице
    renderShops(shops, pagination) {
        const shopsTable = document.getElementById('shops-table-body');
        if (!shopsTable) return;

        if (shops.length === 0) {
            shopsTable.innerHTML = '<tr><td colspan="8" class="text-center">Магазины не найдены</td></tr>';
            return;
        }

        shopsTable.innerHTML = shops.map(shop => {
            const license = shop.license || null;
            const hasLicense = shop.hasLicense || false;
            const licenseStatus = license ? this.getLicenseStatusBadge(license) : '<span class="badge badge-danger">Нет лицензии</span>';
            const daysRemaining = license?.daysRemaining !== null && license?.daysRemaining !== undefined 
                ? license.daysRemaining 
                : license?.subscriptionType === 'lifetime' ? '∞' : '-';

            return `
                <tr>
                    <td>
                        <div style="font-weight: 600;">${shop.name || 'Без названия'}</div>
                        <small style="color: #666;">${shop.email || '-'}</small>
                    </td>
                    <td>
                        <div>${shop.owner?.name || '-'}</div>
                        <small style="color: #666;">${shop.owner?.email || '-'}</small>
                    </td>
                    <td>${shop.productsCount || 0}</td>
                    <td>${shop.subscribersCount || 0}</td>
                    <td>${licenseStatus}</td>
                    <td>
                        ${hasLicense && license ? `
                            <div>${daysRemaining !== '∞' ? daysRemaining + ' дн.' : 'Бессрочно'}</div>
                            <small style="color: #666;">${license.expiresAt ? new Date(license.expiresAt).toLocaleDateString('ru-RU') : '-'}</small>
                        ` : '-'}
                    </td>
                    <td>
                        ${hasLicense && license ? `
                            <div style="font-weight: 600; font-family: monospace; font-size: 12px;">${license.licenseKey || '-'}</div>
                        ` : '-'}
                    </td>
                    <td>
                        <div style="display: flex; gap: 5px; flex-wrap: wrap;">
                            ${!hasLicense ? `
                                <button class="btn btn-sm btn-success" onclick="window.shops.generateLicense('${shop.id}')" title="Создать лицензию">
                                    <i class="fas fa-key"></i>
                                </button>
                            ` : ''}
                            ${hasLicense && license ? `
                                <button class="btn btn-sm btn-primary" onclick="window.shops.extendLicense('${license.id}')" title="Продлить лицензию">
                                    <i class="fas fa-calendar-plus"></i>
                                </button>
                                <button class="btn btn-sm btn-info" onclick="window.shops.viewLicense('${license.id}')" title="Просмотр лицензии">
                                    <i class="fas fa-eye"></i>
                                </button>
                                <button class="btn btn-sm btn-danger" onclick="window.shops.deleteLicense('${license.id}', '${shop.name}')" title="Удалить лицензию">
                                    <i class="fas fa-trash"></i>
                                </button>
                            ` : ''}
                        </div>
                    </td>
                </tr>
            `;
        }).join('');

        // Обновляем пагинацию
        this.updatePagination(pagination);
    },

    // Бейдж статуса лицензии
    getLicenseStatusBadge(license) {
        if (!license) return '<span class="badge badge-danger">Нет</span>';
        
        if (license.isExpired) {
            return '<span class="badge badge-danger">Истекла</span>';
        } else if (license.isValid) {
            return '<span class="badge badge-success">Активна</span>';
        } else if (license.subscriptionStatus === 'pending') {
            return '<span class="badge badge-warning">Ожидает</span>';
        } else {
            return '<span class="badge badge-secondary">Неактивна</span>';
        }
    },

    // Обновление пагинации
    updatePagination(pagination) {
        const paginationContainer = document.getElementById('shops-pagination');
        if (!paginationContainer || !pagination) return;

        const { page, limit, total, totalPages } = pagination;
        
        let paginationHTML = '<div style="display: flex; justify-content: space-between; align-items: center; margin-top: 20px;">';
        paginationHTML += `<div style="color: #666;">Всего: ${total} магазинов</div>`;
        paginationHTML += '<div style="display: flex; gap: 5px;">';

        // Кнопка "Назад"
        if (page > 1) {
            paginationHTML += `<button class="btn btn-sm" onclick="window.shops.loadShops(${page - 1}, window.shops.currentFilter)"><i class="fas fa-chevron-left"></i></button>`;
        }

        // Номера страниц
        for (let i = Math.max(1, page - 2); i <= Math.min(totalPages, page + 2); i++) {
            paginationHTML += `<button class="btn btn-sm ${i === page ? 'btn-primary' : ''}" onclick="window.shops.loadShops(${i}, window.shops.currentFilter)">${i}</button>`;
        }

        // Кнопка "Вперед"
        if (page < totalPages) {
            paginationHTML += `<button class="btn btn-sm" onclick="window.shops.loadShops(${page + 1}, window.shops.currentFilter)"><i class="fas fa-chevron-right"></i></button>`;
        }

        paginationHTML += '</div></div>';
        paginationContainer.innerHTML = paginationHTML;
    },

    // Генерация лицензии для магазина
    async generateLicense(shopId) {
        console.log('🔑 Генерация лицензии для магазина:', shopId);
        
        // Показываем модальное окно для выбора параметров лицензии
        const modal = document.getElementById('generate-license-modal');
        if (modal) {
            document.getElementById('generate-license-shop-id').value = shopId;
            modal.style.display = 'block';
        } else {
            // Если модального окна нет, используем простой prompt
            const months = prompt('Введите количество месяцев (1, 3, 6, 12) или "lifetime" для бессрочной:');
            if (!months) return;

            const subscriptionType = months === 'lifetime' ? 'lifetime' : 'monthly';
            const durationMonths = months === 'lifetime' ? 0 : parseInt(months);

            await this.createLicense(shopId, {
                subscriptionType: subscriptionType,
                paymentAmount: 0,
                paymentCurrency: 'USD',
                paymentProvider: 'manual',
                paymentTransactionId: '',
                autoRenew: false,
                notes: `Создано вручную админом, ${months === 'lifetime' ? 'бессрочно' : months + ' мес.'}`
            });
        }
    },

    // Создание лицензии
    async createLicense(shopId, licenseData) {
        try {
            const token = window.storage?.getAdminToken() || localStorage.getItem('adminToken');
            if (!token) {
                throw new Error('Токен не найден');
            }

            const response = await fetch(`${getApiUrl(`/api/v1/admin/licenses/shops/${shopId}/generate`)}`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(licenseData)
            });

            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.error || 'Ошибка создания лицензии');
            }

            const result = await response.json();
            console.log('✅ Лицензия создана:', result);

            alert('Лицензия успешно создана!');
            this.loadShops(this.currentPage, this.currentFilter);
            
            // Закрываем модальное окно если есть
            const modal = document.getElementById('generate-license-modal');
            if (modal) modal.style.display = 'none';
        } catch (error) {
            console.error('❌ Ошибка создания лицензии:', error);
            alert('Ошибка создания лицензии: ' + error.message);
        }
    },

    // Продление лицензии
    async extendLicense(licenseId) {
        console.log('📅 Продление лицензии:', licenseId);
        
        const months = prompt('Введите количество месяцев для продления:');
        if (!months || isNaN(months) || parseInt(months) < 1) {
            alert('Введите корректное количество месяцев (минимум 1)');
            return;
        }

        try {
            const token = window.storage?.getAdminToken() || localStorage.getItem('adminToken');
            if (!token) {
                throw new Error('Токен не найден');
            }

            const response = await fetch(`${getApiUrl(`/api/v1/admin/licenses/${licenseId}/extend`)}`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    months: parseInt(months)
                })
            });

            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.error || 'Ошибка продления лицензии');
            }

            const result = await response.json();
            console.log('✅ Лицензия продлена:', result);

            alert(`Лицензия успешно продлена на ${months} месяцев!`);
            this.loadShops(this.currentPage, this.currentFilter);
        } catch (error) {
            console.error('❌ Ошибка продления лицензии:', error);
            alert('Ошибка продления лицензии: ' + error.message);
        }
    },

    // Просмотр лицензии
    async viewLicense(licenseId) {
        console.log('👁️ Просмотр лицензии:', licenseId);
        
        try {
            const token = window.storage?.getAdminToken() || localStorage.getItem('adminToken');
            if (!token) {
                throw new Error('Токен не найден');
            }

            const response = await fetch(`${getApiUrl(`/api/v1/admin/licenses/${licenseId}`)}`, {
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                }
            });

            if (!response.ok) {
                throw new Error('Ошибка загрузки лицензии');
            }

            const result = await response.json();
            const license = result.data;

            // Показываем информацию о лицензии
            const info = `
Лицензия: ${license.licenseKey}
Статус: ${license.subscriptionStatus}
Тип: ${license.subscriptionType}
Активирована: ${license.activatedAt ? new Date(license.activatedAt).toLocaleString('ru-RU') : '-'}
Истекает: ${license.expiresAt ? new Date(license.expiresAt).toLocaleString('ru-RU') : 'Бессрочно'}
Осталось дней: ${license.daysRemaining !== null ? license.daysRemaining : '∞'}
Провайдер: ${license.paymentProvider || '-'}
Сумма: ${license.paymentAmount || 0} ${license.paymentCurrency || 'USD'}
Активна: ${license.isActive ? 'Да' : 'Нет'}
Валидна: ${license.isValid ? 'Да' : 'Нет'}
Истекла: ${license.isExpired ? 'Да' : 'Нет'}
            `;

            alert(info);
        } catch (error) {
            console.error('❌ Ошибка просмотра лицензии:', error);
            alert('Ошибка загрузки лицензии: ' + error.message);
        }
    },

    // Применение фильтров
    applyFilters() {
        const filters = {};
        
        const hasLicense = document.getElementById('shop-license-filter')?.value;
        if (hasLicense) filters.hasLicense = hasLicense;
        
        const search = document.getElementById('shop-search')?.value.trim();
        if (search) filters.search = search;

        this.loadShops(1, filters);
    },

    // Удаление лицензии
    async deleteLicense(licenseId, shopName) {
        console.log('🗑️ Удаление лицензии:', licenseId);
        
        const confirmed = confirm(`Вы уверены, что хотите удалить лицензию для магазина "${shopName}"?\n\nЭто действие нельзя отменить!`);
        if (!confirmed) {
            return;
        }

        try {
            const token = window.storage?.getAdminToken() || localStorage.getItem('adminToken');
            if (!token) {
                throw new Error('Токен не найден');
            }

            const response = await fetch(`${getApiUrl(`/api/v1/admin/licenses/${licenseId}`)}`, {
                method: 'DELETE',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                }
            });

            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.error || 'Ошибка удаления лицензии');
            }

            const result = await response.json();
            console.log('✅ Лицензия удалена:', result);

            alert('Лицензия успешно удалена!');
            this.loadShops(this.currentPage, this.currentFilter);
        } catch (error) {
            console.error('❌ Ошибка удаления лицензии:', error);
            alert('Ошибка удаления лицензии: ' + error.message);
        }
    }
};

// Инициализация при загрузке вкладки
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
        console.log('✅ shops.js загружен');
    });
} else {
    console.log('✅ shops.js загружен');
}

