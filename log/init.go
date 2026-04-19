package log

import (
	"context"

	"cloud.google.com/go/logging"
	"github.com/toyo/gcp/gce"
)

var client *logging.Client

func init() {
	var err error
	ctx := context.Background()
	client, err = logging.NewClient(ctx, `projects/`+gce.GetProjectID(ctx))
	if err != nil {
		panic(err.Error())
	}
}
