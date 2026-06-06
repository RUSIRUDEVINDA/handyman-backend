package dashboard

import "github.com/gofiber/fiber/v2"

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router fiber.Router, auth fiber.Handler, adminOnly fiber.Handler) {
	routes := router.Group("/dashboards", auth)
	routes.Get("/customer", h.customer)
	routes.Get("/handyman", h.handyman)
	routes.Get("/admin", adminOnly, h.admin)
}

func (h *Handler) customer(c *fiber.Ctx) error {
	dashboard, err := h.service.Customer(c.UserContext(), c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to load customer dashboard"})
	}
	return c.JSON(dashboard)
}

func (h *Handler) handyman(c *fiber.Ctx) error {
	dashboard, err := h.service.Handyman(c.UserContext(), c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to load handyman dashboard"})
	}
	return c.JSON(dashboard)
}

func (h *Handler) admin(c *fiber.Ctx) error {
	dashboard, err := h.service.Admin(c.UserContext())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to load admin dashboard"})
	}
	return c.JSON(dashboard)
}
