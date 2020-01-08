// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by gapic-generator. DO NOT EDIT.

package automl

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"time"

	"cloud.google.com/go/longrunning"
	lroauto "cloud.google.com/go/longrunning/autogen"
	gax "github.com/googleapis/gax-go/v2"
	"google.golang.org/api/option"
	"google.golang.org/api/transport"
	automlpb "google.golang.org/genproto/googleapis/cloud/automl/v1"
	longrunningpb "google.golang.org/genproto/googleapis/longrunning"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// PredictionCallOptions contains the retry settings for each method of PredictionClient.
type PredictionCallOptions struct {
	Predict      []gax.CallOption
	BatchPredict []gax.CallOption
}

func defaultPredictionClientOptions() []option.ClientOption {
	return []option.ClientOption{
		option.WithEndpoint("automl.googleapis.com:443"),
		option.WithScopes(DefaultAuthScopes()...),
		option.WithGRPCDialOption(grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(math.MaxInt32))),
	}
}

func defaultPredictionCallOptions() *PredictionCallOptions {
	retry := map[[2]string][]gax.CallOption{}
	return &PredictionCallOptions{
		Predict:      retry[[2]string{"default", "non_idempotent"}],
		BatchPredict: retry[[2]string{"default", "non_idempotent"}],
	}
}

// PredictionClient is a client for interacting with Cloud AutoML API.
//
// Methods, except Close, may be called concurrently. However, fields must not be modified concurrently with method calls.
type PredictionClient struct {
	// The connection to the service.
	conn *grpc.ClientConn

	// The gRPC API client.
	predictionClient automlpb.PredictionServiceClient

	// LROClient is used internally to handle longrunning operations.
	// It is exposed so that its CallOptions can be modified if required.
	// Users should not Close this client.
	LROClient *lroauto.OperationsClient

	// The call options for this service.
	CallOptions *PredictionCallOptions

	// The x-goog-* metadata to be sent with each request.
	xGoogMetadata metadata.MD
}

// NewPredictionClient creates a new prediction service client.
//
// AutoML Prediction API.
//
// On any input that is documented to expect a string parameter in
// snake_case or kebab-case, either of those cases is accepted.
func NewPredictionClient(ctx context.Context, opts ...option.ClientOption) (*PredictionClient, error) {
	conn, err := transport.DialGRPC(ctx, append(defaultPredictionClientOptions(), opts...)...)
	if err != nil {
		return nil, err
	}
	c := &PredictionClient{
		conn:        conn,
		CallOptions: defaultPredictionCallOptions(),

		predictionClient: automlpb.NewPredictionServiceClient(conn),
	}
	c.setGoogleClientInfo()

	c.LROClient, err = lroauto.NewOperationsClient(ctx, option.WithGRPCConn(conn))
	if err != nil {
		// This error "should not happen", since we are just reusing old connection
		// and never actually need to dial.
		// If this does happen, we could leak conn. However, we cannot close conn:
		// If the user invoked the function with option.WithGRPCConn,
		// we would close a connection that's still in use.
		// TODO(pongad): investigate error conditions.
		return nil, err
	}
	return c, nil
}

// Connection returns the client's connection to the API service.
func (c *PredictionClient) Connection() *grpc.ClientConn {
	return c.conn
}

// Close closes the connection to the API service. The user should invoke this when
// the client is no longer required.
func (c *PredictionClient) Close() error {
	return c.conn.Close()
}

// setGoogleClientInfo sets the name and version of the application in
// the `x-goog-api-client` header passed on each request. Intended for
// use by Google-written clients.
func (c *PredictionClient) setGoogleClientInfo(keyval ...string) {
	kv := append([]string{"gl-go", versionGo()}, keyval...)
	kv = append(kv, "gapic", versionClient, "gax", gax.Version, "grpc", grpc.Version)
	c.xGoogMetadata = metadata.Pairs("x-goog-api-client", gax.XGoogHeader(kv...))
}

// Predict perform an online prediction. The prediction result will be directly
// returned in the response.
// Available for following ML problems, and their expected request payloads:
//
//   Image Classification - Image in .JPEG, .GIF or .PNG format, image_bytes
//   up to 30MB.
//
//   Image Object Detection - Image in .JPEG, .GIF or .PNG format, image_bytes
//   up to 30MB.
//
//   Text Classification - TextSnippet, content up to 60,000 characters,
//   UTF-8 encoded.
//
//   Text Extraction - TextSnippet, content up to 30,000 characters,
//   UTF-8 NFC encoded.
//
//   Translation - TextSnippet, content up to 25,000 characters, UTF-8
//   encoded.
//
//   Text Sentiment - TextSnippet, content up 500 characters, UTF-8
//   encoded.
func (c *PredictionClient) Predict(ctx context.Context, req *automlpb.PredictRequest, opts ...gax.CallOption) (*automlpb.PredictResponse, error) {
	md := metadata.Pairs("x-goog-request-params", fmt.Sprintf("%s=%v", "name", url.QueryEscape(req.GetName())))
	ctx = insertMetadata(ctx, c.xGoogMetadata, md)
	opts = append(c.CallOptions.Predict[0:len(c.CallOptions.Predict):len(c.CallOptions.Predict)], opts...)
	var resp *automlpb.PredictResponse
	err := gax.Invoke(ctx, func(ctx context.Context, settings gax.CallSettings) error {
		var err error
		resp, err = c.predictionClient.Predict(ctx, req, settings.GRPC...)
		return err
	}, opts...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// BatchPredict perform a batch prediction. Unlike the online
// [Predict][google.cloud.automl.v1.PredictionService.Predict], batch
// prediction result won't be immediately available in the response. Instead,
// a long running operation object is returned. User can poll the operation
// result via [GetOperation][google.longrunning.Operations.GetOperation]
// method. Once the operation is done,
// [BatchPredictResult][google.cloud.automl.v1.BatchPredictResult] is returned
// in the [response][google.longrunning.Operation.response] field. Available
// for following ML problems:
//
//   Image Classification
//
//   Image Object Detection
//
//   Text Extraction
func (c *PredictionClient) BatchPredict(ctx context.Context, req *automlpb.BatchPredictRequest, opts ...gax.CallOption) (*BatchPredictOperation, error) {
	md := metadata.Pairs("x-goog-request-params", fmt.Sprintf("%s=%v", "name", url.QueryEscape(req.GetName())))
	ctx = insertMetadata(ctx, c.xGoogMetadata, md)
	opts = append(c.CallOptions.BatchPredict[0:len(c.CallOptions.BatchPredict):len(c.CallOptions.BatchPredict)], opts...)
	var resp *longrunningpb.Operation
	err := gax.Invoke(ctx, func(ctx context.Context, settings gax.CallSettings) error {
		var err error
		resp, err = c.predictionClient.BatchPredict(ctx, req, settings.GRPC...)
		return err
	}, opts...)
	if err != nil {
		return nil, err
	}
	return &BatchPredictOperation{
		lro: longrunning.InternalNewOperation(c.LROClient, resp),
	}, nil
}

// BatchPredictOperation manages a long-running operation from BatchPredict.
type BatchPredictOperation struct {
	lro *longrunning.Operation
}

// BatchPredictOperation returns a new BatchPredictOperation from a given name.
// The name must be that of a previously created BatchPredictOperation, possibly from a different process.
func (c *PredictionClient) BatchPredictOperation(name string) *BatchPredictOperation {
	return &BatchPredictOperation{
		lro: longrunning.InternalNewOperation(c.LROClient, &longrunningpb.Operation{Name: name}),
	}
}

// Wait blocks until the long-running operation is completed, returning the response and any errors encountered.
//
// See documentation of Poll for error-handling information.
func (op *BatchPredictOperation) Wait(ctx context.Context, opts ...gax.CallOption) (*automlpb.BatchPredictResult, error) {
	var resp automlpb.BatchPredictResult
	if err := op.lro.WaitWithInterval(ctx, &resp, 5000*time.Millisecond, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Poll fetches the latest state of the long-running operation.
//
// Poll also fetches the latest metadata, which can be retrieved by Metadata.
//
// If Poll fails, the error is returned and op is unmodified. If Poll succeeds and
// the operation has completed with failure, the error is returned and op.Done will return true.
// If Poll succeeds and the operation has completed successfully,
// op.Done will return true, and the response of the operation is returned.
// If Poll succeeds and the operation has not completed, the returned response and error are both nil.
func (op *BatchPredictOperation) Poll(ctx context.Context, opts ...gax.CallOption) (*automlpb.BatchPredictResult, error) {
	var resp automlpb.BatchPredictResult
	if err := op.lro.Poll(ctx, &resp, opts...); err != nil {
		return nil, err
	}
	if !op.Done() {
		return nil, nil
	}
	return &resp, nil
}

// Metadata returns metadata associated with the long-running operation.
// Metadata itself does not contact the server, but Poll does.
// To get the latest metadata, call this method after a successful call to Poll.
// If the metadata is not available, the returned metadata and error are both nil.
func (op *BatchPredictOperation) Metadata() (*automlpb.OperationMetadata, error) {
	var meta automlpb.OperationMetadata
	if err := op.lro.Metadata(&meta); err == longrunning.ErrNoMetadata {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &meta, nil
}

// Done reports whether the long-running operation has completed.
func (op *BatchPredictOperation) Done() bool {
	return op.lro.Done()
}

// Name returns the name of the long-running operation.
// The name is assigned by the server and is unique within the service from which the operation is created.
func (op *BatchPredictOperation) Name() string {
	return op.lro.Name()
}
