package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tihe/susi-proto/auth"
	"github.com/tihe/susi-shared/discovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func JWTAuthMiddleware(registry discovery.ServiceDiscovery) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Authorization header required",
			})
			return
		}

		// Extract from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid authorization header format",
			})
			return
		}

		token := parts[1]

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		validateReq := &auth.ValidateTokenRequest{
			Token: token,
			// RequiredPermissions // TODO: add specific permissions if needed
		}

		serviceURL, err := registry.GetServiceURL("auth-service")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error":   "Notfound",
				"message": "Auth service not found",
			})
			return
		}

		conn, err := grpc.NewClient(serviceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error":   "Connectionfailed",
				"message": "Failed to connect to auth service",
			})
		}
		defer conn.Close()

		client := auth.NewAuthServiceClient(conn)

		validateResp, err := client.ValidateToken(ctx, validateReq)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Token validation failed",
			})
			return
		}

		if !validateResp.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Account is not active",
			})
			return
		}

		if validateResp.User != nil {
			user := validateResp.User
			c.Set("user_id", user.UserId)
			c.Set("user_email", user.Email)
			c.Set("user_role", user.Role.String())
			c.Set("user_permissions", user.Permissions)
		}

		// Token is valid, continue to next handler
		c.Next()
	}
}

// func RoleAuthMiddleware(authClient auth.AuthService, requiredRoles ...string) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		// Get token from Authorization header
// 		authHeader := c.GetHeader("Authorization")
// 		if authHeader == "" {
// 			c.JSON(http.StatusUnauthorized, gin.H{
// 				"error":   "Unauthorized",
// 				"message": "Authorization header required",
// 			})
// 			c.Abort()
// 			return
// 		}

// 		parts := strings.Split(authHeader, " ")
// 		if len(parts) != 2 || parts[0] != "Bearer" {
// 			c.JSON(http.StatusUnauthorized, gin.H{
// 				"error":   "Unauthorized",
// 				"message": "Invalid authorization header format",
// 			})
// 			c.Abort()
// 			return
// 		}

// 		token := parts[1]

// 		// Validate token with role requirements
// 		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 		defer cancel()

// 		validateReq := &auth.ValidateTokenRequest{
// 			Token: token,
// 		}

// 		validateResp, err := authClient.ValidateToken(ctx, validateReq)
// 		if err != nil || !validateResp.Valid {
// 			c.JSON(http.StatusUnauthorized, gin.H{
// 				"error":   "Unauthorized",
// 				"message": "Token validation failed",
// 			})
// 			c.Abort()
// 			return
// 		}

// 		// Check if user has required role
// 		userRole := validateResp.User.Role.String()
// 		hasRequiredRole := false

// 		for _, requiredRole := range requiredRoles {
// 			if userRole == requiredRole {
// 				hasRequiredRole = true
// 				break
// 			}
// 		}

// 		if !hasRequiredRole {
// 			c.JSON(http.StatusForbidden, gin.H{
// 				"error":   "Forbidden",
// 				"message": "Insufficient privileges",
// 			})
// 			c.Abort()
// 			return
// 		}

// 		// Add user info to context
// 		c.Set("user_id", validateResp.User.UserId)
// 		c.Set("user_email", validateResp.User.Email)
// 		c.Set("user_role", validateResp.User.Role.String())
// 		c.Set("user_permissions", validateResp.User.Permissions)

// 		c.Next()
// 	}
// }

// // PermissionAuthMiddleware - middleware to check specific permissions
// func PermissionAuthMiddleware(authClient auth.AuthService, requiredPermissions ...string) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		authHeader := c.GetHeader("Authorization")
// 		if authHeader == "" {
// 			c.JSON(http.StatusUnauthorized, gin.H{
// 				"error":   "Unauthorized",
// 				"message": "Authorization header required",
// 			})
// 			c.Abort()
// 			return
// 		}

// 		parts := strings.Split(authHeader, " ")
// 		if len(parts) != 2 || parts[0] != "Bearer" {
// 			c.JSON(http.StatusUnauthorized, gin.H{
// 				"error":   "Unauthorized",
// 				"message": "Invalid authorization header format",
// 			})
// 			c.Abort()
// 			return
// 		}

// 		token := parts[1]

// 		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 		defer cancel()

// 		validateReq := &auth.ValidateTokenRequest{
// 			Token:               token,
// 			RequiredPermissions: requiredPermissions, // Check specific permissions
// 		}

// 		validateResp, err := authClient.ValidateToken(ctx, validateReq)
// 		if err != nil || !validateResp.Valid {
// 			c.JSON(http.StatusUnauthorized, gin.H{
// 				"error":   "Unauthorized",
// 				"message": "Token validation failed",
// 			})
// 			c.Abort()
// 			return
// 		}

// 		// The auth service should have already checked permissions
// 		// but we can double-check here if needed
// 		userPermissions := validateResp.User.Permissions

// 		for _, requiredPerm := range requiredPermissions {
// 			hasPermission := false
// 			for _, userPerm := range userPermissions {
// 				if userPerm == requiredPerm {
// 					hasPermission = true
// 					break
// 				}
// 			}

// 			if !hasPermission {
// 				c.JSON(http.StatusForbidden, gin.H{
// 					"error":   "Forbidden",
// 					"message": "Insufficient permissions",
// 				})
// 				c.Abort()
// 				return
// 			}
// 		}

// 		c.Set("user_id", validateResp.User.UserId)
// 		c.Set("user_email", validateResp.User.Email)
// 		c.Set("user_role", validateResp.User.Role.String())
// 		c.Set("user_permissions", validateResp.User.Permissions)

// 		c.Next()
// 	}
// }
