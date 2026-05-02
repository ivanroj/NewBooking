import { calculateStats } from '../admin.js';

describe('Admin Stats Calculation', () => {
    test('calculates correct totals', () => {
        const mockBookings = [
            { id: 1, status: 'active' },
            { id: 2, status: 'confirmed' },
            { id: 3, status: 'cancelled' },
            { id: 4, status: 'completed' }
        ];

        const stats = calculateStats(mockBookings, []);

        expect(stats.total).toBe(4);
        expect(stats.active).toBe(2);
    });

    test('handles empty bookings', () => {
        const stats = calculateStats([], []);
        expect(stats.total).toBe(0);
        expect(stats.active).toBe(0);
    });
});
