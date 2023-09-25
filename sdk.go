package dcsdk

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"sort"
	"strings"
	"sync"
	"time"

	dcv1 "github.com/doublecloud/go-genproto/doublecloud/v1"
	"github.com/doublecloud/go-sdk/gen/clickhouse"
	"github.com/doublecloud/go-sdk/gen/kafka"
	"github.com/doublecloud/go-sdk/gen/network"
	"github.com/doublecloud/go-sdk/gen/transfer"
	"github.com/doublecloud/go-sdk/gen/visualization"
	"github.com/doublecloud/go-sdk/iamkey"
	"github.com/doublecloud/go-sdk/operation"
	"github.com/doublecloud/go-sdk/pkg/grpcclient"
	"github.com/doublecloud/go-sdk/pkg/sdkerrors"
	"github.com/google/uuid"
	"golang.org/x/sync/singleflight"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Endpoint string
type APIEndpoint struct {
	Id      Endpoint
	Address string
}

const (
	DefaultPageSize int64 = 1000

	ClickHouseServiceID    Endpoint = "clickhouse"
	KafkaServiceID         Endpoint = "kafka"
	VpcServiceID           Endpoint = "vpc"
	TransferServiceID      Endpoint = "transfer"
	VisualizationServiceID Endpoint = "visualization"
)

// Config is a config that is used to create SDK instance.
type Config struct {
	// Credentials are used to authenticate the client. See Credentials for more info.
	Credentials Credentials
	// DialContextTimeout specifies timeout of dial on API endpoint that
	// is used when building an SDK instance.
	// DialContextTimeout time.Duration
	// TLSConfig is optional tls.Config that one can use in order to tune TLS options.
	TLSConfig *tls.Config

	// Endpoint is an API endpoint of DoubleCloud against which the SDK is used.
	// Most users won't need to explicitly set it.
	Endpoint         string
	OverrideEndpoint bool
	Plaintext        bool
}

// SDK is a DoubleCloud SDK
type SDK struct {
	conf      Config
	cc        grpcclient.ConnContext
	endpoints struct {
		initDone bool
		mu       sync.Mutex
		ep       map[Endpoint]*APIEndpoint
	}

	initErr  error
	initCall singleflight.Group
	muErr    sync.Mutex
}

// Build creates an SDK instance
func Build(ctx context.Context, conf Config, customOpts ...grpc.DialOption) (*SDK, error) {
	if conf.Credentials == nil {
		return nil, errors.New("credentials required")
	}

	const defaultEndpoint = "api.double.cloud:443"
	if conf.Endpoint == "" {
		conf.Endpoint = defaultEndpoint
	}
	const DefaultTimeout = 20 * time.Second

	switch creds := conf.Credentials.(type) {
	case ExchangeableCredentials, NonExchangeableCredentials:
	default:
		return nil, fmt.Errorf("unsupported credentials type %T", creds)
	}
	sdk := &SDK{
		cc:   nil, // Later
		conf: conf,
	}
	tokenMiddleware := NewIAMTokenMiddleware(sdk, now)
	var dialOpts []grpc.DialOption
	dialOpts = append(dialOpts,
		grpc.WithChainUnaryInterceptor(tokenMiddleware.InterceptUnary),
		grpc.WithChainStreamInterceptor(tokenMiddleware.InterceptStream),
	)

	if conf.Plaintext {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		tlsConfig := conf.TLSConfig
		if tlsConfig == nil {
			tlsConfig = &tls.Config{}
		}
		creds := credentials.NewTLS(tlsConfig)
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(creds))
	}
	// Append custom options after default, to allow to customize dialer and etc.
	dialOpts = append(dialOpts, customOpts...)
	sdk.cc = grpcclient.NewLazyConnContext(grpcclient.DialOptions(dialOpts...))
	return sdk, nil
}

// Shutdown shutdowns SDK and closes all open connections.
func (sdk *SDK) Shutdown(ctx context.Context) error {
	return sdk.cc.Shutdown(ctx)
}

func (sdk *SDK) CheckEndpointConnection(ctx context.Context, endpoint Endpoint) error {
	_, err := sdk.getConn(endpoint)(ctx)
	return err
}

// WrapOperation wraps operation proto message to handy structure
func (sdk *SDK) WrapOperation(o *dcv1.Operation, err error) (*operation.Operation, error) {
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(o.Id, operation.CLICKHOUSE_OPERATION_PREFIX) {
		return operation.New(sdk.ClickHouse().Operation(), o), nil
	}
	if strings.HasPrefix(o.Id, operation.KAFKA_OPERATION_PREFIX) {
		return operation.New(sdk.Kafka().Operation(), o), nil
	}
	if strings.HasPrefix(o.Id, operation.TRANSFER_ENDPOINTS_OPERATION_PREFIX) || strings.HasPrefix(o.Id, operation.TRANSFER_OPERATION_PREFIX) {
		return operation.New(sdk.Transfer().Operation(), o), nil
	}
	if _, err := uuid.Parse(o.Id); err == nil {
		return operation.New(sdk.Network().Operation(), o), nil
	}
	return nil, sdkerrors.WithMessage(fmt.Errorf("opID: %q", o.Id), "Unknown operation type")
}

func (sdk *SDK) getConn(serviceID Endpoint) func(ctx context.Context) (*grpc.ClientConn, error) {
	return func(ctx context.Context) (*grpc.ClientConn, error) {
		if !sdk.initDone() {
			sdk.initCall.Do("init", func() (any, error) {
				sdk.muErr.Lock()
				sdk.initErr = sdk.initConns(ctx)
				sdk.muErr.Unlock()
				return nil, nil
			})
			if err := sdk.InitErr(); err != nil {
				return nil, err
			}
		}
		endpoint, endpointExist := sdk.Endpoint(serviceID)
		if !endpointExist {
			return nil, &ServiceIsNotAvailableError{
				ServiceID:           serviceID,
				APIEndpoint:         sdk.conf.Endpoint,
				availableServiceIDs: sdk.KnownServices(),
			}
		}
		return sdk.cc.GetConn(ctx, endpoint.Address)
	}
}

type ServiceIsNotAvailableError struct {
	ServiceID           Endpoint
	APIEndpoint         string
	availableServiceIDs []string
}

func (s *ServiceIsNotAvailableError) Error() string {
	return fmt.Sprintf("Service \"%v\" is not available at Cloud API endpoint \"%v\". Available services: %v",
		s.ServiceID,
		s.APIEndpoint,
		s.availableServiceIDs,
	)
}

var _ error = &ServiceIsNotAvailableError{}

func (sdk *SDK) initDone() (b bool) {
	sdk.endpoints.mu.Lock()
	b = sdk.endpoints.initDone
	sdk.endpoints.mu.Unlock()
	return
}

func (sdk *SDK) KnownServices() []string {
	sdk.endpoints.mu.Lock()
	result := make([]string, 0, len(sdk.endpoints.ep))
	for k := range sdk.endpoints.ep {
		result = append(result, string(k))
	}
	sdk.endpoints.mu.Unlock()
	sort.Strings(result)
	return result
}

func (sdk *SDK) Endpoint(endpointName Endpoint) (ep *APIEndpoint, exist bool) {
	sdk.endpoints.mu.Lock()
	defer sdk.endpoints.mu.Unlock()
	ep, exist = sdk.endpoints.ep[endpointName]
	return
}

func (sdk *SDK) InitErr() error {
	sdk.muErr.Lock()
	defer sdk.muErr.Unlock()
	return sdk.initErr
}

func endpointsMap(baseEndpoint string, overrideEndpoint bool) map[Endpoint]*APIEndpoint {
	m := make(map[Endpoint]*APIEndpoint)
	for _, v := range []Endpoint{ClickHouseServiceID, KafkaServiceID, VpcServiceID, TransferServiceID, VisualizationServiceID} {
		var endpoint string
		if overrideEndpoint {
			endpoint = baseEndpoint
		} else {
			endpoint = fmt.Sprintf("%v.%s", v, baseEndpoint)
		}
		m[v] = &APIEndpoint{
			Id:      v,
			Address: endpoint,
		}
	}
	return m
}

func (sdk *SDK) initConns(ctx context.Context) error {
	sdk.endpoints.mu.Lock()
	defer sdk.endpoints.mu.Unlock()
	sdk.endpoints.ep = endpointsMap(sdk.conf.Endpoint, sdk.conf.OverrideEndpoint)

	sdk.endpoints.initDone = true
	return nil
}

func exchangeJWT2IAM(ctx context.Context, request *iamkey.CreateIamTokenRequest) (*iamkey.CreateIamTokenResponse, error) {
	client := http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   time.Second, // One second should be enough for localhost connection.
				KeepAlive: -1,          // No keep alive. Near token per hour requested.
			}).DialContext,
		},
	}
	data := fmt.Sprintf("grant_type=urn:ietf:params:oauth:grant-type:jwt-bearer&assertion=%v", request.GetJwt())

	req, err := http.NewRequest("POST", "https://auth.double.cloud/oauth/token", strings.NewReader(data))
	if err != nil {
		return nil, sdkerrors.WithMessage(err, "request make failed")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// TODO: define user-agent version
	req.Header.Set("User-Agent", "doublecloud-go-sdk")
	reqDump, _ := httputil.DumpRequestOut(req, false)
	grpclog.Infof("Going to request SA token:\n%s", reqDump)
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, sdkerrors.WithMessage(err, "failed to exchange JWT")
	}
	defer resp.Body.Close()
	respDump, _ := httputil.DumpResponse(resp, false)
	grpclog.Infof("response (without body, because contains sensitive token):\n%s", respDump)

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("%s.\n"+
			"Is this IAM running using Service Account? That is, Instance.service_account_id should not be empty.",
			resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		if err != nil {
			body = []byte(fmt.Sprintf("Failed response body read failed: %s", err.Error()))
		}
		grpclog.Errorf("IAM service token get failed: %s. Body:\n%s", resp.Status, body)
		return nil, fmt.Errorf("%s", resp.Status)
	}
	if err != nil {
		return nil, fmt.Errorf("reponse read failed: %s", err)
	}

	var tokenResponse struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}

	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		grpclog.Errorf("Failed to unmarshal SA token response body.\nError: %s\nBody:\n%s", err, body)
		return nil, sdkerrors.WithMessage(err, "body unmarshal failed")
	}
	expiresAt := timestamppb.Now()
	expiresAt.Seconds += tokenResponse.ExpiresIn - 1
	expiresAt.Nanos = 0 // Truncate is for readability.
	return &iamkey.CreateIamTokenResponse{
		IamToken:  tokenResponse.AccessToken,
		ExpiresAt: expiresAt,
	}, nil
}

func (sdk *SDK) CreateIAMToken(ctx context.Context) (*iamkey.CreateIamTokenResponse, error) {
	creds := sdk.conf.Credentials
	switch creds := creds.(type) {
	case ExchangeableCredentials:
		req, err := creds.IAMTokenRequest()
		if err != nil {
			return nil, sdkerrors.WithMessage(err, "IAM token request build failed")
		}
		return exchangeJWT2IAM(ctx, req)
	case NonExchangeableCredentials:
		return creds.IAMToken(ctx)
	default:
		return nil, fmt.Errorf("credentials type %T is not supported yet", creds)
	}
}

func (sdk *SDK) CreateIAMTokenForServiceAccount(ctx context.Context, serviceAccountID string) (*iamkey.CreateIamTokenResponse, error) {
	// Later
	return nil, nil
}

var now = time.Now

func (sdk *SDK) Kafka() *kafka.Kafka {
	return kafka.NewKafka(sdk.getConn(KafkaServiceID))
}

func (sdk *SDK) Network() *network.Network {
	return network.NewNetwork(sdk.getConn(VpcServiceID))
}

func (sdk *SDK) ClickHouse() *clickhouse.ClickHouse {
	return clickhouse.NewClickHouse(sdk.getConn(ClickHouseServiceID))
}

func (sdk *SDK) Transfer() *transfer.Transfer {
	return transfer.NewTransfer(sdk.getConn(TransferServiceID))
}

func (sdk *SDK) Visualization() *visualization.Visualization {
	return visualization.NewVisualization(sdk.getConn(VisualizationServiceID))
}
