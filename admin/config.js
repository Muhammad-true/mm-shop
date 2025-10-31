// ===== КОНФИГУРАЦИЯ АДМИН ПАНЕЛИ =====

const CONFIG = {
  // === API НАСТРОЙКИ ===
  API: {
    // Базовый URL для разработки
    DEV_BASE_URL: '',  // Пусто для same-origin через Nginx

    // Базовый URL для продакшена — пусто, чтобы ходить на тот же хост (same-origin)
    PROD_BASE_URL: '',

    // Автоматический выбор URL
    get BASE_URL() {
      // Везде используем относительные пути, чтобы ходить через Nginx
      return this.PROD_BASE_URL; // Same-origin через прокси
    },

    // Эндпоинты API
    ENDPOINTS: {
      AUTH: {
        LOGIN: '/api/v1/auth/login',
        PROFILE: '/api/v1/users/profile',
      },
      PRODUCTS: {
        LIST: '/api/v1/products/',
        SHOP_LIST: '/api/v1/shop/products/',
        CREATE: '/api/v1/shop/products/',
        UPDATE: (id) => `/api/v1/shop/products/${id}`,
        DELETE: (id) => `/api/v1/shop/products/${id}`,
        GET:    (id) => `/api/v1/products/${id}`,
      },
      CATEGORIES: {
        LIST: '/api/v1/categories/',
        CREATE: '/api/v1/admin/categories/',
        UPDATE: (id) => `/api/v1/admin/categories/${id}`,
        DELETE: (id) => `/api/v1/admin/categories/${id}`,
        GET:    (id) => `/api/v1/categories/${id}`,
      },
      USERS: {
        LIST: '/api/v1/admin/users/',
        CREATE: '/api/v1/admin/users/',
        UPDATE: (id) => `/api/v1/admin/users/${id}`,
        DELETE: (id) => `/api/v1/admin/users/${id}`,
        GET:    (id) => `/api/v1/admin/users/${id}`,
      },
      ROLES: {
        LIST: '/api/v1/admin/roles/',
        CREATE: '/api/v1/admin/roles/',
        UPDATE: (id) => `/api/v1/admin/roles/${id}`,
        DELETE: (id) => `/api/v1/admin/roles/${id}`,
        GET:    (id) => `/api/v1/admin/roles/${id}`,
      },
      ORDERS: {
        LIST: '/api/v1/admin/orders/', // со слешем для совместимости с Go Gin
        SHOP_LIST: '/api/v1/shop/orders/',
      },
      UPLOAD: {
        IMAGE: '/api/v1/upload/image',
        DELETE_IMAGE: (filename) => `/api/v1/upload/image/${filename}`,
      },
      HEALTH: '/api/health',
    },
  },

  // === НАСТРОЙКИ ИЗОБРАЖЕНИЙ ===
  IMAGES: {
    FOLDERS: {
      PRODUCTS: 'products',
      VARIATIONS: 'variations',
      CATEGORIES: 'categories',
      USERS: 'users',
    },
    MAX_FILE_SIZE: 10 * 1024 * 1024, // 10MB
    ALLOWED_TYPES: ['image/jpeg', 'image/jpg', 'image/png', 'image/gif', 'image/webp'],
    IP_REPLACEMENTS: { // оставим на всякий случай; не используется в новой версии getImageUrl
      '0.0.0.0': 'localhost',
      '127.0.0.1': 'localhost',
    },
  },

  // === НАСТРОЙКИ СЕССИИ ===
  SESSION: {
    TOKEN_LIFETIME: 24 * 60 * 60 * 1000, // 24 часа
    STORAGE_KEYS: {
      TOKEN: 'adminToken',
      USER_ROLE: 'userRole',
      USER_DATA: 'userData',
      LAST_ACTIVITY: 'lastActivity',
      API_URL: 'api_url',
    },
  },

  // === НАСТРОЙКИ UI ===
  UI: {
    ANIMATION_DELAYS: { TAB_SWITCH: 150, DATA_LOAD: 100, PRODUCTS_LOAD: 200 },
    MOBILE: { MIN_BUTTON_SIZE: 44, FONT_SIZE: 16 },
  },

  // === НАСТРОЙКИ ЛОГИРОВАНИЯ ===
  LOGGING: {
    ENABLED: true,
    LEVELS: { ERROR:'❌', WARNING:'⚠️', INFO:'ℹ️', SUCCESS:'✅', DEBUG:'🔍' },
  },
};

// ===== УТИЛИТЫ =====

// Логирование
function log(level, message, data = null) {
  if (!CONFIG.LOGGING.ENABLED) return;
  const prefix = CONFIG.LOGGING.LEVELS[level] || 'ℹ️';
  const timestamp = new Date().toLocaleTimeString();
  data ? console.log(`${prefix} [${timestamp}] ${message}`, data)
       : console.log(`${prefix} [${timestamp}] ${message}`);
}

// Картинка: ВСЕГДА same-origin (относительный путь)
window.getImageUrl = function(url) {
  if (!url) return '';

  // Если абсолютный URL с /images/ — оставим только путь
  const m = String(url).match(/^https?:\/\/[^/]+(?::\d+)?(\/images\/.+)$/i);
  if (m) return m[1];

  // Если относительный путь — добавляем / если нужно
  if (!/^https?:\/\//i.test(url)) {
    return url.startsWith('/') ? url : '/' + url;
  }

  // Срежем «плохие» хосты у /images/, включая порт 3000
  let imageUrl = url;
  ['http://0.0.0.0:8080','http://127.0.0.1:8080','http://localhost:8080','http://localhost','https://localhost',
   'http://159.89.99.252:3000','http://159.89.99.252:8080']
    .forEach((h) => { if (imageUrl.startsWith(h + '/images/')) imageUrl = imageUrl.replace(h, ''); });

  return imageUrl;
};

// URL для API
window.getApiUrl = function(endpoint) {
  return (CONFIG.API.BASE_URL || '') + endpoint;
};

// Экспорт
window.CONFIG = CONFIG;
window.log = log;

// Отладочный лог
log('INFO', 'Конфигурация загружена', {
  baseUrl: CONFIG.API.BASE_URL,
  environment: (location.hostname === 'localhost' || location.hostname === '127.0.0.1') ? 'development' : 'production',
});
