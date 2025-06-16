package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/utils"
	"github.com/ONSdigital/dp-authorisation/v2/authorisation"
	"github.com/ONSdigital/dp-permissions-api/sdk"
)

// TokenBearerPrefix is the standard prefix for bearer authentication header values
const TokenBearerPrefix = "Bearer "

// AuthorisationMiddleware wraps the dp-authorisation middleware and provides additional common functionality
type AuthorisationMiddleware interface {
	authorisation.Middleware
	GetJWTEntityData(r *http.Request) (*sdk.EntityData, *models.Error)
}

// AuthMiddleware is a concrete implementation of AuthorisationMiddleware
type AuthMiddleware struct {
	authorisation.Middleware
}

// Interface check
var _ AuthorisationMiddleware = (*AuthMiddleware)(nil)

// CreateAuthorisationMiddlewareFromConfig creates and configures the authorisation middleware using the provided config
//
// Parameters:
// - cfg: authorisation config to use for constructing the middleware
// - useKeys: whether to use the JWTVerificationPublicKeys from the Config or not. Used in the feature tests.
func CreateAuthorisationMiddlewareFromConfig(ctx context.Context, cfg *authorisation.Config, useKeys bool) (AuthorisationMiddleware, error) {
	var middleware authorisation.Middleware
	var err error

	if cfg.Enabled {
		middleware, err = getAuthorisationMiddleware(ctx, cfg, useKeys)
	} else {
		middleware, err = getFeatureFlaggedMiddleware(ctx, cfg)
	}

	if err != nil {
		return nil, err
	}

	return &AuthMiddleware{
		middleware,
	}, nil
}

// getAuthorisationMiddleware creates the dp-authorisation middleware with authorisation enabled
func getAuthorisationMiddleware(ctx context.Context, cfg *authorisation.Config, useKeys bool) (authorisation.Middleware, error) {
	if !useKeys {
		return authorisation.NewMiddlewareFromConfig(ctx, cfg, nil)
	}

	return authorisation.NewMiddlewareFromConfig(ctx, cfg, cfg.JWTVerificationPublicKeys)
}

// getFeatureFlaggedMiddleware creates the dp-authorisation middleware depending on the configured feature flag
func getFeatureFlaggedMiddleware(ctx context.Context, cfg *authorisation.Config) (authorisation.Middleware, error) {
	return authorisation.NewFeatureFlaggedMiddleware(ctx, cfg, nil)
}

// GetJWTEntityData extracts JWT entity data from the HTTP request's authoriation header, using the dp-authorisastion middleware's existing parse method.
func (a *AuthMiddleware) GetJWTEntityData(r *http.Request) (*sdk.EntityData, *models.Error) {
	bearerTokenValue := getBearerTokenValue(r)

	if bearerTokenValue == "" {
		return nil, models.CreateModelError(models.CodeBadRequest, apierrors.ErrorDescriptionNoTokenFound)
	}

	JWTEntityData, err := a.Parse(bearerTokenValue)
	if err != nil {
		return nil, models.CreateModelError(models.CodeInternalServerError, apierrors.ErrorDescriptionUserIdentityParseFailed)
	}

	return JWTEntityData, nil
}

// getBearerTokenValue strips the "Bearer " prefix from the authorisation header and returns the trimmed value
func getBearerTokenValue(r *http.Request) string {
	authHeader := r.Header.Get(utils.HeaderAuthorization)
	if authHeader == "" {
		return authHeader
	}

	if !strings.HasPrefix(authHeader, TokenBearerPrefix) {
		return ""
	}

	return strings.TrimPrefix(r.Header.Get(utils.HeaderAuthorization), TokenBearerPrefix)
}
