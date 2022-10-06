package internal

import "time"

const (
	HeaderHumanitecDriverCookie    = "Humanitec-Driver-Cookie"
	HeaderHumanitecDriverCookieSet = "Set-Humanitec-Driver-Cookie"
)

type ResourceCookie struct {
	GUResID   string    `json:"id"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`

	Region          string `json:"region,omitempty"`
	HostedZoneID    string `json:"hosted_zone_id,omitempty"`
	AWSAccessKeyId  string `json:"aws_access_key_id"`
	AWSAccessSecret string `json:"aws_secret_access_key"`

	Resource *ValuesSecrets `json:"resource"`
}

type ValuesSecrets struct {
	// Values section of the data set.
	Values map[string]interface{} `json:"values,omitempty"`
	// Secrets section of the data set.
	Secrets map[string]interface{} `json:"secrets,omitempty"`
}

type DriverInputs struct {
	// The type of the resource to generate.
	// oas:required: true
	Type string `json:"type"`

	// The resource-related parameters passed from the deployment set (if any).
	// oas:required: true
	Resource map[string]interface{} `json:"resource"`

	// The driver-specific parameters passed from the Resource Definition for the Environment.
	// oas:required: true
	Driver *ValuesSecrets `json:"driver"`
}

type DriverOutputs struct {
	// The resource GUResID.
	// oas:required: true
	GUResID string `json:"id"`

	// The type of the resource.
	// oas:required: true
	Type string `json:"type"`

	// The resource usage parameters and secrets.
	// oas:required: true
	Resource *ValuesSecrets `json:"resource"`

	// The resource definition manifests (if any).
	// oas:required: true
	Manifests []Manifest `json:"manifests"`
}

type Manifest struct {
	// Location to inject the Manifest at.
	// oas:required: true
	Location string `json:"location"`

	// Manifest data to inject.
	// oas:required: true
	Data interface{} `json:"data"`
}
