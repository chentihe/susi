package app

import (
	"github.com/gin-gonic/gin"
	"github.com/tihe/susi/backend/handlers"
	"github.com/tihe/susi/backend/services"
)

type AppRouter struct {
	Engine     *gin.Engine
	AppContext *AppContext
}

func NewAppRouter(ctx *AppContext) *AppRouter {
	r := gin.Default()
	apiV1 := r.Group("/api/v1")

	// Initialize handlers with their services (assume services are in ctx)
	apartmentHandler := handlers.NewApartmentHandler(ctx.ApartmentService)
	handlers.RegisterApartmentRoutes(apiV1, apartmentHandler)

	roomHandler := handlers.NewRoomHandler(ctx.RoomService)
	handlers.RegisterRoomRoutes(apiV1, roomHandler)

	tenantHandler := handlers.NewTenantHandler(ctx.TenantService)
	handlers.RegisterTenantRoutes(apiV1, tenantHandler)

	renovationHandler := handlers.NewRenovationHandler(ctx.RenovationService)
	handlers.RegisterRenovationRoutes(apiV1, renovationHandler)

	landLordHandler := handlers.NewLandLordHandler(ctx.LandLordService)
	handlers.RegisterLandLordRoutes(apiV1, landLordHandler)

	adminHandler := handlers.NewAdminHandler(ctx.AdminService)
	handlers.RegisterAdminRoutes(apiV1, adminHandler)

	return &AppRouter{
		Engine:     r,
		AppContext: ctx,
	}
} 