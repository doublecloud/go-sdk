package main

import (
	"context"
	"flag"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"log"
	"os"

	"github.com/doublecloud/go-genproto/doublecloud/visualization/v1"
	dc "github.com/doublecloud/go-sdk"
	"github.com/doublecloud/go-sdk/iamkey"
)

var (
	saPath = flag.String(
		"sa-path",
		"sa.json",
		`Path to the service account key JSON file.
			This file can be created using UI:
				Members -> Service Accounts -> Create and then create authorized keys`,
	)
	workbookID   = flag.String("workbook-id", "", "ID of your workbood to load config")
	workbookPath = flag.String("workbook-path", "workbook.json", "Path to workbook json file, for load-upload, depends on direction")
	direction    = flag.String("direction", "download", "Either download / upload. Direction for workbook")
)

func main() {
	flag.Parse()
	ctx := context.Background()
	key, err := iamkey.ReadFromJSONFile(*saPath)
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
		panic(err)
	}
	switch *direction {
	case "download":
		res, err := sdk.Visualization().Workbook().Get(ctx, &visualization.GetWorkbookRequest{
			WorkbookId: *workbookID,
		})
		if err != nil {
			panic(err)
		}
		log.Println(res.Title)
		dd, _ := res.Workbook.Config.MarshalJSON()
		if err := os.WriteFile(*workbookPath, dd, 0600); err != nil {
			panic(err)
		}
	case "upload":
		data, err := os.ReadFile(*workbookPath)
		if err != nil {
			panic(err)
		}
		strpb := structpb.NewNullValue()
		if err := strpb.UnmarshalJSON(data); err != nil {
			panic(err)
		}
		_, err = sdk.Visualization().Workbook().Update(ctx, &visualization.UpdateWorkbookRequest{
			WorkbookId:   *workbookID,
			Workbook:     &visualization.Workbook{Config: strpb},
			ForceRewrite: &wrapperspb.BoolValue{Value: true},
		})
		if err != nil {
			panic(err)
		}
	default:
		panic("unknown direction " + *direction)
	}
}
