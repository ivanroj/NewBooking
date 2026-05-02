/**
 * @jest-environment jsdom
 */

import { jest } from '@jest/globals';
import { API, switchView } from '../app.js';

describe('Frontend App Logic', () => {
    beforeEach(() => {
        document.body.innerHTML = `
            <nav id="navBar">
                <button class="nav-btn active" data-target="roomsView">Помещения</button>
                <button class="nav-btn" data-target="historyView">Мои Брони</button>
            </nav>
            <main>
                <section id="loadingView" class="view active"></section>
                <section id="roomsView" class="view"></section>
                <section id="workspacesView" class="view"></section>
                <section id="historyView" class="view"></section>
            </main>
        `;
        // Setup mock for global fetch
        global.fetch = jest.fn();
    });

    test('switchView changes active class correctly', () => {
        // Initially loadingView is active
        expect(document.getElementById('loadingView').classList.contains('active')).toBe(true);
        expect(document.getElementById('roomsView').classList.contains('active')).toBe(false);

        // Switch to rooms
        switchView('rooms');

        expect(document.getElementById('loadingView').classList.contains('active')).toBe(false);
        expect(document.getElementById('roomsView').classList.contains('active')).toBe(true);
        expect(document.querySelector('.nav-btn[data-target="roomsView"]').classList.contains('active')).toBe(true);
    });

    test('API wrapper sets token and constructs headers correctly', async () => {
        const api = new API();
        api.setToken('test-jwt-token');

        global.fetch.mockResolvedValueOnce({
            ok: true,
            json: async () => ({ id: 1, name: 'Room 1' })
        });

        const res = await api.request('/test');

        expect(global.fetch).toHaveBeenCalledWith('/api/test', expect.objectContaining({
            headers: expect.objectContaining({
                'Content-Type': 'application/json',
                'Authorization': 'Bearer test-jwt-token'
            })
        }));
        expect(res.id).toBe(1);
    });

    test('API wrapper throws parsed error on 4xx', async () => {
        const api = new API();
        global.fetch.mockResolvedValueOnce({
            ok: false,
            status: 409,
            json: async () => ({ error: 'conflict' })
        });

        await expect(api.request('/bookings')).rejects.toMatchObject({
            status: 409,
            error: 'conflict'
        });
    });
});
