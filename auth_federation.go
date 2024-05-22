package dcsdk

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/doublecloud/go-sdk/iamkey"
	"github.com/doublecloud/go-sdk/pkg/browser"
	"github.com/doublecloud/go-sdk/pkg/server"
	"net/url"
	"os"
	"path"
	"time"
)

const beforeBrowserOpenNote = `
You are going to be authenticated via federation-id '%s'.
Your federation authentication web site will be opened.
After your successful authentication, you will be redirected to '%s'.
Press Enter

`

func federationURL(federationID, federationEndpoint, federationPath, redirectURL string) (string, error) {
	baseURL, err := url.Parse(federationEndpoint)
	if err != nil {
		return "", err
	}

	baseURL.Path = path.Join(baseURL.Path, federationPath, federationID)

	params := url.Values{}
	params.Add("redirectUrl", redirectURL)

	baseURL.RawQuery = params.Encode() // Escape Query Parameters

	return baseURL.String(), nil
}

type FederationConfig struct {
	// Token store temporary IAM token in memory, so browser will not open too often
	cachedToken *iamkey.CreateIamTokenResponse

	FederationID string

	// Endpoint is an API endpoint of DoubleCloud against which the SDK authorize federated user.
	// Most users won't need to explicitly set it.
	// Default value: https://auth.double.cloud
	FederationEndpoint string
	// FederationPath is a path where authorizer API located
	// Most users won't need to explicitly set it.
	// Default value: federations
	FederationPath string
	// TokenCachePath where to store IAM token between SDK-calls
	// Default value: ~/.dcsdk
	TokenCachePath    string
	DisableTokenCache bool
}

type FederationCredentials struct {
	cfg *FederationConfig
}

func (ft *FederationCredentials) urlRequest(serverAddr string) error {
	requestURL, err := federationURL(
		ft.cfg.FederationID,
		ft.cfg.FederationEndpoint,
		ft.cfg.FederationPath,
		fmt.Sprintf("http://%s", serverAddr))
	if err != nil {
		return err
	}

	if err := browser.OpenURL(requestURL); err != nil {
		return err
	}

	return nil
}

func (ft *FederationCredentials) consoleURL() (string, error) {
	endpointURL, err := url.Parse(ft.cfg.FederationEndpoint)
	if err != nil {
		return "", err
	}
	endpointURL.Path = ""
	return endpointURL.String(), nil
}

func (ft *FederationCredentials) DCAPICredentials() {}

func (ft *FederationCredentials) IAMToken(ctx context.Context) (*iamkey.CreateIamTokenResponse, error) {
	oldToken := ft.cfg.cachedToken
	if oldToken != nil && oldToken.ExpiresAt.AsTime().After(time.Now()) {
		return &iamkey.CreateIamTokenResponse{
			IamToken:  oldToken.IamToken,
			ExpiresAt: oldToken.ExpiresAt,
		}, nil
	}

	consoleURL, err := ft.consoleURL()
	if err != nil {
		return nil, err
	}
	fmt.Printf(beforeBrowserOpenNote, ft.cfg.FederationID, consoleURL)
	token, err := server.GetToken(ctx, ft.urlRequest, consoleURL)
	if err != nil {
		return nil, err
	}

	ft.cfg.cachedToken = &iamkey.CreateIamTokenResponse{
		IamToken:  token.IamToken,
		ExpiresAt: token.ExpiresAt,
	}
	tokenData, err := json.Marshal(ft.cfg.cachedToken)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(
		ft.cfg.TokenCachePath,
		0660,
	); err != nil && !os.IsExist(err) {
		return nil, err
	}

	if err := os.WriteFile(
		fmt.Sprintf(
			"%s/%s.json",
			ft.cfg.TokenCachePath,
			ft.cfg.FederationID,
		),
		tokenData,
		0660,
	); err != nil {
		return nil, err
	}

	fmt.Printf("Federation successfully finished, token will expire at %s\n", token.ExpiresAt.AsTime())
	iamToken := &iamkey.CreateIamTokenResponse{
		IamToken:  token.IamToken,
		ExpiresAt: token.ExpiresAt,
	}

	return iamToken, nil
}

func NewFederationCredentials(cfg *FederationConfig) NonExchangeableCredentials {
	if cfg.FederationEndpoint == "" {
		cfg.FederationEndpoint = "https://auth.double.cloud"
	}
	if cfg.FederationPath == "" {
		cfg.FederationPath = "federations"
	}
	if cfg.TokenCachePath == "" {
		home, _ := os.UserHomeDir()
		cfg.TokenCachePath = fmt.Sprintf("%s/.dcsdk", home)
	}
	if !cfg.DisableTokenCache {
		cachedToken, err := os.ReadFile(fmt.Sprintf("%s/%s.json", cfg.TokenCachePath, cfg.FederationID))
		if err != nil {
			return &FederationCredentials{
				cfg: cfg,
			}
		}
		if err := json.Unmarshal(cachedToken, &cfg.cachedToken); err != nil {
			return &FederationCredentials{
				cfg: cfg,
			}
		}
	}
	return &FederationCredentials{
		cfg: cfg,
	}
}
