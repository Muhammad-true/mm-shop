// ===== UI.JS - UI утилиты =====

// Показ сообщений
function showMessage(text, type = 'success') {
    // Удаляем предыдущие сообщения
    const existingMessages = document.querySelectorAll('.message');
    existingMessages.forEach(msg => msg.remove());
    
    const message = document.createElement('div');
    message.className = `message ${type}`;
    
    // Выбираем иконку в зависимости от типа сообщения
    let icon = 'ℹ️';
    if (type === 'success') icon = '✅';
    else if (type === 'error') icon = '❌';
    else if (type === 'warning') icon = '⚠️';
    
    message.innerHTML = `
        <div class="message-content">
            <span class="message-icon">${icon}</span>
            <span class="message-text">${text}</span>
            <button class="message-close" onclick="this.parentElement.parentElement.remove()">×</button>
        </div>
    `;
    
    // Добавляем в начало body для лучшей видимости
    document.body.insertBefore(message, document.body.firstChild);
    
    // Автоматически скрываем через 4 секунды
    setTimeout(() => {
        if (message.parentElement) {
            message.remove();
        }
    }, 4000);
}

// Показ модального окна
function showModal(title, content) {
    // Создаем элемент модального окна
    const modal = document.createElement('div');
    modal.className = 'modal';
    modal.style.display = 'block';
    modal.innerHTML = `
        <div class="modal-content">
            <div class="modal-header">
                <h3>${title}</h3>
                <span class="close" onclick="this.closest('.modal').remove()">&times;</span>
            </div>
            <div class="modal-body">
                ${content}
            </div>
        </div>
    `;
    
    document.body.appendChild(modal);
    
    // Закрытие по клику вне модального окна
    modal.addEventListener('click', function(event) {
        if (event.target === modal) {
            modal.remove();
        }
    });
    
    return modal;
}

// Диалог подтверждения
function showConfirmDialog(title, message, description = '', confirmText = 'Да', cancelText = 'Нет') {
    return new Promise((resolve) => {
        const modalDiv = document.createElement('div');
        modalDiv.className = 'modal';
        modalDiv.style.display = 'block';
        modalDiv.style.zIndex = '9999';
        
        const handleConfirm = () => {
            modalDiv.remove();
            resolve(true);
        };
        
        const handleCancel = () => {
            modalDiv.remove();
            resolve(false);
        };
        
        modalDiv.innerHTML = `
            <div class="modal-content" style="max-width: 400px; margin: 100px auto;">
                <div class="modal-header">
                    <h3 style="color: #e74c3c;"><i class="fas fa-exclamation-triangle"></i> ${title}</h3>
                    <span class="close" onclick="handleCancel()">&times;</span>
                </div>
                <div style="padding: 20px;">
                    <p style="font-size: 16px; margin-bottom: 10px; color: #2c3e50;">${message}</p>
                    ${description ? `<p style="font-size: 14px; color: #7f8c8d; margin-bottom: 20px;">${description}</p>` : ''}
                    <div style="display: flex; gap: 10px; justify-content: flex-end;">
                        <button class="btn btn-secondary" id="cancel-confirm-btn">${cancelText}</button>
                        <button class="btn btn-danger" id="confirm-btn">${confirmText}</button>
                    </div>
                </div>
            </div>
        `;
        
        document.body.appendChild(modalDiv);
        
        // Привязываем обработчики
        const confirmBtn = modalDiv.querySelector('#confirm-btn');
        const cancelBtn = modalDiv.querySelector('#cancel-confirm-btn');
        
        confirmBtn.addEventListener('click', handleConfirm);
        cancelBtn.addEventListener('click', handleCancel);
        
        // Закрытие по клику вне модального окна
        modalDiv.addEventListener('click', (e) => {
            if (e.target === modalDiv) {
                handleCancel();
            }
        });
        
        // Обновляем глобальные функции для onclick
        window.handleConfirm = handleConfirm;
        window.handleCancel = handleCancel;
    });
}

// Экспорт функций
window.ui = {
    showMessage,
    showModal,
    showConfirmDialog
};

