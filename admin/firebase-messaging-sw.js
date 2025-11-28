// Service Worker для обработки push-уведомлений FCM
// Этот файл должен быть доступен по пути /admin/firebase-messaging-sw.js

// Импортируем Firebase SDK (если используется)
// importScripts('https://www.gstatic.com/firebasejs/10.7.1/firebase-app-compat.js');
// importScripts('https://www.gstatic.com/firebasejs/10.7.1/firebase-messaging-compat.js');

// Обработка push-уведомлений
self.addEventListener('push', function(event) {
    console.log('📨 Push уведомление получено:', event);
    
    let notificationData = {
        title: 'Уведомление',
        body: 'У вас новое уведомление',
        icon: '/admin/favicon.ico',
        badge: '/admin/favicon.ico',
        data: {}
    };

    if (event.data) {
        try {
            const payload = event.data.json();
            notificationData = {
                title: payload.notification?.title || payload.title || 'Уведомление',
                body: payload.notification?.body || payload.body || 'У вас новое уведомление',
                icon: payload.notification?.icon || '/admin/favicon.ico',
                badge: '/admin/favicon.ico',
                data: payload.data || {}
            };
        } catch (e) {
            console.error('Ошибка парсинга данных push-уведомления:', e);
        }
    }

    const options = {
        body: notificationData.body,
        icon: notificationData.icon,
        badge: notificationData.badge,
        tag: notificationData.data.notificationId || 'default',
        data: notificationData.data,
        requireInteraction: false
    };

    event.waitUntil(
        self.registration.showNotification(notificationData.title, options)
    );
});

// Обработка клика по уведомлению
self.addEventListener('notificationclick', function(event) {
    console.log('👆 Клик по уведомлению:', event);
    
    event.notification.close();

    const actionUrl = event.notification.data?.action_url || '/admin#dashboard';
    
    event.waitUntil(
        clients.matchAll({ type: 'window', includeUncontrolled: true }).then(function(clientList) {
            // Если есть открытое окно, фокусируемся на нем
            for (let i = 0; i < clientList.length; i++) {
                const client = clientList[i];
                if (client.url.includes('/admin') && 'focus' in client) {
                    client.focus();
                    client.navigate(actionUrl);
                    return;
                }
            }
            // Если окно не открыто, открываем новое
            if (clients.openWindow) {
                return clients.openWindow(actionUrl);
            }
        })
    );
});

