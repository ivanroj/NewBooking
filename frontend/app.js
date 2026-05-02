// API Client Wrapper
class API {
    constructor() {
        this.baseUrl = '/api';
        this.token = localStorage.getItem('jwt_token');
    }

    setToken(token) {
        this.token = token;
        localStorage.setItem('jwt_token', token);
    }

    async request(endpoint, options = {}) {
        const headers = { 'Content-Type': 'application/json' };
        if (this.token) {
            headers['Authorization'] = `Bearer ${this.token}`;
        }
        
        const config = { ...options, headers: { ...headers, ...options.headers } };
        
        try {
            const response = await fetch(`${this.baseUrl}${endpoint}`, config);
            const data = await response.json().catch(() => ({}));
            
            if (!response.ok) {
                throw { status: response.status, ...data };
            }
            return data;
        } catch (error) {
            if (error.status === 401) {
                localStorage.removeItem('jwt_token');
                this.token = null;
            }
            throw error;
        }
    }
}

const api = new API();
const tg = window.Telegram?.WebApp;

// State
let state = {
    currentRoomId: null,
    selectedWorkspace: null,
    bookingDate: '',
    startTime: '',
    endTime: ''
};

// DOM Elements
const views = {
    loading: document.getElementById('loadingView'),
    rooms: document.getElementById('roomsView'),
    workspaces: document.getElementById('workspacesView'),
    history: document.getElementById('historyView')
};

const navBar = document.getElementById('navBar');
const modal = document.getElementById('bookingModal');

// Init
async function initApp() {
    if (tg) {
        tg.ready();
        tg.expand();
    }

    // Set default dates for inputs
    const now = new Date();
    document.getElementById('bookingDate').value = now.toISOString().split('T')[0];
    document.getElementById('startTime').value = `${String(now.getHours()+1).padStart(2, '0')}:00`;
    document.getElementById('endTime').value = `${String(now.getHours()+3).padStart(2, '0')}:00`;

    setupEventListeners();

    try {
        // Try to auth via Telegram WebApp data
        if (tg && tg.initData) {
            const { access_token } = await api.request('/auth/student/telegram', {
                method: 'POST',
                body: JSON.stringify({ init_data: tg.initData })
            });
            api.setToken(access_token);
        } else if (!api.token) {
            // No token, no Telegram environment
            document.getElementById('loadingText').textContent = "Пожалуйста, откройте приложение через Telegram.";
            document.querySelector('.spinner').style.display = 'none';
            return;
        }

        // Load initial data
        await loadRooms();
        switchView('rooms');
        navBar.classList.remove('hidden');
    } catch (err) {
        document.getElementById('loadingText').textContent = "Ошибка авторизации.";
        document.querySelector('.spinner').style.display = 'none';
        if (tg) tg.showAlert("Не удалось авторизоваться: " + (err.error || err.message));
    }
}

// Navigation
function switchView(viewName) {
    Object.values(views).forEach(v => v.classList.remove('active'));
    views[viewName].classList.add('active');

    // Update nav buttons
    document.querySelectorAll('.nav-btn').forEach(btn => {
        if (btn.dataset.target === `${viewName}View`) {
            btn.classList.add('active');
        } else {
            btn.classList.remove('active');
        }
    });

    if (viewName === 'history') {
        loadHistory();
    } else if (viewName === 'rooms') {
        loadRooms();
    }
}

function setupEventListeners() {
    document.querySelectorAll('.nav-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            const target = e.target.dataset.target.replace('View', '');
            switchView(target);
        });
    });

    document.getElementById('backToRooms').addEventListener('click', () => {
        switchView('rooms');
    });

    document.getElementById('searchWorkspacesBtn').addEventListener('click', () => {
        loadWorkspaces(state.currentRoomId);
    });

    document.getElementById('modalCancel').addEventListener('click', () => {
        modal.classList.add('hidden');
    });

    document.getElementById('modalConfirm').addEventListener('click', confirmBooking);
}

// Logic: Rooms
async function loadRooms() {
    try {
        const rooms = await api.request('/rooms');
        const container = document.getElementById('roomsList');
        container.innerHTML = '';
        
        rooms.forEach(room => {
            const card = document.createElement('div');
            card.className = 'glass-card room-card';
            card.innerHTML = `
                <h3>${escapeHTML(room.name)}</h3>
                <p>${escapeHTML(room.description || '')}</p>
            `;
            card.addEventListener('click', () => {
                state.currentRoomId = room.id;
                document.getElementById('roomTitle').textContent = room.name;
                switchView('workspaces');
                loadWorkspaces(room.id);
            });
            container.appendChild(card);
        });
    } catch (err) {
        if (tg) tg.showAlert("Ошибка загрузки помещений");
    }
}

// Logic: Workspaces
async function loadWorkspaces(roomId) {
    const date = document.getElementById('bookingDate').value;
    const start = document.getElementById('startTime').value;
    const end = document.getElementById('endTime').value;

    if (!date || !start || !end) {
        if (tg) tg.showAlert("Выберите дату и время");
        return;
    }

    state.bookingDate = date;
    state.startTime = start;
    state.endTime = end;

    const startISO = new Date(`${date}T${start}:00Z`).toISOString();
    const endISO = new Date(`${date}T${end}:00Z`).toISOString();

    try {
        const list = await api.request(`/rooms/${roomId}/workspaces?start=${startISO}&end=${endISO}`);
        const container = document.getElementById('workspacesList');
        container.innerHTML = '';

        if (list.length === 0) {
            container.innerHTML = '<p>В этом помещении пока нет мест.</p>';
            return;
        }

        list.forEach(ws => {
            const card = document.createElement('div');
            card.className = `glass-card ws-card ${ws.available ? 'available' : 'unavailable'}`;
            card.textContent = ws.name;
            
            if (ws.available) {
                card.addEventListener('click', () => {
                    state.selectedWorkspace = ws.id;
                    document.getElementById('modalText').textContent = `Забронировать ${ws.name} на ${start}-${end}?`;
                    modal.classList.remove('hidden');
                });
            }
            container.appendChild(card);
        });
    } catch (err) {
        if (tg) tg.showAlert("Ошибка загрузки схемы мест");
    }
}

// Logic: Booking
async function confirmBooking() {
    modal.classList.add('hidden');
    const startISO = new Date(`${state.bookingDate}T${state.startTime}:00Z`).toISOString();
    const endISO = new Date(`${state.bookingDate}T${state.endTime}:00Z`).toISOString();

    try {
        await api.request('/bookings', {
            method: 'POST',
            body: JSON.stringify({
                workspace_id: state.selectedWorkspace,
                start_time: startISO,
                end_time: endISO
            })
        });
        if (tg) tg.showAlert("Успешно забронировано!");
        switchView('history');
    } catch (err) {
        let msg = "Ошибка бронирования.";
        if (err.error === 'conflict') msg = "Место уже занято на это время.";
        if (err.error === 'limit_exceeded') msg = "Вы превысили лимит активных бронирований.";
        if (err.error === 'invalid_time_window') msg = "Неверно указано время.";
        if (tg) tg.showAlert(msg);
    }
}

// Logic: History
async function loadHistory() {
    try {
        const res = await api.request('/bookings/my');
        const container = document.getElementById('historyList');
        container.innerHTML = '';

        if (!res.items || res.items.length === 0) {
            container.innerHTML = '<p>У вас нет бронирований.</p>';
            return;
        }

        res.items.forEach(b => {
            const sTime = new Date(b.start_time).toLocaleString();
            const eTime = new Date(b.end_time).toLocaleTimeString();
            
            const card = document.createElement('div');
            card.className = 'glass-card';
            card.innerHTML = `
                <h3>Место #${b.workspace_id}</h3>
                <p>${sTime} - ${eTime}</p>
                <p>Статус: <strong>${b.status}</strong></p>
            `;

            if (b.status === 'active') {
                const cancelBtn = document.createElement('button');
                cancelBtn.className = 'btn danger mt-10';
                cancelBtn.textContent = 'Отменить';
                cancelBtn.addEventListener('click', () => cancelBooking(b.id));
                card.appendChild(cancelBtn);
            }
            container.appendChild(card);
        });
    } catch (err) {
        if (tg) tg.showAlert("Ошибка загрузки истории");
    }
}

async function cancelBooking(id) {
    if (tg) {
        tg.showConfirm("Вы уверены, что хотите отменить бронь?", async (ok) => {
            if (ok) await doCancel(id);
        });
    } else {
        if (confirm("Вы уверены?")) await doCancel(id);
    }
}

async function doCancel(id) {
    try {
        await api.request(`/bookings/${id}/cancel`, { method: 'PATCH' });
        loadHistory();
    } catch (err) {
        if (tg) tg.showAlert("Ошибка отмены");
    }
}

// Utils
function escapeHTML(str) {
    const div = document.createElement('div');
    div.textContent = str;
    return div.innerHTML;
}

// Start
window.addEventListener('DOMContentLoaded', initApp);

// Export for tests
export { API, switchView, loadRooms, loadWorkspaces, confirmBooking, loadHistory };
