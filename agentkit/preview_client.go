package agentkit

import (
	"net/http"
	"strings"

	"github.com/AgoraIO/agora-agents-go/v2/core"
	"github.com/AgoraIO/agora-agents-go/v2/option"
)

// Preview endpoint support.
//
// Preview providers are detected from the resolved start body. AgentSession
// then pins their gate header and base URL to every request in that session.
//
// Everything in preview_client.go and vendors/preview.go is temporary. When
// these providers ship on the production gateway, delete both files and move
// the vendor types into vendors/stt.go.

// PreviewAPIBaseURL is the base URL that serves the preview providers.
const PreviewAPIBaseURL = "https://partner.ai.agora.io/preview/api/conversational-ai-agent"

// PreviewFeatureHeader opts a request into a preview provider family.
//
// This is the header the preview gateway routes on. A request that reaches the
// gateway without it is not rejected — it is routed to the production
// environment, where the preview providers do not exist.
const PreviewFeatureHeader = "agora-feature"

// PreviewFeatureGeminiLive gates the Gemini 3.5 Transcribe ASR provider.
const PreviewFeatureGeminiLive = "gemini-live"

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

func previewRequestOptions(features []string, inner core.HTTPClient) []option.RequestOption {
	if len(features) == 0 {
		return nil
	}
	if inner == nil {
		inner = http.DefaultClient
	}
	return []option.RequestOption{
		option.WithBaseURL(PreviewAPIBaseURL),
		option.WithHTTPClient(&previewGateClient{inner: inner, features: strings.Join(features, ",")}),
	}
}

// previewASRVendors are served only by the preview endpoint.
var previewASRVendors = map[string]struct{}{
	"gemini": {},
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

	if asr, ok := properties["asr"].(map[string]interface{}); ok {
		if vendor, ok := asr["vendor"].(string); ok {
			if _, found := previewASRVendors[vendor]; found {
				add(PreviewFeatureGeminiLive)
			}
		}
	}

	return features
}
