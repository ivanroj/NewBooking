// ── Admin API Client ───────────────────────────────────────────────────────
export class AdminAPI {
    constructor() {
        this.baseUrl = '/api';
    }

    get token() {
        return localStorage.getItem('adminToken');
    }

    async request(endpoint, options = {}) {
        const headers = { 'Content-Type': 'application/json', ...options.headers };
        if (this.token) {
            headers['Authorization'] = `Bearer ${this.token}`;
        }
        const response = await fetch(`${this.baseUrl}${endpoint}`, { ...options, headers });
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

    async getRooms() { return this.request('/rooms'); }

    async createRoom(name, description) {
        return this.request('/rooms', { method: 'POST', body: JSON.stringify({ name, description }) });
    }

    async updateRoom(id, name, description) {
        return this.request(`/rooms/${id}`, { method: 'PATCH', body: JSON.stringify({ name, description }) });
    }

    async deleteRoom(id) {
        return this.request(`/rooms/${id}`, { method: 'DELETE' });
    }

    async getWorkspaces(roomId) { return this.request(`/rooms/${roomId}/workspaces`); }

    async createWorkspace(roomId, name, gridX, gridY) {
        return this.request(`/rooms/${roomId}/workspaces`, {
            method: 'POST',
            body: JSON.stringify({ name, grid_x: gridX, grid_y: gridY })
        });
    }

    async updateWorkspace(roomId, wsId, name, gridX, gridY) {
        return this.request(`/rooms/${roomId}/workspaces/${wsId}`, {
            method: 'PATCH',
            body: JSON.stringify({ name, grid_x: gridX, grid_y: gridY })
        });
    }

    async deleteWorkspace(roomId, wsId) {
        return this.request(`/rooms/${roomId}/workspaces/${wsId}`, { method: 'DELETE' });
    }

    async getBookings(page, limit, status) {
        let url = `/admin/bookings?page=${page || 1}&limit=${limit || 20}`;
        if (status) url += `&status=${status}`;
        return this.request(url);
    }

    async cancelBooking(bookingId) {
        return this.request(`/admin/bookings/${bookingId}`, {
            method: 'PATCH',
            body: JSON.stringify({ status: 'canceled' })
        });
    }

    async getStats(from, to) {
        let url = '/admin/stats';
        const params = [];
        if (from) params.push(`from=${from}`);
        if (to) params.push(`to=${to}`);
        if (params.length) url += '?' + params.join('&');
        return this.request(url);
    }
}

// ── Stats Transformations (exported for testing) ───────────────────────────
const DAY_NAMES = ['Вс', 'Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб'];

export function calculateStats(bookings) {
    const total = bookings.length;
    const active = bookings.filter(b => b.status === 'active' || b.status === 'confirmed').length;
    return { total, active };
}

export function computeOccupancyPercent(count, maxCount) {
    if (maxCount <= 0) return 0;
    return Math.round((count / maxCount) * 100);
}

export function dayName(dayIndex) {
    return DAY_NAMES[dayIndex] || '?';
}

export function formatDateRange(start, end) {
    const s = new Date(start);
    const e = new Date(end);
    const datePart = s.toLocaleDateString('ru-RU', { day: '2-digit', month: '2-digit', year: 'numeric' });
    const startTime = s.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' });
    const endTime = e.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' });
    return `${datePart} ${startTime}\u2013${endTime}`;
}

export function aggregateByRoom(bookings) {
    const map = {};
    bookings.forEach(b => {
        const key = b.room_name || `Room ${b.workspace_id}`;
        map[key] = (map[key] || 0) + 1;
    });
    return Object.entries(map)
        .map(([name, count]) => ({ name, count }))
        .sort((a, b) => b.count - a.count);
}

// ── Canvas Schema Editor ───────────────────────────────────────────────────
const CELL = 80;
const DESK_W = 60;
const DESK_H = 40;

class CanvasEditor {
    constructor(canvas, api, roomSelectEl) {
        this.canvas = canvas;
        this.ctx = canvas.getContext('2d');
        this.api = api;
        this.roomSelect = roomSelectEl;
        this.workspaces = [];
        this.dragging = null;
        this.dragOffsetX = 0;
        this.dragOffsetY = 0;
        this.addMode = false;
        this.selectedRoomId = null;

        this.canvas.addEventListener('mousedown', (e) => this.onMouseDown(e));
        this.canvas.addEventListener('mousemove', (e) => this.onMouseMove(e));
        this.canvas.addEventListener('mouseup', (e) => this.onMouseUp(e));
        this.canvas.addEventListener('dblclick', (e) => this.onDblClick(e));
        this.canvas.addEventListener('contextmenu', (e) => this.onContextMenu(e));
    }

    getMousePos(e) {
        const rect = this.canvas.getBoundingClientRect();
        return {
            x: (e.clientX - rect.left) * (this.canvas.width / rect.width),
            y: (e.clientY - rect.top) * (this.canvas.height / rect.height)
        };
    }

    findWorkspaceAt(x, y) {
        for (let i = this.workspaces.length - 1; i >= 0; i--) {
            const ws = this.workspaces[i];
            const wx = ws.grid_x * CELL + (CELL - DESK_W) / 2;
            const wy = ws.grid_y * CELL + (CELL - DESK_H) / 2;
            if (x >= wx && x <= wx + DESK_W && y >= wy && y <= wy + DESK_H) {
                return ws;
            }
        }
        return null;
    }

    onMouseDown(e) {
        if (e.button !== 0) return;
        const pos = this.getMousePos(e);

        if (this.addMode) {
            const gx = Math.floor(pos.x / CELL);
            const gy = Math.floor(pos.y / CELL);
            this.addMode = false;
            document.getElementById('canvasHint').classList.add('hidden');
            this.promptAddWorkspace(gx, gy);
            return;
        }

        const ws = this.findWorkspaceAt(pos.x, pos.y);
        if (ws) {
            this.dragging = ws;
            this.dragOffsetX = pos.x - ws.grid_x * CELL;
            this.dragOffsetY = pos.y - ws.grid_y * CELL;
            this.canvas.style.cursor = 'grabbing';
        }
    }

    onMouseMove(e) {
        if (!this.dragging) {
            const pos = this.getMousePos(e);
            const ws = this.findWorkspaceAt(pos.x, pos.y);
            this.canvas.style.cursor = ws ? 'grab' : (this.addMode ? 'crosshair' : 'default');
            return;
        }
        const pos = this.getMousePos(e);
        const gx = Math.max(0, Math.floor((pos.x - this.dragOffsetX + CELL / 2) / CELL));
        const gy = Math.max(0, Math.floor((pos.y - this.dragOffsetY + CELL / 2) / CELL));
        this.dragging.grid_x = gx;
        this.dragging.grid_y = gy;
        this.draw();
    }

    async onMouseUp(_e) {
        if (!this.dragging) return;
        const ws = this.dragging;
        this.dragging = null;
        this.canvas.style.cursor = 'default';
        try {
            await this.api.updateWorkspace(this.selectedRoomId, ws.id, ws.name, ws.grid_x, ws.grid_y);
        } catch (err) {
            alert('Ошибка сохранения позиции: ' + err.message);
        }
    }

    onDblClick(e) {
        const pos = this.getMousePos(e);
        const ws = this.findWorkspaceAt(pos.x, pos.y);
        if (ws) {
            document.getElementById('workspaceId').value = ws.id;
            document.getElementById('workspaceName').value = ws.name;
            document.getElementById('workspaceGridX').value = ws.grid_x;
            document.getElementById('workspaceGridY').value = ws.grid_y;
            document.getElementById('workspaceModalTitle').textContent = 'Редактировать место';
            document.getElementById('workspaceModal').classList.remove('hidden');
        }
    }

    onContextMenu(e) {
        e.preventDefault();
        const pos = this.getMousePos(e);
        const ws = this.findWorkspaceAt(pos.x, pos.y);
        if (ws && confirm(`Удалить место "${ws.name}"?`)) {
            this.api.deleteWorkspace(this.selectedRoomId, ws.id)
                .then(() => this.loadWorkspaces(this.selectedRoomId))
                .catch(err => alert('Ошибка: ' + err.message));
        }
    }

    promptAddWorkspace(gx, gy) {
        document.getElementById('workspaceId').value = '';
        document.getElementById('workspaceName').value = '';
        document.getElementById('workspaceGridX').value = gx;
        document.getElementById('workspaceGridY').value = gy;
        document.getElementById('workspaceModalTitle').textContent = 'Добавить место';
        document.getElementById('workspaceModal').classList.remove('hidden');
    }

    async loadWorkspaces(roomId) {
        this.selectedRoomId = roomId;
        try {
            this.workspaces = await this.api.getWorkspaces(roomId) || [];
        } catch (_err) {
            this.workspaces = [];
        }
        this.draw();
    }

    draw() {
        const ctx = this.ctx;
        const w = this.canvas.width;
        const h = this.canvas.height;

        ctx.clearRect(0, 0, w, h);

        // Background grid
        ctx.strokeStyle = 'rgba(255,255,255,0.06)';
        ctx.lineWidth = 1;
        for (let x = 0; x <= w; x += CELL) {
            ctx.beginPath(); ctx.moveTo(x, 0); ctx.lineTo(x, h); ctx.stroke();
        }
        for (let y = 0; y <= h; y += CELL) {
            ctx.beginPath(); ctx.moveTo(0, y); ctx.lineTo(w, y); ctx.stroke();
        }

        // Room boundary
        ctx.strokeStyle = 'rgba(59,130,246,0.3)';
        ctx.lineWidth = 2;
        ctx.strokeRect(2, 2, w - 4, h - 4);

        // Workspaces
        this.workspaces.forEach(ws => {
            const x = ws.grid_x * CELL + (CELL - DESK_W) / 2;
            const y = ws.grid_y * CELL + (CELL - DESK_H) / 2;

            // Desk shape
            ctx.fillStyle = ws === this.dragging
                ? 'rgba(59,130,246,0.5)'
                : 'rgba(59,130,246,0.2)';
            ctx.strokeStyle = 'rgba(59,130,246,0.8)';
            ctx.lineWidth = 1.5;
            this.roundRect(ctx, x, y, DESK_W, DESK_H, 6);
            ctx.fill();
            ctx.stroke();

            // Label
            ctx.fillStyle = '#e2e8f0';
            ctx.font = '12px Inter, sans-serif';
            ctx.textAlign = 'center';
            ctx.textBaseline = 'middle';
            ctx.fillText(ws.name, x + DESK_W / 2, y + DESK_H / 2);
        });

        if (this.workspaces.length === 0 && this.selectedRoomId) {
            ctx.fillStyle = 'rgba(148,163,184,0.5)';
            ctx.font = '16px Inter, sans-serif';
            ctx.textAlign = 'center';
            ctx.textBaseline = 'middle';
            ctx.fillText('Нет рабочих мест. Нажмите "Добавить место" и кликните на план.', w / 2, h / 2);
        }
    }

    roundRect(ctx, x, y, w, h, r) {
        ctx.beginPath();
        ctx.moveTo(x + r, y);
        ctx.lineTo(x + w - r, y);
        ctx.quadraticCurveTo(x + w, y, x + w, y + r);
        ctx.lineTo(x + w, y + h - r);
        ctx.quadraticCurveTo(x + w, y + h, x + w - r, y + h);
        ctx.lineTo(x + r, y + h);
        ctx.quadraticCurveTo(x, y + h, x, y + h - r);
        ctx.lineTo(x, y + r);
        ctx.quadraticCurveTo(x, y, x + r, y);
        ctx.closePath();
    }
}

// ── UI Controller ──────────────────────────────────────────────────────────
window.addEventListener('load', () => {
    const api = new AdminAPI();

    const loginView = document.getElementById('loginView');
    const dashboardView = document.getElementById('dashboardView');
    const loginForm = document.getElementById('adminLoginForm');
    const loginError = document.getElementById('loginError');

    const navButtons = document.querySelectorAll('.nav-btn');
    const tabContents = document.querySelectorAll('.tab-content');

    // ── Navigation ─────────────────────────────────────────────────────────
    function switchTab(targetId) {
        navButtons.forEach(btn => btn.classList.remove('active'));
        tabContents.forEach(tab => { tab.classList.add('hidden'); tab.classList.remove('active'); });
        const btn = document.querySelector(`.nav-btn[data-target="${targetId}"]`);
        if (btn) btn.classList.add('active');
        const tab = document.getElementById(targetId);
        if (tab) { tab.classList.remove('hidden'); tab.classList.add('active'); }
        loadTabData(targetId);
    }

    navButtons.forEach(btn => {
        btn.addEventListener('click', (e) => switchTab(e.target.dataset.target));
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

    // ── Login ──────────────────────────────────────────────────────────────
    loginForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const email = document.getElementById('email').value;
        const password = document.getElementById('password').value;
        loginError.textContent = '';
        try {
            await api.login(email, password);
            checkAuth();
        } catch (_err) {
            loginError.textContent = 'Неверный email или пароль';
        }
    });

    document.getElementById('logoutBtn').addEventListener('click', () => {
        localStorage.removeItem('adminToken');
        checkAuth();
    });

    // ── Modal close buttons ────────────────────────────────────────────────
    document.querySelectorAll('.modal-cancel').forEach(btn => {
        btn.addEventListener('click', (e) => {
            document.getElementById(e.target.dataset.modal).classList.add('hidden');
        });
    });

    // ── Rooms Logic ────────────────────────────────────────────────────────
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
                    <td>${escapeHTML(room.name)}</td>
                    <td>${escapeHTML(room.description || '-')}</td>
                    <td>
                        <button class="btn secondary edit-room-btn" data-id="${room.id}" data-name="${escapeAttr(room.name)}" data-desc="${escapeAttr(room.description || '')}">Изменить</button>
                        <button class="btn danger del-room-btn" data-id="${room.id}">Удалить</button>
                    </td>
                `;
                roomsTableBody.appendChild(tr);

                const option = document.createElement('option');
                option.value = room.id;
                option.textContent = room.name;
                schemaSelect.appendChild(option);
            });

            roomsTableBody.querySelectorAll('.del-room-btn').forEach(btn => {
                btn.addEventListener('click', async (ev) => {
                    if (confirm('Точно удалить?')) {
                        await api.deleteRoom(ev.target.dataset.id);
                        loadRooms();
                    }
                });
            });

            roomsTableBody.querySelectorAll('.edit-room-btn').forEach(btn => {
                btn.addEventListener('click', (ev) => {
                    document.getElementById('roomId').value = ev.target.dataset.id;
                    document.getElementById('roomName').value = ev.target.dataset.name;
                    document.getElementById('roomDesc').value = ev.target.dataset.desc;
                    document.getElementById('roomModalTitle').textContent = 'Редактировать помещение';
                    roomModal.classList.remove('hidden');
                });
            });
        } catch (err) {
            roomsTableBody.innerHTML = `<tr><td colspan="4" class="error-text">Ошибка загрузки: ${escapeHTML(err.message)}</td></tr>`;
        }
    }

    // ── Canvas Schema Editor ───────────────────────────────────────────────
    const schemaCanvas = document.getElementById('schemaCanvas');
    const schemaRoomSelect = document.getElementById('schemaRoomSelect');
    const addWsBtn = document.getElementById('addWorkspaceCanvasBtn');
    const canvasHint = document.getElementById('canvasHint');
    const workspaceModal = document.getElementById('workspaceModal');
    const workspaceForm = document.getElementById('workspaceForm');
    const editor = new CanvasEditor(schemaCanvas, api, schemaRoomSelect);

    schemaRoomSelect.addEventListener('change', (e) => {
        const roomId = e.target.value;
        addWsBtn.disabled = !roomId;
        if (roomId) {
            editor.loadWorkspaces(roomId);
        } else {
            editor.workspaces = [];
            editor.draw();
        }
    });

    addWsBtn.addEventListener('click', () => {
        editor.addMode = true;
        canvasHint.classList.remove('hidden');
        schemaCanvas.style.cursor = 'crosshair';
    });

    workspaceForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const roomId = schemaRoomSelect.value;
        const id = document.getElementById('workspaceId').value;
        const name = document.getElementById('workspaceName').value;
        const gx = parseInt(document.getElementById('workspaceGridX').value, 10) || 0;
        const gy = parseInt(document.getElementById('workspaceGridY').value, 10) || 0;

        try {
            if (id) {
                await api.updateWorkspace(roomId, id, name, gx, gy);
            } else {
                await api.createWorkspace(roomId, name, gx, gy);
            }
            workspaceModal.classList.add('hidden');
            editor.loadWorkspaces(roomId);
        } catch (err) {
            alert('Ошибка: ' + err.message);
        }
    });

    // ── Bookings Tab ───────────────────────────────────────────────────────
    let bookingsPage = 1;
    const bookingsStatusFilter = document.getElementById('bookingsStatusFilter');

    bookingsStatusFilter.addEventListener('change', () => {
        bookingsPage = 1;
        loadBookings();
    });

    async function loadBookings() {
        const tbody = document.getElementById('bookingsTableBody');
        tbody.innerHTML = '<tr><td colspan="7">Загрузка...</td></tr>';
        try {
            const status = bookingsStatusFilter.value;
            const result = await api.getBookings(bookingsPage, 20, status);
            const bookings = result.items || [];
            const total = result.total || 0;
            tbody.innerHTML = '';

            if (bookings.length === 0) {
                tbody.innerHTML = '<tr><td colspan="7">Нет бронирований</td></tr>';
                renderPagination(0, 0);
                return;
            }

            bookings.forEach(b => {
                const tr = document.createElement('tr');
                tr.innerHTML = `
                    <td>${b.id}</td>
                    <td>${escapeHTML(b.workspace_name)}</td>
                    <td>${escapeHTML(b.room_name)}</td>
                    <td>${b.user_id}</td>
                    <td>${formatDateRange(b.start_time, b.end_time)}</td>
                    <td><span class="status-badge status-${b.status}">${statusLabel(b.status)}</span></td>
                    <td>${b.status === 'active'
                        ? `<button class="btn danger cancel-book-btn" data-id="${b.id}">Отменить</button>`
                        : '-'}</td>
                `;
                tbody.appendChild(tr);
            });

            tbody.querySelectorAll('.cancel-book-btn').forEach(btn => {
                btn.addEventListener('click', async (ev) => {
                    if (confirm('Отменить бронирование?')) {
                        await api.cancelBooking(ev.target.dataset.id);
                        loadBookings();
                    }
                });
            });

            renderPagination(total, bookingsPage);
        } catch (err) {
            tbody.innerHTML = `<tr><td colspan="7" class="error-text">Ошибка: ${escapeHTML(err.message)}</td></tr>`;
        }
    }

    function renderPagination(total, currentPage) {
        const container = document.getElementById('bookingsPagination');
        container.innerHTML = '';
        const pages = Math.ceil(total / 20);
        if (pages <= 1) return;
        for (let i = 1; i <= pages; i++) {
            const btn = document.createElement('button');
            btn.className = `btn ${i === currentPage ? 'primary' : 'secondary'} pagination-btn`;
            btn.textContent = i;
            btn.addEventListener('click', () => {
                bookingsPage = i;
                loadBookings();
            });
            container.appendChild(btn);
        }
    }

    function statusLabel(status) {
        const labels = { active: 'Активно', canceled: 'Отменено', completed: 'Завершено' };
        return labels[status] || status;
    }

    // ── Stats Tab ──────────────────────────────────────────────────────────
    const statsFromEl = document.getElementById('statsFrom');
    const statsToEl = document.getElementById('statsTo');

    // Set default date range: last 30 days
    const now = new Date();
    const thirtyDaysAgo = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000);
    statsFromEl.value = thirtyDaysAgo.toISOString().split('T')[0];
    statsToEl.value = now.toISOString().split('T')[0];

    document.getElementById('applyStatsFilter').addEventListener('click', () => loadStats());

    async function loadStats() {
        try {
            const stats = await api.getStats(statsFromEl.value, statsToEl.value);
            document.getElementById('statTotalBookings').textContent = stats.total_bookings;
            document.getElementById('statActiveBookings').textContent = stats.active_bookings;
            renderBarChart('roomLoadChart', (stats.room_load || []).map(r => ({ label: r.room_name, value: r.count })));
            renderBarChart('topWorkspacesChart', (stats.top_workspaces || []).map(w => ({ label: `${w.workspace_name} (${w.room_name})`, value: w.count })));
            renderBarChart('dayDistChart', (stats.day_distribution || []).map(d => ({ label: dayName(d.day), value: d.count })));
        } catch (err) {
            console.error('Stats error', err);
        }
    }

    function renderBarChart(containerId, data) {
        const container = document.getElementById(containerId);
        container.innerHTML = '';
        if (!data.length) {
            container.innerHTML = '<p class="hint-text">Нет данных</p>';
            return;
        }
        const maxVal = Math.max(...data.map(d => d.value), 1);
        data.forEach(item => {
            const row = document.createElement('div');
            row.className = 'bar-row';
            const pct = computeOccupancyPercent(item.value, maxVal);
            row.innerHTML = `
                <span class="bar-label">${escapeHTML(item.label)}</span>
                <div class="bar-track">
                    <div class="bar-fill" style="width:${pct}%"></div>
                </div>
                <span class="bar-value">${item.value}</span>
            `;
            container.appendChild(row);
        });
    }

    // ── Tab data loader ────────────────────────────────────────────────────
    async function loadTabData(targetId) {
        if (targetId === 'roomsTab') loadRooms();
        if (targetId === 'schemaTab') {
            await loadRooms();
            addWsBtn.disabled = !schemaRoomSelect.value;
            if (schemaRoomSelect.value) editor.loadWorkspaces(schemaRoomSelect.value);
            else editor.draw();
        }
        if (targetId === 'bookingsTab') loadBookings();
        if (targetId === 'statsTab') loadStats();
    }

    checkAuth();
});

// ── Helpers ────────────────────────────────────────────────────────────────
function escapeHTML(str) {
    const div = document.createElement('div');
    div.textContent = str;
    return div.innerHTML;
}

function escapeAttr(str) {
    return str.replace(/&/g, '&amp;').replace(/"/g, '&quot;').replace(/'/g, '&#39;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
}
