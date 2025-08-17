package gateway

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hashicorp/consul/api"
	"github.com/tihe/susi-proto/auth"
	"github.com/tihe/susi-shared/discovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var requriedServices = []string{
	os.Getenv("AUTH_SERVICE_NAME"),
}

func NewGateway(ctx context.Context, registry discovery.ServiceDiscovery) (http.Handler, error) {
	mux := runtime.NewServeMux()

	services, err := waitForRequiredServices(ctx, registry, requriedServices, 30*time.Second)
	if err != nil {
		return nil, err
	}

	for _, service := range services {
		if err := registerServiceHandler(ctx, mux, service); err != nil {
			return nil, err
		}
	}

	return mux, nil
}

func waitForRequiredServices(ctx context.Context, registry discovery.ServiceDiscovery, required []string, timeout time.Duration) (map[string]*api.AgentService, error) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-timeoutCtx.Done():
			return nil, fmt.Errorf("timeout waiting for required services: %v", required)
		case <-ticker.C:
			services, err := registry.GetServices()
			if err != nil {
				log.Printf("Error getting services: %v", err)
				continue
			}

			availableServices := make(map[string]*api.AgentService)
			for _, service := range services {
				availableServices[service.Service] = service
			}

			allAvailable := true
			var missing []string
			for _, req := range required {
				if _, exists := availableServices[req]; !exists {
					allAvailable = false
					missing = append(missing, req)
				}
			}

			if allAvailable {
				log.Printf("All required services are now available")
				return services, nil
			}

			log.Printf("Strill waiting for services: %v", missing)
		}
	}
}

func registerServiceHandler(ctx context.Context, gatewayMux *runtime.ServeMux, service *api.AgentService) error {
	serviceURL := fmt.Sprintf("%s:%d", service.Address, service.Port)
	conn, err := dial(serviceURL)
	if err != nil {
		return err
	}
	defer conn.Close()

	switch service.Service {
	case os.Getenv("AUTH_SERVICE_NAME"):
		return auth.RegisterAuthServiceHandler(ctx, gatewayMux, conn)
	default:
		return nil
	}
}

func dial(addr string) (*grpc.ClientConn, error) {
	return grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
}
