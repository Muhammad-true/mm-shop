// ===== UPDATES.JS - Управление обновлениями =====

(function() {
    const updatesModule = {
        async init() {
            const form = document.getElementById('update-upload-form');
            if (form) {
                form.addEventListener('submit', this.handleUpload.bind(this));
            }
            await this.loadUpdates();
        },

        getToken() {
            if (window.storage && window.storage.getAdminToken) return window.storage.getAdminToken();
            return null;
        },

        async handleUpload(e) {
            e.preventDefault();
            const platform = document.getElementById('update-platform').value;
            const version = document.getElementById('update-version').value.trim();
            const notes = document.getElementById('update-notes').value.trim();
            const fileInput = document.getElementById('update-file');
            const file = fileInput.files[0];

            if (!file) {
                window.ui?.showMessage ? window.ui.showMessage('Выберите файл обновления', 'error') : alert('Выберите файл');
                return;
            }

            const formData = new FormData();
            formData.append('platform', platform);
            formData.append('version', version);
            formData.append('releaseNotes', notes);
            formData.append('file', file);

            try {
                const token = this.getToken();
                const response = await fetch(window.getApiUrl('/api/v1/admin/updates/upload'), {
                    method: 'POST',
                    headers: {
                        ...(token ? { 'Authorization': `Bearer ${token}` } : {})
                    },
                    body: formData
                });

                const data = await response.json();
                if (!response.ok || !data.success) {
                    throw new Error(data.error || data.message || 'Ошибка загрузки');
                }

                window.ui?.showMessage ? window.ui.showMessage('Обновление загружено', 'success') : alert('Обновление загружено');
                e.target.reset();
                await this.loadUpdates();
            } catch (err) {
                console.error('Ошибка загрузки обновления:', err);
                window.ui?.showMessage ? window.ui.showMessage(err.message || 'Ошибка загрузки', 'error') : alert(err.message || 'Ошибка');
            }
        },

        async loadUpdates() {
            const container = document.getElementById('updates-table');
            if (!container) return;
            container.innerHTML = '<p>Загрузка...</p>';

            try {
                const data = await fetchData('/api/v1/admin/updates');
                const updates = data.data || data.updates || [];
                container.innerHTML = this.renderTable(updates);
            } catch (err) {
                console.error('Ошибка загрузки обновлений:', err);
                container.innerHTML = `<p style="color:red;">Ошибка загрузки: ${err.message}</p>`;
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
        }
    };

    window.updates = updatesModule;
})();

