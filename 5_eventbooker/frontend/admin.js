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

// Store created events
let createdEvents = JSON.parse(localStorage.getItem('createdEvents')) || [];

function saveCreatedEvents() {
    localStorage.setItem('createdEvents', JSON.stringify(createdEvents));
}

function displayCreatedEvents() {
    const eventsContainer = document.getElementById('created-events');
    
    if (createdEvents.length === 0) {
        eventsContainer.innerHTML = '<p class="info-message">Мероприятия еще не созданы</p>';
        return;
    }
    
    eventsContainer.innerHTML = createdEvents.map(event => `
        <div class="event-item">
            <h4>${event.name}</h4>
            <p><strong>ID:</strong> <span class="event-id">${event.id}</span></p>
            <p><strong>Дата:</strong> ${formatDateTime(event.date)}</p>
            <p><strong>Вместимость:</strong> ${event.capacity} мест</p>
            <p><strong>TTL бронирования:</strong> ${event.booking_ttl || 300} секунд</p>
            <div class="booking-actions">
                <button onclick="viewEvent('${event.id}')" class="btn btn-primary">Посмотреть</button>
                <button onclick="copyEventId('${event.id}')" class="btn btn-secondary">Копировать ID</button>
            </div>
        </div>
    `).join('');
}

function copyEventId(eventId) {
    navigator.clipboard.writeText(eventId).then(() => {
        showMessage('ID мероприятия скопирован в буфер обмена', 'success');
    }).catch(() => {
        showMessage('Не удалось скопировать ID', 'error');
    });
}

function viewEvent(eventId) {
    window.open(`event.html?id=${eventId}`, '_blank');
}

// Create event
document.getElementById('create-event-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const formData = {
        name: document.getElementById('event-name').value,
        date: new Date(document.getElementById('event-date').value).toISOString(),
        capacity: parseInt(document.getElementById('event-capacity').value),
        payment_ttl: parseInt(document.getElementById('payment-ttl').value) || 300
    };
    
    try {
        const response = await fetch(`${API_BASE_URL}/events`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(formData)
        });
        
        const data = await response.json();
        
        if (response.ok) {
            showMessage('Мероприятие успешно создано!', 'success');
            
            // Store created event
            const newEvent = {
                id: data.id,
                name: formData.name,
                date: formData.date,
                capacity: formData.capacity,
                payment_ttl: formData.payment_ttl
            };
            
            createdEvents.unshift(newEvent);
            saveCreatedEvents();
            displayCreatedEvents();
            
            // Reset form
            document.getElementById('create-event-form').reset();
            document.getElementById('payment-ttl').value = '300';
        } else {
            showMessage(data.error || 'Ошибка при создании мероприятия', 'error');
        }
    } catch (error) {
        console.error('Error:', error);
        showMessage('Ошибка соединения с сервером', 'error');
    }
});

// View event by ID
document.getElementById('view-event-btn').addEventListener('click', () => {
    const eventId = document.getElementById('event-id-input').value.trim();
    
    if (!eventId) {
        showMessage('Введите ID мероприятия', 'warning');
        return;
    }
    
    window.open(`event.html?id=${eventId}`, '_blank');
});

// Initialize page
document.addEventListener('DOMContentLoaded', () => {
    displayCreatedEvents();
});
