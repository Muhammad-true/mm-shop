// ===== STORAGE.JS - Управление локальным хранилищем =====

// Глобальные переменные состояния
let API_BASE_URL = CONFIG.API.BASE_URL;
let currentProductId = null;
let currentCategoryId = null;
let adminToken = null;
let userRole = null;

// Глобальные переменные для загрузки изображений
let uploadedImages = [];
let imageUrls = [];

// Глобальные переменные для вариаций
let variations = [];

// Глобальная переменная для хранения всех товаров
let allProducts = [];

// Геттеры и сеттеры
function getAdminToken() {
    return adminToken;
}

function setAdminToken(token) {
    adminToken = token;
    if (token) {
        localStorage.setItem('adminToken', token);
    } else {
        localStorage.removeItem('adminToken');
    }
}

function getUserRole() {
    return userRole;
}

function setUserRole(role) {
    userRole = role;
    if (role) {
        localStorage.setItem('userRole', role);
    } else {
        localStorage.removeItem('userRole');
    }
}

function updateLastActivity() {
    localStorage.setItem('lastActivity', Date.now().toString());
    console.log('🕐 Время активности обновлено');
}

function clearAllStorage() {
    adminToken = null;
    userRole = null;
    currentProductId = null;
    currentCategoryId = null;
    uploadedImages = [];
    imageUrls = [];
    variations = [];
    allProducts = [];
    
    localStorage.removeItem('adminToken');
    localStorage.removeItem('userRole');
    localStorage.removeItem('lastActivity');
    localStorage.removeItem('userData');
}

// Геттеры и сеттеры для других данных
function getAllProducts() {
    return allProducts;
}

function setAllProducts(products) {
    allProducts = products;
}

function getVariations() {
    return variations;
}

function setVariations(vars) {
    variations = vars;
}

function addVariation(variation) {
    variations.push(variation);
}

function removeVariation(index) {
    variations.splice(index, 1);
}

function getCurrentProductId() {
    return currentProductId;
}

function setCurrentProductId(id) {
    currentProductId = id;
}

function getCurrentCategoryId() {
    return currentCategoryId;
}

function setCurrentCategoryId(id) {
    currentCategoryId = id;
}

// Возвращаем глобальный доступ
window.storage = {
    getAdminToken,
    setAdminToken,
    getUserRole,
    setUserRole,
    updateLastActivity,
    clearAllStorage,
    getAllProducts,
    setAllProducts,
    getVariations,
    setVariations,
    addVariation,
    removeVariation,
    getCurrentProductId,
    setCurrentProductId,
    getCurrentCategoryId,
    setCurrentCategoryId
};

// Для обратной совместимости
window.getAdminToken = getAdminToken;
window.setAdminToken = setAdminToken;

