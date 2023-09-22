package dcsdk

import (
	"context"
	"crypto"
	"crypto/rsa"
	"errors"
	"fmt" //nolint:staticcheck
	"time"

	"github.com/doublecloud/go-sdk/iamkey"
	"github.com/doublecloud/go-sdk/pkg/sdkerrors"
	jwt "github.com/golang-jwt/jwt/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Credentials is an abstraction of API authorization credentials.
// See https://double.cloud/docs/en/public-api/get-iam-token for details.
// Note that functions that return Credentials may return different Credentials implementation
// in next SDK version, and this is not considered breaking change.
type Credentials interface {
	DCAPICredentials()
}

// ExchangeableCredentials can be exchanged for IAM Token in IAM Token Service, that can be used
// to authorize API calls.
// See https://double.cloud/docs/en/public-api/get-iam-token for details.
type ExchangeableCredentials interface {
	Credentials
	// IAMTokenRequest returns request for fresh IAM token or error.
	IAMTokenRequest() (*iamkey.CreateIamTokenRequest, error)
	// NonExchangeableCredentials
	// IAMToken(ctx context.Context) (*iamkey.CreateIamTokenResponse, error)
}

// NonExchangeableCredentials allows to get IAM Token without calling IAM Token Service.
type NonExchangeableCredentials interface {
	Credentials
	// IAMToken returns IAM Token.
	IAMToken(ctx context.Context) (*iamkey.CreateIamTokenResponse, error)
}

// ServiceAccountKey returns credentials for the given IAM Key. The key is used to sign JWT tokens.
// JWT tokens are exchanged for IAM Tokens used to authorize API calls.
// This authorization method is not supported for IAM Keys issued for User Accounts.
func ServiceAccountKey(key *iamkey.Key) (Credentials, error) {
	jwtBuilder, err := newServiceAccountJWTBuilder(key)
	if err != nil {
		return nil, err
	}
	return exchangeableCredentialsFunc(func() (*iamkey.CreateIamTokenRequest, error) {
		signedJWT, err := jwtBuilder.SignedToken()
		if err != nil {
			return nil, sdkerrors.WithMessage(err, "JWT sign failed")
		}
		return &iamkey.CreateIamTokenRequest{
			Identity: &iamkey.CreateIamTokenRequest_Jwt{
				Jwt: signedJWT,
			},
		}, nil
	}), nil
}

func newServiceAccountJWTBuilder(key *iamkey.Key) (*serviceAccountJWTBuilder, error) {
	err := validateServiceAccountKey(key)
	if err != nil {
		return nil, sdkerrors.WithMessage(err, "key validation failed")
	}
	rsaPrivateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(key.PrivateKey))
	if err != nil {
		return nil, sdkerrors.WithMessage(err, "private key parsing failed")
	}
	return &serviceAccountJWTBuilder{
		key:           key,
		rsaPrivateKey: rsaPrivateKey,
	}, nil
}

func validateServiceAccountKey(key *iamkey.Key) error {
	if key.Id == "" {
		return errors.New("key id is missing")
	}
	if key.GetServiceAccountId() == "" {
		return fmt.Errorf("key should de issued for service account, but subject is %#v", key.Subject)
	}
	return nil
}

type serviceAccountJWTBuilder struct {
	key           *iamkey.Key
	rsaPrivateKey *rsa.PrivateKey
}

func (b *serviceAccountJWTBuilder) SignedToken() (string, error) {
	return b.issueToken().SignedString(b.rsaPrivateKey)
}

func (b *serviceAccountJWTBuilder) issueToken() *jwt.Token {
	issuedAt := time.Now()
	token := jwt.NewWithClaims(jwtSigningMethodPS256WithSaltLengthEqualsHash, jwt.RegisteredClaims{
		Issuer:    b.key.GetServiceAccountId(),
		Subject:   b.key.GetServiceAccountId(),
		IssuedAt:  jwt.NewNumericDate(issuedAt),
		ExpiresAt: jwt.NewNumericDate(issuedAt.Add(time.Hour)),
		Audience:  jwt.ClaimStrings{"https://auth.double.cloud/oauth/token"},
	})
	token.Header["kid"] = b.key.Id
	return token
}

// Should be removed after https://github.com/dgrijalva/jwt-go/issues/285 fix.
var jwtSigningMethodPS256WithSaltLengthEqualsHash = &jwt.SigningMethodRSAPSS{
	SigningMethodRSA: jwt.SigningMethodPS256.SigningMethodRSA,
	Options: &rsa.PSSOptions{
		Hash:       crypto.SHA256,
		SaltLength: rsa.PSSSaltLengthEqualsHash,
	},
}

type exchangeableCredentialsFunc func() (iamTokenReq *iamkey.CreateIamTokenRequest, err error)

var _ ExchangeableCredentials = (exchangeableCredentialsFunc)(nil)

func (exchangeableCredentialsFunc) DCAPICredentials() {}

func (f exchangeableCredentialsFunc) IAMTokenRequest() (iamTokenReq *iamkey.CreateIamTokenRequest, err error) {
	return f()
}

// NoCredentials implements Credentials, it allows to create unauthenticated connections
type NoCredentials struct{}

func (creds NoCredentials) DCAPICredentials() {}

// IAMToken always returns gRPC error with status UNAUTHENTICATED
func (creds NoCredentials) IAMToken(ctx context.Context) (*iamkey.CreateIamTokenResponse, error) {
	return nil, status.Error(codes.Unauthenticated, "unauthenticated connection")
}

// IAMTokenCredentials implements Credentials with IAM token as-is
type IAMTokenCredentials struct {
	iamToken string
}

func (creds IAMTokenCredentials) DCAPICredentials() {}

func (creds IAMTokenCredentials) IAMToken(ctx context.Context) (*iamkey.CreateIamTokenResponse, error) {
	return &iamkey.CreateIamTokenResponse{
		IamToken: creds.iamToken,
	}, nil
}

func NewIAMTokenCredentials(iamToken string) NonExchangeableCredentials {
	return &IAMTokenCredentials{
		iamToken: iamToken,
	}
}
