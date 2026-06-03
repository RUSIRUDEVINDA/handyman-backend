package customer

import "github.com/gofiber/fiber/v2"

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router fiber.Router, auth fiber.Handler) {
	routes := router.Group("/customers", auth)
	routes.Get("/me", h.getMe)
	routes.Put("/me", h.upsertMe)
}

func (h *Handler) getMe(c *fiber.Ctx) error {
	profile, err := h.service.GetProfile(c.UserContext(), c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "customer profile not found"})
	}
	return c.JSON(profile)
}

func (h *Handler) upsertMe(c *fiber.Ctx) error {
	var req UpsertCustomerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	profile, err := h.service.UpsertProfile(c.UserContext(), c.Locals("user_id").(string), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to save customer profile"})
	}
	return c.JSON(profile)
}
