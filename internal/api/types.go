package api

import (
	"time"
)

// ResourceCookie stores provisioned resource details
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

// oas:schema
// AsyncProgressHint describes ongoing asynchronous operation.
//
// oas:example
//  {
//    "status": "pending|deleting",
//    "estimate": "2021-02-22T09:00:10Z",
//    "deadline": "2021-02-22T09:02:00Z",
//    "timestamp": "2021-02-22T09:00:00Z"
//  }
type AsyncProgressHint struct {
	// Current status.
	// oas:required: true
	// oas:pattern: ^pending|deleting$
	Status string `json:"status"`

	// Estimated date and time of when the operation is expected to be completed.
	Estimate time.Time `json:"estimate,omitempty"`

	// The date and time of when the operation will be aborted if not yet finished.
	Deadline time.Time `json:"deadline,omitempty"`

	// Current (reference) date and time.
	Timestamp time.Time `json:"timestamp,omitempty"`
}

// oas:schema
// ValuesSecrets stores data that should be passed around split by sensitivity.
//
// oas:example
//  {
//    "values": {
//      "host": "127.0.0.1",
//      "name": "my-database"
//    },
//    "secrets": {
//      "user": "<secret>",
//      "password": "<secret>"
//    }
//  }
type ValuesSecrets struct {
	// Values section of the data set.
	Values map[string]interface{} `json:"values,omitempty"`
	// Secrets section of the data set.
	Secrets map[string]interface{} `json:"secrets,omitempty"`
}

// oas:schema
// Manifest represents a complete or a partial Kubernetes manifest, and a location for its injection.
//
// oas:example
//
//  {
//    "location": "/place/me/there",
//    "data": {
//      "key": "value"
//    }
//  }
type Manifest struct {
	// Location to inject the Manifest at.
	// oas:required: true
	Location string `json:"location"`

	// Manifest data to inject.
	// oas:required: true
	Data interface{} `json:"data"`
}

// oas:schema
// DriverInputs describes the resource and the input paramaters for the driver.
//
// oas:example
//  {
//    "type": "postgres",
//    "resource": {},
//    "driver": {
//      "values": {
//        "instance": "database"
//      },
//      "secrets": {
//        "account": {
//           "username": "<secret>",
//           "password": "<secret>"
//        },
//        "secret": "<secret>"
//      }
//    }
//  }
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

// oas:schema
// DriverOutputs stores all the necessary information about the provisioned resource.
//
// oas:example
//  {
//    "id": "{guresid}",
//    "type": "postgres",
//    "resource": {
//      "values": {
//        "host": "127.0.0.1",
//        "name": "my-database"
//      },
//      "secrets": {
//        "username": "<secret>",
//        "password": "<secret>"
//      }
//    },
//    "manifests" : []
//  }
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
