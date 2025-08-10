module github.com/tihe/susi-property-service

go 1.23.0

toolchain go1.24.5

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/golang-jwt/jwt/v5 v5.0.0
	github.com/segmentio/kafka-go v0.4.48
	github.com/tihe/susi-shared v0.0.0
	golang.org/x/crypto v0.40.0
	gorm.io/driver/postgres v1.6.0
	gorm.io/gorm v1.30.0
)

// Use local shared module for development
replace github.com/tihe/susi-shared => ../../shared 