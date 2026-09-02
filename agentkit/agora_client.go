package agentkit

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	Agora "github.com/AgoraIO/agora-agents-go/v2"
	"github.com/AgoraIO/agora-agents-go/v2/agentmanagement"
	"github.com/AgoraIO/agora-agents-go/v2/agents"
	"github.com/AgoraIO/agora-agents-go/v2/client"
	"github.com/AgoraIO/agora-agents-go/v2/core"
	"github.com/AgoraIO/agora-agents-go/v2/option"
	"github.com/AgoraIO/agora-agents-go/v2/phonenumbers"
	"github.com/AgoraIO/agora-agents-go/v2/telephony"
)

// AuthMode represents the authentication mode for the Agora client.
type AuthMode string

const (
	// AuthModeAppCredentials uses AppID + AppCertificate to auto-generate a
	// short-lived ConvoAI token per API call — no Customer Secret required.
	AuthModeAppCredentials AuthMode = "app-credentials"
	// AuthModeBasic uses HTTP Basic Auth with a Customer ID and Customer Secret.
	AuthModeBasic AuthMode = "basic"
	// AuthModeToken uses a pre-built Agora token passed as the Authorization header.
	AuthModeToken AuthMode = "token"
)

// AgoraClientOptions contains configuration for creating an AgoraClient.
type AgoraClientOptions struct {
	// Area is the geographic region for the client (US, EU, AP, CN).
	Area option.Area
	// AppID is the Agora App ID, used as the path parameter and for token generation.
	AppID string
	// AppCertificate is used to sign tokens. Keep this value secret.
	AppCertificate string
	// CustomerID and CustomerSecret are used for Basic auth (from Agora Console).
	// Mutually exclusive with Token.
	CustomerID     string
	CustomerSecret string
	// Token is a pre-built raw Agora token for REST API authentication.
	// The SDK sets "Authorization: agora token=<Token>" automatically.
	// Mutually exclusive with CustomerID/CustomerSecret.
	Token string
}

// AgoraClient wraps the Fern-generated client with AppID, AppCertificate, and
// auth mode tracking—providing the same client API as the TypeScript and Python SDKs.
//
// Agents, AgentManagement, Telephony, and PhoneNumbers expose the underlying
// REST clients for advanced usage.
type AgoraClient struct {
	Agents          *agents.Client
	AgentManagement *agentmanagement.Client
	Telephony       *telephony.Client
	PhoneNumbers    *phonenumbers.Client
	appID           string
	appCertificate  string
	AuthMode        AuthMode
}

func (c *AgoraClient) AgentsClient() *agents.Client {
	return c.Agents
}

func (c *AgoraClient) AgentManagementClient() *agentmanagement.Client {
	return c.AgentManagement
}

func (c *AgoraClient) AppID() string {
	return c.appID
}

func (c *AgoraClient) AppCertificate() string {
	return c.appCertificate
}

func (c *AgoraClient) IsAppCredentialsMode() bool {
	return c.AuthMode == AuthModeAppCredentials
}

// NewAgoraClient creates a new AgoraClient with the given options.
//
// App credentials mode (recommended) — no Customer Secret needed:
//
//	c := agentkit.NewAgoraClient(agentkit.AgoraClientOptions{
//	    Area:           option.AreaUS,
//	    AppID:          "your-app-id",
//	    AppCertificate: "your-app-certificate",
//	})
//
// Basic auth mode:
//
//	c := agentkit.NewAgoraClient(agentkit.AgoraClientOptions{
//	    Area:           option.AreaUS,
//	    AppID:          "your-app-id",
//	    AppCertificate: "your-app-certificate",
//	    CustomerID:     "your-customer-id",
//	    CustomerSecret: "your-customer-secret",
//	})
//
// Token auth mode (pre-built token):
//
//	c := agentkit.NewAgoraClient(agentkit.AgoraClientOptions{
//	    Area:           option.AreaUS,
//	    AppID:          "your-app-id",
//	    AppCertificate: "your-app-certificate",
//	    Token:          "your-raw-token",
//	})
func NewAgoraClient(opts AgoraClientOptions) *AgoraClient {
	reqOpts := []option.RequestOption{option.WithArea(opts.Area)}

	authMode := AuthModeAppCredentials
	if opts.CustomerID != "" {
		authMode = AuthModeBasic
		reqOpts = append(reqOpts, option.WithBasicAuth(opts.CustomerID, opts.CustomerSecret))
	} else if opts.Token != "" {
		authMode = AuthModeToken
		reqOpts = append(reqOpts, option.WithToken(opts.Token))
	}

	c := client.NewClient(reqOpts...)
	return &AgoraClient{
		Agents:          c.Agents,
		AgentManagement: c.AgentManagement,
		Telephony:       c.Telephony,
		PhoneNumbers:    c.PhoneNumbers,
		appID:           opts.AppID,
		appCertificate:  opts.AppCertificate,
		AuthMode:        authMode,
	}
}

// StopAgent stops an agent by its ID without requiring an AgentSession reference.
//
// Use this when handling a stop request from your client app (e.g., an end-call
// button). Create a client with the same credentials used to start the agent and
// call this method.
//
// If the agent has already stopped (e.g., timed out or crashed), StopAgent
// returns nil rather than an error.
//
//	c := agentkit.NewAgoraClient(agentkit.AgoraClientOptions{
//	    Area:           option.AreaUS,
//	    AppID:          "your-app-id",
//	    AppCertificate: "your-app-certificate",
//	})
//	if err := c.StopAgent(ctx, agentID); err != nil {
//	    log.Printf("stop failed: %v", err)
//	}
func (c *AgoraClient) StopAgent(ctx context.Context, agentID string) error {
	var reqOpts []option.RequestOption
	if c.AuthMode == AuthModeAppCredentials && c.appCertificate != "" {
		token, err := GenerateConvoAIToken(GenerateConvoAITokenOptions{
			AppID:          c.appID,
			AppCertificate: c.appCertificate,
			ChannelName:    "stop",
			UID:            0,
		})
		if err != nil {
			return fmt.Errorf("agentkit: failed to generate token for StopAgent: %w", err)
		}
		reqOpts = []option.RequestOption{option.WithToken(token)}
	}

	err := c.Agents.Stop(ctx, &Agora.StopAgentsRequest{
		Appid:   c.appID,
		AgentID: agentID,
	}, reqOpts...)
	if err != nil {
		var apiErr *core.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
			return nil // Agent already stopped — treat as success
		}
		return err
	}
	return nil
}
