package cloudtask

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"cloud.google.com/go/cloudtasks/apiv2/cloudtaskspb"
	"github.com/toyo/gcp/cloudrun"
	"github.com/toyo/gcp/gce"
	"github.com/toyo/gcp/googlecloudlogging"
)

// QueueID is taskqueue id.
const QueueID = "default"

func MiddlewareFunc(next http.HandlerFunc) http.HandlerFunc {

	const cloudtaskHeader = "X-Cloudtasks-Taskname" // "X-Appengine-Taskname"

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = googlecloudlogging.ContextInit(ctx, r)

		t, ok := r.Header[cloudtaskHeader]
		if !ok || len(t[0]) == 0 {
			slog.ErrorContext(ctx, "Invalid Cloudtask: No "+cloudtaskHeader+" request header found")
			http.Error(w, "Bad Request - Invalid Cloudtask", http.StatusBadRequest)
			return
		} else {
			slog.DebugContext(ctx, `Valid Cloudtask Header`, `Header`, r.Header)
			next(w, r)
		}
	}
}

//

func Middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(MiddlewareFunc(handler.ServeHTTP))
}

func taskOidcToken(ctx context.Context) *cloudtaskspb.HttpRequest_OidcToken {
	if serviceaccount, err := cloudrun.GetServiceAccount(ctx); err == nil {
		return &cloudtaskspb.HttpRequest_OidcToken{OidcToken: &cloudtaskspb.OidcToken{ServiceAccountEmail: serviceaccount}}
	} else {
		return nil
	}
}

// CreateTaskJSON creates task.
func CreateTaskJSON(ctx context.Context, cloudtasksclient *cloudtasks.Client, uri, taskID string, argint interface{}) (err error) {

	var b []byte
	if b, err = json.Marshal(argint); err != nil {
		return
	}

	// Build the Task payload.
	// https://godoc.org/google.golang.org/genproto/googleapis/cloud/tasks/v2#CreateTaskRequest
	taskreq := cloudtaskspb.CreateTaskRequest{
		Parent: "projects/" + gce.GetProjectID() + "/locations/" + gce.GetRegion() + "/queues/" + QueueID,
		Task: &cloudtaskspb.Task{
			MessageType: &cloudtaskspb.Task_HttpRequest{
				HttpRequest: &cloudtaskspb.HttpRequest{
					HttpMethod:          cloudtaskspb.HttpMethod_POST,
					Url:                 uri,
					Body:                b,
					Headers:             map[string]string{"Content-Type": "application/json"},
					AuthorizationHeader: taskOidcToken(ctx),
				},
			},
		},
	}

	if taskID != `` {
		taskreq.Task.Name = taskreq.Parent + "/tasks/" + taskID
	}

	if _, err = cloudtasksclient.CreateTask(ctx, &taskreq); err == nil {
		slog.DebugContext(ctx, `cloudtasks.CreateTask POST`, `URL`, uri, `POST`, string(b))
	} else {
		/*
			if sts, ok := status.FromError(err); ok && sts.Code() == codes.AlreadyExists {
				duplicate = true
				err = nil
			} else {
				err = errors.Wrapf(err, "cloudtasks.CreateTask %s, %s", uri, taskID)
			}
		*/
	}

	return
}

// CreateTaskGET creates appengine task.
func CreateTaskGET(ctx context.Context, cloudtasksclient *cloudtasks.Client, uri, taskID string, arg url.Values) (err error) {

	// Build the Task payload.
	// https://godoc.org/google.golang.org/genproto/googleapis/cloud/tasks/v2#CreateTaskRequest
	taskreq := cloudtaskspb.CreateTaskRequest{
		Parent: "projects/" + gce.GetProjectID() + "/locations/" + gce.GetRegion() + "/queues/" + QueueID,
		Task: &cloudtaskspb.Task{
			MessageType: &cloudtaskspb.Task_HttpRequest{
				HttpRequest: &cloudtaskspb.HttpRequest{
					HttpMethod:          cloudtaskspb.HttpMethod_GET,
					Url:                 uri + `?` + arg.Encode(),
					AuthorizationHeader: taskOidcToken(ctx),
				},
			},
		},
	}

	if taskID != `` {
		taskreq.Task.Name = taskreq.Parent + "/tasks/" + taskID
	}

	if _, err = cloudtasksclient.CreateTask(ctx, &taskreq); err == nil {
		slog.DebugContext(ctx, `cloudtasks.CreateTask GET`, `URL`, uri, `GET`, arg.Encode())
	} else { /*
			if sts, ok := status.FromError(err); ok && sts.Code() == codes.AlreadyExists {
				duplicate = true
				err = nil
			} else {
				err = errors.Wrapf(err, "cloudtasks.CreateTask %s?%s, %s", uri, arg.Encode(), taskID)
			}*/
	}

	return
}
