package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/doublecloud/go-genproto/doublecloud/transfer/v1"
	"github.com/doublecloud/go-genproto/doublecloud/transfer/v1/endpoint"
	endpoint_airbyte "github.com/doublecloud/go-genproto/doublecloud/transfer/v1/endpoint/airbyte"
	dc "github.com/doublecloud/go-sdk"

	"github.com/doublecloud/go-sdk/iamkey"
	"github.com/doublecloud/go-sdk/operation"
)

func createS3SourceEndpoint(ctx context.Context, dc *dc.SDK, flags *cmdFlags) (*operation.Operation, error) {
	op, err := dc.WrapOperation(dc.Transfer().Endpoint().Create(ctx, &transfer.CreateEndpointRequest{
		ProjectId: *flags.projectID,
		Name:      fmt.Sprint("s3-src-", *flags.name),
		Settings: &transfer.EndpointSettings{
			Settings: &transfer.EndpointSettings_S3Source{
				S3Source: &endpoint_airbyte.S3Source{
					Dataset:     "test",
					PathPattern: "test",
					Schema:      "test",
					Format:      &endpoint_airbyte.S3Source_Format{Format: &endpoint_airbyte.S3Source_Format_Csv{}},
					Provider:    &endpoint_airbyte.S3Source_Provider{Bucket: "test"},
				},
			},
		},
	}))
	if err != nil {
		return op, err
	}
	err = op.Wait(ctx)
	return op, err
}

func createCHDstEndpoint(ctx context.Context, dc *dc.SDK, flags *cmdFlags) (*operation.Operation, error) {
	op, err := dc.WrapOperation(dc.Transfer().Endpoint().Create(ctx, &transfer.CreateEndpointRequest{
		ProjectId: *flags.projectID,
		Name:      fmt.Sprint("s3-dst-", *flags.name),
		Settings: &transfer.EndpointSettings{
			Settings: &transfer.EndpointSettings_ClickhouseTarget{
				ClickhouseTarget: &endpoint.ClickhouseTarget{
					Connection: &endpoint.ClickhouseConnection{
						Connection: &endpoint.ClickhouseConnection_ConnectionOptions{
							ConnectionOptions: &endpoint.ClickhouseConnectionOptions{
								Address:  &endpoint.ClickhouseConnectionOptions_MdbClusterId{MdbClusterId: "cluster_id"},
								Database: "default",
								User:     "user",
								Password: &endpoint.Secret{Value: &endpoint.Secret_Raw{Raw: "98s*%^P!3Bw38"}},
							},
						},
					},
				},
			},
		},
	}))
	if err != nil {
		return op, err
	}
	err = op.Wait(ctx)
	return op, err
}

func deleteEndpoint(ctx context.Context, dc *dc.SDK, endpointId string) (*operation.Operation, error) {
	op, err := dc.WrapOperation(dc.Transfer().Endpoint().Delete(ctx, &transfer.DeleteEndpointRequest{EndpointId: endpointId}))
	if err != nil {
		log.Fatal(err)
	}
	err = op.Wait(ctx)
	return op, err
}

func createTransfer(ctx context.Context, dc *dc.SDK, flags *cmdFlags, src string, dst string) (*operation.Operation, error) {
	op, err := dc.WrapOperation(dc.Transfer().Transfer().Create(ctx, &transfer.CreateTransferRequest{
		SourceId:  src,
		TargetId:  dst,
		Name:      *flags.name,
		ProjectId: *flags.projectID,
		Type:      transfer.TransferType_SNAPSHOT_ONLY,
	}))
	if err != nil {
		return op, err
	}
	err = op.Wait(ctx)
	return op, err
}

func activateTransfer(ctx context.Context, dc *dc.SDK, transferId string) (*operation.Operation, error) {
	op, err := dc.WrapOperation(dc.Transfer().Transfer().Activate(ctx, &transfer.ActivateTransferRequest{
		TransferId: transferId,
	}))
	if err != nil {
		return op, err
	}
	err = op.Wait(ctx)
	return op, err
}

func deactivateTransfer(ctx context.Context, dc *dc.SDK, transferId string) (*operation.Operation, error) {
	op, err := dc.WrapOperation(dc.Transfer().Transfer().Deactivate(ctx, &transfer.DeactivateTransferRequest{
		TransferId: transferId,
	}))
	if err != nil {
		return op, err
	}
	err = op.Wait(ctx)
	return op, err
}

func deleteTransfer(ctx context.Context, dc *dc.SDK, transferId string) (*operation.Operation, error) {
	op, err := dc.WrapOperation(dc.Transfer().Transfer().Delete(ctx, &transfer.DeleteTransferRequest{
		TransferId: transferId,
	}))
	if err != nil {
		return op, err
	}
	err = op.Wait(ctx)
	return op, err
}

func main() {
	flags := parseCmd()
	ctx := context.Background()

	key, err := iamkey.ReadFromJSONFile(*flags.saPath)
	if err != nil {
		panic(err)
	}
	creds, err := dc.ServiceAccountKey(key)
	if err != nil {
		panic(err)
	}

	sdk, err := dc.Build(ctx, dc.Config{
		Credentials: creds,
	})
	if err != nil {
		log.Fatal(err)
	}

	op, err := createS3SourceEndpoint(ctx, sdk, flags)
	if err != nil {
		log.Panic(err, "Failed to create s3 source endpoint")
	}
	srcEndpointId := op.ResourceId()

	log.Println("Created s3 source endpoint: ", srcEndpointId)
	op, err = createCHDstEndpoint(ctx, sdk, flags)
	if err != nil {
		log.Panic(err, "Failed to create s3 dst endpoint")
	}
	dstEndpointId := op.ResourceId()
	log.Println("Created ClickHouse destination endpoint: ", dstEndpointId)

	op, err = createTransfer(ctx, sdk, flags, srcEndpointId, dstEndpointId)
	if err != nil {
		log.Panic(err, "Failed to create transfer")
	}
	transferId := op.ResourceId()

	log.Println("Wonderful! 🚀 Check out created transfer\n\thttps://app.double.cloud/data-transfer/" + transferId)

	log.Println("Activating transfer", transferId)
	op, err = activateTransfer(ctx, sdk, transferId)
	if err != nil {
		log.Println("Activate failed, because we specified unexisted cluster, nothing to look here")
	}

	log.Println("Press F to respect and delete all created resources ...")
	fmt.Scanln()

	log.Println("Deactivating transfer", transferId)
	op, err = deactivateTransfer(ctx, sdk, transferId)
	if err != nil {
		log.Panic(err, "Failed to activate")
	}

	log.Println("Deleting transfer", transferId)
	op, err = deleteTransfer(ctx, sdk, transferId)
	if err != nil {
		log.Panic(err, "Failed to delete transfer")
	}

	log.Println("Deleting s3 endpoint", srcEndpointId)
	op, err = deleteEndpoint(ctx, sdk, srcEndpointId)
	if err != nil {
		log.Panic(err, "Failed to delete s3 source endpoint")
	}
	log.Println("Deleting clickhouse destination endpoint", dstEndpointId)
	op, err = deleteEndpoint(ctx, sdk, dstEndpointId)
	if err != nil {
		log.Panic(err, "Failed to delete clickhouse destination endpoint")
	}
}

type cmdFlags struct {
	saPath    *string
	projectID *string
	name      *string
}

func parseCmd() (ret *cmdFlags) {
	ret = &cmdFlags{}

	ret.saPath = flag.String("saPath", "authorized_key.json", "Path to the service account key JSON file.\nThis file can be created using UI:\n"+
		"Members -> Service Accounts -> Create and then create authorized keys")
	ret.projectID = flag.String("projectID", "", "Your project id")
	ret.name = flag.String("name", "go-example", "Name for your service")

	flag.Parse()
	return
}
