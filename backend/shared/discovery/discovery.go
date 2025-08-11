package discovery

type ServiceDiscovery interface {
	Register(serviceName, serviceAddress string, servicePort int) error
	Deregister(serviceName, serviceAddress string, servicePort int) error
	GetServiceURL(serviceName string) (string, error)
	HealthCheck(instanceId string) error
}
