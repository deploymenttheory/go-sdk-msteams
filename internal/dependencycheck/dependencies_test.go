// Package dependencycheck compiles the agreed dependencies on every supported OS.
// These references pin compatibility before their production adapters are added.
// They do not instantiate credentials, authenticate, or claim provider coverage.
package dependencycheck

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
	"github.com/go-resty/resty/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var (
	_ *resty.Response
	_ *zap.Logger
	_ trace.TracerProvider
	_ = otel.Version
	_ public.Client
	_ confidential.Client
	_ *azidentity.ManagedIdentityCredential
	_ *azidentity.WorkloadIdentityCredential
)
