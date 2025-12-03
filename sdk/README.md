# dis-bundle-api SDK

## Overview

This API contains a client with functions for interacting with the Bundle API from other applications. It provides methods to retrieve and update bundle information.

### Available methods

| Method | Description |
|--------|-------------|
| [GetBundles](#getbundles) | Retrieves a paginated list of bundles, optionally filtered by scheduled date |
| [GetBundle](#getbundle) | Retrieves a single bundle by ID |
| [PutBundleState](#putbundlestate) | Updates the state of a bundle (DRAFT, IN_REVIEW, APPROVED, PUBLISHED) |
| [Checker](#checker) | Performs a health check against the Bundle API endpoint |
| [Health](#health) | Returns the underlying health check client |
| [URL](#url) | Returns the base URL of the Bundle API |

## Example Use of the Client

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/ONSdigital/dis-bundle-api/models"
    "github.com/ONSdigital/dis-bundle-api/sdk"
    sdkErrors "github.com/ONSdigital/dis-bundle-api/sdk/errors"
)

func main() {
    client := sdk.New("http://localhost:29800")
    
    headers := sdk.Headers{
        ServiceAuthToken: "your-service-token",
    }
    
    bundles, err := client.GetBundles(context.Background(), headers, &time.Time{}, nil)
    if err != nil {
        statusCode := sdkErrors.ErrorStatus(err)
        message := sdkErrors.ErrorMessage(err)
        log.Printf("Failed to get bundles: %s (status: %d)", message, statusCode)
        return
    }
    
    log.Printf("Retrieved %d bundles", len(bundles.Items))
}
```

### With Health Client

```go
    import (
        "github.com/ONSdigital/dis-bundle-api/sdk"
        "github.com/ONSdigital/dp-api-clients-go/v2/health"
    )

    hcClient := health.NewClient("my-service", "http://localhost:29800")
    client := sdk.NewWithHealthClient(hcClient)
    bundles, err := client.GetBundles(context.Background(), headers, &time.Time{}, nil)

```

## Available Functionality

### GetBundles
Retrieves a paginated list of bundles, optionally filtered by scheduled date.

```go
queryParams := &sdk.QueryParams{Limit: 10, Offset: 0}
bundles, err := client.GetBundles(ctx, headers, &time.Time{}, queryParams)
```

### GetBundle
Retrieves a single bundle by ID.

```go
respInfo, err := client.GetBundle(ctx, headers, "bundle-id")
```

### PutBundleState
Updates the state of a bundle (DRAFT, IN_REVIEW, APPROVED, PUBLISHED).

```go
headers := sdk.Headers{
    ServiceAuthToken: "token",
    IfMatch:          "etag-value",
}
bundle, err := client.PutBundleState(ctx, headers, "bundle-id", models.BundleStateApproved)
```

### Checker
Performs a health check against the Bundle API endpoint.

```go
check := &health.CheckState{}
err := client.Checker(ctx, check)
```

### Health
Returns the underlying health check client.

```go
hcClient := client.Health()
```

### URL
Returns the base URL of the Bundle API.

```go
url := client.URL()
```