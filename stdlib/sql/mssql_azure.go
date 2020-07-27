package sql

import (
	"database/sql"
	neturl "net/url"
	"os"

	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/denisenkom/go-mssqldb"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
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
// See https://docs.microsoft.com/en-us/azure/developer/go/azure-sdk-authorization for details.
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

// Azure resource ie. Azure SQL Server
const (
	mssqlAzureResource = "https://database.windows.net/"
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
		var spt *adal.ServicePrincipalToken
		spt, err := mssqlAzureAuthToken(cfg.AzureAuth, cfg.AzureConfig)
		if err != nil {
			return nil, err
		}
		connector, err := mssql.NewAccessTokenConnector(dataSourceName, func() (string, error) {
			if e := spt.EnsureFresh(); e != nil {
				return "", e
			}
			t := spt.OAuthToken()
			return t, nil
		})
		if err != nil {
			return nil, err
		}
		db := sql.OpenDB(connector)
		return db, nil
	}
}

func mssqlAzureAuthToken(method string, cfg *AzureConfig) (*adal.ServicePrincipalToken, error) {
	fromEnvSettings := func(settings auth.EnvironmentSettings) (*adal.ServicePrincipalToken, error) { // see auth.EnvironmentSettings.GetAuthorizer()
		if c, err := settings.GetClientCredentials(); err == nil {
			return c.ServicePrincipalToken()
		}
		if c, err := settings.GetClientCertificate(); err == nil {
			return c.ServicePrincipalToken()
		}
		if c, err := settings.GetUsernamePassword(); err == nil {
			return c.ServicePrincipalToken()
		}
		return settings.GetMSI().ServicePrincipalToken()
	}
	switch method {
	case mssqlAzureAuthConfig:
		settings := auth.EnvironmentSettings{
			Values: map[string]string{
				auth.Resource: mssqlAzureResource,
			},
			Environment: azure.PublicCloud,
		}
		settings.Values[auth.TenantID] = cfg.TenantId
		settings.Values[auth.ClientID] = cfg.ClientId
		settings.Values[auth.ClientSecret] = cfg.ClientSecret
		settings.Values[auth.CertificatePath] = cfg.CertificatePath
		settings.Values[auth.CertificatePassword] = cfg.CertificatePassword
		settings.Values[auth.Username] = cfg.Username
		settings.Values[auth.Password] = cfg.Password
		return fromEnvSettings(settings)
	case mssqlAzureAuthMsi:
		mc := auth.NewMSIConfig()
		mc.Resource = mssqlAzureResource
		return mc.ServicePrincipalToken()
	case mssqlAzureAuthEnv:
		settings, err := auth.GetSettingsFromEnvironment()
		if err != nil {
			return nil, err
		}
		return fromEnvSettings(settings)
	case mssqlAzureAuthFile: // see auth.GetSettingsFromFile()
		os.Setenv("AZURE_AUTH_LOCATION", cfg.Location)
		defer os.Unsetenv("AZURE_AUTH_LOCATION")
		settings, err := auth.GetSettingsFromFile()
		if err != nil {
			return nil, err
		}
		if t, err := settings.ServicePrincipalTokenFromClientCredentialsWithResource(mssqlAzureResource); err == nil {
			return t, nil
		}
		if t, err := settings.ServicePrincipalTokenFromClientCertificateWithResource(mssqlAzureResource); err == nil {
			return t, nil
		}
		return nil, errors.Newf(codes.Invalid, "only client credentials and certificate authentication is supported with authentication file")

	}
	return nil, errors.Newf(codes.Invalid, "unsupported authentication")
}
