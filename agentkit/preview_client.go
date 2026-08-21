package agentkit

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors"
	"github.com/AgoraIO/agora-agents-go/v2/client"
	"github.com/AgoraIO/agora-agents-go/v2/core"
	"github.com/AgoraIO/agora-agents-go/v2/option"
)

// Preview endpoint support.
//
// Preview providers are not served by the production Conversational AI gateway.
// They live behind a single partner host and are gated by an "agora-feature"
// header. NewAgoraPreviewClient returns an *AgoraClient pointed at that host
// with the header pinned, so the rest of agentkit (Agent, AgentSession, vendor
// classes) works unchanged.
//
// Everything in preview_client.go and vendors/preview.go is temporary. When
// these providers ship on the production gateway, delete both files and move
// the vendor types into vendors/mllm.go.

// PreviewAPIBaseURL is the base URL that serves the preview providers.
const PreviewAPIBaseURL = "https://partner.ai.agora.io/preview/api/conversational-ai-agent"

// PreviewFeatureHeader opts a request into a preview provider family.
//
// This is the header the preview gateway routes on. A request that reaches the
// gateway without it is not rejected — it is routed to the production
// environment, where the preview providers do not exist.
const PreviewFeatureHeader = "agora-feature"

// PreviewFeatureGeminiLive gates Gemini 3.5 Transcribe ASR and the Gemini
// walkie-talkie / clever-chatter MLLMs.
const PreviewFeatureGeminiLive = "gemini-live"

var defaultPreviewFeatures = []string{PreviewFeatureGeminiLive}

// AgoraPreviewClientOptions configures a preview client. It is AgoraClientOptions
// plus the preview-only Features and BaseURL overrides.
type AgoraPreviewClientOptions struct {
	AgoraClientOptions

	// Features are the preview families to request via the agora-feature header.
	// Defaults to ["gemini-live"].
	Features []string
	// BaseURL overrides the preview base URL. For internal testing against a
	// staging host.
	BaseURL string
	// HTTPClient is wrapped rather than replaced, so the gate header survives.
	HTTPClient core.HTTPClient
}

// previewGateClient pins the gate header onto every request.
//
// The header is applied here rather than via option.WithHTTPHeader because that
// option replaces the entire header map: a per-call option.WithHTTPHeader would
// otherwise silently drop the gate and route the request to production.
type previewGateClient struct {
	inner    core.HTTPClient
	features string
}

func (p *previewGateClient) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set(PreviewFeatureHeader, p.features)
	return p.inner.Do(req)
}

// NewAgoraPreviewClient creates a client for the Agora Conversational AI preview
// endpoint.
//
// Use it exactly like NewAgoraClient — same auth modes, same Agent and
// AgentSession API. It differs only in where requests go and in the
// agora-feature header it attaches.
//
//	c := agentkit.NewAgoraPreviewClient(agentkit.AgoraPreviewClientOptions{
//	    AgoraClientOptions: agentkit.AgoraClientOptions{
//	        Area:           option.AreaUS,
//	        AppID:          appID,
//	        AppCertificate: appCertificate,
//	    },
//	})
//	agent := agentkit.NewAgent(c)
//
// It returns *AgoraClient rather than a distinct type so it is accepted
// everywhere a production client is — NewAgent takes *AgoraClient.
func NewAgoraPreviewClient(opts AgoraPreviewClientOptions) *AgoraClient {
	features := opts.Features
	if len(features) == 0 {
		features = defaultPreviewFeatures
	}
	baseURL := opts.BaseURL
	if baseURL == "" {
		baseURL = PreviewAPIBaseURL
	}

	var inner core.HTTPClient = opts.HTTPClient
	if inner == nil {
		inner = http.DefaultClient
	}

	reqOpts := []option.RequestOption{
		option.WithArea(opts.Area),
		option.WithBaseURL(baseURL),
		option.WithHTTPClient(&previewGateClient{inner: inner, features: strings.Join(features, ",")}),
	}

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
		previewFeatures: features,
		previewBaseURL:  baseURL,
	}
}

// previewMLLMModels are the preview MLLM model names.
//
// mllm.vendor is "gemini" for both the production Gemini Live provider and the
// preview ones, so the model name is one of two things that mark a config as
// preview-only; see hasPreviewMLLMEnvelope for the other.
//
// Both the codenames and their post-rename IDs are listed, so recognition keeps
// working through Google's rollout whichever name a caller sends.
var previewMLLMModels = map[string]struct{}{
	"walkie-talkie":            {},
	"clever-chatter":           {},
	"gemini-3.5-live":          {},
	"gemini-3.5-live-thinking": {},
}

// hasPreviewMLLMEnvelope reports whether a config carries the envelope the
// preview MLLM vendor types emit: the API key repeated inside params, and a url
// on the Gemini Developer API host. The production GeminiLive vendor emits
// neither — it leaves params.api_key unset and defaults url to "" or a
// WebSocket endpoint.
//
// This is the second recognition path, and it exists because keying only off
// previewMLLMModels makes an unrecognised model name fail silently:
// ApplyPreviewShape would stop retargeting greeting_message, the greeting would
// land in a field these models ignore, and the agent would simply never greet.
// A model name we have not listed yet is reachable by following this SDK's own
// advice to override Model when Google renames one ahead of a release, so the
// failure has to not be silent.
func hasPreviewMLLMEnvelope(mllm map[string]interface{}) bool {
	params, ok := mllm["params"].(map[string]interface{})
	if !ok {
		return false
	}
	if _, ok := params["api_key"].(string); !ok {
		return false
	}
	url, ok := mllm["url"].(string)
	return ok && strings.HasPrefix(url, vendors.GeminiPreviewMLLMURL)
}

// isPreviewMLLM reports whether an MLLM config targets a preview model —
// by model name, or by the wire envelope only the preview vendor types produce.
func isPreviewMLLM(properties map[string]interface{}) (map[string]interface{}, bool) {
	mllm, ok := properties["mllm"].(map[string]interface{})
	if !ok {
		return nil, false
	}
	if vendor, _ := mllm["vendor"].(string); vendor != "gemini" {
		return nil, false
	}
	if params, ok := mllm["params"].(map[string]interface{}); ok {
		if model, ok := params["model"].(string); ok {
			if _, found := previewMLLMModels[model]; found {
				return mllm, true
			}
		}
	}
	if hasPreviewMLLMEnvelope(mllm) {
		return mllm, true
	}
	return nil, false
}

// previewMLLMFieldRenames are MLLM wire keys the preview route spells
// differently from the Agora schema, as production spelling -> preview spelling.
//
// "failure_message" is deliberately absent: it is an Agora engine feature
// rather than a Gemini one, so it keeps its schema spelling.
var previewMLLMFieldRenames = map[string]string{
	"greeting_message": "greeting",
}

// ApplyPreviewShape retargets MLLM fields the shared builder wrote with
// production spellings, so a preview config follows the preview route's
// structure.
//
// The Agent fills mllm.greeting_message from an agent-level greeting whenever
// the vendor has not set that key — correct for every GA vendor, but the preview
// Gemini models read "greeting", so the value would land in a field they ignore
// and the agent would silently never greet.
//
// Rather than teach the shared builder about preview providers, the translation
// lives here and disappears with this file at GA. The vendor's own value wins;
// the production-spelled one is the fallback, which also migrates a hand-written
// greeting_message onto the preview key so an existing config keeps working
// after only swapping the model.
//
// Mutates the mllm map in place. Safe because the builder hands this a fresh
// clone of the MLLM config rather than the Agent's stored one.
func ApplyPreviewShape(properties map[string]interface{}) {
	mllm, ok := isPreviewMLLM(properties)
	if !ok {
		return
	}
	for production, preview := range previewMLLMFieldRenames {
		value, found := mllm[production]
		if !found {
			continue
		}
		if _, exists := mllm[preview]; !exists {
			mllm[preview] = value
		}
		delete(mllm, production)
	}
}

// RequiredPreviewFeatures returns the preview features a start request needs.
//
// Derived from the request body rather than from the vendor types, so
// hand-written configs are covered too.
func RequiredPreviewFeatures(properties map[string]interface{}) []string {
	var features []string
	add := func(feature string) {
		for _, existing := range features {
			if existing == feature {
				return
			}
		}
		features = append(features, feature)
	}

	if _, found := isPreviewMLLM(properties); found {
		add(PreviewFeatureGeminiLive)
	}

	return features
}

// AssertPreviewSupport fails fast when preview providers are paired with a
// client that cannot reach them, instead of surfacing an opaque gateway error —
// or worse, silently succeeding against the production environment.
func AssertPreviewSupport(enabledFeatures []string, properties map[string]interface{}) error {
	required := RequiredPreviewFeatures(properties)
	if len(required) == 0 {
		return nil
	}

	var missing []string
	for _, feature := range required {
		found := false
		for _, enabled := range enabledFeatures {
			if enabled == feature {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, feature)
		}
	}
	if len(missing) == 0 {
		return nil
	}

	joined := strings.Join(missing, ", ")
	return fmt.Errorf(
		"this agent uses preview providers that require the preview endpoint (%s): "+
			"create the session with NewAgoraPreviewClient instead of NewAgoraClient, "+
			"and make sure its Features include: %s",
		joined, joined,
	)
}
