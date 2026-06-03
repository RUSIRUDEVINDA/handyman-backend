package review

import "github.com/gofiber/fiber/v2"

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router fiber.Router, auth fiber.Handler) {
	routes := router.Group("/reviews")
	routes.Get("/handymen/:handyman_id", h.listForHandyman)
	routes.Post("/", auth, h.create)
}

func (h *Handler) create(c *fiber.Ctx) error {
	var req CreateReviewRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	review, err := h.service.CreateReview(c.UserContext(), c.Locals("user_id").(string), req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(review)
}

func (h *Handler) listForHandyman(c *fiber.Ctx) error {
	reviews, err := h.service.ListForHandyman(c.UserContext(), c.Params("handyman_id"))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list reviews"})
	}
	return c.JSON(reviews)
}
