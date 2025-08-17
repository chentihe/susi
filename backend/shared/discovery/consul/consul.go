package consul

import (
	"fmt"
	"log"

	"github.com/hashicorp/consul/api"
)

type ConsulClient struct {
	client *api.Client
}

func NewConsulClient(consulURL string) (*ConsulClient, error) {
	client, err := api.NewClient(&api.Config{Address: consulURL})
	if err != nil {
		return nil, err
	}
	return &ConsulClient{
		client: client,
	}, nil
}

func (c *ConsulClient) Register(serviceName, serviceAddress string, servicePort int) error {
	registration := &api.AgentServiceRegistration{
		ID:      fmt.Sprintf("%s:%s:%d", serviceName, serviceAddress, servicePort),
		Name:    serviceName, // Service name for discovery
		Port:    servicePort,
		Address: serviceAddress, // Container name for Docker network resolution
		Check: &api.AgentServiceCheck{
			CheckID: fmt.Sprintf("%s:%s:%d", serviceName, serviceAddress, servicePort),
			TTL:     "5s",
		},
	}

	if err := c.client.Agent().ServiceRegister(registration); err != nil {
		return err
	}

	return nil
}

func (c *ConsulClient) Deregister(serviceName, serviceAddress string, servicePort int) error {
	if err := c.client.Agent().ServiceDeregister(fmt.Sprintf("%s:%s:%d", serviceName, serviceAddress, servicePort)); err != nil {
		return err
	}
	return nil
}

func (c *ConsulClient) GetServiceURL(serviceName string) (string, error) {
	services, _, err := c.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return "", err
	}
	if len(services) == 0 {
		return "", fmt.Errorf("no available instance found for service: %s", serviceName)
	}

	// var instances []string
	// for _, entry := range services {
	// 	instances = append(instances, fmt.Sprintf("%s:%d", entry.Service.Address, entry.Service.Port))
	// }
	return fmt.Sprintf("%s:%d", services[0].Service.Address, services[0].Service.Port), nil
}

func (c *ConsulClient) GetServices() (map[string]*api.AgentService, error) {
	services, err := c.client.Agent().Services()
	if err != nil {
		return nil, err
	}
	for name, service := range services {
		log.Printf("name: %s, service: %v", name, service)
	}
	return services, nil
}

// HealthCheck performs an actual health check on the service
func (c *ConsulClient) HealthCheck(instanceId string) error {
	err := c.client.Agent().UpdateTTL(instanceId, "", "pass")
	return err
}
