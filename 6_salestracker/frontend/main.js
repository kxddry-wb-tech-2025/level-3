// API Base URL
const API_BASE_URL = 'http://localhost:8080/api/v1';

// Global state
let allItems = [];
let filteredItems = [];
let currentPage = 1;
let itemsPerPage = 10;
let currentEditId = null;
let currentDeleteId = null;
let sortColumn = 'date';
let sortDirection = 'desc';

// DOM elements
const tabButtons = document.querySelectorAll('.tab-button');
const tabContents = document.querySelectorAll('.tab-content');
const loading = document.getElementById('loading');
const alertContainer = document.getElementById('alert-container');

// Initialize the application
document.addEventListener('DOMContentLoaded', function() {
    initializeEventListeners();
    loadDashboardData();
    setMaxDateTime();
});

// Event Listeners
function initializeEventListeners() {
    // Tab switching
    tabButtons.forEach(button => {
        button.addEventListener('click', () => switchTab(button.dataset.tab));
    });

    // Form submissions
    document.getElementById('item-form').addEventListener('submit', handleAddItem);
    document.getElementById('edit-form').addEventListener('submit', handleUpdateItem);

    // Search and filters
    document.getElementById('search-input').addEventListener('input', debounce(handleSearch, 300));
    document.getElementById('category-filter').addEventListener('change', handleCategoryFilter);

    // Table sorting
    document.querySelectorAll('th[data-sort]').forEach(th => {
        th.addEventListener('click', () => handleSort(th.dataset.sort));
    });

    // Pagination
    document.getElementById('prev-page').addEventListener('click', () => changePage(currentPage - 1));
    document.getElementById('next-page').addEventListener('click', () => changePage(currentPage + 1));

    // Export
    document.getElementById('export-csv').addEventListener('click', exportToCSV);

    // Analytics filters
    document.getElementById('apply-filter').addEventListener('click', applyAnalyticsFilter);
    document.getElementById('reset-filter').addEventListener('click', resetAnalyticsFilter);

    // Modal controls
    setupModalControls();

    // Form reset on tab switch
    document.getElementById('cancel-btn').addEventListener('click', resetForm);
}

// Tab Management
function switchTab(tabName) {
    // Update active tab button
    tabButtons.forEach(btn => btn.classList.remove('active'));
    document.querySelector(`[data-tab="${tabName}"]`).classList.add('active');

    // Update active tab content
    tabContents.forEach(content => content.classList.remove('active'));
    document.getElementById(tabName).classList.add('active');

    // Load data based on tab
    switch(tabName) {
        case 'dashboard':
            loadDashboardData();
            break;
        case 'items':
            loadAllItems();
            break;
        case 'analytics':
            loadAnalytics();
            break;
        case 'add-item':
            resetForm();
            break;
    }
}

// API Functions
async function apiRequest(endpoint, options = {}) {
    showLoading();
    
    try {
        const response = await fetch(`${API_BASE_URL}${endpoint}`, {
            headers: {
                'Content-Type': 'application/json',
                ...options.headers
            },
            ...options
        });

        if (!response.ok) {
            const errorData = await response.json().catch(() => ({}));
            throw new Error(errorData.error || `HTTP ${response.status}: ${response.statusText}`);
        }

        const data = response.status === 202 ? null : await response.json();
        return data;
    } catch (error) {
        console.error(`API Error (${endpoint}):`, error);
        showAlert(`Error: ${error.message}`, 'error');
        throw error;
    } finally {
        hideLoading();
    }
}

async function getAllItems() {
    return await apiRequest('/items');
}

async function createItem(item) {
    return await apiRequest('/items', {
        method: 'POST',
        body: JSON.stringify(item)
    });
}

async function updateItem(id, item) {
    return await apiRequest(`/items/${id}`, {
        method: 'PUT',
        body: JSON.stringify(item)
    });
}

async function deleteItem(id) {
    return await apiRequest(`/items/${id}`, {
        method: 'DELETE'
    });
}

async function getAnalytics(fromDate = null, toDate = null) {
    let endpoint = '/analytics';
    const params = new URLSearchParams();
    
    if (fromDate) params.append('from', fromDate);
    if (toDate) params.append('to', toDate);
    
    if (params.toString()) {
        endpoint += '?' + params.toString();
    }
    
    return await apiRequest(endpoint);
}

// Dashboard Functions
async function loadDashboardData() {
    try {
        const [items, analytics] = await Promise.all([
            getAllItems(),
            getAnalytics()
        ]);

        updateDashboardStats(analytics);
        displayRecentItems(items?.slice(0, 5) || []);
    } catch (error) {
        console.error('Failed to load dashboard data:', error);
    }
}

function updateDashboardStats(analytics) {
    if (!analytics) {
        analytics = { sum: 0, count: 0, average: 0, median: 0 };
    }

    document.getElementById('total-revenue').textContent = formatCurrency(analytics.sum || 0);
    document.getElementById('total-items').textContent = analytics.count || 0;
    document.getElementById('average-price').textContent = formatCurrency(analytics.average || 0);
    document.getElementById('median-price').textContent = formatCurrency(analytics.median || 0);
}

function displayRecentItems(items) {
    const container = document.getElementById('recent-items-list');
    
    if (!items || items.length === 0) {
        container.innerHTML = '<p class="text-center" style="color: #64748b; padding: 2rem;">No recent items found</p>';
        return;
    }

    container.innerHTML = items.map(item => `
        <div class="recent-item">
            <div class="recent-item-info">
                <h4>${escapeHtml(item.title)}</h4>
                <p>${escapeHtml(item.category)} • ${formatDate(item.date)}</p>
            </div>
            <div class="recent-item-price">${formatCurrency(item.price)}</div>
        </div>
    `).join('');
}

// Items Management
async function loadAllItems() {
    try {
        allItems = await getAllItems() || [];
        populateCategoryFilter();
        applyFilters();
        displayItems();
    } catch (error) {
        console.error('Failed to load items:', error);
        allItems = [];
        displayItems();
    }
}

function populateCategoryFilter() {
    const categories = [...new Set(allItems.map(item => item.category))].sort();
    const select = document.getElementById('category-filter');
    
    // Keep current selection
    const currentValue = select.value;
    
    select.innerHTML = '<option value="">All Categories</option>';
    categories.forEach(category => {
        const option = document.createElement('option');
        option.value = category;
        option.textContent = category;
        if (category === currentValue) option.selected = true;
        select.appendChild(option);
    });
}

function applyFilters() {
    const searchTerm = document.getElementById('search-input').value.toLowerCase();
    const categoryFilter = document.getElementById('category-filter').value;

    filteredItems = allItems.filter(item => {
        const matchesSearch = !searchTerm || 
            item.title.toLowerCase().includes(searchTerm) ||
            item.description.toLowerCase().includes(searchTerm) ||
            item.category.toLowerCase().includes(searchTerm);
        
        const matchesCategory = !categoryFilter || item.category === categoryFilter;
        
        return matchesSearch && matchesCategory;
    });

    // Apply sorting
    filteredItems.sort((a, b) => {
        let aVal, bVal;
        
        switch(sortColumn) {
            case 'price':
                aVal = parseFloat(a.price);
                bVal = parseFloat(b.price);
                break;
            case 'date':
                aVal = new Date(a.date);
                bVal = new Date(b.date);
                break;
            default:
                aVal = a[sortColumn]?.toString().toLowerCase() || '';
                bVal = b[sortColumn]?.toString().toLowerCase() || '';
        }
        
        if (aVal < bVal) return sortDirection === 'asc' ? -1 : 1;
        if (aVal > bVal) return sortDirection === 'asc' ? 1 : -1;
        return 0;
    });

    currentPage = 1;
}

function displayItems() {
    const tbody = document.getElementById('items-tbody');
    const startIndex = (currentPage - 1) * itemsPerPage;
    const endIndex = startIndex + itemsPerPage;
    const pageItems = filteredItems.slice(startIndex, endIndex);

    if (pageItems.length === 0) {
        tbody.innerHTML = `
            <tr>
                <td colspan="6" class="text-center" style="padding: 2rem; color: #64748b;">
                    No items found
                </td>
            </tr>
        `;
    } else {
        tbody.innerHTML = pageItems.map(item => `
            <tr>
                <td>${escapeHtml(item.title)}</td>
                <td>${formatCurrency(item.price)}</td>
                <td>${escapeHtml(item.category)}</td>
                <td>${formatDate(item.date)}</td>
                <td>${escapeHtml(item.description || '')}</td>
                <td>
                    <div class="item-actions">
                        <button class="btn btn-secondary btn-small" onclick="editItem('${item.id}')">
                            <i class="fas fa-edit"></i>
                        </button>
                        <button class="btn btn-danger btn-small" onclick="confirmDelete('${item.id}')">
                            <i class="fas fa-trash"></i>
                        </button>
                    </div>
                </td>
            </tr>
        `).join('');
    }

    updatePagination();
    updateSortIndicators();
}

function updatePagination() {
    const totalPages = Math.ceil(filteredItems.length / itemsPerPage);
    
    document.getElementById('page-info').textContent = `Page ${currentPage} of ${totalPages}`;
    document.getElementById('prev-page').disabled = currentPage <= 1;
    document.getElementById('next-page').disabled = currentPage >= totalPages;
}

function updateSortIndicators() {
    document.querySelectorAll('th[data-sort] i').forEach(i => {
        i.className = 'fas fa-sort';
    });
    
    const currentHeader = document.querySelector(`th[data-sort="${sortColumn}"] i`);
    if (currentHeader) {
        currentHeader.className = `fas fa-sort-${sortDirection === 'asc' ? 'up' : 'down'}`;
    }
}

// Event Handlers
function handleSearch() {
    applyFilters();
    displayItems();
}

function handleCategoryFilter() {
    applyFilters();
    displayItems();
}

function handleSort(column) {
    if (sortColumn === column) {
        sortDirection = sortDirection === 'asc' ? 'desc' : 'asc';
    } else {
        sortColumn = column;
        sortDirection = 'asc';
    }
    
    applyFilters();
    displayItems();
}

function changePage(page) {
    const totalPages = Math.ceil(filteredItems.length / itemsPerPage);
    if (page >= 1 && page <= totalPages) {
        currentPage = page;
        displayItems();
    }
}

async function handleAddItem(e) {
    e.preventDefault();
    
    const formData = new FormData(e.target);
    const item = {
        title: formData.get('title').trim(),
        price: parseFloat(formData.get('price')),
        category: formData.get('category').trim(),
        date: formData.get('date'),
        description: formData.get('description')?.trim() || ''
    };

    // Client-side validation
    if (!validateItem(item)) return;

    try {
        item.date = new Date(item.date).toISOString();
        await createItem(item);
        showAlert('Item created successfully!', 'success');
        resetForm();
        
        // Refresh data if on items tab
        if (document.getElementById('items').classList.contains('active')) {
            await loadAllItems();
        }
        
        // Switch to items tab to show the new item
        switchTab('items');
    } catch (error) {
        console.error('Failed to create item:', error);
    }
}

async function handleUpdateItem(e) {
    e.preventDefault();
    
    const formData = new FormData(e.target);
    const item = {
        title: formData.get('title').trim(),
        price: parseFloat(formData.get('price')),
        category: formData.get('category').trim(),
        date: formData.get('date'),
        description: formData.get('description')?.trim() || ''
    };

    if (!validateItem(item)) return;

    try {
        item.date = new Date(item.date).toISOString();
        await updateItem(currentEditId, item);
        showAlert('Item updated successfully!', 'success');
        closeModal('edit-modal');
        await loadAllItems();
    } catch (error) {
        console.error('Failed to update item:', error);
    }
}

function validateItem(item) {
    if (!item.title) {
        showAlert('Title is required', 'error');
        return false;
    }
    
    if (isNaN(item.price) || item.price < 0) {
        showAlert('Price must be a positive number', 'error');
        return false;
    }
    
    if (!item.category) {
        showAlert('Category is required', 'error');
        return false;
    }
    
    if (!item.date) {
        showAlert('Date is required', 'error');
        return false;
    }

    // Check if date is in the future
    const itemDate = new Date(item.date);
    const now = new Date();
    if (itemDate > now) {
        showAlert('Date cannot be in the future', 'error');
        return false;
    }
    
    return true;
}

function editItem(id) {
    const item = allItems.find(item => item.id === id);
    if (!item) return;

    currentEditId = id;
    
    // Populate form
    document.getElementById('edit-id').value = item.id;
    document.getElementById('edit-title').value = item.title;
    document.getElementById('edit-price').value = item.price;
    document.getElementById('edit-category').value = item.category;
    document.getElementById('edit-description').value = item.description || '';
    
    // Format date for datetime-local input
    const date = new Date(item.date);
    const formattedDate = date.toISOString().slice(0, 16);
    document.getElementById('edit-date').value = formattedDate;
    
    showModal('edit-modal');
}

function confirmDelete(id) {
    currentDeleteId = id;
    showModal('delete-modal');
}

async function executeDelete() {
    if (!currentDeleteId) return;

    try {
        await deleteItem(currentDeleteId);
        showAlert('Item deleted successfully!', 'success');
        closeModal('delete-modal');
        await loadAllItems();
        currentDeleteId = null;
    } catch (error) {
        console.error('Failed to delete item:', error);
    }
}

// Analytics Functions
async function loadAnalytics() {
    try {
        const analytics = await getAnalytics();
        displayAnalytics(analytics);
    } catch (error) {
        console.error('Failed to load analytics:', error);
        displayAnalytics(null);
    }
}

async function applyAnalyticsFilter() {
    const fromDate = document.getElementById('from-date').value;
    const toDate = document.getElementById('to-date').value;

    try {
        const analytics = await getAnalytics(fromDate || null, toDate || null);
        displayAnalytics(analytics);
    } catch (error) {
        console.error('Failed to apply analytics filter:', error);
    }
}

function resetAnalyticsFilter() {
    document.getElementById('from-date').value = '';
    document.getElementById('to-date').value = '';
    loadAnalytics();
}

function displayAnalytics(analytics) {
    if (!analytics) {
        analytics = { sum: 0, count: 0, average: 0, median: 0, percentile_90: 0 };
    }

    document.getElementById('analytics-sum').textContent = formatCurrency(analytics.sum || 0);
    document.getElementById('analytics-count').textContent = analytics.count || 0;
    document.getElementById('analytics-average').textContent = formatCurrency(analytics.average || 0);
    document.getElementById('analytics-median').textContent = formatCurrency(analytics.median || 0);
    document.getElementById('analytics-percentile').textContent = formatCurrency(analytics.percentile_90 || 0);
}

// Export Functions
function exportToCSV() {
    if (!filteredItems || filteredItems.length === 0) {
        showAlert('No data to export', 'warning');
        return;
    }

    const headers = ['Title', 'Price', 'Category', 'Date', 'Description'];
    const csvContent = [
        headers.join(','),
        ...filteredItems.map(item => [
            `"${escapeCSV(item.title)}"`,
            item.price,
            `"${escapeCSV(item.category)}"`,
            `"${formatDate(item.date)}"`,
            `"${escapeCSV(item.description || '')}"`
        ].join(','))
    ].join('\n');

    const blob = new Blob([csvContent], { type: 'text/csv' });
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = `salestracker-export-${new Date().toISOString().split('T')[0]}.csv`;
    link.click();
    window.URL.revokeObjectURL(url);
    
    showAlert('Data exported successfully!', 'success');
}

// Modal Functions
function setupModalControls() {
    // Edit modal
    document.getElementById('close-edit-modal').addEventListener('click', () => closeModal('edit-modal'));
    document.getElementById('cancel-edit').addEventListener('click', () => closeModal('edit-modal'));
    
    // Delete modal
    document.getElementById('close-delete-modal').addEventListener('click', () => closeModal('delete-modal'));
    document.getElementById('cancel-delete').addEventListener('click', () => closeModal('delete-modal'));
    document.getElementById('confirm-delete').addEventListener('click', executeDelete);

    // Close modal when clicking outside
    document.getElementById('edit-modal').addEventListener('click', (e) => {
        if (e.target === e.currentTarget) closeModal('edit-modal');
    });
    
    document.getElementById('delete-modal').addEventListener('click', (e) => {
        if (e.target === e.currentTarget) closeModal('delete-modal');
    });
}

function showModal(modalId) {
    const modal = document.getElementById(modalId);
    modal.classList.add('active');
    document.body.style.overflow = 'hidden';
}

function closeModal(modalId) {
    const modal = document.getElementById(modalId);
    modal.classList.remove('active');
    document.body.style.overflow = '';
    
    if (modalId === 'edit-modal') {
        currentEditId = null;
    }
    if (modalId === 'delete-modal') {
        currentDeleteId = null;
    }
}

// Form Functions
function resetForm() {
    const form = document.getElementById('item-form');
    form.reset();
    
    // Reset form state
    document.getElementById('form-title').textContent = 'Add New Item';
    document.getElementById('submit-btn').innerHTML = '<i class="fas fa-plus"></i> Add Item';
    document.getElementById('cancel-btn').style.display = 'none';
    
    // Set max datetime to now
    setMaxDateTime();
}

function setMaxDateTime() {
    const now = new Date();
    const maxDateTime = now.toISOString().slice(0, 16);
    document.getElementById('date').setAttribute('max', maxDateTime);
    document.getElementById('edit-date').setAttribute('max', maxDateTime);
}

// Utility Functions
function showLoading() {
    loading.style.display = 'flex';
}

function hideLoading() {
    loading.style.display = 'none';
}

function showAlert(message, type = 'info') {
    const alert = document.createElement('div');
    alert.className = `alert ${type}`;
    alert.innerHTML = `
        ${message}
        <button class="alert-close" onclick="this.parentElement.remove()">
            <i class="fas fa-times"></i>
        </button>
    `;
    
    alertContainer.appendChild(alert);
    
    // Auto-remove after 5 seconds
    setTimeout(() => {
        if (alert.parentElement) {
            alert.remove();
        }
    }, 5000);
}

function formatCurrency(amount) {
    return new Intl.NumberFormat('en-US', {
        style: 'currency',
        currency: 'USD'
    }).format(amount);
}

function formatDate(dateString) {
    const date = new Date(dateString);
    return date.toLocaleString('en-US', {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
    });
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function escapeCSV(text) {
    return text.replace(/"/g, '""');
}

function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func.apply(this, args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

// Global functions for onclick handlers
window.editItem = editItem;
window.confirmDelete = confirmDelete;
