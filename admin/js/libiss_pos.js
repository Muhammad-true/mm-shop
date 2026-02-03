// ===== LIBISS_POS.JS - Управление файлами программ libiss_pos =====

const libissPos = {
    files: [],
    currentFilter: '',

    // Инициализация
    init: function() {
        console.log('🔧 Инициализация управления файлами libiss_pos');
        this.setupEventListeners();
        this.loadFiles();
    },

    // Настройка обработчиков событий
    setupEventListeners: function() {
        const uploadForm = document.getElementById('libiss-pos-upload-form');
        if (uploadForm) {
            uploadForm.addEventListener('submit', (e) => this.handleUpload(e));
        }

        // Обработчик изменения платформы для обновления accept атрибута файла
        const platformSelect = document.getElementById('libiss-pos-platform');
        if (platformSelect) {
            platformSelect.addEventListener('change', (e) => {
                const fileInput = document.getElementById('libiss-pos-file-input');
                const fileLabel = document.getElementById('libiss-pos-file-label');
                const fileHint = document.getElementById('libiss-pos-file-hint');
                
                if (e.target.value === 'windows') {
                    fileInput.accept = '.exe';
                    fileLabel.textContent = 'Файл программы Windows (.exe)';
                    fileHint.textContent = 'Разрешены только .exe файлы';
                } else if (e.target.value === 'android') {
                    fileInput.accept = '.apk';
                    fileLabel.textContent = 'Файл программы Android (.apk)';
                    fileHint.textContent = 'Разрешены только .apk файлы';
                } else {
                    fileInput.accept = '.exe,.apk';
                    fileLabel.textContent = 'Файл программы';
                    fileHint.textContent = 'Выберите платформу для отображения допустимых форматов';
                }
            });
        }

        const filterTypeSelect = document.getElementById('libiss-pos-filter-type');
        if (filterTypeSelect) {
            filterTypeSelect.addEventListener('change', () => {
                this.loadFiles();
            });
        }

        const filterPlatformSelect = document.getElementById('libiss-pos-filter-platform');
        if (filterPlatformSelect) {
            filterPlatformSelect.addEventListener('change', () => {
                this.loadFiles();
            });
        }
    },

    // Загрузка списка файлов
    loadFiles: async function() {
        try {
            const token = window.storage?.getAdminToken() || localStorage.getItem('adminToken');
            if (!token) {
                console.error('❌ Токен не найден');
                return;
            }

            const apiBaseUrl = window.getApiUrl ? window.getApiUrl('') : (window.API_BASE_URL || 'http://localhost:8080');
            
            // Собираем параметры фильтрации
            const filterType = document.getElementById('libiss-pos-filter-type')?.value || '';
            const filterPlatform = document.getElementById('libiss-pos-filter-platform')?.value || '';
            
            let url = `${apiBaseUrl}/api/v1/admin/libiss-pos`;
            const params = [];
            if (filterType) params.push(`type=${filterType}`);
            if (filterPlatform) params.push(`platform=${filterPlatform}`);
            if (params.length > 0) url += '?' + params.join('&');

            const response = await fetch(url, {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                }
            });

            const data = await response.json();
            
            if (data.success && data.data) {
                this.files = data.data;
                this.renderFiles();
            } else {
                console.error('❌ Ошибка загрузки файлов:', data.error);
                window.ui?.showNotification('Ошибка загрузки файлов: ' + (data.error || 'Неизвестная ошибка'), 'error');
            }
        } catch (error) {
            console.error('❌ Ошибка при загрузке файлов:', error);
            window.ui?.showNotification('Ошибка при загрузке файлов', 'error');
        }
    },

    // Обработка загрузки файла
    handleUpload: async function(e) {
        e.preventDefault();
        
        const form = e.target;
        const formData = new FormData(form);
        
        const fileInput = form.querySelector('input[type="file"]');
        if (!fileInput.files || !fileInput.files[0]) {
            window.ui?.showNotification('Выберите файл для загрузки', 'error');
            return;
        }

        const file = fileInput.files[0];
        const platformSelect = form.querySelector('select[name="platform"]');
        const platform = platformSelect?.value;
        
        if (platform === 'windows' && !file.name.endsWith('.exe')) {
            window.ui?.showNotification('Для Windows разрешены только .exe файлы', 'error');
            return;
        }
        if (platform === 'android' && !file.name.endsWith('.apk')) {
            window.ui?.showNotification('Для Android разрешены только .apk файлы', 'error');
            return;
        }

            // Показываем индикатор загрузки
            const submitBtn = form.querySelector('button[type="submit"]');
            const originalText = submitBtn.textContent;
            submitBtn.disabled = true;
            submitBtn.textContent = 'Загрузка...';

            try {
                const token = window.storage?.getAdminToken() || localStorage.getItem('adminToken');
                if (!token) {
                    throw new Error('Токен не найден');
                }

            formData.append('file', file);
            
            // Добавляем isPublic из checkbox
            const isPublicCheckbox = form.querySelector('input[name="isPublic"]');
            if (isPublicCheckbox && isPublicCheckbox.checked) {
                formData.append('isPublic', 'true');
            } else {
                formData.append('isPublic', 'false');
            }

            const apiBaseUrl = window.getApiUrl ? window.getApiUrl('') : (window.API_BASE_URL || 'http://localhost:8080');
            const response = await fetch(`${apiBaseUrl}/api/v1/admin/libiss-pos/upload`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`
                },
                body: formData
            });

            const data = await response.json();

            if (data.success) {
                window.ui?.showNotification('Файл успешно загружен', 'success');
                form.reset();
                this.loadFiles();
            } else {
                window.ui?.showNotification('Ошибка загрузки: ' + (data.error || 'Неизвестная ошибка'), 'error');
            }
        } catch (error) {
            console.error('❌ Ошибка при загрузке файла:', error);
            window.ui?.showNotification('Ошибка при загрузке файла', 'error');
        } finally {
            submitBtn.disabled = false;
            submitBtn.textContent = originalText;
        }
    },

    // Отображение списка файлов
    renderFiles: function() {
        const container = document.getElementById('libiss-pos-files-list');
        if (!container) return;

        if (this.files.length === 0) {
            container.innerHTML = '<div class="empty-state"><p>Файлы не найдены</p></div>';
            return;
        }

        container.innerHTML = this.files.map(file => this.renderFileCard(file)).join('');
    },

    // Отображение карточки файла
    renderFileCard: function(file) {
        const typeNames = {
            'full': 'Полный пакет (Касса1)',
            'cassa2': 'Программа для Касса2',
            'server_only': 'Программа + сервер без MySQL'
        };

        const platformNames = {
            'windows': 'Windows',
            'android': 'Android'
        };

        const formatSize = (bytes) => {
            if (bytes < 1024) return bytes + ' B';
            if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(2) + ' KB';
            return (bytes / (1024 * 1024)).toFixed(2) + ' MB';
        };

        const formatDate = (dateStr) => {
            const date = new Date(dateStr);
            return date.toLocaleString('ru-RU');
        };

        return `
            <div class="file-card">
                <div class="file-card-header">
                    <h3>${file.originalName || file.fileName}</h3>
                    <div class="file-badges">
                        <span class="badge badge-${file.type}">${typeNames[file.type] || file.type}</span>
                        <span class="badge badge-${file.platform === 'android' ? 'android' : 'windows'}">${platformNames[file.platform] || file.platform}</span>
                        ${file.isPublic ? '<span class="badge badge-success">Публичный</span>' : '<span class="badge badge-secondary">Приватный</span>'}
                        ${file.isActive ? '<span class="badge badge-info">Активен</span>' : '<span class="badge badge-warning">Неактивен</span>'}
                    </div>
                </div>
                <div class="file-card-body">
                    <div class="file-info">
                        <p><strong>Версия:</strong> ${file.version}</p>
                        <p><strong>Платформа:</strong> ${platformNames[file.platform] || file.platform}</p>
                        <p><strong>Размер:</strong> ${formatSize(file.fileSize)}</p>
                        <p><strong>Загрузок:</strong> ${file.downloadCount || 0}</p>
                        <p><strong>Загружен:</strong> ${formatDate(file.createdAt)}</p>
                        ${file.description ? `<p><strong>Описание:</strong> ${file.description}</p>` : ''}
                        <p><strong>SHA256:</strong> <code>${file.checksumSha256.substring(0, 16)}...</code></p>
                    </div>
                    <div class="file-actions">
                        <a href="${window.getApiUrl ? window.getApiUrl('') : (window.API_BASE_URL || 'http://localhost:8080')}${file.fileUrl}" 
                           class="btn btn-sm btn-primary" 
                           download
                           target="_blank">
                            <i class="fas fa-download"></i> Скачать (авторизованным)
                        </a>
                        ${file.isPublic ? `
                            <a href="${window.getApiUrl ? window.getApiUrl('') : (window.API_BASE_URL || 'http://localhost:8080')}${file.publicUrl}" 
                               class="btn btn-sm btn-success" 
                               download
                               target="_blank">
                                <i class="fas fa-globe"></i> Публичное скачивание
                            </a>
                        ` : ''}
                        <button class="btn btn-sm btn-danger" onclick="libissPos.deleteFile('${file.id}')">
                            <i class="fas fa-trash"></i> Удалить
                        </button>
                    </div>
                </div>
            </div>
        `;
    },

    // Удаление файла
    deleteFile: async function(fileId) {
        if (!confirm('Вы уверены, что хотите удалить этот файл? Это действие нельзя отменить.')) {
            return;
        }

        try {
            const token = window.storage?.getAdminToken() || localStorage.getItem('adminToken');
            if (!token) {
                throw new Error('Токен не найден');
            }

            const apiBaseUrl = window.getApiUrl ? window.getApiUrl('') : (window.API_BASE_URL || 'http://localhost:8080');
            const response = await fetch(`${apiBaseUrl}/api/v1/admin/libiss-pos/${fileId}`, {
                method: 'DELETE',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                }
            });

            const data = await response.json();

            if (data.success) {
                window.ui?.showNotification('Файл успешно удален', 'success');
                this.loadFiles();
            } else {
                window.ui?.showNotification('Ошибка удаления: ' + (data.error || 'Неизвестная ошибка'), 'error');
            }
        } catch (error) {
            console.error('❌ Ошибка при удалении файла:', error);
            window.ui?.showNotification('Ошибка при удалении файла', 'error');
        }
    }
};

// Экспорт
window.libissPos = libissPos;

