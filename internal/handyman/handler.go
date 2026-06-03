package handyman

import "github.com/gofiber/fiber/v2"

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router fiber.Router, auth fiber.Handler) {
	public := router.Group("/handymen")
	public.Get("/", h.search)

	private := router.Group("/handymen", auth)
	private.Get("/me", h.getMe)
	private.Put("/me", h.upsertMe)
}

func (h *Handler) search(c *fiber.Ctx) error {
	handymen, err := h.service.Search(c.UserContext(), c.Query("skill"))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to search handymen"})
	}
	return c.JSON(handymen)
}

func (h *Handler) getMe(c *fiber.Ctx) error {
	profile, err := h.service.GetProfile(c.UserContext(), c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "handyman profile not found"})
	}
	return c.JSON(profile)
}

func (h *Handler) upsertMe(c *fiber.Ctx) error {
	var req UpsertHandymanRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	profile, err := h.service.UpsertProfile(c.UserContext(), c.Locals("user_id").(string), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to save handyman profile"})
	}
	return c.JSON(profile)
}
