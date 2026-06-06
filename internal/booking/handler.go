package booking

import "github.com/gofiber/fiber/v2"

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
	routes.Put("/:id/approve", h.approve)
	routes.Put("/:id/reject", h.reject)
	routes.Put("/:id/start", h.start)
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

func (h *Handler) approve(c *fiber.Ctx) error {
	booking, err := h.service.Approve(c.UserContext(), c.Params("id"), c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(booking)
}

func (h *Handler) reject(c *fiber.Ctx) error {
	booking, err := h.service.Reject(c.UserContext(), c.Params("id"), c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(booking)
}

func (h *Handler) start(c *fiber.Ctx) error {
	booking, err := h.service.Start(c.UserContext(), c.Params("id"), c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(booking)
}

func (h *Handler) complete(c *fiber.Ctx) error {
	booking, err := h.service.Complete(c.UserContext(), c.Params("id"), c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(booking)
}

func (h *Handler) cancel(c *fiber.Ctx) error {
	booking, err := h.service.Cancel(c.UserContext(), c.Params("id"), c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(booking)
}
