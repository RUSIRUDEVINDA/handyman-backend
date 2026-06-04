package payment

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router fiber.Router, auth fiber.Handler) {
	routes := router.Group("/payments")
	routes.Post("/payhere/checkout", auth, h.createPayHereCheckout)
	routes.Post("/cash", auth, h.createCashPayment)
	routes.Put("/cash/:payment_id/collected", auth, h.markCashCollected)
	routes.Get("/bookings/:booking_id", auth, h.getByBooking)
	routes.Post("/payhere/notify", h.payHereNotify)
}

func (h *Handler) createPayHereCheckout(c *fiber.Ctx) error {
	var req CreatePayHereCheckoutRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	checkout, err := h.service.CreatePayHereCheckout(c.UserContext(), c.Locals("user_id").(string), req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(checkout)
}

func (h *Handler) createCashPayment(c *fiber.Ctx) error {
	var req CreateCashPaymentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	payment, err := h.service.CreateCashPayment(c.UserContext(), c.Locals("user_id").(string), req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(payment)
}

func (h *Handler) markCashCollected(c *fiber.Ctx) error {
	if err := h.service.MarkCashCollected(c.UserContext(), c.Params("payment_id")); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to mark cash payment collected"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) getByBooking(c *fiber.Ctx) error {
	payment, err := h.service.GetByBookingID(c.UserContext(), c.Params("booking_id"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "payment not found"})
	}
	return c.JSON(payment)
}

func (h *Handler) payHereNotify(c *fiber.Ctx) error {
	notification := PayHereNotification{
		MerchantID:      c.FormValue("merchant_id"),
		OrderID:         c.FormValue("order_id"),
		PaymentID:       c.FormValue("payment_id"),
		PayHereAmount:   c.FormValue("payhere_amount"),
		PayHereCurrency: c.FormValue("payhere_currency"),
		StatusCode:      c.FormValue("status_code"),
		MD5Sig:          c.FormValue("md5sig"),
		Method:          c.FormValue("method"),
		StatusMessage:   c.FormValue("status_message"),
		CardHolderName:  c.FormValue("card_holder_name"),
		CardNo:          c.FormValue("card_no"),
		CardExpiry:      c.FormValue("card_expiry"),
	}

	if err := h.service.HandlePayHereNotification(c.UserContext(), notification); err != nil {
		log.Printf("PayHere notification rejected: %v", err)
		return c.SendStatus(fiber.StatusBadRequest)
	}

	return c.SendStatus(fiber.StatusOK)
}
