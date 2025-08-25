// Warehouse Control System - Frontend Application

// Global state
let currentUser = null;
let currentToken = null;
let items = [];
let history = [];

// API configuration
const API_BASE = '/api/v1';

// Role definitions
const ROLES = {
    1: { name: 'Пользователь', class: 'user', permissions: ['read'] },
    2: { name: 'Менеджер', class: 'manager', permissions: ['read', 'write'] },
    3: { name: 'Администратор', class: 'admin', permissions: ['read', 'write', 'delete'] }
};

// Initialize application
document.addEventListener('DOMContentLoaded', function() {
    initializeApp();
});

function initializeApp() {
    // Set up event listeners
    setupEventListeners();
    
    // Check if user is already logged in
    const savedToken = localStorage.getItem('warehouse_token');
    const savedRole = localStorage.getItem('warehouse_role');
    
    if (savedToken && savedRole) {
        currentToken = savedToken;
        currentUser = { role: parseInt(savedRole) };
        updateUI();
        loadItems();
    }
}

function setupEventListeners() {
    // Form submission
    document.getElementById('itemForm').addEventListener('submit', function(e) {
        e.preventDefault();
        saveItem();
    });
    
    // Date inputs for history
    document.getElementById('dateFrom').addEventListener('change', function() {
        if (this.value && document.getElementById('dateTo').value) {
            loadHistory();
        }
    });
    
    document.getElementById('dateTo').addEventListener('change', function() {
        if (this.value && document.getElementById('dateFrom').value) {
            loadHistory();
        }
    });
    
    // Action filter
    document.getElementById('actionFilter').addEventListener('change', loadHistory);
}

// Authentication functions
async function loginAs(role) {
    try {
        showLoading();
        
        const response = await fetch(`${API_BASE}/meta/jwt/${role}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            }
        });
        
        if (!response.ok) {
            throw new Error('Failed to get JWT token');
        }
        
        const data = await response.json();
        currentToken = data.jwt;
        currentUser = { role: role };
        
        // Save to localStorage
        localStorage.setItem('warehouse_token', currentToken);
        localStorage.setItem('warehouse_role', role);
        
        updateUI();
        loadItems();
        loadHistory();
        
        showToast('Успешно', `Вошли как ${ROLES[role].name}`, 'success');
        
    } catch (error) {
        console.error('Login error:', error);
        showToast('Ошибка', 'Не удалось войти в систему', 'error');
    } finally {
        hideLoading();
    }
}

function logout() {
    currentUser = null;
    currentToken = null;
    items = [];
    history = [];
    
    localStorage.removeItem('warehouse_token');
    localStorage.removeItem('warehouse_role');
    
    updateUI();
    showToast('Информация', 'Вы вышли из системы', 'info');
}

// Items management
async function loadItems() {
    if (!currentToken) return;
    
    try {
        showLoading();
        
        const response = await fetch(`${API_BASE}/items`, {
            headers: {
                'Authorization': `Bearer ${currentToken}`
            }
        });
        
        if (!response.ok) {
            throw new Error('Failed to load items');
        }
        
        const data = await response.json();
        items = Array.isArray(data) ? data : data.items || [];
        renderItems();
        
    } catch (error) {
        console.error('Load items error:', error);
        showToast('Ошибка', 'Не удалось загрузить товары', 'error');
    } finally {
        hideLoading();
    }
}

function renderItems() {
    const tbody = document.getElementById('itemsTableBody');
    
    if (!items || items.length === 0) {
        tbody.innerHTML = `
            <tr>
                <td colspan="6" class="text-center text-muted">
                    <div class="empty-state">
                        <i class="bi bi-box"></i>
                        <p>Товары не найдены</p>
                    </div>
                </td>
            </tr>
        `;
        return;
    }
    
    tbody.innerHTML = items.map(item => `
        <tr>
            <td><code>${item.id}</code></td>
            <td><strong>${escapeHtml(item.name)}</strong></td>
            <td>${escapeHtml(item.description || '')}</td>
            <td>
                <span class="badge bg-${item.quantity > 0 ? 'success' : 'danger'}">
                    ${item.quantity}
                </span>
            </td>
            <td>${formatPrice(item.price)} ₽</td>
            <td class="item-actions">
                ${hasPermission('write') ? `
                    <button class="btn btn-sm btn-outline-primary" onclick="editItem('${item.id}')" title="Редактировать">
                        <i class="bi bi-pencil"></i>
                    </button>
                ` : ''}
                ${hasPermission('delete') ? `
                    <button class="btn btn-sm btn-outline-danger" onclick="deleteItem('${item.id}')" title="Удалить">
                        <i class="bi bi-trash"></i>
                    </button>
                ` : ''}
                <button class="btn btn-sm btn-outline-info" onclick="viewItemHistory('${item.id}')" title="История">
                    <i class="bi bi-clock-history"></i>
                </button>
            </td>
        </tr>
    `).join('');
}

function showAddItemModal() {
    document.getElementById('itemModalTitle').textContent = 'Добавить товар';
    document.getElementById('itemForm').reset();
    document.getElementById('itemId').value = '';
    
    const modal = new bootstrap.Modal(document.getElementById('itemModal'));
    modal.show();
}

function editItem(id) {
    const item = items.find(i => i.id === id);
    if (!item) return;
    
    document.getElementById('itemModalTitle').textContent = 'Редактировать товар';
    document.getElementById('itemId').value = item.id;
    document.getElementById('itemName').value = item.name;
    document.getElementById('itemDescription').value = item.description || '';
    document.getElementById('itemQuantity').value = item.quantity;
    document.getElementById('itemPrice').value = item.price;
    
    const modal = new bootstrap.Modal(document.getElementById('itemModal'));
    modal.show();
}

async function saveItem() {
    const id = document.getElementById('itemId').value;
    const name = document.getElementById('itemName').value.trim();
    const description = document.getElementById('itemDescription').value.trim();
    const quantity = parseInt(document.getElementById('itemQuantity').value);
    const price = parseFloat(document.getElementById('itemPrice').value);
    
    if (!name || isNaN(quantity) || isNaN(price)) {
        showToast('Ошибка', 'Пожалуйста, заполните все обязательные поля', 'error');
        return;
    }
    
    try {
        showLoading();
        
        const itemData = {
            name: name,
            description: description,
            quantity: quantity,
            price: price
        };
        
        if (id) {
            // Update existing item
            itemData.id = id;
            
            const response = await fetch(`${API_BASE}/items/${id}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${currentToken}`
                },
                body: JSON.stringify(itemData)
            });
            
            if (!response.ok) {
                throw new Error('Failed to update item');
            }
            
            showToast('Успешно', 'Товар обновлен', 'success');
        } else {
            // Create new item
            const response = await fetch(`${API_BASE}/items`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${currentToken}`
                },
                body: JSON.stringify(itemData)
            });
            
            if (!response.ok) {
                throw new Error('Failed to create item');
            }
            
            showToast('Успешно', 'Товар создан', 'success');
        }
        
        // Close modal and reload items
        const modal = bootstrap.Modal.getInstance(document.getElementById('itemModal'));
        modal.hide();
        
        loadItems();
        loadHistory();
        
    } catch (error) {
        console.error('Save item error:', error);
        showToast('Ошибка', 'Не удалось сохранить товар', 'error');
    } finally {
        hideLoading();
    }
}

async function deleteItem(id) {
    if (!confirm('Вы уверены, что хотите удалить этот товар?')) {
        return;
    }
    
    try {
        showLoading();
        
        const response = await fetch(`${API_BASE}/items/${id}`, {
            method: 'DELETE',
            headers: {
                'Authorization': `Bearer ${currentToken}`
            }
        });
        
        if (!response.ok) {
            throw new Error('Failed to delete item');
        }
        
        showToast('Успешно', 'Товар удален', 'success');
        loadItems();
        loadHistory();
        
    } catch (error) {
        console.error('Delete item error:', error);
        showToast('Ошибка', 'Не удалось удалить товар', 'error');
    } finally {
        hideLoading();
    }
}

// History management
async function loadHistory() {
    if (!currentToken) return;
    
    try {
        showLoading();
        
        const params = new URLSearchParams();
        
        const dateFrom = document.getElementById('dateFrom').value;
        const dateTo = document.getElementById('dateTo').value;
        const action = document.getElementById('actionFilter').value;
        const limit = document.getElementById('limit').value;
        const offset = document.getElementById('offset').value;
        
        if (dateFrom) params.append('date_from', dateFrom + 'T00:00:00Z');
        if (dateTo) params.append('date_to', dateTo + 'T23:59:59Z');
        if (action) params.append('action', action);
        if (limit) params.append('limit', limit);
        if (offset) params.append('offset', offset);
        
        const response = await fetch(`${API_BASE}/meta/history?${params.toString()}`, {
            headers: {
                'Authorization': `Bearer ${currentToken}`
            }
        });
        
        if (!response.ok) {
            throw new Error('Failed to load history');
        }
        
        const data = await response.json();
        history = Array.isArray(data) ? data : data.history || [];
        renderHistory();
        
        // Show export button if there's history
        document.getElementById('exportBtn').style.display = history.length > 0 ? 'block' : 'none';
        
    } catch (error) {
        console.error('Load history error:', error);
        showToast('Ошибка', 'Не удалось загрузить историю', 'error');
    } finally {
        hideLoading();
    }
}

function renderHistory() {
    const container = document.getElementById('historyContainer');
    
    if (!history || history.length === 0) {
        container.innerHTML = `
            <div class="text-center text-muted">
                <i class="bi bi-clock"></i>
                <p>История не найдена</p>
            </div>
        `;
        return;
    }
    
    container.innerHTML = history.map(entry => `
        <div class="history-item ${entry.action}" onclick="viewHistoryDetails('${entry.id}')">
            <div class="d-flex justify-content-between align-items-start">
                <div>
                    <div class="history-action ${entry.action}">
                        ${getActionLabel(entry.action)}
                    </div>
                    <div class="history-user">
                        Пользователь: ${entry.user_id}
                    </div>
                    <div class="history-time">
                        ${formatDateTime(entry.changed_at)}
                    </div>
                </div>
                <div class="text-end">
                    <small class="text-muted">ID: ${entry.item_id}</small>
                </div>
            </div>
        </div>
    `).join('');
}

function viewItemHistory(itemId) {
    // Filter history for specific item
    const itemHistory = history.filter(h => h.item_id === itemId);
    
    if (itemHistory.length === 0) {
        showToast('Информация', 'История для этого товара не найдена', 'info');
        return;
    }
    
    const detailsHtml = `
        <h6>История изменений товара</h6>
        <div class="table-responsive">
            <table class="table table-sm">
                <thead>
                    <tr>
                        <th>Действие</th>
                        <th>Пользователь</th>
                        <th>Дата</th>
                    </tr>
                </thead>
                <tbody>
                    ${itemHistory.map(entry => `
                        <tr>
                            <td>
                                <span class="history-action ${entry.action}">
                                    ${getActionLabel(entry.action)}
                                </span>
                            </td>
                            <td>${entry.user_id}</td>
                            <td>${formatDateTime(entry.changed_at)}</td>
                        </tr>
                    `).join('')}
                </tbody>
            </table>
        </div>
    `;
    
    document.getElementById('historyDetails').innerHTML = detailsHtml;
    const modal = new bootstrap.Modal(document.getElementById('historyModal'));
    modal.show();
}

function viewHistoryDetails(entryId) {
    const entry = history.find(h => h.id === entryId);
    if (!entry) return;
    
    const detailsHtml = `
        <div class="row">
            <div class="col-md-6">
                <h6>Детали записи</h6>
                <table class="table table-sm">
                    <tr><td><strong>ID записи:</strong></td><td>${entry.id}</td></tr>
                    <tr><td><strong>Действие:</strong></td><td>
                        <span class="history-action ${entry.action}">
                            ${getActionLabel(entry.action)}
                        </span>
                    </td></tr>
                    <tr><td><strong>ID товара:</strong></td><td>${entry.item_id}</td></tr>
                    <tr><td><strong>Пользователь:</strong></td><td>${entry.user_id}</td></tr>
                    <tr><td><strong>Дата:</strong></td><td>${formatDateTime(entry.changed_at)}</td></tr>
                </table>
            </div>
            <div class="col-md-6">
                <h6>Информация о товаре</h6>
                ${getItemInfo(entry.item_id)}
            </div>
        </div>
    `;
    
    document.getElementById('historyDetails').innerHTML = detailsHtml;
    const modal = new bootstrap.Modal(document.getElementById('historyModal'));
    modal.show();
}

function exportHistory() {
    if (!history || history.length === 0) {
        showToast('Информация', 'Нет данных для экспорта', 'info');
        return;
    }
    
    const csvContent = [
        ['Действие', 'ID товара', 'Пользователь', 'Дата'],
        ...history.map(entry => [
            getActionLabel(entry.action),
            entry.item_id,
            entry.user_id,
            formatDateTime(entry.changed_at)
        ])
    ].map(row => row.map(cell => `"${cell}"`).join(',')).join('\n');
    
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const link = document.createElement('a');
    const url = URL.createObjectURL(blob);
    
    link.setAttribute('href', url);
    link.setAttribute('download', `warehouse_history_${new Date().toISOString().split('T')[0]}.csv`);
    link.style.visibility = 'hidden';
    
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    
    showToast('Успешно', 'История экспортирована в CSV', 'success');
}

// Utility functions
function updateUI() {
    const roleSpan = document.getElementById('currentRole');
    const addItemBtn = document.getElementById('addItemBtn');
    
    if (currentUser) {
        const role = ROLES[currentUser.role];
        roleSpan.innerHTML = `
            <span class="role-badge ${role.class}">${role.name}</span>
        `;
        
        addItemBtn.style.display = hasPermission('write') ? 'block' : 'none';
    } else {
        roleSpan.textContent = 'Выберите роль';
        addItemBtn.style.display = 'none';
    }
}

function hasPermission(permission) {
    if (!currentUser) return false;
    return ROLES[currentUser.role].permissions.includes(permission);
}

function getActionLabel(action) {
    const labels = {
        'create': 'Создание',
        'update': 'Обновление',
        'delete': 'Удаление'
    };
    return labels[action] || action;
}

function formatDateTime(dateString) {
    const date = new Date(dateString);
    return date.toLocaleString('ru-RU');
}

function formatPrice(price) {
    return new Intl.NumberFormat('ru-RU', {
        minimumFractionDigits: 2,
        maximumFractionDigits: 2
    }).format(price);
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function getItemInfo(itemId) {
    const item = items.find(i => i.id === itemId);
    if (!item) {
        return '<p class="text-muted">Товар не найден</p>';
    }
    
    return `
        <table class="table table-sm">
            <tr><td><strong>Название:</strong></td><td>${escapeHtml(item.name)}</td></tr>
            <tr><td><strong>Описание:</strong></td><td>${escapeHtml(item.description || '')}</td></tr>
            <tr><td><strong>Количество:</strong></td><td>${item.quantity}</td></tr>
            <tr><td><strong>Цена:</strong></td><td>${formatPrice(item.price)} ₽</td></tr>
        </table>
    `;
}

function showToast(title, message, type = 'info') {
    const toast = document.getElementById('toast');
    const toastTitle = document.getElementById('toastTitle');
    const toastBody = document.getElementById('toastBody');
    
    toastTitle.textContent = title;
    toastBody.textContent = message;
    
    // Remove existing classes
    toast.classList.remove('bg-success', 'bg-danger', 'bg-warning', 'bg-info');
    
    // Add appropriate class
    switch (type) {
        case 'success':
            toast.classList.add('bg-success', 'text-white');
            break;
        case 'error':
            toast.classList.add('bg-danger', 'text-white');
            break;
        case 'warning':
            toast.classList.add('bg-warning');
            break;
        default:
            toast.classList.add('bg-info', 'text-white');
    }
    
    const bsToast = new bootstrap.Toast(toast);
    bsToast.show();
}

function showLoading() {
    // You can add a loading indicator here if needed
}

function hideLoading() {
    // Hide loading indicator
}
