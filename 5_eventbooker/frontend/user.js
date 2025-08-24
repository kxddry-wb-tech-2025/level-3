const API_BASE_URL = '/api';

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

function generateUUID() {
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
        const r = Math.random() * 16 | 0;
        const v = c == 'x' ? r : (r & 0x3 | 0x8);
        return v.toString(16);
    });
}

// User info management
function getUserInfo() {
    const userInfo = localStorage.getItem('userInfo');
    return userInfo ? JSON.parse(userInfo) : null;
}

function saveUserInfo(userInfo) {
    localStorage.setItem('userInfo', JSON.stringify(userInfo));
}

function displayUserInfo() {
    const userInfo = getUserInfo();
    if (userInfo) {
        document.getElementById('display-user-id').textContent = userInfo.userId;
        document.getElementById('display-telegram-id').textContent = userInfo.telegramId;
        document.getElementById('user-info-display').classList.remove('hidden');
        document.querySelector('.user-info-section .form').classList.add('hidden');
        
        // Pre-fill form values
        document.getElementById('user-id').value = userInfo.userId;
        document.getElementById('telegram-id').value = userInfo.telegramId;
    }
}

function hideUserInfo() {
    document.getElementById('user-info-display').classList.add('hidden');
    document.querySelector('.user-info-section .form').classList.remove('hidden');
}

// User bookings management
function getUserBookings() {
    const bookings = localStorage.getItem('userBookings');
    return bookings ? JSON.parse(bookings) : [];
}

function saveUserBookings(bookings) {
    localStorage.setItem('userBookings', JSON.stringify(bookings));
}

function addUserBooking(booking) {
    const bookings = getUserBookings();
    bookings.unshift(booking);
    saveUserBookings(bookings);
    displayUserBookings();
}

function updateBookingStatus(bookingId, status) {
    const bookings = getUserBookings();
    const booking = bookings.find(b => b.id === bookingId);
    if (booking) {
        booking.status = status;
        saveUserBookings(bookings);
        displayUserBookings();
    }
}

function displayUserBookings() {
    const bookingsContainer = document.getElementById('user-bookings');
    const bookings = getUserBookings();
    
    if (bookings.length === 0) {
        bookingsContainer.innerHTML = '<p class="info-message">У вас пока нет бронирований</p>';
        return;
    }
    
    const now = new Date();
    
    bookingsContainer.innerHTML = bookings.map(booking => {
        const deadline = new Date(booking.paymentDeadline);
        const isExpired = now > deadline && booking.status === 'pending';
        const isConfirmed = booking.status === 'confirmed';
        
        let statusClass = '';
        let statusText = booking.status;
        
        if (isExpired) {
            statusClass = 'expired';
            statusText = 'Просрочено';
        } else if (isConfirmed) {
            statusClass = 'confirmed';
            statusText = 'Подтверждено';
        } else if (booking.status === 'pending') {
            statusText = 'Ожидает оплаты';
        }
        
        return `
            <div class="booking-item ${statusClass}">
                <h4>Бронирование ${booking.eventName || 'Неизвестное мероприятие'}</h4>
                <p><strong>ID бронирования:</strong> <span class="event-id">${booking.id}</span></p>
                <p><strong>ID мероприятия:</strong> <span class="event-id">${booking.eventId}</span></p>
                <p><strong>Статус:</strong> <span class="status ${statusClass}">${statusText}</span></p>
                <p><strong>Срок оплаты:</strong> ${formatDateTime(booking.paymentDeadline)}</p>
                ${!isExpired && !isConfirmed ? `
                    <div class="booking-actions">
                        <button onclick="confirmPayment('${booking.id}')" class="btn btn-success">Подтвердить оплату</button>
                        <button onclick="viewEventDetails('${booking.eventId}')" class="btn btn-secondary">Мероприятие</button>
                    </div>
                ` : ''}
            </div>
        `;
    }).join('');
}

function viewEventDetails(eventId) {
    window.open(`event.html?id=${eventId}`, '_blank');
}

async function confirmPayment(bookingId) {
    const userInfo = getUserInfo();
    if (!userInfo) {
        showMessage('Сначала введите информацию пользователя', 'warning');
        return;
    }

    try {
        // We need the event ID to confirm payment, let's find it from bookings
        const bookings = getUserBookings();
        const booking = bookings.find(b => b.id === bookingId);
        
        if (!booking) {
            showMessage('Бронирование не найдено', 'error');
            return;
        }

        const response = await fetch(`${API_BASE_URL}/events/${booking.eventId}/confirm`, {
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
            showMessage('Оплата успешно подтверждена!', 'success');
            updateBookingStatus(bookingId, 'confirmed');
        } else {
            showMessage(data.error || 'Ошибка при подтверждении оплаты', 'error');
        }
    } catch (error) {
        console.error('Error:', error);
        showMessage('Ошибка соединения с сервером', 'error');
    }
}

// Event search and booking
async function searchEvent() {
    const eventId = document.getElementById('event-id-input').value.trim();
    
    if (!eventId) {
        showMessage('Введите ID мероприятия', 'warning');
        return;
    }
    
    try {
        const response = await fetch(`${API_BASE_URL}/events/${eventId}`);
        const data = await response.json();
        
        if (response.ok) {
            displayEventInfo(data, eventId);
            document.getElementById('event-info-section').classList.remove('hidden');
        } else {
            showMessage(data.error || 'Мероприятие не найдено', 'error');
            document.getElementById('event-info-section').classList.add('hidden');
        }
    } catch (error) {
        console.error('Error:', error);
        showMessage('Ошибка соединения с сервером', 'error');
    }
}

function displayEventInfo(eventData, eventId) {
    const eventDetails = document.getElementById('event-details');
    
    let statusClass = 'available';
    let statusText = 'Места доступны';
    
    if (eventData.available === 0) {
        statusClass = 'full';
        statusText = 'Мест нет';
    } else if (eventData.available < eventData.capacity * 0.2) {
        statusClass = 'limited';
        statusText = 'Мало мест';
    }
    
    eventDetails.innerHTML = `
        <div class="event-info">
            <h3>${eventData.name}</h3>
            <p><strong>ID мероприятия:</strong> <span class="event-id">${eventId}</span></p>
            <p><strong>Дата и время:</strong> ${formatDateTime(eventData.date)}</p>
            <p><strong>Доступно мест:</strong> ${eventData.available} из ${eventData.capacity}</p>
            <span class="status ${statusClass}">${statusText}</span>
        </div>
    `;
    
    // Update book button state
    const bookBtn = document.getElementById('book-seat-btn');
    if (eventData.available === 0) {
        bookBtn.disabled = true;
        bookBtn.textContent = 'Мест нет';
        bookBtn.classList.remove('btn-success');
        bookBtn.classList.add('btn-secondary');
    } else {
        bookBtn.disabled = false;
        bookBtn.textContent = 'Забронировать место';
        bookBtn.classList.remove('btn-secondary');
        bookBtn.classList.add('btn-success');
    }
    
    // Store event ID for booking
    bookBtn.dataset.eventId = eventId;
    bookBtn.dataset.eventName = eventData.name;
}

async function bookSeat(eventId, eventName) {
    const userInfo = getUserInfo();
    if (!userInfo) {
        showMessage('Сначала введите информацию пользователя', 'warning');
        return;
    }
    
    try {
        const response = await fetch(`${API_BASE_URL}/events/${eventId}/book`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                user_id: userInfo.userId,
                telegram_id: parseInt(userInfo.telegramId)
            })
        });
        
        const data = await response.json();
        
        if (response.ok) {
            showMessage('Место успешно забронировано!', 'success');
            
            // Add booking to user bookings
            addUserBooking({
                id: data.id,
                eventId: eventId,
                eventName: eventName,
                status: data.status,
                paymentDeadline: data.payment_deadline
            });
            
            // Refresh event info
            searchEvent();
        } else {
            showMessage(data.error || 'Ошибка при бронировании', 'error');
        }
    } catch (error) {
        console.error('Error:', error);
        showMessage('Ошибка соединения с сервером', 'error');
    }
}

// Event listeners
document.getElementById('save-user-info').addEventListener('click', () => {
    const userId = document.getElementById('user-id').value.trim();
    const telegramId = document.getElementById('telegram-id').value.trim();
    
    if (!userId || !telegramId) {
        showMessage('Заполните все поля', 'warning');
        return;
    }
    
    // Validate UUID format
    const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
    if (!uuidRegex.test(userId)) {
        showMessage('Некорректный формат User ID (должен быть UUID)', 'warning');
        return;
    }
    
    const userInfo = {
        userId: userId,
        telegramId: telegramId
    };
    
    saveUserInfo(userInfo);
    displayUserInfo();
    showMessage('Информация пользователя сохранена', 'success');
});

document.getElementById('edit-user-info').addEventListener('click', () => {
    hideUserInfo();
});

document.getElementById('search-event-btn').addEventListener('click', searchEvent);

document.getElementById('book-seat-btn').addEventListener('click', (e) => {
    const eventId = e.target.dataset.eventId;
    const eventName = e.target.dataset.eventName;
    if (eventId) {
        bookSeat(eventId, eventName);
    }
});

document.getElementById('view-full-event-btn').addEventListener('click', (e) => {
    const eventId = document.getElementById('book-seat-btn').dataset.eventId;
    if (eventId) {
        window.open(`event.html?id=${eventId}`, '_blank');
    }
});

// Initialize page
document.addEventListener('DOMContentLoaded', () => {
    displayUserInfo();
    displayUserBookings();
});

// Make functions global for onclick handlers
window.confirmPayment = confirmPayment;
window.viewEventDetails = viewEventDetails;

document.getElementById('generate-uuid-btn').addEventListener('click', () => {
    const uuid = generateUUID();
    document.getElementById('event-id-input').value = uuid;
    showMessage('Случайный UUID сгенерирован', 'success');
});
