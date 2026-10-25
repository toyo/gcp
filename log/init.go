package log

import (
	"context"

	"cloud.google.com/go/logging"
	"github.com/toyo/gcp/gce"
)

var (
	client *logging.Client
	// logger is shared. Client.Logger starts a goroutine and a bundler on every call, so calling it per entry leaks.
	logger *logging.Logger
)

func init() {
	var err error
	ctx := context.Background()
	client, err = logging.NewClient(ctx, `projects/`+gce.GetProjectID(ctx))
	if err != nil {
		panic(err.Error())
	}
	logger = client.Logger(`github.com/toyo/gcp/log`)
}
