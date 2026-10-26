package cloudrun

import (
	"context"
	"os"
	"sync"

	runv2 "cloud.google.com/go/run/apiv2"
	runpb "cloud.google.com/go/run/apiv2/runpb"
	"github.com/toyo/gcp/gce"
)

func GetServiceID() string {
	return os.Getenv(`K_SERVICE`)
}

func GetService(ctx context.Context) (*runpb.Service, error) {
	if c, err := runv2.NewServicesClient(ctx); err != nil {
		return nil, err
	} else {
		defer c.Close()

		return c.GetService(ctx,
			&runpb.GetServiceRequest{
				Name: gce.GetSlashedProjectsLocations(ctx) + `/services/` + GetServiceID(),
			})
	}
}

// Cloud Run から取った値は、取れたときだけ覚えておく。並行するリクエストから呼ばれるので mu で守る。
var (
	mu             sync.Mutex
	storedURI      string
	serviceAccount string
)

func GetURI(ctx context.Context) (string, error) { // Need IAM roles/run.viewer
	mu.Lock()
	defer mu.Unlock()

	if storedURI == `` {
		if resp, err := GetService(ctx); err != nil {
			return ``, err
		} else {
			storedURI = resp.GetUri()
		}
	}
	return storedURI, nil
}

func GetServiceAccount(ctx context.Context) (string, error) { // Need IAM roles/run.viewer
	mu.Lock()
	defer mu.Unlock()

	if serviceAccount == `` {
		if resp, err := GetService(ctx); err != nil {
			return ``, err
		} else {
			serviceAccount = resp.GetTemplate().GetServiceAccount()
		}
	}
	return serviceAccount, nil
}
