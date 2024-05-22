package main

import (
	"context"
	"flag"
	"fmt"
	dcsdk "github.com/doublecloud/go-sdk"
)

var (
	federationID = flag.String("federation-id", "", "ID of federation to authorize")
)

func main() {
	flag.Parse()
	if *federationID == "" {
		panic("federation-id is required")
	}
	creds := dcsdk.NewFederationCredentials(&dcsdk.FederationConfig{FederationID: *federationID})
	iamToken, err := creds.IAMToken(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println(iamToken.IamToken)
}
