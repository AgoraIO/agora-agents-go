---
sidebar_position: 6
title: Error Handling
description: Handle API errors with core.APIError and errors.As.
---

# Error Handling

When the API returns a non-success status code (4xx or 5xx response), the SDK returns a `*core.APIError`. Use `errors.As` to inspect it:

```go
package main

import (
    "errors"
    "fmt"
    "log"
    "net/http"

    core "github.com/AgoraIO/agora-agents-go/v2/core"
)

func main() {
    resp, err := client.Agents.Start(ctx, req)
    if err != nil {
        var apiErr *core.APIError
        if errors.As(err, &apiErr) {
            fmt.Println("Status:", apiErr.StatusCode)
            fmt.Println("Message:", apiErr.Error())
            fmt.Println("Headers:", apiErr.Header)
        } else {
            log.Fatal(err)
        }
        return
    }
    _ = resp
}
```

## APIError Fields

| Field | Type | Description |
|-------|------|--------------|
| `StatusCode` | `int` | HTTP status code (e.g., 400, 404, 500) |
| `Header` | `http.Header` | Response headers |
| `Error()` | `string` | Combined status and message |

## v2.7 Reason Codes

v2.7 error responses may include a provider-specific `reason` field in the JSON body. The generated `core.APIError` preserves the raw response body in `Error()`, so you can parse it when you need machine-readable handling:

```go
type agentErrorBody struct {
    Message string `json:"message"`
    Reason  string `json:"reason"`
}

var apiErr *core.APIError
if errors.As(err, &apiErr) {
    var body agentErrorBody
    if json.Unmarshal([]byte(apiErr.Unwrap().Error()), &body) == nil {
        switch body.Reason {
        case "invalid_vendor_config":
            // Prompt the user to verify vendor credentials or model fields.
        case "token_expired":
            // Regenerate the session or avatar token and retry.
        }
    }
}
```

Use `StatusCode` for broad retry behavior and `reason` for actionable user-facing remediation.

## Graceful 404 Handling

For operations like `Stop` where a 404 may indicate the agent was already stopped, check the status code:

```go
err := session.Stop(ctx)
if err != nil {
    var apiErr *core.APIError
    if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
        // Agent already stopped or never existed — treat as success
        return nil
    }
    return err
}
```

Use this for both the agentkit layer (e.g. `session.Start()`, `session.Stop()`) and direct client calls (`client.Agents.Start(...)`).

The v3 preview package has its own error categories (validation, configuration, connection, timeout, and API); see [v3 preview errors](v3-preview.md#errors).
