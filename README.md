# handyman-backend

A modular Go backend for a handyman marketplace. It uses Fiber for HTTP, Supabase/Postgres with PostGIS for data and location search, PayHere sandbox plus cash payments, JWT auth, WebSockets for realtime updates, and an internal event bus for asynchronous module communication.

## Folder Structure

```text
cmd/server              Application entrypoint and dependency wiring
configs                 Environment variable loading
db/schema.sql           Supabase/Postgres schema, including PostGIS
internal/events         In-memory domain event bus
internal/auth           Registration, login, JWT creation
internal/customer       Customer profile APIs
internal/dashboard      Customer, handyman, and admin summary dashboards
internal/handyman       Handyman profile, skills, rating, availability
internal/booking        Booking state machine and domain events
internal/location       PostGIS location updates and nearby search
internal/payment        PayHere sandbox checkout, PayHere notify/IPN, and cash payments
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
4. Add PayHere sandbox keys if you want hosted checkout testing.
5. Install dependencies and run:

```bash
go mod tidy
go run ./cmd/server
```

If you already created the earlier Stripe-shaped `payments` table, run `db/migrate_payments_to_payhere_cash.sql` once in Supabase before using the new PayHere/cash endpoints.

If you already created the earlier booking table, run `db/migrate_booking_approval_flow.sql` once so bookings support the approval-before-payment flow.

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
6. Auth / login handyman
7. Bookings / PUT /bookings/:id/approve
8. Auth / login customer again
9. Payments / POST /payments/payhere/checkout or POST /payments/cash
10. Payments / GET /payments/bookings/:booking_id
11. Bookings / PUT /bookings/:id/start
12. Bookings / PUT /bookings/:id/complete
13. Reviews / POST /reviews
14. Dashboards / GET /dashboards/customer
15. Dashboards / GET /dashboards/handyman
16. Dashboards / GET /dashboards/admin
```

The admin dashboard requires an `ADMIN` JWT. For development, create a normal user and update that user's role to `ADMIN` in Supabase, then login and use that token.

The login request automatically saves `token` as a collection variable. The create booking request automatically saves `booking_id` and `payment_booking_id`.

For handyman-specific checks, register/login the handyman user first, then call `PUT /api/v1/handymen/me` and `PUT /api/v1/locations/me`.

PayHere notify/IPN cannot be tested from localhost by PayHere directly. PayHere's documentation requires the `notify_url` to be publicly accessible, and notifications are sent as `application/x-www-form-urlencoded`.

For local testing, use a tunnel such as ngrok and set `APP_BASE_URL` to the public tunnel URL before starting the server.

## Main Routes

```text
GET    /health
POST   /api/v1/auth/register
POST   /api/v1/auth/login
GET    /api/v1/customers/me
PUT    /api/v1/customers/me
GET    /api/v1/dashboards/customer
GET    /api/v1/dashboards/handyman
GET    /api/v1/dashboards/admin
GET    /api/v1/handymen?skill=plumbing
GET    /api/v1/handymen/me
PUT    /api/v1/handymen/me
POST   /api/v1/bookings
GET    /api/v1/bookings
GET    /api/v1/bookings/:id
PUT    /api/v1/bookings/:id/assign
PUT    /api/v1/bookings/:id/approve
PUT    /api/v1/bookings/:id/reject
PUT    /api/v1/bookings/:id/start
PUT    /api/v1/bookings/:id/complete
PUT    /api/v1/bookings/:id/cancel
PUT    /api/v1/locations/me
GET    /api/v1/locations/nearby?lat=6.9271&lng=79.8612&radius_km=10
POST   /api/v1/payments/payhere/checkout
POST   /api/v1/payments/cash
PUT    /api/v1/payments/cash/:payment_id/collected
GET    /api/v1/payments/bookings/:booking_id
POST   /api/v1/payments/payhere/notify
POST   /api/v1/reviews
GET    /api/v1/reviews/handymen/:handyman_id
GET    /api/v1/ws
```
