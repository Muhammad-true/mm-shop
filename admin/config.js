// ===== КОНФИГУРАЦИЯ АДМИН ПАНЕЛИ =====

const CONFIG = {
  // === API НАСТРОЙКИ ===
  API: {
    // Базовый URL для разработки
    DEV_BASE_URL: 'http://localhost:8080',

    // Базовый URL для продакшена — пусто, чтобы ходить на тот же хост (same-origin)
    PROD_BASE_URL: '',

    // Автоматический выбор URL
    get BASE_URL() {
      return (location.hostname === 'localhost' || location.hostname === '127.0.0.1')
        ? this.DEV_BASE_URL
        : this.PROD_BASE_URL; // в проде даст относительные /api/... /images/...
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
        LIST: '/api/v1/admin/orders', // если у бэка нужен конечный слэш — поставь '/'
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

// Полный URL для API
function getApiUrl(endpoint) {
  return CONFIG.API.BASE_URL + endpoint;
}

// Картинка: ВСЕГДА same-origin
function getImageUrl(url) {
  if (!url) return '';

  // 1) Если API вернул полный URL к /images/ — делаем относительным
  const m = url.match(/^https?:\/\/[^/]+(\/images\/.+)$/i);
  if (m) return m[1]; // -> "/images/variations/.."

  // 2) Иначе — как раньше: добавим BASE_URL при относительном пути
  if (!url.startsWith('http')) {
    return (url.startsWith('/') ? CONFIG.API.BASE_URL + url
                                : CONFIG.API.BASE_URL + '/' + url);
  }

  // 3) На всякий случай: замены «плохих» хостов
  let imageUrl = url;
  const REPL = {
    'http://0.0.0.0:8080': '',
    'http://127.0.0.1:8080': '',
    'http://localhost:8080': '',
    'http://localhost': '',
  };
  Object.entries(REPL).forEach(([from, to]) => {
    if (imageUrl.startsWith(from + '/images/')) imageUrl = imageUrl.replace(from, '');
  });

  return imageUrl;
}
// Логирование
function log(level, message, data = null) {
  if (!CONFIG.LOGGING.ENABLED) return;
  const prefix = CONFIG.LOGGING.LEVELS[level] || 'ℹ️';
  const timestamp = new Date().toLocaleTimeString();
  data ? console.log(`${prefix} [${timestamp}] ${message}`, data)
       : console.log(`${prefix} [${timestamp}] ${message}`);
}

// Экспорт
window.CONFIG = CONFIG;
window.getApiUrl = getApiUrl;
window.getImageUrl = getImageUrl;
window.log = log;

// Отладочный лог
log('INFO', 'Конфигурация загружена', {
  baseUrl: CONFIG.API.BASE_URL,
  environment: (location.hostname === 'localhost' || location.hostname === '127.0.0.1') ? 'development' : 'production',
});


/* ==== RUNTIME OVERRIDES (prod safe) ==== */

/** Формирует URL для API.
 *  В деве:  http://localhost:8080 + endpoint
 *  В проде: '' + endpoint -> получится относительный /api/...
 */
window.getApiUrl = function(endpoint) {
  return (CONFIG.API.BASE_URL || '') + endpoint;
};

/** Нормализует расширение файла к нижнему регистру (JPG -> jpg) */
function normalizeExtLower(p) {
  return String(p).replace(/(\.[a-zA-Z0-9]+)$/, s => s.toLowerCase());
}

/** Делает корректный URL картинки (предпочтительно относительный /images/...) */
window.getImageUrl = function(url) {
  if (!url) return '';

  // Если прилетел абсолютный URL с /images/ — оставим только путь
  const m = String(url).match(/^https?:\/\/[^/]+(\/images\/.+)$/i);
  if (m) return normalizeExtLower(m[1]); // "/images/variations/...."

  // Если относительный путь — добавим BASE_URL (в деве это http://localhost:8080)
  if (!/^https?:\/\//i.test(url)) {
    const rel = url.startsWith('/') ? url : '/' + url;
    return (CONFIG.API.BASE_URL || '') + normalizeExtLower(rel);
  }

  // На всякий случай: срежем "плохие" хосты у /images/
  let imageUrl = url;
  ['http://0.0.0.0:8080','http://127.0.0.1:8080','http://localhost:8080','http://localhost']
    .forEach((h) => { if (imageUrl.startsWith(h + '/images/')) imageUrl = imageUrl.replace(h, ''); });

  // Если всё равно абсолютный — вернём как есть, но с нижним регистром расширения
  return normalizeExtLower(imageUrl);
};

/* ==== Override: сохраняем регистр расширения (важно для Linux FS) ==== */
window.getImageUrl = function(url) {
  if (!url) return '';

  // Если абсолютный URL на /images — вернём только путь (оставляя регистр)
  const m = String(url).match(/^https?:\/\/[^/]+(\/images\/.+)$/i);
  if (m) return m[1];

  // Если относительный — добавим BASE_URL только в dev (в prod BASE_URL пустой)
  if (!/^https?:\/\//i.test(url)) {
    const rel = url.startsWith('/') ? url : '/' + url;
    return (CONFIG.API.BASE_URL || '') + rel;
  }

  // Срежем «плохие» хосты у /images/, если вдруг попались
  let imageUrl = url;
  ['http://0.0.0.0:8080','http://127.0.0.1:8080','http://localhost:8080','http://localhost']
    .forEach((h) => { if (imageUrl.startsWith(h + '/images/')) imageUrl = imageUrl.replace(h, ''); });

  return imageUrl;
};
