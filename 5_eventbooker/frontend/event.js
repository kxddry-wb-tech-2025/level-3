const API_BASE_URL = '/api';

// Get event ID from URL parameters
const urlParams = new URLSearchParams(window.location.search);
const eventId = urlParams.get('id');

// Utility functions
function showMessage(text, type = 'info') {
    const messageEl = document.getElementById('message');
    messageEl.textContent = text;
    messageEl.className = `message ${type} show`;
    
    setTimeout(() => {
        messageEl.classList.remove('show');
    }, 5000);
}

function formatDateTime(dateString) {
    const date = new Date(dateString);
    return date.toLocaleString('ru-RU');
}

// Load and display event information
async function loadEventInfo() {
    if (!eventId) {
        document.getElementById('event-info').innerHTML = `
            <div class="event-info">
                <h3>Ошибка</h3>
                <p>ID мероприятия не указан в URL</p>
            </div>
        `;
        return;
    }
    
    try {
        const response = await fetch(`${API_BASE_URL}/events/${eventId}`);
        const data = await response.json();
        
        if (response.ok) {
            displayEventInfo(data);
        } else {
            document.getElementById('event-info').innerHTML = `
                <div class="event-info">
                    <h3>Ошибка</h3>
                    <p>${data.error || 'Мероприятие не найдено'}</p>
                </div>
            `;
        }
    } catch (error) {
        console.error('Error:', error);
        document.getElementById('event-info').innerHTML = `
            <div class="event-info">
                <h3>Ошибка соединения</h3>
                <p>Не удается загрузить информацию о мероприятии</p>
            </div>
        `;
    }
}

function displayEventInfo(eventData) {
    let statusClass = 'available';
    let statusText = 'Места доступны';
    
    if (eventData.available === 0) {
        statusClass = 'full';
        statusText = 'Мест нет';
    } else if (eventData.available < eventData.capacity * 0.2) {
        statusClass = 'limited';
        statusText = 'Мало мест';
    }
    
    document.getElementById('event-info').innerHTML = `
        <div class="event-info">
            <h3>${eventData.name}</h3>
            <p><strong>ID мероприятия:</strong> <span class="event-id">${eventId}</span></p>
            <p><strong>Дата и время:</strong> ${formatDateTime(eventData.date)}</p>
            <p><strong>Общая вместимость:</strong> ${eventData.capacity} мест</p>
            <p><strong>Доступно мест:</strong> ${eventData.available}</p>
            <p><strong>Забронировано:</strong> ${eventData.capacity - eventData.available}</p>
            <span class="status ${statusClass}">${statusText}</span>
        </div>
    `;
    
    // Update page title
    document.title = `${eventData.name} - Система бронирования`;
}

// Book seat
async function bookSeat() {
    const userId = document.getElementById('user-id').value.trim();
    const telegramId = document.getElementById('telegram-id').value.trim();
    
    if (!userId || !telegramId) {
        showMessage('Заполните все поля для бронирования', 'warning');
        return;
    }
    
    // Validate UUID format
    const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
    if (!uuidRegex.test(userId)) {
        showMessage('Некорректный формат User ID (должен быть UUID)', 'warning');
        return;
    }
    
    try {
        const response = await fetch(`${API_BASE_URL}/events/${eventId}/book`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                user_id: userId,
                telegram_id: parseInt(telegramId)
            })
        });
        
        const data = await response.json();
        
        if (response.ok) {
            showMessage(`Место успешно забронировано! ID брони: ${data.id}`, 'success');
            showMessage(`Срок оплаты: ${formatDateTime(data.payment_deadline)}`, 'info');
            
            // Pre-fill booking ID in confirm form
            document.getElementById('booking-id').value = data.id;
            
            // Refresh event info
            await loadEventInfo();
        } else {
            showMessage(data.error || 'Ошибка при бронировании', 'error');
        }
    } catch (error) {
        console.error('Error:', error);
        showMessage('Ошибка соединения с сервером', 'error');
    }
}

// Confirm payment
async function confirmPayment() {
    const bookingId = document.getElementById('booking-id').value.trim();
    
    if (!bookingId) {
        showMessage('Введите ID бронирования', 'warning');
        return;
    }
    
    // Validate UUID format
    const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
    if (!uuidRegex.test(bookingId)) {
        showMessage('Некорректный формат ID бронирования (должен быть UUID)', 'warning');
        return;
    }
    
    try {
        const response = await fetch(`${API_BASE_URL}/events/${eventId}/confirm`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                booking_id: bookingId
            })
        });
        
        const data = await response.json();
        
        if (response.ok) {
            showMessage(`Оплата подтверждена! Статус: ${data.status}`, 'success');
            
            // Clear booking ID field
            document.getElementById('booking-id').value = '';
            
            // Refresh event info
            await loadEventInfo();
        } else {
            showMessage(data.error || 'Ошибка при подтверждении оплаты', 'error');
        }
    } catch (error) {
        console.error('Error:', error);
        showMessage('Ошибка соединения с сервером', 'error');
    }
}

// Event listeners
document.getElementById('book-btn').addEventListener('click', bookSeat);
document.getElementById('confirm-btn').addEventListener('click', confirmPayment);
document.getElementById('refresh-btn').addEventListener('click', loadEventInfo);

// Initialize page
document.addEventListener('DOMContentLoaded', () => {
    loadEventInfo();
});