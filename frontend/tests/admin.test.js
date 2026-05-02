import { jest } from '@jest/globals';
import {
    calculateStats,
    computeOccupancyPercent,
    dayName,
    formatDateRange,
    aggregateByRoom,
    AdminAPI
} from '../admin.js';

// ── calculateStats ─────────────────────────────────────────────────────────
describe('calculateStats', () => {
    test('calculates correct totals for mixed statuses', () => {
        const bookings = [
            { id: 1, status: 'active' },
            { id: 2, status: 'confirmed' },
            { id: 3, status: 'cancelled' },
            { id: 4, status: 'completed' }
        ];
        const stats = calculateStats(bookings);
        expect(stats.total).toBe(4);
        expect(stats.active).toBe(2);
    });

    test('handles empty bookings', () => {
        const stats = calculateStats([]);
        expect(stats.total).toBe(0);
        expect(stats.active).toBe(0);
    });

    test('counts only active and confirmed as active', () => {
        const bookings = [
            { id: 1, status: 'canceled' },
            { id: 2, status: 'canceled' },
            { id: 3, status: 'completed' }
        ];
        const stats = calculateStats(bookings);
        expect(stats.total).toBe(3);
        expect(stats.active).toBe(0);
    });

    test('all active bookings', () => {
        const bookings = [
            { id: 1, status: 'active' },
            { id: 2, status: 'active' }
        ];
        const stats = calculateStats(bookings);
        expect(stats.total).toBe(2);
        expect(stats.active).toBe(2);
    });
});

// ── computeOccupancyPercent ────────────────────────────────────────────────
describe('computeOccupancyPercent', () => {
    test('returns correct percentage', () => {
        expect(computeOccupancyPercent(50, 100)).toBe(50);
        expect(computeOccupancyPercent(1, 3)).toBe(33);
        expect(computeOccupancyPercent(3, 3)).toBe(100);
    });

    test('returns 0 when maxCount is 0 or negative', () => {
        expect(computeOccupancyPercent(5, 0)).toBe(0);
        expect(computeOccupancyPercent(5, -1)).toBe(0);
    });

    test('handles zero count', () => {
        expect(computeOccupancyPercent(0, 100)).toBe(0);
    });
});

// ── dayName ────────────────────────────────────────────────────────────────
describe('dayName', () => {
    test('returns correct Russian day abbreviations', () => {
        expect(dayName(0)).toBe('Вс');
        expect(dayName(1)).toBe('Пн');
        expect(dayName(5)).toBe('Пт');
        expect(dayName(6)).toBe('Сб');
    });

    test('returns ? for invalid index', () => {
        expect(dayName(7)).toBe('?');
        expect(dayName(-1)).toBe('?');
    });
});

// ── formatDateRange ────────────────────────────────────────────────────────
describe('formatDateRange', () => {
    test('formats date range correctly', () => {
        const result = formatDateRange('2026-05-01T10:00:00Z', '2026-05-01T12:00:00Z');
        expect(result).toContain('2026');
        expect(result).toContain('\u2013');
    });

    test('handles same day range', () => {
        const result = formatDateRange('2026-01-15T08:00:00Z', '2026-01-15T10:00:00Z');
        expect(typeof result).toBe('string');
        expect(result.length).toBeGreaterThan(5);
    });
});

// ── aggregateByRoom ────────────────────────────────────────────────────────
describe('aggregateByRoom', () => {
    test('aggregates bookings by room_name', () => {
        const bookings = [
            { room_name: 'Room A', workspace_id: 1 },
            { room_name: 'Room A', workspace_id: 2 },
            { room_name: 'Room B', workspace_id: 3 }
        ];
        const result = aggregateByRoom(bookings);
        expect(result).toHaveLength(2);
        expect(result[0].name).toBe('Room A');
        expect(result[0].count).toBe(2);
        expect(result[1].name).toBe('Room B');
        expect(result[1].count).toBe(1);
    });

    test('handles empty array', () => {
        expect(aggregateByRoom([])).toEqual([]);
    });

    test('sorts by count descending', () => {
        const bookings = [
            { room_name: 'X', workspace_id: 1 },
            { room_name: 'Y', workspace_id: 2 },
            { room_name: 'Y', workspace_id: 3 },
            { room_name: 'Y', workspace_id: 4 }
        ];
        const result = aggregateByRoom(bookings);
        expect(result[0].name).toBe('Y');
        expect(result[0].count).toBe(3);
    });
});

// ── AdminAPI ───────────────────────────────────────────────────────────────
describe('AdminAPI', () => {
    let api;

    beforeEach(() => {
        api = new AdminAPI();
        localStorage.clear();
        global.fetch = jest.fn();
    });

    afterEach(() => {
        jest.restoreAllMocks();
    });

    test('token getter reads from localStorage', () => {
        expect(api.token).toBeNull();
        localStorage.setItem('adminToken', 'test-token');
        expect(api.token).toBe('test-token');
    });

    test('login saves token', async () => {
        global.fetch = jest.fn().mockResolvedValue({
            ok: true,
            status: 200,
            json: () => Promise.resolve({ access_token: 'jwt-123' })
        });

        await api.login('admin@test.com', 'pass');
        expect(localStorage.getItem('adminToken')).toBe('jwt-123');
    });

    test('request throws on 401 and clears token', async () => {
        localStorage.setItem('adminToken', 'old');
        // Mock location.reload to prevent actual reload
        delete window.location;
        window.location = { reload: jest.fn() };

        global.fetch = jest.fn().mockResolvedValue({
            ok: false,
            status: 401,
            json: () => Promise.resolve({ error: 'unauthorized' })
        });

        await expect(api.request('/test')).rejects.toThrow('Unauthorized');
        expect(localStorage.getItem('adminToken')).toBeNull();
    });

    test('request includes auth header when token present', async () => {
        localStorage.setItem('adminToken', 'my-jwt');
        global.fetch = jest.fn().mockResolvedValue({
            ok: true,
            status: 200,
            json: () => Promise.resolve({ data: 'ok' })
        });

        await api.request('/rooms');
        const callHeaders = global.fetch.mock.calls[0][1].headers;
        expect(callHeaders['Authorization']).toBe('Bearer my-jwt');
    });
});
