//go:build !fipsonly

package sql

import (
	"context"
	"database/sql"
	"encoding/json"
	neturl "net/url"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	mssql "github.com/microsoft/go-mssqldb"
)

//
// Azure authentication & authorization
//
// There are 4 options to authenticate against Azure for Azure SQL database access:
// 1. client secret
// 2. certificate
// 3. username & password
// 4. managed system identity
//
// There are 4 ways to supply authentication as (ADO style) connection string parameters
// 1. "azure auth=ENV" - authentication info will be retrieved from env variables
// 2. "azure auth=c:\secure\azure.auth" - authentication info will be retrieved from a file
// 3. authentication info is specified directly in the connection string:
//  1) "azure tenant id=77...;azure client id=58...;azure client secret=0cf123.."
//  1) "azure tenant id=77...;azure client id=58...;azure certificate path=C:\secure\...;azure certificate password=xY..."
//  1) "azure tenant id=77...;azure client id=58...;azure username=some@myorg;azure password=a1..."
// 4. "azure auth=MSI" - requires no other info but it works only in Azure VM with managed identity set
//
// See https://learn.microsoft.com/en-us/azure/developer/go/azure-sdk-authentication for details.
//

// Azure authentication config
type AzureConfig struct {
	TenantId            string
	ClientId            string
	ClientSecret        string
	CertificatePath     string
	CertificatePassword string
	Username            string `json:"Username (Azure)"`
	Password            string `json:"Password (Azure)"`
	Location            string
}

// Azure authentication options
const (
	mssqlAzureAuthEnv    = "ENV"
	mssqlAzureAuthFile   = "FILE"
	mssqlAzureAuthConfig = "CONFIG"
	mssqlAzureAuthMsi    = "MANAGED_IDENTITY"
)

// Azure SQL scope for OAuth 2.0 authentication
const (
	mssqlAzureSQLScope = "https://database.windows.net/.default"
)

// Connection parameter keys for azure authentication
const (
	mssqlAzureAuthKey                      = "azure auth"
	mssqlAzureClientIdKey                  = "azure client id"
	mssqlAzureTenantIdKey                  = "azure tenant id"
	mssqlAzureClientSecretKey              = "azure client secret"
	mssqlAzureClientCertificatePathKey     = "azure certificate path"
	mssqlAzureClientCertificatePasswordKey = "azure certificate password"
	mssqlAzureUsernameKey                  = "azure username"
	mssqlAzurePasswordKey                  = "azure password"
)

func mssqlSetAzureConfig(params neturl.Values, cfg *mssqlConfig) {
	if aauth := params.Get(mssqlAzureAuthKey); aauth != "" {
		switch aauth {
		case "ENV", "env":
			cfg.AzureAuth = mssqlAzureAuthEnv
		case "MSI", "":
			cfg.AzureAuth = mssqlAzureAuthMsi
		default:
			cfg.AzureAuth = mssqlAzureAuthFile
			cfg.AzureConfig = &AzureConfig{
				Location: aauth,
			}
		}
	}
	if acid := params.Get(mssqlAzureClientIdKey); acid != "" {
		cfg.AzureAuth = mssqlAzureAuthConfig
		cfg.AzureConfig = &AzureConfig{
			TenantId:            params.Get(mssqlAzureTenantIdKey),
			ClientId:            acid,
			ClientSecret:        params.Get(mssqlAzureClientSecretKey),
			CertificatePath:     params.Get(mssqlAzureClientCertificatePathKey),
			CertificatePassword: params.Get(mssqlAzureClientCertificatePasswordKey),
			Username:            params.Get(mssqlAzureUsernameKey),
			Password:            params.Get(mssqlAzurePasswordKey),
		}
	}
}

func mssqlOpenFunction(driverName, dataSourceName string) openFunc {
	cfg, err := mssqlParseDSN(dataSourceName)
	if err != nil {
		return func() (*sql.DB, error) {
			return nil, err
		}
	}
	if cfg.AzureAuth == "" {
		return defaultOpenFunction(driverName, dataSourceName)
	}

	return func() (*sql.DB, error) {
		credential, err := mssqlAzureAuthToken(cfg.AzureAuth, cfg.AzureConfig)
		if err != nil {
			return nil, err
		}
		connector, err := mssql.NewAccessTokenConnector(dataSourceName, func() (string, error) {
			ctx := context.Background()
			token, err := credential.GetToken(ctx, policy.TokenRequestOptions{
				Scopes: []string{mssqlAzureSQLScope},
			})
			if err != nil {
				return "", err
			}
			return token.Token, nil
		})
		if err != nil {
			return nil, err
		}
		db := sql.OpenDB(connector)
		return db, nil
	}
}

func mssqlAzureAuthToken(method string, cfg *AzureConfig) (azcore.TokenCredential, error) {
	switch method {
	case mssqlAzureAuthConfig:
		// Try authentication methods in order based on what credentials are provided
		if cfg.ClientSecret != "" {
			// Client Secret authentication
			return azidentity.NewClientSecretCredential(cfg.TenantId, cfg.ClientId, cfg.ClientSecret, nil)
		}
		if cfg.CertificatePath != "" {
			// Certificate-based authentication
			certData, err := os.ReadFile(cfg.CertificatePath)
			if err != nil {
				return nil, errors.Newf(codes.Invalid, "failed to read certificate file: %v", err)
			}
			certs, key, err := azidentity.ParseCertificates(certData, []byte(cfg.CertificatePassword))
			if err != nil {
				return nil, errors.Newf(codes.Invalid, "failed to parse certificate: %v", err)
			}
			return azidentity.NewClientCertificateCredential(cfg.TenantId, cfg.ClientId, certs, key, nil)
		}
		if cfg.Username != "" && cfg.Password != "" {
			// Username/Password authentication
			return azidentity.NewUsernamePasswordCredential(cfg.TenantId, cfg.ClientId, cfg.Username, cfg.Password, nil)
		}
		return nil, errors.Newf(codes.Invalid, "insufficient authentication credentials provided")

	case mssqlAzureAuthMsi:
		// Managed Identity authentication
		if cfg != nil && cfg.ClientId != "" {
			// User-assigned managed identity
			return azidentity.NewManagedIdentityCredential(&azidentity.ManagedIdentityCredentialOptions{
				ID: azidentity.ClientID(cfg.ClientId),
			})
		}
		// System-assigned managed identity
		return azidentity.NewManagedIdentityCredential(nil)

	case mssqlAzureAuthEnv:
		// Environment-based authentication using DefaultAzureCredential
		// This supports multiple authentication methods via environment variables
		return azidentity.NewDefaultAzureCredential(nil)

	case mssqlAzureAuthFile:
		// File-based authentication
		authData, err := os.ReadFile(cfg.Location)
		if err != nil {
			return nil, errors.Newf(codes.Invalid, "failed to read authentication file: %v", err)
		}

		// Parse the auth file (typically JSON format)
		var authFile struct {
			ClientID                string `json:"clientId"`
			ClientSecret            string `json:"clientSecret"`
			TenantID                string `json:"tenantId"`
			CertificatePath         string `json:"certificatePath"`
			CertificatePassword     string `json:"certificatePassword"`
			ActiveDirectoryEndpoint string `json:"activeDirectoryEndpointUrl"`
		}
		if err := json.Unmarshal(authData, &authFile); err != nil {
			return nil, errors.Newf(codes.Invalid, "failed to parse authentication file: %v", err)
		}

		// Try client secret first
		if authFile.ClientSecret != "" {
			return azidentity.NewClientSecretCredential(authFile.TenantID, authFile.ClientID, authFile.ClientSecret, nil)
		}

		// Try certificate
		if authFile.CertificatePath != "" {
			certData, err := os.ReadFile(authFile.CertificatePath)
			if err != nil {
				return nil, errors.Newf(codes.Invalid, "failed to read certificate file from auth file: %v", err)
			}
			certs, key, err := azidentity.ParseCertificates(certData, []byte(authFile.CertificatePassword))
			if err != nil {
				return nil, errors.Newf(codes.Invalid, "failed to parse certificate from auth file: %v", err)
			}
			return azidentity.NewClientCertificateCredential(authFile.TenantID, authFile.ClientID, certs, key, nil)
		}

		return nil, errors.Newf(codes.Invalid, "only client credentials and certificate authentication is supported with authentication file")
	}

	return nil, errors.Newf(codes.Invalid, "unsupported authentication method")
}
