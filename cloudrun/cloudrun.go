package cloudrun

import (
	"context"
	"os"

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

var storedURI string

func GetURI(ctx context.Context) (string, error) { // Need IAM roles/run.viewer
	if storedURI == `` {
		if resp, err := GetService(ctx); err != nil {
			return ``, err
		} else {
			storedURI = resp.GetUri()
		}
	}
	return storedURI, nil
}

var serviceAccount string

func GetServiceAccount(ctx context.Context) (string, error) { // Need IAM roles/run.viewer
	if serviceAccount == `` {
		if resp, err := GetService(ctx); err != nil {
			return ``, err
		} else {
			serviceAccount = resp.GetTemplate().GetServiceAccount()
		}
	}
	return serviceAccount, nil
}
