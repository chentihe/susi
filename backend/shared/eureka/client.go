package eureka

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type EurekaClient struct {
	serverURL string
	client    *http.Client
}

type Instance struct {
	InstanceID                    string            `json:"instanceId"`
	HostName                      string            `json:"hostName"`
	App                           string            `json:"app"`
	IPAddr                        string            `json:"ipAddr"`
	Status                        string            `json:"status"`
	OverriddenStatus              string            `json:"overriddenstatus"`
	Port                          Port              `json:"port"`
	SecurePort                    Port              `json:"securePort"`
	CountryID                     int               `json:"countryId"`
	DataCenterInfo                DataCenterInfo    `json:"dataCenterInfo"`
	LeaseInfo                     LeaseInfo         `json:"leaseInfo"`
	Metadata                      map[string]string `json:"metadata"`
	HomePageURL                   string            `json:"homePageUrl"`
	StatusPageURL                 string            `json:"statusPageUrl"`
	HealthCheckURL                string            `json:"healthCheckUrl"`
	VIPAddress                    string            `json:"vipAddress"`
	SecureVIPAddress              string            `json:"secureVipAddress"`
	IsCoordinatingDiscoveryServer bool              `json:"isCoordinatingDiscoveryServer"`
	LastUpdatedTimestamp          string            `json:"lastUpdatedTimestamp"`
	LastDirtyTimestamp            string            `json:"lastDirtyTimestamp"`
	ActionType                    string            `json:"actionType"`
}

type Port struct {
	Port    int  `json:"$"`
	Enabled bool `json:"@enabled"`
}

type DataCenterInfo struct {
	Class string `json:"@class"`
	Name  string `json:"name"`
}

type LeaseInfo struct {
	RenewalIntervalInSecs int   `json:"renewalIntervalInSecs"`
	DurationInSecs        int   `json:"durationInSecs"`
	RegistrationTimestamp int64 `json:"registrationTimestamp"`
	LastRenewalTimestamp  int64 `json:"lastRenewalTimestamp"`
	EvictionTimestamp     int64 `json:"evictionTimestamp"`
	ServiceUpTimestamp    int64 `json:"serviceUpTimestamp"`
}

type Application struct {
	Name     string     `json:"name"`
	Instance []Instance `json:"instance"`
}

type Applications struct {
	Application []Application `json:"application"`
}

type EurekaResponse struct {
	Applications Applications `json:"applications"`
}

func NewEurekaClient(serverURL string) *EurekaClient {
	return &EurekaClient{
		serverURL: serverURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (e *EurekaClient) Register(appName, hostName string, port int) error {
	instance := Instance{
		InstanceID: fmt.Sprintf("%s:%s:%d", hostName, appName, port),
		HostName:   hostName,
		App:        appName,
		IPAddr:     hostName,
		Status:     "UP",
		Port: Port{
			Port:    port,
			Enabled: true,
		},
		SecurePort: Port{
			Port:    443,
			Enabled: false,
		},
		CountryID: 1,
		DataCenterInfo: DataCenterInfo{
			Class: "com.netflix.appinfo.InstanceInfo$DefaultDataCenterInfo",
			Name:  "MyOwn",
		},
		LeaseInfo: LeaseInfo{
			RenewalIntervalInSecs: 30,
			DurationInSecs:        90,
		},
		Metadata: map[string]string{
			"management.port": fmt.Sprintf("%d", port),
		},
		HomePageURL:                   fmt.Sprintf("http://%s:%d/", hostName, port),
		StatusPageURL:                 fmt.Sprintf("http://%s:%d/health", hostName, port),
		HealthCheckURL:                fmt.Sprintf("http://%s:%d/health", hostName, port),
		VIPAddress:                    appName,
		SecureVIPAddress:              appName,
		IsCoordinatingDiscoveryServer: false,
		LastUpdatedTimestamp:          fmt.Sprintf("%d", time.Now().Unix()),
		LastDirtyTimestamp:            fmt.Sprintf("%d", time.Now().Unix()),
		ActionType:                    "ADDED",
	}

	url := fmt.Sprintf("%s/apps/%s", e.serverURL, appName)
	jsonData, err := json.Marshal(map[string]Instance{"instance": instance})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := e.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to register with Eureka: %d", resp.StatusCode)
	}

	return nil
}

func (e *EurekaClient) Deregister(appName, instanceID string) error {
	url := fmt.Sprintf("%s/apps/%s/%s", e.serverURL, appName, instanceID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to deregister from Eureka: %d", resp.StatusCode)
	}

	return nil
}

func (e *EurekaClient) GetServiceURL(appName string) (string, error) {
	url := fmt.Sprintf("%s/apps/%s", e.serverURL, appName)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/json")
	resp, err := e.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get service from Eureka: %d", resp.StatusCode)
	}

	var eurekaResp EurekaResponse
	if err := json.NewDecoder(resp.Body).Decode(&eurekaResp); err != nil {
		return "", err
	}

	for _, app := range eurekaResp.Applications.Application {
		if app.Name == appName && len(app.Instance) > 0 {
			instance := app.Instance[0]
			if instance.Status == "UP" {
				return fmt.Sprintf("http://%s:%d", instance.HostName, instance.Port.Port), nil
			}
		}
	}

	return "", fmt.Errorf("no available instance found for service: %s", appName)
}

func (e *EurekaClient) Heartbeat(appName, instanceID string) error {
	url := fmt.Sprintf("%s/apps/%s/%s", e.serverURL, appName, instanceID)
	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return err
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send heartbeat to Eureka: %d", resp.StatusCode)
	}

	return nil
}

// HealthCheck performs an actual health check on the service
func (e *EurekaClient) HealthCheck(serviceURL string) error {
	healthURL := serviceURL + "/health"
	req, err := http.NewRequest("GET", healthURL, nil)
	if err != nil {
		return err
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}

	return nil
}

// GetServiceURLWithHealthCheck gets service URL and verifies it's healthy
func (e *EurekaClient) GetServiceURLWithHealthCheck(appName string) (string, error) {
	url, err := e.GetServiceURL(appName)
	if err != nil {
		return "", err
	}

	// Perform health check
	if err := e.HealthCheck(url); err != nil {
		return "", fmt.Errorf("service %s is not healthy: %v", appName, err)
	}

	return url, nil
}
