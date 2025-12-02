# dis-bundle-api SDK

## Overview

This API contains a client with functions for interacting with the Bundle API from other applications. It provides methods to retrieve and update bundle information.

## Example Use of the Client

```go
package main

import (
    "context"
    "time"

    "github.com/ONSdigital/dis-bundle-api/models"
    "github.com/ONSdigital/dis-bundle-api/sdk"
)

func main() {
    client := sdk.New("http://localhost:29800")
    
    headers := sdk.Headers{
        ServiceAuthToken: "your-service-token",
    }
    
    bundles, err := client.GetBundles(context.Background(), headers, &time.Time{}, nil)
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