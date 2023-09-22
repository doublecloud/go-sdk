package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/doublecloud/go-genproto/doublecloud/network/v1"
	dc "github.com/doublecloud/go-sdk"
	"github.com/doublecloud/go-sdk/iamkey"
	"github.com/doublecloud/go-sdk/operation"
)

func createNetwork(ctx context.Context, dc *dc.SDK, flags *cmdFlags) (*operation.Operation, error) {
	x, err := dc.Network().Network().Create(ctx, &network.CreateNetworkRequest{
		ProjectId:     *flags.projectID,
		Name:          *flags.name,
		CloudType:     *flags.cloudType,
		RegionId:      *flags.region,
		Ipv4CidrBlock: *flags.ipv4CIDR,
	})
	if err != nil {
		return nil, err
	}
	log.Println("Creating network ...")
	op, err := dc.WrapOperation(x, err)
	if err != nil {
		panic(err)
	}
	err = op.Wait(ctx)
	return op, err
}

func deleteNetwork(ctx context.Context, dc *dc.SDK, networkId string) (*operation.Operation, error) {
	op, err := dc.WrapOperation(dc.Network().Network().Delete(ctx, &network.DeleteNetworkRequest{NetworkId: networkId}))
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

	op, err := createNetwork(ctx, sdk, flags)
	if err != nil {
		log.Panic(err, "Failed to create network")
	}
	networkId := op.ResourceId()

	log.Println("Wonderful! 🚀 Check out created network\n\thttps://app.double.cloud/vpc/network/" + networkId)

	log.Println("Press F to respect and delete all created resources ...")
	fmt.Scanln()

	log.Println("Deleting network", networkId)
	op, err = deleteNetwork(ctx, sdk, networkId)
	if err != nil {
		log.Panic(err, "Failed to delete network")
	}

}

type cmdFlags struct {
	saPath    *string
	projectID *string
	cloudType *string
	region    *string
	name      *string
	ipv4CIDR  *string
}

func parseCmd() (ret *cmdFlags) {
	ret = &cmdFlags{}

	ret.saPath = flag.String("saPath", "authorized_key.json", "Path to the service account key JSON file.\nThis file can be created using UI:\n"+
		"Members -> Service Accounts -> Create and then create authorized keys")
	ret.projectID = flag.String("projectID", "", "Your project id")
	ret.name = flag.String("name", "go-example", "Name for your service")
	ret.cloudType = flag.String("cloudType", "aws", "One of cloud: aws or gcp")
	ret.region = flag.String("region", "eu-central-1", "Region to deploy to.")
	ret.ipv4CIDR = flag.String("ipv4CIDR", "10.0.0.0/16", "IPv4 block for network")

	flag.Parse()
	return
}
