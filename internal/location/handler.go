package location

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router fiber.Router, auth fiber.Handler) {
	routes := router.Group("/locations")
	routes.Get("/nearby", h.nearby)
	routes.Put("/me", auth, h.updateMine)
}

func (h *Handler) updateMine(c *fiber.Ctx) error {
	var req UpdateLocationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	location, err := h.service.UpdateLocation(c.UserContext(), c.Locals("user_id").(string), req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(location)
}

func (h *Handler) nearby(c *fiber.Ctx) error {
	lat, err := parseFloatQuery(c, "lat")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "lat query parameter is required"})
	}
	lng, err := parseFloatQuery(c, "lng")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "lng query parameter is required"})
	}
	radiusKm, _ := strconv.ParseFloat(c.Query("radius_km", "10"), 64)

	nearby, err := h.service.FindNearby(c.UserContext(), lat, lng, radiusKm)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to find nearby handymen"})
	}
	return c.JSON(nearby)
}

func parseFloatQuery(c *fiber.Ctx, name string) (float64, error) {
	return strconv.ParseFloat(c.Query(name), 64)
}
