// ===== UPDATES.JS - Управление обновлениями =====

(function() {
    const updatesModule = {
        isInitialized: false,
        isUploading: false, // Оставляем для совместимости, но больше не используется
        async init() {
            // Форма загрузки удалена - загрузка только через FTP
            // Оставляем функции загрузки в коде для совместимости, но они не вызываются
            this.isInitialized = true;
            await this.loadUpdates();
        },

        getToken() {
            if (window.storage && window.storage.getAdminToken) return window.storage.getAdminToken();
            return null;
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
                const url = window.getApiUrl('/api/v1/admin/updates/upload');
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
            if (!container) {
                console.warn('⚠️ [loadUpdates] Контейнер updates-table не найден');
                return;
            }
            container.innerHTML = '<p>Загрузка...</p>';

            try {
                console.log('📡 [loadUpdates] Начало загрузки обновлений...');
                const data = await fetchData('/api/v1/admin/updates/');
                console.log('✅ [loadUpdates] Данные получены:', data);
                
                const updates = data.data || data.updates || [];
                console.log(`📦 [loadUpdates] Найдено обновлений: ${updates.length}`, updates);
                
                container.innerHTML = this.renderTable(updates);
                console.log('✅ [loadUpdates] Таблица обновлена');
            } catch (err) {
                console.error('❌ [loadUpdates] Ошибка загрузки обновлений:', err);
                const message = this.formatErrorMessage(err, 'list');
                container.innerHTML = `<p style="color:red;">Ошибка загрузки: ${message}</p>`;
            }
        },

        renderTable(updates) {
            if (!updates || updates.length === 0) {
                return `
                    <div style="text-align: center; padding: 40px; color: #666;">
                        <i class="fas fa-inbox" style="font-size: 48px; color: #ccc; margin-bottom: 15px;"></i>
                        <p style="font-size: 16px; margin: 10px 0;">Обновления отсутствуют</p>
                        <p style="font-size: 14px; color: #999;">Загрузи первое обновление через FTP в папку <code>/var/ftp/uploads/</code></p>
                        <p style="font-size: 13px; color: #999; margin-top: 10px;">
                            <a href="/docs/FTP_UPLOAD_GUIDE.md" target="_blank" style="color: #2196F3;">📖 Инструкция по загрузке через FTP</a>
                        </p>
                    </div>
                `;
            }

            const rows = updates.map(u => `
                <tr data-update-id="${u.id}">
                    <td>${u.platform}</td>
                    <td>${u.version}</td>
                    <td>${this.formatSize(u.fileSize)}</td>
                    <td><a href="${u.fileUrl}" target="_blank">${u.fileName}</a></td>
                    <td><code title="${u.checksumSha256 || u.checksumSHA256 || ''}">${(u.checksumSha256 || u.checksumSHA256 || '').substring(0, 16)}...</code></td>
                    <td>${u.releaseNotes ? `<div class="notes" title="${u.releaseNotes}">${u.releaseNotes.length > 50 ? u.releaseNotes.substring(0, 50) + '...' : u.releaseNotes}</div>` : '-'}</td>
                    <td>${new Date(u.createdAt).toLocaleString('ru-RU')}</td>
                    <td>
                        <button class="btn btn-danger btn-sm" onclick="window.updates.deleteUpdate('${u.id}', '${u.platform}', '${u.version}')" title="Удалить обновление">
                            <i class="fas fa-trash"></i> Удалить
                        </button>
                    </td>
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
                                <th>Действия</th>
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

        async deleteUpdate(updateId, platform, version) {
            if (!updateId) {
                window.ui?.showMessage ? window.ui.showMessage('ID обновления не указан', 'error') : alert('ID обновления не указан');
                return;
            }

            const confirmMessage = `Вы уверены, что хотите удалить обновление?\n\nПлатформа: ${platform}\nВерсия: ${version}\n\nЭто действие нельзя отменить!`;
            if (!confirm(confirmMessage)) {
                return;
            }

            try {
                const token = this.getToken();
                const url = window.getApiUrl(`/api/v1/admin/updates/${updateId}`);
                
                const response = await fetch(url, {
                    method: 'DELETE',
                    headers: {
                        'Authorization': `Bearer ${token}`,
                        'Content-Type': 'application/json'
                    }
                });

                const data = await response.json();

                if (!response.ok || !data.success) {
                    throw new Error(data.error || data.message || 'Ошибка удаления');
                }

                window.ui?.showMessage ? window.ui.showMessage('Обновление удалено', 'success') : alert('Обновление удалено');
                
                // Обновляем список обновлений
                await this.loadUpdates();
            } catch (err) {
                console.error('Ошибка удаления обновления:', err);
                const message = this.formatErrorMessage(err, 'delete');
                window.ui?.showMessage ? window.ui.showMessage(message, 'error') : alert(message);
            }
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
            if (context === 'delete') {
                return `Ошибка удаления: ${rawMessage}`;
            }
            return rawMessage;
        }
    };

    window.updates = updatesModule;
})();

