// Исправления для отображения пользователя
// Добавить в script.js после функции updateUserInfo

// Функция для восстановления данных пользователя из localStorage
function restoreUserData() {
    const savedUserData = localStorage.getItem('userData');
    if (savedUserData) {
        try {
            const userData = JSON.parse(savedUserData);
            console.log('🔄 Восстанавливаем данные пользователя из localStorage:', userData);
            
            // Обновляем header с сохраненными данными
            const headerUserName = document.getElementById('header-user-name');
            const headerUserEmail = document.getElementById('header-user-email');
            const headerUserRole = document.getElementById('header-user-role');
            
            if (headerUserName) {
                if (userData.name && userData.name.trim() !== '') {
                    headerUserName.textContent = userData.name;
                } else if (userData.email && userData.email.trim() !== '') {
                    headerUserName.textContent = userData.email.split('@')[0];
                }
            }
            
            if (headerUserEmail) {
                headerUserEmail.textContent = userData.email || '';
            }
            
            if (headerUserRole) {
                const role = userData.role?.name || 'admin';
                switch (role) {
                    case 'admin':
                        headerUserRole.textContent = 'Администратор';
                        break;
                    case 'shop_owner':
                        headerUserRole.textContent = 'Владелец магазина';
                        break;
                    default:
                        headerUserRole.textContent = 'Пользователь';
                }
            }
            
            console.log('✅ Данные пользователя восстановлены из localStorage');
        } catch (error) {
            console.error('❌ Ошибка восстановления данных пользователя:', error);
        }
    }
}

// Вызываем восстановление при загрузке страницы
document.addEventListener('DOMContentLoaded', function() {
    // Небольшая задержка для инициализации
    setTimeout(restoreUserData, 100);
});
