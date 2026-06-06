CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL CHECK (role IN ('CUSTOMER', 'HANDYMAN', 'ADMIN')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS customers (
    id UUID PRIMARY KEY,
    user_id UUID UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    full_name VARCHAR(255) NOT NULL,
    phone VARCHAR(50) NOT NULL DEFAULT '',
    address TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS handymen (
    id UUID PRIMARY KEY,
    user_id UUID UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    full_name VARCHAR(255) NOT NULL,
    phone VARCHAR(50) NOT NULL DEFAULT '',
    bio TEXT NOT NULL DEFAULT '',
    skills TEXT[] NOT NULL DEFAULT '{}',
    hourly_rate INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(50) NOT NULL DEFAULT 'OFFLINE' CHECK (status IN ('AVAILABLE', 'BUSY', 'OFFLINE')),
    rating NUMERIC(3, 2) NOT NULL DEFAULT 0,
    review_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS handyman_locations (
    handyman_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    geom GEOMETRY(Point, 4326) NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS handyman_locations_geom_idx
ON handyman_locations
USING GIST (geom);

CREATE TABLE IF NOT EXISTS bookings (
    id UUID PRIMARY KEY,
    customer_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    handyman_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    service_name VARCHAR(255) NOT NULL,
    notes TEXT NOT NULL DEFAULT '',
    status VARCHAR(50) NOT NULL CHECK (
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
    ),
    amount_cents BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS bookings_customer_id_idx ON bookings(customer_id);
CREATE INDEX IF NOT EXISTS bookings_handyman_id_idx ON bookings(handyman_id);
CREATE INDEX IF NOT EXISTS bookings_status_idx ON bookings(status);

CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY,
    booking_id UUID NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    customer_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    method VARCHAR(50) NOT NULL CHECK (method IN ('PAYHERE', 'CASH')),
    provider VARCHAR(50) NOT NULL CHECK (provider IN ('PAYHERE', 'CASH')),
    provider_payment_id VARCHAR(255) NULL,
    provider_order_id VARCHAR(255) UNIQUE NULL,
    amount_cents BIGINT NOT NULL CHECK (amount_cents > 0),
    currency VARCHAR(10) NOT NULL DEFAULT 'LKR',
    status VARCHAR(50) NOT NULL CHECK (
        status IN (
            'PENDING',
            'SUCCEEDED',
            'CANCELED',
            'FAILED',
            'CHARGEDBACK'
        )
    ),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS payments_booking_id_idx ON payments(booking_id);
CREATE INDEX IF NOT EXISTS payments_customer_id_idx ON payments(customer_id);
CREATE INDEX IF NOT EXISTS payments_provider_payment_id_idx ON payments(provider_payment_id);
CREATE INDEX IF NOT EXISTS payments_status_idx ON payments(status);

CREATE TABLE IF NOT EXISTS reviews (
    id UUID PRIMARY KEY,
    booking_id UUID NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    customer_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    handyman_id UUID NOT NULL REFERENCES handymen(id) ON DELETE CASCADE,
    rating INTEGER NOT NULL CHECK (rating BETWEEN 1 AND 5),
    comment TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (booking_id, customer_id)
);

CREATE INDEX IF NOT EXISTS reviews_handyman_id_idx ON reviews(handyman_id);
