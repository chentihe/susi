package discovery

import "github.com/hashicorp/consul/api"

type ServiceDiscovery interface {
	Register(serviceName, serviceAddress string, servicePort int) error
	Deregister(serviceName, serviceAddress string, servicePort int) error
	GetServiceURL(serviceName string) (string, error)
	GetServices() (map[string]*api.AgentService, error)
	HealthCheck(instanceId string) error
}
