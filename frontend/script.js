// Configuration
const API_BASE_URL = "http://localhost:8080";
const CITIES = [
    "Москва", "Санкт-Петербург", "Новосибирск", "Екатеринбург", "Казань",
    "Нижний Новгород", "Челябинск", "Самара", "Уфа", "Ростов-на-Дону",
    "Краснодар", "Пермь", "Воронеж", "Волгоград", "Красноярск"
];

const TAGS = [
    "fragile", "dangerous", "heavy", "light", "urgent", "express",
    "refrigerated", "oversized", "upright", "flat", "stackable"
];

// State
let currentPage = 1;
let currentLimit = 20;
let totalResults = 0;

// DOM elements
const searchBtn = document.getElementById("search-btn");
const resultsContainer = document.getElementById("results");
const loadingElement = document.getElementById("loading");
const noResultsElement = document.getElementById("no-results");
const resultsCountElement = document.getElementById("results-count");
const pageInfoElement = document.getElementById("page-info");
const prevPageBtn = document.getElementById("prev-page");
const nextPageBtn = document.getElementById("next-page");

// Event listeners
document.addEventListener("DOMContentLoaded", function() {
    searchBtn.addEventListener("click", performSearch);
    prevPageBtn.addEventListener("click", () => changePage(-1));
    nextPageBtn.addEventListener("click", () => changePage(1));
    
    // Auto-search on Enter key
    document.addEventListener("keypress", function(e) {
        if (e.key === "Enter") {
            performSearch();
        }
    });
    
    // Initial search
    performSearch();
});

// Search function
async function performSearch() {
    showLoading();
    
    try {
        const params = buildSearchParams();
        const url = `${API_BASE_URL}/v1/orders?${params}`;
        
        const response = await fetch(url);
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        const data = await response.json();
        displayResults(data);
        
    } catch (error) {
        console.error("Search error:", error);
        showError("Ошибка при загрузке данных. Попробуйте позже.");
    }
}

// Build search parameters
function buildSearchParams() {
    const params = new URLSearchParams();
    
    // Location filters
    const fromLocation = document.getElementById("from").value.trim();
    const toLocation = document.getElementById("to").value.trim();
    
    if (fromLocation) params.append("from", fromLocation);
    if (toLocation) params.append("to", toLocation);
    
    // Weight filters
    const minWeight = document.getElementById("min-weight").value;
    const maxWeight = document.getElementById("max-weight").value;
    
    if (minWeight) params.append("min_weight", minWeight);
    if (maxWeight) params.append("max_weight", maxWeight);
    
    // Price filters
    const minPrice = document.getElementById("min-price").value;
    const maxPrice = document.getElementById("max-price").value;
    
    if (minPrice) params.append("min_price", minPrice);
    if (maxPrice) params.append("max_price", maxPrice);
    
    // Tags
    const tags = document.getElementById("tags").value.trim();
    if (tags) params.append("tags", tags);
    
    // Sorting
    const sortBy = document.getElementById("sort").value;
    params.append("sort_by", sortBy);
    params.append("sort_order", "asc");
    
    // Pagination
    params.append("page", currentPage.toString());
    params.append("limit", currentLimit.toString());
    
    return params.toString();
}

// Display results
function displayResults(data) {
    hideLoading();
    
    currentPage = data.page;
    totalResults = data.total;
    
    updateResultsCount();
    updatePagination();
    
    if (data.orders.length === 0) {
        showNoResults();
        return;
    }
    
    hideNoResults();
    
    const resultsHTML = data.orders.map(order => createOrderCard(order)).join("");
    resultsContainer.innerHTML = resultsHTML;
}

// Create order card HTML
function createOrderCard(order) {
    const tagsHTML = order.tags.map(tag => 
        `<span class="order-tag">${tag}</span>`
    ).join("");
    
    const dimensions = [];
    if (order.length_cm) dimensions.push(`Д: ${order.length_cm}см`);
    if (order.width_cm) dimensions.push(`Ш: ${order.width_cm}см`);
    if (order.height_cm) dimensions.push(`В: ${order.height_cm}см`);
    
    const customerInfo = order.customer ? `
        <div class="order-customer">
            <strong>Заказчик:</strong> ${order.customer.name}
            ${order.customer.telegram_tag ? `<br>Telegram: @${order.customer.telegram_tag}` : ""}
        </div>
    ` : "";
    
    return `
        <div class="order-card">
            <div class="order-title">${escapeHtml(order.title)}</div>
            
            <div class="order-details">
                <div class="order-detail">
                    <span>⚖️</span>
                    <strong>${order.weight_kg}</strong> кг
                </div>
                <div class="order-detail">
                    <span>💰</span>
                    <strong>${formatPrice(order.price)}</strong> ₽
                </div>
                ${order.from_location ? `
                    <div class="order-detail">
                        <span>��</span>
                        <strong>Откуда:</strong> ${escapeHtml(order.from_location)}
                    </div>
                ` : ""}
                ${order.to_location ? `
                    <div class="order-detail">
                        <span>🎯</span>
                        <strong>Куда:</strong> ${escapeHtml(order.to_location)}
                    </div>
                ` : ""}
                ${dimensions.length > 0 ? `
                    <div class="order-detail">
                        <span>📏</span>
                        <strong>Размеры:</strong> ${dimensions.join(", ")}
                    </div>
                ` : ""}
                ${order.available_from ? `
                    <div class="order-detail">
                        <span>📅</span>
                        <strong>Доступен с:</strong> ${formatDate(order.available_from)}
                    </div>
                ` : ""}
            </div>
            
            ${order.description ? `
                <div class="order-description">
                    ${escapeHtml(order.description)}
                </div>
            ` : ""}
            
            ${tagsHTML ? `
                <div class="order-tags">
                    ${tagsHTML}
                </div>
            ` : ""}
            
            <div class="order-price">
                ${formatPrice(order.price)} ₽
            </div>
            
            ${customerInfo}
        </div>
    `;
}

// Pagination functions
function changePage(delta) {
    const newPage = currentPage + delta;
    if (newPage >= 1 && newPage <= Math.ceil(totalResults / currentLimit)) {
        currentPage = newPage;
        performSearch();
    }
}

function updatePagination() {
    const totalPages = Math.ceil(totalResults / currentLimit);
    
    pageInfoElement.textContent = `Страница ${currentPage} из ${totalPages}`;
    
    prevPageBtn.disabled = currentPage <= 1;
    nextPageBtn.disabled = currentPage >= totalPages;
}

// Utility functions
function showLoading() {
    loadingElement.classList.remove("hidden");
    resultsContainer.innerHTML = "";
    hideNoResults();
}

function hideLoading() {
    loadingElement.classList.add("hidden");
}

function showNoResults() {
    noResultsElement.classList.remove("hidden");
    resultsContainer.innerHTML = "";
}

function hideNoResults() {
    noResultsElement.classList.add("hidden");
}

function updateResultsCount() {
    resultsCountElement.textContent = totalResults;
}

function showError(message) {
    hideLoading();
    resultsContainer.innerHTML = `
        <div class="error-message" style="text-align: center; color: #dc3545; padding: 20px;">
            <p>❌ ${message}</p>
        </div>
    `;
}

function formatPrice(price) {
    return new Intl.NumberFormat("ru-RU").format(price);
}

function formatDate(dateString) {
    const date = new Date(dateString);
    return date.toLocaleDateString("ru-RU");
}

function escapeHtml(text) {
    const div = document.createElement("div");
    div.textContent = text;
    return div.innerHTML;
}

// PWA support
if ("serviceWorker" in navigator) {
    window.addEventListener("load", function() {
        navigator.serviceWorker.register("/sw.js")
            .then(function(registration) {
                console.log("SW registered: ", registration);
            })
            .catch(function(registrationError) {
                console.log("SW registration failed: ", registrationError);
            });
    });
}

// Add to home screen prompt
let deferredPrompt;
window.addEventListener("beforeinstallprompt", (e) => {
    e.preventDefault();
    deferredPrompt = e;
    
    // Show install button if needed
    // You can add a custom install button here
});

// Offline support
window.addEventListener("online", function() {
    console.log("App is online");
    // Refresh data if needed
});

window.addEventListener("offline", function() {
    console.log("App is offline");
    showError("Нет подключения к интернету. Проверьте соединение.");
});
