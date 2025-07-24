package app

import (
	"github.com/tihe/susi/backend/events"
	"github.com/tihe/susi/backend/models"
	"github.com/tihe/susi/backend/services"
	"gorm.io/gorm"
)

type AppContext struct {
	DB               *gorm.DB
	KafkaProducer    *events.KafkaProducer
	ApartmentService services.ApartmentService
	RoomService      services.RoomService
	TenantService    services.TenantService
	RenovationService services.RenovationService
	LandLordService  services.LandLordService
	AdminService     services.AdminService
}

func NewAppContext(db *gorm.DB, producer *events.KafkaProducer,
	apartmentService services.ApartmentService,
	roomService services.RoomService,
	tenantService services.TenantService,
	renovationService services.RenovationService,
	landLordService services.LandLordService,
	adminService services.AdminService) *AppContext {
	return &AppContext{
		DB:               db,
		KafkaProducer:    producer,
		ApartmentService: apartmentService,
		RoomService:      roomService,
		TenantService:    tenantService,
		RenovationService: renovationService,
		LandLordService:  landLordService,
		AdminService:     adminService,
	}
} 