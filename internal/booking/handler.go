package booking

import (
	"context"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router fiber.Router, auth fiber.Handler) {
	routes := router.Group("/bookings", auth)
	routes.Post("/", h.create)
	routes.Get("/", h.listMine)
	routes.Get("/:id", h.get)
	routes.Put("/:id/assign", h.assign)
	routes.Put("/:id/confirm", h.confirm)
	routes.Put("/:id/complete", h.complete)
	routes.Put("/:id/cancel", h.cancel)
}

func (h *Handler) create(c *fiber.Ctx) error {
	var req CreateBookingRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	booking, err := h.service.CreateBooking(c.UserContext(), c.Locals("user_id").(string), req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(booking)
}

func (h *Handler) listMine(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	if c.Query("as") == "handyman" {
		bookings, err := h.service.ListHandymanBookings(c.UserContext(), userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list bookings"})
		}
		return c.JSON(bookings)
	}

	bookings, err := h.service.ListCustomerBookings(c.UserContext(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list bookings"})
	}
	return c.JSON(bookings)
}

func (h *Handler) get(c *fiber.Ctx) error {
	booking, err := h.service.GetBooking(c.UserContext(), c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "booking not found"})
	}
	return c.JSON(booking)
}

func (h *Handler) assign(c *fiber.Ctx) error {
	var req AssignHandymanRequest
	if err := c.BodyParser(&req); err != nil || req.HandymanID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "handyman_id is required"})
	}
	booking, err := h.service.AssignHandyman(c.UserContext(), c.Params("id"), req.HandymanID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to assign handyman"})
	}
	return c.JSON(booking)
}

func (h *Handler) confirm(c *fiber.Ctx) error {
	return h.transition(c, h.service.Confirm)
}

func (h *Handler) complete(c *fiber.Ctx) error {
	return h.transition(c, h.service.Complete)
}

func (h *Handler) cancel(c *fiber.Ctx) error {
	return h.transition(c, h.service.Cancel)
}

func (h *Handler) transition(c *fiber.Ctx, fn func(ctx context.Context, id string) (*Booking, error)) error {
	booking, err := fn(c.UserContext(), c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update booking"})
	}
	return c.JSON(booking)
}
