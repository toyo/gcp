package gce

import (
	"context"
	"strings"
	"sync"

	"cloud.google.com/go/compute/metadata"
)

// メタデータサーバーから取った値は、取れたときだけ覚えておく。並行するリクエストから呼ばれるので mu で守る。
var (
	mu              sync.Mutex
	storedProjectID string
	storedRegion    string
)

func GetProjectID(ctx context.Context) string {
	mu.Lock()
	defer mu.Unlock()

	if storedProjectID == `` {
		if !metadata.OnGCEWithContext(ctx) {
			return ``
		}

		projectID, err := metadata.ProjectIDWithContext(ctx)
		if err != nil {
			panic(err.Error())
		}
		storedProjectID = projectID
	}
	return storedProjectID
	//return os.Getenv(`GOOGLE_CLOUD_PROJECT`)
}

func GetRegion(ctx context.Context) string { // Get region from metadata server
	mu.Lock()
	defer mu.Unlock()

	if storedRegion == `` {
		if !metadata.OnGCEWithContext(ctx) {
			panic("this process is not running on GCE")
		}

		if instanceRegion, err := metadata.GetWithContext(ctx, `instance/region`); err != nil {
			panic(err.Error())
		} else {
			splittedInstanceRegion := strings.Split(instanceRegion, `/`)
			for i := 0; i+1 < len(splittedInstanceRegion); i += 2 {
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
func GetSlashedProjectsLocations(ctx context.Context) string {
	return `projects/` + GetProjectID(ctx) + `/locations/` + GetRegion(ctx)
}

func GetServiceAccount(ctx context.Context) (string, error) { // not tested
	if !metadata.OnGCEWithContext(ctx) {
		panic("this process is not running on GCE")
	}

	if email, err := metadata.GetWithContext(ctx, `instance/service-accounts/email`); err != nil {
		panic(err.Error())
	} else {
		return email, err
	}

}
