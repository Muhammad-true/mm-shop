// ===== UPDATES.JS - Управление обновлениями =====

(function() {
    const updatesModule = {
        isInitialized: false,
        isUploading: false,
        async init() {
            const form = document.getElementById('update-upload-form');
            if (form && !this.isInitialized) {
                form.addEventListener('submit', this.handleUpload.bind(this));
                this.isInitialized = true;
            }
            await this.loadUpdates();
        },

        getToken() {
            if (window.storage && window.storage.getAdminToken) return window.storage.getAdminToken();
            return null;
        },

        // Получить прямой URL для загрузки (обход Cloudflare)
        // Cloudflare может обрывать соединения при загрузке больших файлов (>100 секунд)
        getDirectUploadUrl(endpoint) {
            // Используем прямой IP только в продакшене и только для загрузки файлов
            const isProduction = location.hostname !== 'localhost' && location.hostname !== '127.0.0.1';
            if (!isProduction) return null;
            
            // Прямой IP сервера (обход Cloudflare)
            // ВАЖНО: убедись, что на сервере настроен SSL для этого IP или используй HTTP
            // Если используешь HTTP, убедись, что nginx слушает на порту 80 для прямого IP
            const DIRECT_SERVER_IP = '159.89.99.252';
            const DIRECT_SERVER_PORT = '443'; // HTTPS через nginx
            const DIRECT_SERVER_PROTOCOL = 'https';
            
            // Формируем URL с прямым IP
            // Используем Host header для правильной маршрутизации в nginx
            return `${DIRECT_SERVER_PROTOCOL}://${DIRECT_SERVER_IP}:${DIRECT_SERVER_PORT}${endpoint}`;
        },

        async handleUpload(e) {
            e.preventDefault();
            if (this.isUploading) {
                this.setUploadStatus('Загрузка уже выполняется. Пожалуйста, подождите...', 'info');
                return;
            }

            const platform = document.getElementById('update-platform').value;
            const version = document.getElementById('update-version').value.trim();
            const notes = document.getElementById('update-notes').value.trim();
            const fileInput = document.getElementById('update-file');
            const file = fileInput.files[0];

            if (!platform || !version) {
                this.setUploadStatus('Укажите платформу и версию', 'error');
                return;
            }

            if (!file) {
                this.setUploadStatus('Выберите файл обновления', 'error');
                window.ui?.showMessage ? window.ui.showMessage('Выберите файл обновления', 'error') : alert('Выберите файл');
                return;
            }

            const formData = new FormData();
            formData.append('platform', platform);
            formData.append('version', version);
            formData.append('releaseNotes', notes);
            formData.append('file', file);

            try {
                this.isUploading = true;
                this.setUploadButtonState(true, 'Загрузка...');
                this.setUploadStatus('Загрузка файла...', 'info');
                this.resetProgress();

                const token = this.getToken();
                
                // ОБХОД CLOUDFLARE: используем прямой IP для загрузки больших файлов
                // Cloudflare имеет таймаут 100 секунд на бесплатном тарифе
                // Прямой IP обходит Cloudflare и позволяет загружать файлы без ограничений
                const directIpUrl = this.getDirectUploadUrl('/api/v1/admin/updates/upload');
                const url = directIpUrl || window.getApiUrl('/api/v1/admin/updates/upload');
                
                const data = await this.uploadWithProgress(url, formData, token);

                if (!data || data.success === false) {
                    throw new Error(data?.error || data?.message || 'Ошибка загрузки');
                }

                this.setUploadStatus('Обновление загружено', 'success');
                window.ui?.showMessage ? window.ui.showMessage('Обновление загружено', 'success') : alert('Обновление загружено');
                e.target.reset();
                await this.loadUpdates();
            } catch (err) {
                console.error('Ошибка загрузки обновления:', err);
                const message = this.formatErrorMessage(err, 'upload');
                this.setUploadStatus(message, 'error');
                window.ui?.showMessage ? window.ui.showMessage(message, 'error') : alert(message);
            } finally {
                this.isUploading = false;
                this.setUploadButtonState(false);
                this.hideProgress();
            }
        },

        async loadUpdates() {
            const container = document.getElementById('updates-table');
            if (!container) return;
            container.innerHTML = '<p>Загрузка...</p>';

            try {
                const data = await fetchData('/api/v1/admin/updates/');
                const updates = data.data || data.updates || [];
                container.innerHTML = this.renderTable(updates);
            } catch (err) {
                console.error('Ошибка загрузки обновлений:', err);
                const message = this.formatErrorMessage(err, 'list');
                container.innerHTML = `<p style="color:red;">Ошибка загрузки: ${message}</p>`;
            }
        },

        renderTable(updates) {
            if (!updates || updates.length === 0) {
                return '<p>Обновления отсутствуют</p>';
            }

            const rows = updates.map(u => `
                <tr>
                    <td>${u.platform}</td>
                    <td>${u.version}</td>
                    <td>${this.formatSize(u.fileSize)}</td>
                    <td><a href="${u.fileUrl}" target="_blank">${u.fileName}</a></td>
                    <td><code>${u.checksumSha256 || u.checksumSHA256 || ''}</code></td>
                    <td>${u.releaseNotes ? `<div class="notes">${u.releaseNotes}</div>` : '-'}</td>
                    <td>${new Date(u.createdAt).toLocaleString('ru-RU')}</td>
                </tr>
            `).join('');

            return `
                <div class="table-responsive">
                    <table class="data-table">
                        <thead>
                            <tr>
                                <th>Платформа</th>
                                <th>Версия</th>
                                <th>Размер</th>
                                <th>Файл</th>
                                <th>SHA256</th>
                                <th>Описание</th>
                                <th>Загружено</th>
                            </tr>
                        </thead>
                        <tbody>
                            ${rows}
                        </tbody>
                    </table>
                </div>
            `;
        },

        formatSize(bytes) {
            if (!bytes || bytes <= 0) return '0 B';
            const units = ['B', 'KB', 'MB', 'GB'];
            const i = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1);
            return (bytes / Math.pow(1024, i)).toFixed(2) + ' ' + units[i];
        },

        setUploadStatus(message, type = 'info') {
            const status = document.getElementById('update-upload-status');
            if (!status) return;
            status.style.display = message ? 'block' : 'none';
            status.textContent = message || '';

            let color = '#666';
            if (type === 'error') color = '#d9534f';
            if (type === 'success') color = '#28a745';
            status.style.color = color;
        },

        setUploadButtonState(disabled, label) {
            const button = document.getElementById('update-upload-submit');
            if (!button) return;
            if (!button.dataset.originalHtml) {
                button.dataset.originalHtml = button.innerHTML;
            }
            button.disabled = !!disabled;
            button.innerHTML = disabled ? `<i class="fas fa-spinner fa-spin"></i> ${label || 'Загрузка...'}` : button.dataset.originalHtml;
        },

        resetProgress() {
            const wrap = document.getElementById('update-upload-progress');
            const bar = document.getElementById('update-upload-progress-bar');
            const text = document.getElementById('update-upload-progress-text');
            const time = document.getElementById('update-upload-progress-time');
            if (!wrap || !bar || !text || !time) return;
            wrap.style.display = 'block';
            bar.style.width = '0%';
            text.textContent = 'Загрузка: 0%';
            time.textContent = '';
        },

        updateProgress(percent, elapsedMs) {
            const bar = document.getElementById('update-upload-progress-bar');
            const text = document.getElementById('update-upload-progress-text');
            const time = document.getElementById('update-upload-progress-time');
            if (!bar || !text || !time) return;
            const pct = Math.max(0, Math.min(100, Math.round(percent)));
            bar.style.width = `${pct}%`;
            text.textContent = `Загрузка: ${pct}%`;
            time.textContent = elapsedMs ? `Прошло: ${this.formatElapsed(elapsedMs)}` : '';
        },

        hideProgress() {
            const wrap = document.getElementById('update-upload-progress');
            if (wrap) wrap.style.display = 'none';
        },

        formatElapsed(ms) {
            const totalSeconds = Math.floor(ms / 1000);
            const minutes = Math.floor(totalSeconds / 60);
            const seconds = totalSeconds % 60;
            if (minutes <= 0) return `${seconds}с`;
            return `${minutes}м ${seconds}с`;
        },

        uploadWithProgress(url, formData, token) {
            return new Promise((resolve, reject) => {
                const xhr = new XMLHttpRequest();
                const startTime = Date.now();

                // Вычисляем динамический таймаут на основе размера файла
                // Базовый таймаут: 5 минут + 1 минута на каждый МБ файла (минимум 30 минут для больших файлов)
                const fileInput = document.getElementById('update-file');
                const file = fileInput?.files?.[0];
                let timeoutMs = 30 * 60 * 1000; // 30 минут по умолчанию
                
                if (file && file.size) {
                    const fileSizeMB = file.size / (1024 * 1024);
                    // Минимум 30 минут, максимум 60 минут для очень больших файлов
                    timeoutMs = Math.max(30 * 60 * 1000, Math.min(60 * 60 * 1000, (5 + fileSizeMB) * 60 * 1000));
                    console.log(`⏱️ Установлен таймаут ${Math.round(timeoutMs / 60000)} минут для файла ${fileSizeMB.toFixed(2)} МБ`);
                }

                xhr.open('POST', url, true);
                xhr.timeout = timeoutMs;
                
                // Если используем прямой IP, добавляем Host header для правильной маршрутизации в nginx
                if (url.includes('159.89.99.252')) {
                    xhr.setRequestHeader('Host', 'api.libiss.com');
                }
                
                if (token) {
                    xhr.setRequestHeader('Authorization', `Bearer ${token}`);
                }

                xhr.upload.onprogress = (event) => {
                    if (!event.lengthComputable) return;
                    const percent = (event.loaded / event.total) * 100;
                    this.updateProgress(percent, Date.now() - startTime);
                };

                xhr.onerror = () => {
                    reject(new Error('Сетевая ошибка. Проверьте интернет и доступность API.'));
                };

                xhr.ontimeout = () => {
                    reject(new Error('Истек таймаут загрузки. Сервер не ответил вовремя.'));
                };

                xhr.onload = () => {
                    let responseData = null;
                    try {
                        responseData = JSON.parse(xhr.responseText || '{}');
                    } catch (e) {
                        // игнор
                    }

                    if (xhr.status < 200 || xhr.status >= 300) {
                        const serverMessage = responseData?.error || responseData?.message;
                        const raw = !serverMessage && xhr.responseText ? xhr.responseText : '';
                        const statusText = xhr.statusText ? ` ${xhr.statusText}` : '';
                        const message = serverMessage || raw || 'Неизвестная ошибка сервера';
                        reject(new Error(`Сервер вернул ${xhr.status}${statusText}: ${message}`));
                        return;
                    }
                    resolve(responseData);
                };

                xhr.send(formData);
            });
        },

        formatErrorMessage(err, context) {
            const rawMessage = err?.message || 'Неизвестная ошибка';
            if (rawMessage.includes('Failed to fetch')) {
                return 'Сетевая ошибка: запрос не выполнен (CORS/нет соединения/сервер недоступен).';
            }
            if (rawMessage.toLowerCase().includes('timeout')) {
                return `Сетевая ошибка: ${rawMessage}`;
            }
            if (rawMessage.startsWith('Сервер вернул')) {
                return `Ошибка сервера: ${rawMessage}`;
            }
            if (context === 'upload') {
                return `Ошибка загрузки: ${rawMessage}`;
            }
            return rawMessage;
        }
    };

    window.updates = updatesModule;
})();

