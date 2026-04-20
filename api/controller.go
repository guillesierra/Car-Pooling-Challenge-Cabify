package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gitlab-hiring.cabify.tech/cabify/interviewing/car-pooling-challenge-go/service"
	"gitlab-hiring.cabify.tech/cabify/interviewing/car-pooling-challenge-go/service/model"
)

type Controller struct {
	service *service.CarPool // Domain service.
	engine  *gin.Engine      // HTTP router.
}

// NewController builds the API controller and routes.
func NewController(service *service.CarPool) *Controller {
	c := &Controller{
		service: service,
		engine:  gin.New(),
	}
	c.engine.GET("/status", c.getStatus)
	c.engine.Any("/cars", c.putCars)
	c.engine.Any("/journey", c.postJourney)
	c.engine.Any("/dropoff", c.postDropoff)
	c.engine.Any("/locate", c.postLocate)
	return c
}

// Run starts the HTTP server.
func (c *Controller) Run() {
	_ = c.engine.Run("0.0.0.0:8080")
}

// getStatus returns service readiness.
func (c *Controller) getStatus(ctx *gin.Context) {
	ctx.Status(http.StatusOK)
}

// putCars replaces the fleet and resets state.
func (c *Controller) putCars(ctx *gin.Context) {
	if ctx.Request.Method != http.MethodPut {
		ctx.AbortWithStatus(http.StatusMethodNotAllowed)
		return
	}
	if !isJSONContentType(ctx.ContentType()) {
		ctx.Status(http.StatusBadRequest)
		return
	}

	var cars []*model.Car
	if err := ctx.ShouldBindJSON(&cars); err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}

	if err := c.service.ResetCars(cars); err != nil {
		switch err {
		case service.ErrDuplicatedID, service.ErrInvalidCars:
			ctx.Status(http.StatusBadRequest)
		default:
			ctx.Status(http.StatusInternalServerError)
		}
		return
	}
	ctx.Status(http.StatusOK)
}

// postJourney registers a journey request.
func (c *Controller) postJourney(ctx *gin.Context) {
	if ctx.Request.Method != http.MethodPost {
		ctx.AbortWithStatus(http.StatusMethodNotAllowed)
		return
	}
	if !isJSONContentType(ctx.ContentType()) {
		ctx.Status(http.StatusBadRequest)
		return
	}

	var journey model.Journey
	if err := ctx.ShouldBindJSON(&journey); err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}

	assigned, err := c.service.NewJourney(&journey)
	if err != nil {
		switch err {
		case service.ErrDuplicatedID, service.ErrInvalidJourney:
			ctx.Status(http.StatusBadRequest)
		default:
			ctx.Status(http.StatusInternalServerError)
		}
		return
	}

	if assigned {
		ctx.Status(http.StatusOK)
		return
	}
	ctx.Status(http.StatusAccepted)
}

// postDropoff removes a journey by ID.
func (c *Controller) postDropoff(ctx *gin.Context) {
	if ctx.Request.Method != http.MethodPost {
		ctx.AbortWithStatus(http.StatusMethodNotAllowed)
		return
	}
	if !isFormContentType(ctx.ContentType()) {
		ctx.Status(http.StatusBadRequest)
		return
	}

	var dropoff struct {
		ID uint `form:"ID" binding:"required"` // Journey ID.
	}
	if err := ctx.ShouldBind(&dropoff); err != nil || dropoff.ID == 0 {
		ctx.Status(http.StatusBadRequest)
		return
	}

	err := c.service.Dropoff(dropoff.ID)
	if err == service.ErrNotFound {
		ctx.Status(http.StatusNotFound)
		return
	}
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// postLocate returns the assigned car for a journey.
func (c *Controller) postLocate(ctx *gin.Context) {
	if ctx.Request.Method != http.MethodPost {
		ctx.AbortWithStatus(http.StatusMethodNotAllowed)
		return
	}
	if !isFormContentType(ctx.ContentType()) {
		ctx.Status(http.StatusBadRequest)
		return
	}

	var locate struct {
		ID uint `form:"ID" binding:"required"` // Journey ID.
	}
	if err := ctx.ShouldBind(&locate); err != nil || locate.ID == 0 {
		ctx.Status(http.StatusBadRequest)
		return
	}

	car, waiting, err := c.service.Locate(locate.ID)
	if err == service.ErrNotFound {
		ctx.Status(http.StatusNotFound)
		return
	}
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}
	if waiting {
		ctx.Status(http.StatusNoContent)
		return
	}

	ctx.JSON(http.StatusOK, model.Car{ID: car.ID, Seats: car.Seats})
}

// isJSONContentType validates JSON content type.
func isJSONContentType(contentType string) bool {
	return strings.HasPrefix(strings.ToLower(contentType), "application/json")
}

// isFormContentType validates form content type.
func isFormContentType(contentType string) bool {
	return strings.HasPrefix(strings.ToLower(contentType), "application/x-www-form-urlencoded")
}
