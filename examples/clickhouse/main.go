package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/doublecloud/go-genproto/doublecloud/clickhouse/v1"
	dc "github.com/doublecloud/go-sdk"
	"github.com/doublecloud/go-sdk/iamkey"
	"github.com/doublecloud/go-sdk/operation"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func createCluster(ctx context.Context, dc *dc.SDK, flags *cmdFlags) (*operation.Operation, error) {
	x, err := dc.ClickHouse().Cluster().Create(ctx, &clickhouse.CreateClusterRequest{
		ProjectId: *flags.projectID,
		CloudType: "aws",
		RegionId:  *flags.region,
		Name:      *flags.name,
		Resources: &clickhouse.ClusterResources{
			Clickhouse: &clickhouse.ClusterResources_Clickhouse{
				ResourcePresetId: "s2-c2-m4",
				DiskSize:         wrapperspb.Int64(32 * 2 << 30),
				ReplicaCount:     wrapperspb.Int64(1),
			},
		},
		NetworkId: *flags.networkID,
	})
	if err != nil {
		return nil, err
	}
	log.Println("Creating clickhouse cluster ...")
	log.Println("https://app.double.cloud/clickhouse/" + x.ResourceId + "/operations")
	op, err := dc.WrapOperation(x, err)
	if err != nil {
		panic(err)
	}
	err = op.Wait(ctx)
	return op, err
}

func deleteCluster(ctx context.Context, dc *dc.SDK, clusterID string) (*operation.Operation, error) {
	op, err := dc.WrapOperation(dc.ClickHouse().Cluster().Delete(ctx, &clickhouse.DeleteClusterRequest{ClusterId: clusterID}))
	if err != nil {
		log.Fatal(err)
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

	op, err := createCluster(ctx, sdk, flags)
	if err != nil {
		log.Panic(err, "Failed to create cluster")
	}
	clusterID := op.ResourceId()

	log.Println("Wonderful! 🚀 Check out created cluster\n\thttps://app.double.cloud/clickhouse/" + clusterID)

	log.Println("Press F to respect and delete all created resources ...")
	fmt.Scanln()

	log.Println("Deleting cluster", clusterID)
	op, err = deleteCluster(ctx, sdk, clusterID)
	if err != nil {
		log.Panic(err, "Failed to delete cluster")
	}
}

type cmdFlags struct {
	saPath    *string
	projectID *string
	region    *string
	name      *string
	networkID *string
}

func parseCmd() (ret *cmdFlags) {
	ret = &cmdFlags{}

	ret.saPath = flag.String("saPath", "authorized_key.json", "Path to the service account key JSON file.\nThis file can be created using UI:\n"+
		"Members -> Service Accounts -> Create and then create authorized keys")
	ret.projectID = flag.String("projectID", "", "Your project id")
	ret.name = flag.String("name", "go-example", "Name for your service")
	ret.region = flag.String("region", "eu-central-1", "Region to deploy to.")
	ret.networkID = flag.String("networkID", "10.0.0.0/16", "Network of the cluster.")

	flag.Parse()
	return
}
