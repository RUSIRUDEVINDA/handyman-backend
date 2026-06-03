package payment

import (
	"encoding/json"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/webhook"
)

type Handler struct {
	service       Service
	webhookSecret string
}

func NewHandler(service Service, webhookSecret string) *Handler {
	return &Handler{service: service, webhookSecret: webhookSecret}
}

func (h *Handler) RegisterRoutes(router fiber.Router, auth fiber.Handler) {
	routes := router.Group("/payments")
	routes.Post("/intents", auth, h.createIntent)
	routes.Get("/bookings/:booking_id", auth, h.getByBooking)
	routes.Post("/webhook", h.webhook)
}

func (h *Handler) createIntent(c *fiber.Ctx) error {
	var req CreateIntentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	intent, err := h.service.CreatePaymentIntent(c.UserContext(), c.Locals("user_id").(string), req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(intent)
}

func (h *Handler) getByBooking(c *fiber.Ctx) error {
	payment, err := h.service.GetByBookingID(c.UserContext(), c.Params("booking_id"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "payment not found"})
	}
	return c.JSON(payment)
}

func (h *Handler) webhook(c *fiber.Ctx) error {
	payload := c.Body()
	signature := c.Get("Stripe-Signature")

	event, err := webhook.ConstructEvent(payload, signature, h.webhookSecret)
	if err != nil {
		log.Printf("stripe webhook verification failed: %v", err)
		return c.SendStatus(fiber.StatusBadRequest)
	}

	switch event.Type {
	case "payment_intent.succeeded":
		var intent stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &intent); err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}
		if err := h.service.MarkPaymentSucceeded(c.UserContext(), intent.ID, intent.Amount); err != nil {
			log.Printf("failed to update payment status: %v", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
	default:
		log.Printf("unhandled stripe event type: %s", event.Type)
	}

	return c.SendStatus(fiber.StatusOK)
}
