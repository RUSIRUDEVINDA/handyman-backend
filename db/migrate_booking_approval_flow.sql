-- Development migration for projects that already created the earlier booking status check.
-- Run this once if your bookings table still uses PENDING/CONFIRMED/COMPLETED/CANCELLED only.

ALTER TABLE bookings DROP CONSTRAINT IF EXISTS bookings_status_check;

UPDATE bookings
SET status = 'REQUESTED'
WHERE status = 'PENDING';

ALTER TABLE bookings
    ADD CONSTRAINT bookings_status_check CHECK (
        status IN (
            'REQUESTED',
            'APPROVED',
            'REJECTED',
            'PAYMENT_PENDING',
            'CONFIRMED',
            'IN_PROGRESS',
            'COMPLETED',
            'CANCELLED'
        )
    );
