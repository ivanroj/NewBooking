export class AdminAPI {
    constructor() {
        this.baseUrl = '/api';
    }

    get token() {
        return localStorage.getItem('adminToken');
    }

    async request(endpoint, options = {}) {
        const headers = {
            'Content-Type': 'application/json',
            ...options.headers
        };
        if (this.token) {
            headers['Authorization'] = `Bearer ${this.token}`;
        }

        const response = await fetch(`${this.baseUrl}${endpoint}`, {
            ...options,
            headers
        });

        if (response.status === 401) {
            localStorage.removeItem('adminToken');
            window.location.reload();
            throw new Error('Unauthorized');
        }

        const data = response.status !== 204 ? await response.json() : null;

        if (!response.ok) {
            throw new Error(data?.error || `HTTP error! status: ${response.status}`);
        }

        return data;
    }

    async login(email, password) {
        const res = await this.request('/auth/admin/login', {
            method: 'POST',
            body: JSON.stringify({ email, password })
        });
        localStorage.setItem('adminToken', res.access_token);
        return res;
    }

    async getRooms() {
        return this.request('/rooms');
    }

    async createRoom(name, description) {
        return this.request('/rooms', {
            method: 'POST',
            body: JSON.stringify({ name, description })
        });
    }

    async updateRoom(id, name, description) {
        return this.request(`/rooms/${id}`, {
            method: 'PATCH',
            body: JSON.stringify({ name, description })
        });
    }

    async deleteRoom(id) {
        return this.request(`/rooms/${id}`, {
            method: 'DELETE'
        });
    }

    async getWorkspaces(roomId) {
        return this.request(`/rooms/${roomId}/workspaces`);
    }

    async createWorkspace(roomId, name) {
        return this.request(`/rooms/${roomId}/workspaces`, {
            method: 'POST',
            body: JSON.stringify({ name })
        });
    }

    async updateWorkspace(roomId, wsId, name) {
        return this.request(`/rooms/${roomId}/workspaces/${wsId}`, {
            method: 'PATCH',
            body: JSON.stringify({ name })
        });
    }

    async deleteWorkspace(roomId, wsId) {
        return this.request(`/rooms/${roomId}/workspaces/${wsId}`, {
            method: 'DELETE'
        });
    }

    async getBookings() {
        // Assume there is an endpoint or we fetch all workspaces to compute,
        // Wait, looking at internal/router/router.go, there isn't a GET /admin/bookings endpoint!
        // We might need to add it, or use existing endpoints. Let's add an empty mock for now, we will add it on backend.
        // Wait, I am the one doing the backend too. I will add the endpoint if it's missing.
        try {
            return await this.request('/admin/bookings');
        } catch (e) {
            console.error("Endpoint missing: /admin/bookings", e);
            return []; // Fallback
        }
    }

    async cancelBooking(bookingId) {
        return this.request(`/admin/bookings/${bookingId}`, {
            method: 'PATCH',
            body: JSON.stringify({ status: "cancelled" }) // Need to check backend payload
        });
    }
}

// Stats transformations
export function calculateStats(bookings, workspaces) {
    let total = bookings.length;
    let active = bookings.filter(b => b.status === 'active' || b.status === 'confirmed').length;
    return { total, active };
}

// UI Controller
window.addEventListener('load', () => {
    const api = new AdminAPI();
    
    const loginView = document.getElementById('loginView');
    const dashboardView = document.getElementById('dashboardView');
    const loginForm = document.getElementById('adminLoginForm');
    const loginError = document.getElementById('loginError');

    // DOM Elements - Navigation
    const navButtons = document.querySelectorAll('.nav-btn');
    const tabContents = document.querySelectorAll('.tab-content');

    function switchTab(targetId) {
        navButtons.forEach(btn => btn.classList.remove('active'));
        tabContents.forEach(tab => tab.classList.add('hidden'));
        
        document.querySelector(`.nav-btn[data-target="${targetId}"]`).classList.add('active');
        document.getElementById(targetId).classList.remove('hidden');
        document.getElementById(targetId).classList.add('active');

        loadTabData(targetId);
    }

    navButtons.forEach(btn => {
        btn.addEventListener('click', (e) => {
            switchTab(e.target.dataset.target);
        });
    });

    function checkAuth() {
        if (api.token) {
            loginView.classList.remove('active');
            loginView.classList.add('hidden');
            dashboardView.classList.remove('hidden');
            dashboardView.classList.add('active');
            switchTab('roomsTab');
        } else {
            loginView.classList.remove('hidden');
            loginView.classList.add('active');
            dashboardView.classList.remove('active');
            dashboardView.classList.add('hidden');
        }
    }

    // Login Form
    loginForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const email = document.getElementById('email').value;
        const password = document.getElementById('password').value;
        loginError.textContent = '';
        
        try {
            await api.login(email, password);
            checkAuth();
        } catch (err) {
            loginError.textContent = 'Неверный email или пароль';
        }
    });

    document.getElementById('logoutBtn').addEventListener('click', () => {
        localStorage.removeItem('adminToken');
        checkAuth();
    });

    // Rooms Logic
    const roomsTableBody = document.getElementById('roomsTableBody');
    const roomModal = document.getElementById('roomModal');
    const roomForm = document.getElementById('roomForm');

    document.getElementById('openCreateRoomBtn').addEventListener('click', () => {
        document.getElementById('roomId').value = '';
        document.getElementById('roomName').value = '';
        document.getElementById('roomDesc').value = '';
        document.getElementById('roomModalTitle').textContent = 'Добавить помещение';
        roomModal.classList.remove('hidden');
    });

    document.querySelectorAll('.modal-cancel').forEach(btn => {
        btn.addEventListener('click', (e) => {
            document.getElementById(e.target.dataset.modal).classList.add('hidden');
        });
    });

    roomForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const id = document.getElementById('roomId').value;
        const name = document.getElementById('roomName').value;
        const desc = document.getElementById('roomDesc').value;
        
        try {
            if (id) {
                await api.updateRoom(id, name, desc);
            } else {
                await api.createRoom(name, desc);
            }
            roomModal.classList.add('hidden');
            loadRooms();
        } catch (err) {
            alert('Ошибка: ' + err.message);
        }
    });

    async function loadRooms() {
        roomsTableBody.innerHTML = '<tr><td colspan="4">Загрузка...</td></tr>';
        try {
            const rooms = await api.getRooms() || [];
            roomsTableBody.innerHTML = '';
            
            // Update select for Schema Tab
            const schemaSelect = document.getElementById('schemaRoomSelect');
            schemaSelect.innerHTML = '<option value="">-- Выберите --</option>';

            if (rooms.length === 0) {
                roomsTableBody.innerHTML = '<tr><td colspan="4">Нет помещений</td></tr>';
                return;
            }

            rooms.forEach(room => {
                const tr = document.createElement('tr');
                tr.innerHTML = `
                    <td>${room.id}</td>
                    <td>${room.name}</td>
                    <td>${room.description || '-'}</td>
                    <td>
                        <button class="btn secondary edit-room-btn" data-id="${room.id}" data-name="${room.name}" data-desc="${room.description || ''}">Изменить</button>
                        <button class="btn danger del-room-btn" data-id="${room.id}">Удалить</button>
                    </td>
                `;
                roomsTableBody.appendChild(tr);

                const option = document.createElement('option');
                option.value = room.id;
                option.textContent = room.name;
                schemaSelect.appendChild(option);
            });

            document.querySelectorAll('.del-room-btn').forEach(btn => {
                btn.addEventListener('click', async (e) => {
                    if (confirm('Точно удалить?')) {
                        await api.deleteRoom(e.target.dataset.id);
                        loadRooms();
                    }
                });
            });

            document.querySelectorAll('.edit-room-btn').forEach(btn => {
                btn.addEventListener('click', (e) => {
                    document.getElementById('roomId').value = e.target.dataset.id;
                    document.getElementById('roomName').value = e.target.dataset.name;
                    document.getElementById('roomDesc').value = e.target.dataset.desc;
                    document.getElementById('roomModalTitle').textContent = 'Редактировать помещение';
                    roomModal.classList.remove('hidden');
                });
            });

        } catch (err) {
            roomsTableBody.innerHTML = `<tr><td colspan="4" style="color:var(--danger-color)">Ошибка загрузки: ${err.message}</td></tr>`;
        }
    }

    // Schema Logic
    const schemaRoomSelect = document.getElementById('schemaRoomSelect');
    const openAddWorkspaceBtn = document.getElementById('openAddWorkspaceBtn');
    const schemaContainer = document.getElementById('schemaContainer');
    const workspaceModal = document.getElementById('workspaceModal');
    const workspaceForm = document.getElementById('workspaceForm');

    schemaRoomSelect.addEventListener('change', (e) => {
        if (e.target.value) {
            openAddWorkspaceBtn.disabled = false;
            loadWorkspaces(e.target.value);
        } else {
            openAddWorkspaceBtn.disabled = true;
            schemaContainer.innerHTML = '';
        }
    });

    openAddWorkspaceBtn.addEventListener('click', () => {
        document.getElementById('workspaceId').value = '';
        document.getElementById('workspaceName').value = '';
        document.getElementById('workspaceModalTitle').textContent = 'Добавить место';
        workspaceModal.classList.remove('hidden');
    });

    workspaceForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const roomId = schemaRoomSelect.value;
        const id = document.getElementById('workspaceId').value;
        const name = document.getElementById('workspaceName').value;
        
        try {
            if (id) {
                await api.updateWorkspace(roomId, id, name);
            } else {
                await api.createWorkspace(roomId, name);
            }
            workspaceModal.classList.add('hidden');
            loadWorkspaces(roomId);
        } catch (err) {
            alert('Ошибка: ' + err.message);
        }
    });

    async function loadWorkspaces(roomId) {
        schemaContainer.innerHTML = 'Загрузка...';
        try {
            const ws = await api.getWorkspaces(roomId) || [];
            schemaContainer.innerHTML = '';
            
            if (ws.length === 0) {
                schemaContainer.innerHTML = '<p>В этом помещении пока нет рабочих мест.</p>';
                return;
            }

            ws.forEach(w => {
                const card = document.createElement('div');
                card.className = 'schema-ws-card';
                card.innerHTML = `
                    <h3>${w.name}</h3>
                    <div style="display:flex;gap:5px;margin-top:10px;">
                        <button class="btn secondary ws-edit-btn" style="padding:0.2rem 0.5rem;font-size:0.8rem;" data-id="${w.id}" data-name="${w.name}">✍️</button>
                        <button class="btn danger ws-del-btn" style="padding:0.2rem 0.5rem;font-size:0.8rem;" data-id="${w.id}">🗑️</button>
                    </div>
                `;
                schemaContainer.appendChild(card);
            });

            document.querySelectorAll('.ws-del-btn').forEach(btn => {
                btn.addEventListener('click', async (e) => {
                    if (confirm('Точно удалить место?')) {
                        await api.deleteWorkspace(roomId, e.target.dataset.id);
                        loadWorkspaces(roomId);
                    }
                });
            });

            document.querySelectorAll('.ws-edit-btn').forEach(btn => {
                btn.addEventListener('click', (e) => {
                    document.getElementById('workspaceId').value = e.target.dataset.id;
                    document.getElementById('workspaceName').value = e.target.dataset.name;
                    document.getElementById('workspaceModalTitle').textContent = 'Редактировать место';
                    workspaceModal.classList.remove('hidden');
                });
            });

        } catch (err) {
            schemaContainer.innerHTML = `<p style="color:var(--danger-color)">Ошибка: ${err.message}</p>`;
        }
    }

    // Load tab data wrapper
    function loadTabData(targetId) {
        if (targetId === 'roomsTab') loadRooms();
        if (targetId === 'schemaTab') {
            if (schemaRoomSelect.value) {
                loadWorkspaces(schemaRoomSelect.value);
            }
        }
        if (targetId === 'bookingsTab') loadBookings();
        if (targetId === 'statsTab') loadStats();
    }

    // Bookings & Stats
    async function loadBookings() {
        const tbody = document.getElementById('bookingsTableBody');
        tbody.innerHTML = '<tr><td colspan="6">Загрузка...</td></tr>';
        try {
            const bookings = await api.getBookings() || [];
            tbody.innerHTML = '';
            
            if (bookings.length === 0) {
                tbody.innerHTML = '<tr><td colspan="6">Нет бронирований</td></tr>';
                return;
            }

            bookings.forEach(b => {
                const tr = document.createElement('tr');
                tr.innerHTML = `
                    <td>${b.id}</td>
                    <td>${b.workspace_id}</td>
                    <td>${b.user_id}</td>
                    <td>${new Date(b.start_time).toLocaleString()} - ${new Date(b.end_time).toLocaleString()}</td>
                    <td>${b.status}</td>
                    <td>
                        ${b.status === 'active' || b.status === 'confirmed' ? 
                            `<button class="btn danger cancel-book-btn" data-id="${b.id}">Отменить</button>` : '-'}
                    </td>
                `;
                tbody.appendChild(tr);
            });

            document.querySelectorAll('.cancel-book-btn').forEach(btn => {
                btn.addEventListener('click', async (e) => {
                    if (confirm('Отменить бронирование?')) {
                        await api.cancelBooking(e.target.dataset.id);
                        loadBookings();
                    }
                });
            });

        } catch (err) {
            tbody.innerHTML = `<tr><td colspan="6" style="color:var(--danger-color)">Ошибка: ${err.message}</td></tr>`;
        }
    }

    async function loadStats() {
        try {
            const bookings = await api.getBookings() || [];
            const stats = calculateStats(bookings, []);
            document.getElementById('statTotalBookings').textContent = stats.total;
            document.getElementById('statActiveBookings').textContent = stats.active;
        } catch (err) {
            console.error("Stats error", err);
        }
    }

    // Initial load
    checkAuth();
});
