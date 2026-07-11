package http

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/user/go-laos-geo/domain"
)

type GeoHandler struct {
	geoUsecase domain.GeoUsecase
}

func NewGeoHandler(app *fiber.App, us domain.GeoUsecase) {
	handler := &GeoHandler{
		geoUsecase: us,
	}

	api := app.Group("/api/v1")
	api.Get("/provinces", handler.GetProvinces)
	api.Get("/provinces/:id/districts", handler.GetDistrictsByProvince)
	api.Get("/districts/:id/villages", handler.GetVillagesByDistrict)
}

func getPaginationParams(c *fiber.Ctx) (int, int) {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100 // Prevent excessively large queries
	}

	return page, limit
}

func (h *GeoHandler) GetProvinces(c *fiber.Ctx) error {
	ctx := c.Context()
	page, limit := getPaginationParams(c)

	response, err := h.geoUsecase.GetProvinces(ctx, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response)
	}
	return c.JSON(response)
}

func (h *GeoHandler) GetDistrictsByProvince(c *fiber.Ctx) error {
	ctx := c.Context()
	provinceID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		errStr := "Invalid province ID"
		return c.Status(fiber.StatusBadRequest).JSON(domain.StandardResponse{
			Status: 0,
			Error:  &errStr,
		})
	}

	page, limit := getPaginationParams(c)
	response, err := h.geoUsecase.GetDistrictsByProvince(ctx, provinceID, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response)
	}
	return c.JSON(response)
}

func (h *GeoHandler) GetVillagesByDistrict(c *fiber.Ctx) error {
	ctx := c.Context()
	districtID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		errStr := "Invalid district ID"
		return c.Status(fiber.StatusBadRequest).JSON(domain.StandardResponse{
			Status: 0,
			Error:  &errStr,
		})
	}

	page, limit := getPaginationParams(c)
	response, err := h.geoUsecase.GetVillagesByDistrict(ctx, districtID, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response)
	}
	return c.JSON(response)
}
