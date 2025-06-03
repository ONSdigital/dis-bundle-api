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

const TokenBearerPrefix = "Bearer "

type AuthMiddleware struct {
	authorisation.Middleware
}

type AuthorisationMiddleware interface {
	authorisation.Middleware
	GetJWTEntityData(r *http.Request) (*sdk.EntityData, *models.Error)
}

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

func getAuthorisationMiddleware(ctx context.Context, cfg *authorisation.Config, useKeys bool) (authorisation.Middleware, error) {
	if !useKeys {
		return authorisation.NewMiddlewareFromConfig(ctx, cfg, nil)
	}

	return authorisation.NewMiddlewareFromConfig(ctx, cfg, cfg.JWTVerificationPublicKeys)
}

func getFeatureFlaggedMiddleware(ctx context.Context, cfg *authorisation.Config) (authorisation.Middleware, error) {
	return authorisation.NewFeatureFlaggedMiddleware(ctx, cfg, nil)
}

var _ AuthorisationMiddleware = (*AuthMiddleware)(nil)

func (a *AuthMiddleware) GetJWTEntityData(r *http.Request) (*sdk.EntityData, *models.Error) {
	JWTEntityData, err := a.Parse(strings.TrimPrefix(r.Header.Get(utils.HeaderAuthorization), TokenBearerPrefix))
	if err != nil {
		return nil, models.CreateModelError(models.CodeInternalServerError, apierrors.ErrorDescriptionUserIdentityParseFailed)
	}

	return JWTEntityData, nil
}
