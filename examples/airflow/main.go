package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/doublecloud/go-genproto/doublecloud/airflow/v1"
	dc "github.com/doublecloud/go-sdk"
	"github.com/doublecloud/go-sdk/iamkey"
	"github.com/doublecloud/go-sdk/operation"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func createCluster(ctx context.Context, dc *dc.SDK, flags *cmdFlags) (*operation.Operation, error) {
	x, err := dc.Airflow().Cluster().Create(ctx, &airflow.CreateClusterRequest{
		ProjectId: *flags.projectID,
		CloudType: "aws",
		RegionId:  *flags.region,
		Name:      *flags.name,
		Resources: &airflow.ClusterResources{
			Airflow: &airflow.ClusterResources_Airflow{
				MaxWorkerCount:    wrapperspb.Int64(1),
				EnvironmentFlavor: "dev_test",
				MinWorkerCount:    wrapperspb.Int64(1),
				WorkerConcurrency: wrapperspb.Int64(16),
				WorkerDiskSize:    wrapperspb.Int64(10),
				WorkerPreset:      "small",
			},
		},
		Config: &airflow.CreateClusterRequest_AirflowConfig{
			VersionId: "2.9.0",
			GitSync: &airflow.SyncConfig{
				RepoUrl:  "https://github.com/apache/airflow",
				Branch:   "main",
				DagsPath: "airflow/example_dags",
			},
		},
		NetworkId: *flags.networkID,
	})
	if err != nil {
		return nil, err
	}

	log.Println("Creating airflow cluster ...")
	log.Println("https:://app.double.cloud/airflow/" + x.ResourceId + "/operations")
	op, err := dc.WrapOperation(x, err)
	if err != nil {
		panic(err)
	}
	err = op.Wait(ctx)
	return op, err
}

func deleteCluster(ctx context.Context, dc *dc.SDK, clusterID string) (*operation.Operation, error) {
	op, err := dc.WrapOperation(dc.Airflow().Cluster().Delete(ctx, &airflow.DeleteClusterRequest{ClusterId: clusterID}))
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

	log.Println("Wonderful! 🚀 Check out created cluster\n\thttps://app.double.cloud/airflow/" + clusterID)

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
	ret.projectID = flag.String("projectID", "mdb-junk", "Your project id")
	ret.name = flag.String("name", "go-example-airflow", "Name for your service")
	ret.region = flag.String("region", "eu-central-1", "Region to deploy to.")
	ret.networkID = flag.String("networkID", "23ad0fcc-0a40-4329-868d-bb5caec0f90a", "Network of the cluster.")

	flag.Parse()
	return
}
