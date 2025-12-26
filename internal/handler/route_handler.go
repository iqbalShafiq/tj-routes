package handler

import (
	"net/http"
	"strconv"

	"tj-routes/internal/models"
	"tj-routes/internal/service"

	"github.com/gin-gonic/gin"
)

type RouteHandler struct {
	routeService service.RouteService
}

type CreateRouteRequest struct {
	RouteNumber string   `json:"route_number" binding:"required"`
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	Status      string   `json:"status"`
	StopIDs     []uint   `json:"stop_ids" binding:"required,min=2"`
}

type UpdateRouteStopsRequest struct {
	StopIDs []uint `json:"stop_ids" binding:"required,min=2"`
}

func NewRouteHandler(routeService service.RouteService) *RouteHandler {
	return &RouteHandler{
		routeService: routeService,
	}
}

func (h *RouteHandler) ListRoutes(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	filters := make(map[string]interface{})
	if status := c.Query("status"); status != "" {
		filters["status"] = models.Status(status)
	}
	if routeNumber := c.Query("route_number"); routeNumber != "" {
		filters["route_number"] = routeNumber
	}
	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}

	routes, total, err := h.routeService.ListRoutes(offset, limit, filters)
	if err != nil {
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"routes": routes,
		"total":  total,
		"page":   page,
		"limit":  limit,
	})
}

func (h *RouteHandler) GetRoute(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	// Check if client wants statistics
	includeStats := c.Query("stats") == "true"

	if includeStats {
		response, err := h.routeService.GetRouteWithStats(uint(id))
		if err != nil {
			NotFound(c, err)
			return
		}
		SuccessResponse(c, http.StatusOK, response)
		return
	}

	route, err := h.routeService.GetRouteByID(uint(id))
	if err != nil {
		NotFound(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, route)
}

func (h *RouteHandler) CreateRoute(c *gin.Context) {
	var req CreateRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err)
		return
	}

	userID, _ := c.Get("user_id")
	route := &models.Route{
		RouteNumber: req.RouteNumber,
		Name:        req.Name,
		Description: req.Description,
		Status:      models.Status(req.Status),
	}
	if route.Status == "" {
		route.Status = models.StatusActive
	}

	if err := h.routeService.CreateRoute(route, req.StopIDs, userID.(uint)); err != nil {
		BadRequest(c, err)
		return
	}

	SuccessResponse(c, http.StatusCreated, route)
}

func (h *RouteHandler) UpdateRoute(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	var route models.Route
	if err := c.ShouldBindJSON(&route); err != nil {
		BadRequest(c, err)
		return
	}

	route.ID = uint(id)
	if err := h.routeService.UpdateRoute(&route); err != nil {
		BadRequest(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, route)
}

func (h *RouteHandler) UpdateRouteStops(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	var req UpdateRouteStopsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err)
		return
	}

	userID, _ := c.Get("user_id")
	if err := h.routeService.UpdateRouteStops(uint(id), req.StopIDs, userID.(uint)); err != nil {
		BadRequest(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Route stops updated successfully")
}

func (h *RouteHandler) DeleteRoute(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, err)
		return
	}

	if err := h.routeService.DeleteRoute(uint(id)); err != nil {
		InternalServerError(c, err)
		return
	}

	MessageResponse(c, http.StatusOK, "Route deleted successfully")
}

