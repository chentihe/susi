package app

import (
	"github.com/gin-gonic/gin"
	"github.com/tihe/susi/backend/handlers"
	"github.com/tihe/susi/backend/services"
	"github.com/tihe/susi/backend/middleware"
)

type AppRouter struct {
	Engine     *gin.Engine
	AppContext *AppContext
}

func NewAppRouter(ctx *AppContext) *AppRouter {
	r := gin.Default()
	apiV1 := r.Group("/api/v1")

	// --- Public Auth Endpoints ---
	authHandler := handlers.NewAuthHandler(ctx.DB, ctx.KafkaProducer, ctx.AdminService)
	handlers.RegisterAuthRoutes(apiV1, authHandler)

	// --- Protected Endpoints (require JWT) ---
	protected := apiV1.Group("")
	protected.Use(middleware.JWTAuthMiddleware())

	apartmentHandler := handlers.NewApartmentHandler(ctx.ApartmentService)
	handlers.RegisterApartmentRoutes(protected, apartmentHandler)

	roomHandler := handlers.NewRoomHandler(ctx.RoomService)
	handlers.RegisterRoomRoutes(protected, roomHandler)

	tenantHandler := handlers.NewTenantHandler(ctx.TenantService)
	handlers.RegisterTenantRoutes(protected, tenantHandler)

	renovationHandler := handlers.NewRenovationHandler(ctx.RenovationService)
	handlers.RegisterRenovationRoutes(protected, renovationHandler)

	landLordHandler := handlers.NewLandLordHandler(ctx.LandLordService)
	handlers.RegisterLandLordRoutes(protected, landLordHandler)

	adminHandler := handlers.NewAdminHandler(ctx.AdminService)
	handlers.RegisterAdminRoutes(protected, adminHandler)

	return &AppRouter{
		Engine:     r,
		AppContext: ctx,
	}
} 