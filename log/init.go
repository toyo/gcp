package log

import (
	"context"

	"cloud.google.com/go/logging"
	"github.com/toyo/gcp/gce"
)

var client *logging.Client

func init() {
	var err error
	client, err = logging.NewClient(context.Background(), `projects/`+gce.GetProjectID())
	if err != nil {
		panic(err.Error())
	}
}
