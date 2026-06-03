# handyman-backend

A modular Go backend for a handyman marketplace. It uses Fiber for HTTP, Supabase/Postgres with PostGIS for data and location search, Stripe Payment Intents for payments, JWT auth, WebSockets for realtime updates, and an internal event bus for asynchronous module communication.

## Folder Structure

```text
cmd/server              Application entrypoint and dependency wiring
configs                 Environment variable loading
db/schema.sql           Supabase/Postgres schema, including PostGIS
internal/events         In-memory domain event bus
internal/auth           Registration, login, JWT creation
internal/customer       Customer profile APIs
internal/handyman       Handyman profile, skills, rating, availability
internal/booking        Booking state machine and domain events
internal/location       PostGIS location updates and nearby search
internal/payment        Stripe Payment Intent and webhook endpoints
internal/notification   Background worker that reacts to events
internal/review         Reviews and rating aggregation
internal/websocket      Realtime event broadcasting
pkg/database            Shared Postgres connection pool
pkg/logger              Small logging helpers
pkg/middleware          JWT middleware
```

## Goroutines

Yes, this project uses goroutines in several places:

- Fiber handles incoming HTTP requests concurrently.
- `cmd/server/main.go` starts the HTTP server in a goroutine.
- `internal/notification/worker.go` listens for events and spawns goroutines to process notifications.
- `internal/websocket/hub.go` starts event subscription goroutines for realtime broadcasts.
- `internal/events/bus.go` uses buffered channels for async event delivery.

## Setup

1. Copy `.env.example` to `.env`.
2. Add your Supabase `DB_URL`.
3. Run `db/schema.sql` in the Supabase SQL editor.
4. Add Stripe keys if you want payment intent creation and webhooks.
5. Install dependencies and run:

```bash
go mod tidy
go run ./cmd/server
```

## Postman Manual Testing

Import this collection into Postman:

```text
docs/postman_collection.json
```

Recommended order:

```text
1. Health / GET /health
2. Auth / register customer
3. Auth / login customer
4. Customers / PUT /customers/me
5. Bookings / POST /bookings
6. Payments / POST /payments/intents
7. Payments / GET /payments/bookings/:booking_id
```

The login request automatically saves `token` as a collection variable. The create booking request automatically saves `booking_id` and `payment_booking_id`.

For handyman-specific checks, register/login the handyman user first, then call `PUT /api/v1/handymen/me` and `PUT /api/v1/locations/me`.

Stripe webhooks cannot be tested with a normal fake Postman body because the endpoint verifies `Stripe-Signature`. Use Stripe CLI instead:

```bash
stripe listen --forward-to localhost:3000/api/v1/payments/webhook
```

## Main Routes

```text
GET    /health
POST   /api/v1/auth/register
POST   /api/v1/auth/login
GET    /api/v1/customers/me
PUT    /api/v1/customers/me
GET    /api/v1/handymen?skill=plumbing
GET    /api/v1/handymen/me
PUT    /api/v1/handymen/me
POST   /api/v1/bookings
GET    /api/v1/bookings
GET    /api/v1/bookings/:id
PUT    /api/v1/bookings/:id/assign
PUT    /api/v1/bookings/:id/confirm
PUT    /api/v1/bookings/:id/complete
PUT    /api/v1/bookings/:id/cancel
PUT    /api/v1/locations/me
GET    /api/v1/locations/nearby?lat=6.9271&lng=79.8612&radius_km=10
POST   /api/v1/payments/intents
GET    /api/v1/payments/bookings/:booking_id
POST   /api/v1/payments/webhook
POST   /api/v1/reviews
GET    /api/v1/reviews/handymen/:handyman_id
GET    /api/v1/ws
```
