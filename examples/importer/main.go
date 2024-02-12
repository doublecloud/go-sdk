package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/doublecloud/go-genproto/doublecloud/transfer/v1"
	"github.com/doublecloud/go-genproto/doublecloud/transfer/v1/endpoint"
	airbyte "github.com/doublecloud/go-genproto/doublecloud/transfer/v1/endpoint/airbyte"
	dc "github.com/doublecloud/go-sdk"
	"github.com/doublecloud/go-sdk/iamkey"
	"github.com/doublecloud/go-sdk/operation"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"strings"
)

type cmdFlags struct {
	saPath          *string
	projectID       *string
	targetClusterID *string
	configPath      *string
}

// importConfig structure of user defined import data configuration
type importConfig struct {
	Sources []struct {
		Table          string            `yaml:"table"`
		TableNamespace string            `yaml:"table_namespace"`
		ColumnMapping  map[string]string `yaml:"column_mapping"`
		Type           string            `yaml:"type"`
		Conf           any               `yaml:"conf"`
	} `yaml:"sources"`
}

// sourceEndpoint define endpoint ID with required mapper parameters
type sourceEndpoint struct {
	ID             string
	Type           string
	Table          string
	TableNamespace string
	ColumnMapping  map[string]string
}

func mapJson(src any, dst protoreflect.ProtoMessage) error {
	raw, err := json.Marshal(src)
	if err != nil {
		return err
	}
	if err := protojson.Unmarshal(raw, dst); err != nil {
		return err
	}
	return nil
}

func createTargetEndpoint(ctx context.Context, dc *dc.SDK, flags *cmdFlags) (*operation.Operation, error) {
	op, err := dc.WrapOperation(dc.Transfer().Endpoint().Create(ctx, &transfer.CreateEndpointRequest{
		ProjectId: *flags.projectID,
		Name:      "dwh-import-target",
		Settings: &transfer.EndpointSettings{
			Settings: &transfer.EndpointSettings_ClickhouseTarget{
				ClickhouseTarget: &endpoint.ClickhouseTarget{
					Connection: &endpoint.ClickhouseConnection{
						Connection: &endpoint.ClickhouseConnection_ConnectionOptions{
							ConnectionOptions: &endpoint.ClickhouseConnectionOptions{
								Address:  &endpoint.ClickhouseConnectionOptions_MdbClusterId{MdbClusterId: *flags.targetClusterID},
								Database: "default",
								User:     "admin",
								Password: &endpoint.Secret{Value: &endpoint.Secret_Raw{Raw: ""}}, // will be auto inferred
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

func createSourceEndpoint(ctx context.Context, dc *dc.SDK, flags *cmdFlags) ([]sourceEndpoint, error) {
	yamlFile, err := ioutil.ReadFile(*flags.configPath)
	if err != nil {
		return nil, err
	}

	var config importConfig

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		panic(err)
	}
	var endpoints []sourceEndpoint
	for _, source := range config.Sources {
		settings := &transfer.EndpointSettings{}
		switch source.Type {
		case "bigquery":
			model := new(airbyte.BigQuerySource)
			if err := mapJson(source.Conf, model); err != nil {
				return nil, err
			}
			settings.Settings = &transfer.EndpointSettings_BigQuerySource{BigQuerySource: model}
		case "redshift":
			model := new(airbyte.RedshiftSource)
			if err := mapJson(source.Conf, model); err != nil {
				return nil, err
			}
			settings.Settings = &transfer.EndpointSettings_RedshiftSource{RedshiftSource: model}
		case "postgres":
			model := new(endpoint.PostgresSource)
			if err := mapJson(source.Conf, model); err != nil {
				return nil, err
			}
			settings.Settings = &transfer.EndpointSettings_PostgresSource{PostgresSource: model}
		}

		op, err := dc.WrapOperation(dc.Transfer().Endpoint().Create(ctx, &transfer.CreateEndpointRequest{
			ProjectId: *flags.projectID,
			Name:      fmt.Sprintf("dwh-%s-source", source.Type),
			Settings:  settings,
		}))
		if err != nil {
			return nil, err
		}
		err = op.Wait(ctx)
		if err != nil {
			return nil, err
		}
		endpoints = append(endpoints, sourceEndpoint{
			ID:             op.ResourceId(),
			Type:           source.Type,
			Table:          source.Table,
			TableNamespace: source.TableNamespace,
			ColumnMapping:  source.ColumnMapping,
		})
	}
	return endpoints, err
}

func createTransfer(
	ctx context.Context,
	dc *dc.SDK,
	flags *cmdFlags,
	src sourceEndpoint,
	dst string,
) (*operation.Operation, error) {
	dataObject := fmt.Sprintf("%s.%s", src.TableNamespace, src.Table)
	if src.TableNamespace == "" {
		dataObject = src.Table
	}
	var queryCols []string
	for k, v := range src.ColumnMapping {
		queryCols = append(queryCols, fmt.Sprintf("%s as %s", v, k))
	}
	op, err := dc.WrapOperation(dc.Transfer().Transfer().Create(ctx, &transfer.CreateTransferRequest{
		SourceId:  src.ID,
		TargetId:  dst,
		Name:      fmt.Sprintf("dwh-import-from-%s", src.Type),
		ProjectId: *flags.projectID,
		Type:      transfer.TransferType_SNAPSHOT_ONLY,
		RegularSnapshot: &transfer.RegularSnapshot{
			Mode: &transfer.RegularSnapshot_Disabled{Disabled: new(transfer.RegularSnapshotDisabled)},
		},
		Transformation: &transfer.Transformation{Transformers: []*transfer.Transformer{
			{
				Transformer: &transfer.Transformer_RenameTables{
					RenameTables: &transfer.RenameTablesTransformer{
						RenameTables: []*transfer.RenameTable{
							{
								OriginalName: &transfer.Table{
									NameSpace: src.TableNamespace,
									Name:      src.Table,
								},
								NewName: &transfer.Table{
									NameSpace: "",
									Name:      "aggregated_table",
								},
							},
						},
					},
				},
			},
			{
				Transformer: &transfer.Transformer_Sql{
					Sql: &transfer.SQLTransformer{
						Tables: nil,
						Query:  fmt.Sprintf(`select %s from table`, strings.Join(queryCols, ",")),
					},
				},
			},
		}},
		DataObjects: &transfer.DataObjects{IncludeObjects: []string{
			dataObject,
		}},
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

	endpoints, err := createSourceEndpoint(ctx, sdk, flags)
	if err != nil {
		log.Panic(err, "Failed to create source endpoint")
	}
	op, err := createTargetEndpoint(ctx, sdk, flags)
	if err != nil {
		log.Panic(err, "Failed to create s3 dst endpoint")
	}
	dstEndpointId := op.ResourceId()
	log.Println("Created ClickHouse destination endpoint: ", dstEndpointId)

	for _, source := range endpoints {
		log.Println("Created source endpoint: ", source.ID)

		op, err = createTransfer(
			ctx,
			sdk,
			flags,
			source,
			dstEndpointId,
		)
		if err != nil {
			log.Panic(err, "Failed to create transfer")
		}
		transferId := op.ResourceId()

		log.Println("Wonderful! 🚀 Check out created transfer\n\thttps://app.double.cloud/data-transfer/" + transferId)
	}
}

func parseCmd() (ret *cmdFlags) {
	ret = &cmdFlags{}

	ret.saPath = flag.String("sa-path", "authorized_key.json", "Path to the service account key JSON file.\nThis file can be created using UI:\n"+
		"Members -> Service Accounts -> Create and then create authorized keys")
	ret.projectID = flag.String("project-id", "", "Your project id")
	ret.targetClusterID = flag.String("name", "go-example", "Name for your service")
	ret.configPath = flag.String("import-config-path", "examples/importer/import.yaml", "Path to import config file.\nThis file contains all user define configuration with corresponding mappings")

	flag.Parse()
	return
}
