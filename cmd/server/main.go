package main

import (
	"context"
	"errors"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/RUSIRUDEVINDA/handyman-backend/configs"
	"github.com/RUSIRUDEVINDA/handyman-backend/internal/auth"
	"github.com/RUSIRUDEVINDA/handyman-backend/internal/booking"
	"github.com/RUSIRUDEVINDA/handyman-backend/internal/customer"
	"github.com/RUSIRUDEVINDA/handyman-backend/internal/dashboard"
	"github.com/RUSIRUDEVINDA/handyman-backend/internal/events"
	"github.com/RUSIRUDEVINDA/handyman-backend/internal/handyman"
	"github.com/RUSIRUDEVINDA/handyman-backend/internal/location"
	"github.com/RUSIRUDEVINDA/handyman-backend/internal/notification"
	"github.com/RUSIRUDEVINDA/handyman-backend/internal/payment"
	"github.com/RUSIRUDEVINDA/handyman-backend/internal/review"
	realtime "github.com/RUSIRUDEVINDA/handyman-backend/internal/websocket"
	"github.com/RUSIRUDEVINDA/handyman-backend/pkg/database"
	appmiddleware "github.com/RUSIRUDEVINDA/handyman-backend/pkg/middleware"
	"github.com/gofiber/fiber/v2"
)

func main() {
	cfg, err := configs.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	eventBus := events.NewBus()
	defer eventBus.Close()

	app := fiber.New(fiber.Config{
		AppName:      "Handyman API",
		ServerHeader: "handyman-api",
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	api := app.Group("/api/v1")
	authGuard := appmiddleware.RequireAuth(cfg.JWTSecret)
	adminGuard := appmiddleware.RequireRole("ADMIN")

	authRepo := auth.NewRepository(db)
	authService := auth.NewService(authRepo, cfg.JWTSecret)
	auth.NewHandler(authService).RegisterRoutes(api)

	customerRepo := customer.NewRepository(db)
	customerService := customer.NewService(customerRepo)
	customer.NewHandler(customerService).RegisterRoutes(api, authGuard)

	handymanRepo := handyman.NewRepository(db)
	handymanService := handyman.NewService(handymanRepo, eventBus)
	handyman.NewHandler(handymanService).RegisterRoutes(api, authGuard)

	bookingRepo := booking.NewRepository(db)
	bookingService := booking.NewService(bookingRepo, eventBus)
	booking.NewHandler(bookingService).RegisterRoutes(api, authGuard)

	locationRepo := location.NewRepository(db)
	locationService := location.NewService(locationRepo, eventBus)
	location.NewHandler(locationService).RegisterRoutes(api, authGuard)

	paymentRepo := payment.NewRepository(db)
	paymentService := payment.NewService(paymentRepo, bookingRepo, eventBus, payment.PayHereConfig{
		MerchantID:     cfg.PayHereMerchantID,
		MerchantSecret: cfg.PayHereMerchantSecret,
		CheckoutURL:    cfg.PayHereCheckoutURL,
		ReturnURL:      cfg.PayHereReturnURL,
		CancelURL:      cfg.PayHereCancelURL,
		NotifyURL:      cfg.AppBaseURL + "/api/v1/payments/payhere/notify",
	})
	payment.NewHandler(paymentService).RegisterRoutes(api, authGuard)

	reviewRepo := review.NewRepository(db)
	reviewService := review.NewService(reviewRepo, bookingRepo, handymanRepo, eventBus)
	review.NewHandler(reviewService).RegisterRoutes(api, authGuard)

	dashboardRepo := dashboard.NewRepository(db)
	dashboardService := dashboard.NewService(dashboardRepo)
	dashboard.NewHandler(dashboardService).RegisterRoutes(api, authGuard, adminGuard)

	notificationWorker := notification.NewWorker(eventBus, notification.NewService())
	go notificationWorker.Start(ctx)

	hub := realtime.NewHub(eventBus)
	hub.Start(ctx)
	realtime.NewHandler(hub).RegisterRoutes(api)

	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("server listening on :%s", cfg.Port)
		serverErrors <- app.Listen(":" + cfg.Port)
	}()

	select {
	case <-ctx.Done():
		log.Println("shutdown requested")
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, fiber.ErrServiceUnavailable) {
			log.Fatalf("server failed: %v", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Printf("server shutdown failed: %v", err)
	}
}
