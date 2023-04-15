package gce

import (
	"context"
	"strings"

	"cloud.google.com/go/compute/metadata"
)

var storedProjectID string

func GetProjectID() string {
	if storedProjectID == `` {
		if !metadata.OnGCE() {
			panic("this process is not running on GCE")
		}

		var err error
		if storedProjectID, err = metadata.ProjectID(); err != nil {
			panic(err.Error())
		}
	}
	return storedProjectID
	//return os.Getenv(`GOOGLE_CLOUD_PROJECT`)
}

var storedRegion string

func GetRegion() string { // Get region from metadata server
	if storedRegion == `` {
		if !metadata.OnGCE() {
			panic("this process is not running on GCE")
		}

		if instanceRegion, err := metadata.Get(`instance/region`); err != nil {
			panic(err.Error())
		} else {
			splittedInstanceRegion := strings.Split(instanceRegion, `/`)
			for i := 0; i < len(splittedInstanceRegion); i += 2 {
				if splittedInstanceRegion[i] == `regions` || splittedInstanceRegion[i] == `zones` { // Fallback: get "zone" from metadata server (running on VM e.g. Cloud Run for Anthos)
					storedRegion = splittedInstanceRegion[i+1]
					return storedRegion
				}
			}
			panic(`cannot get from metadata.google.internal`)
		}
	}
	return storedRegion
}
func GetSlashedProjectsLocations() string {
	return `projects/` + GetProjectID() + `/locations/` + GetRegion()
}

func GetServiceAccount(ctx context.Context) (string, error) { // not tested
	if !metadata.OnGCE() {
		panic("this process is not running on GCE")
	}

	if email, err := metadata.Get(`instance/service-accounts/email`); err != nil {
		panic(err.Error())
	} else {
		return email, err
	}

}
