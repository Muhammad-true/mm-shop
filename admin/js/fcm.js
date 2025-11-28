// ===== FCM.JS - Работа с Firebase Cloud Messaging для веб =====

// Упрощенная версия: регистрация токена через API без Firebase SDK
// Для полноценной работы нужна настройка Firebase в config.js

let fcmToken = null;

// Инициализация FCM (упрощенная версия)
async function initFCM() {
    // Проверяем поддержку уведомлений
    if (!('Notification' in window)) {
        console.log('⚠️ Браузер не поддерживает уведомления');
        return false;
    }

    // Проверяем Service Worker поддержку
    if (!('serviceWorker' in navigator)) {
        console.log('⚠️ Браузер не поддерживает Service Worker');
        return false;
    }

    try {
        // Запрашиваем разрешение на уведомления
        const permission = await Notification.requestPermission();
        if (permission === 'granted') {
            console.log('✅ Разрешение на уведомления получено');
            
            // Для веб-версии без Firebase SDK используем упрощенный подход
            // Генерируем токен на основе устройства и пользователя
            await generateAndRegisterToken();
            
            return true;
        } else {
            console.log('❌ Разрешение на уведомления отклонено');
            return false;
        }
    } catch (error) {
        console.error('❌ Ошибка инициализации FCM:', error);
        return false;
    }
}

// Генерация и регистрация токена (упрощенная версия)
async function generateAndRegisterToken() {
    try {
        // Генерируем уникальный токен на основе устройства
        const deviceId = getDeviceId();
        const userData = JSON.parse(localStorage.getItem('userData') || '{}');
        const userId = userData.id || userData.userId || 'unknown';
        
        // Создаем токен в формате, который будет работать с FCM
        // В реальном приложении этот токен должен быть получен от Firebase
        // Здесь мы используем упрощенный подход для демонстрации
        // Для production нужно использовать реальный FCM токен от Firebase
        fcmToken = `web_${deviceId}_${userId}_${Date.now()}`;
        
        console.log('✅ FCM токен сгенерирован:', fcmToken.substring(0, 30) + '...');
        console.log('ℹ️ Для production используйте реальный FCM токен от Firebase');
        
        // Сохраняем токен в localStorage
        localStorage.setItem('fcmToken', fcmToken);
        
        // Регистрируем токен на сервере
        await registerDeviceToken(fcmToken);
        
        return fcmToken;
    } catch (error) {
        console.error('❌ Ошибка генерации токена:', error);
        return null;
    }
}

// Получение FCM токена (для совместимости)
async function getFCMToken() {
    const savedToken = localStorage.getItem('fcmToken');
    if (savedToken) {
        return savedToken;
    }
    return await generateAndRegisterToken();
}

// Регистрация токена устройства на сервере
async function registerDeviceToken(token) {
    try {
        const adminToken = window.storage?.getAdminToken?.() || localStorage.getItem('adminToken');
        if (!adminToken) {
            console.log('⚠️ Токен авторизации отсутствует, пропускаем регистрацию FCM токена');
            return false;
        }

        const response = await window.api.fetchData('/api/v1/device-tokens/', {
            method: 'POST',
            body: JSON.stringify({
                token: token,
                platform: 'web',
                deviceId: getDeviceId()
            })
        });

        if (response?.success || response?.data) {
            console.log('✅ FCM токен зарегистрирован на сервере');
            return true;
        } else {
            console.error('❌ Ошибка регистрации FCM токена:', response);
            return false;
        }
    } catch (error) {
        console.error('❌ Ошибка регистрации FCM токена:', error);
        return false;
    }
}

// Получение уникального ID устройства
function getDeviceId() {
    let deviceId = localStorage.getItem('deviceId');
    if (!deviceId) {
        deviceId = 'web_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
        localStorage.setItem('deviceId', deviceId);
    }
    return deviceId;
}

// Настройка обработчика входящих сообщений (для Service Worker)
async function setupMessageHandler() {
    try {
        // Регистрируем Service Worker для обработки push-уведомлений
        if ('serviceWorker' in navigator) {
            const registration = await navigator.serviceWorker.register('/admin/firebase-messaging-sw.js');
            console.log('✅ Service Worker зарегистрирован для push-уведомлений');
            
            // Слушаем сообщения от Service Worker
            navigator.serviceWorker.addEventListener('message', (event) => {
                console.log('📨 Получено сообщение от Service Worker:', event.data);
                if (event.data && event.data.notification) {
                    showNotification(event.data.notification.title, {
                        body: event.data.notification.body,
                        icon: event.data.notification.icon || '/admin/favicon.ico',
                        data: event.data.data
                    });
                }
            });
        }
    } catch (error) {
        console.error('❌ Ошибка настройки Service Worker:', error);
    }
}

// Показ уведомления
function showNotification(title, options = {}) {
    if (Notification.permission === 'granted') {
        const notification = new Notification(title, {
            icon: options.icon || '/admin/favicon.ico',
            badge: options.badge || '/admin/favicon.ico',
            body: options.body || '',
            tag: options.tag || 'default',
            requireInteraction: false,
            ...options
        });

        // Обработка клика по уведомлению
        notification.onclick = function(event) {
            event.preventDefault();
            window.focus();
            
            // Если есть action_url, переходим по нему
            if (options.data?.action_url) {
                window.location.href = options.data.action_url;
            }
            
            notification.close();
        };

        // Автоматически закрываем через 5 секунд
        setTimeout(() => {
            notification.close();
        }, 5000);
    }
}

// Проверка и регистрация токена при входе
async function checkAndRegisterFCMToken() {
    const adminToken = window.storage?.getAdminToken?.() || localStorage.getItem('adminToken');
    
    if (!adminToken) {
        console.log('⚠️ Пользователь не авторизован, пропускаем регистрацию FCM');
        return;
    }

    // Проверяем, есть ли уже токен
    const savedToken = localStorage.getItem('fcmToken');
    
    if (savedToken) {
        console.log('✅ FCM токен уже сохранен, проверяем актуальность...');
        // Проверяем разрешение на уведомления
        if (Notification.permission === 'granted') {
            await registerDeviceToken(savedToken);
        } else {
            // Если разрешение потеряно, запрашиваем снова
            console.log('🔄 Разрешение на уведомления потеряно, запрашиваем снова...');
            await initFCM();
        }
        return;
    }

    // Если токена нет, инициализируем FCM
    console.log('🔄 Инициализация FCM для регистрации токена...');
    await initFCM();
    
    // Настраиваем обработчик сообщений
    await setupMessageHandler();
}

// Удаление токена при выходе
async function unregisterFCMToken() {
    const token = localStorage.getItem('fcmToken');
    if (token) {
        try {
            await window.api.fetchData(`/api/v1/device-tokens/${encodeURIComponent(token)}`, {
                method: 'DELETE'
            });
            localStorage.removeItem('fcmToken');
            console.log('✅ FCM токен удален');
        } catch (error) {
            console.error('❌ Ошибка удаления FCM токена:', error);
        }
    }
}

// Экспорт
window.fcm = {
    initFCM,
    getFCMToken,
    registerDeviceToken,
    checkAndRegisterFCMToken,
    unregisterFCMToken,
    showNotification
};

