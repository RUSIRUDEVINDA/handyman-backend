-- Development migration for projects that already created the earlier Stripe-shaped payments table.
-- Review existing data before running this. It preserves rows where possible, but old Stripe IDs become provider_payment_id.

ALTER TABLE payments DROP CONSTRAINT IF EXISTS payments_status_check;
ALTER TABLE payments DROP CONSTRAINT IF EXISTS payments_stripe_payment_intent_id_key;

ALTER TABLE payments
    RENAME COLUMN stripe_payment_intent_id TO provider_payment_id;

ALTER TABLE payments
    ADD COLUMN IF NOT EXISTS method VARCHAR(50),
    ADD COLUMN IF NOT EXISTS provider VARCHAR(50),
    ADD COLUMN IF NOT EXISTS provider_order_id VARCHAR(255);

UPDATE payments
SET
    method = COALESCE(method, 'PAYHERE'),
    provider = COALESCE(provider, 'PAYHERE'),
    provider_order_id = COALESCE(provider_order_id, id::text),
    currency = UPPER(currency),
    status = CASE status
        WHEN 'REQUIRES_PAYMENT_METHOD' THEN 'PENDING'
        WHEN 'REQUIRES_CONFIRMATION' THEN 'PENDING'
        WHEN 'REQUIRES_ACTION' THEN 'PENDING'
        WHEN 'PROCESSING' THEN 'PENDING'
        WHEN 'SUCCEEDED' THEN 'SUCCEEDED'
        WHEN 'CANCELED' THEN 'CANCELED'
        ELSE 'FAILED'
    END;

ALTER TABLE payments
    ALTER COLUMN method SET NOT NULL,
    ALTER COLUMN provider SET NOT NULL,
    ALTER COLUMN provider_payment_id DROP NOT NULL,
    ALTER COLUMN currency SET DEFAULT 'LKR';

CREATE UNIQUE INDEX IF NOT EXISTS payments_provider_order_id_key ON payments(provider_order_id);
CREATE INDEX IF NOT EXISTS payments_provider_payment_id_idx ON payments(provider_payment_id);

ALTER TABLE payments
    ADD CONSTRAINT payments_method_check CHECK (method IN ('PAYHERE', 'CASH')),
    ADD CONSTRAINT payments_provider_check CHECK (provider IN ('PAYHERE', 'CASH')),
    ADD CONSTRAINT payments_status_check CHECK (status IN ('PENDING', 'SUCCEEDED', 'CANCELED', 'FAILED', 'CHARGEDBACK'));
