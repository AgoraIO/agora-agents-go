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
// Preview registrations are temporary. When a provider ships on the production
// gateway, remove its registration and move its implementation into the
// corresponding production vendor module.

// PreviewAPIBaseURL is the base URL that serves the preview providers.
const PreviewAPIBaseURL = "https://partner.ai.agora.io/preview/api/conversational-ai-agent"

// PreviewFeatureHeader opts a request into a preview provider family.
//
// This is the header the preview gateway routes on. A request that reaches the
// gateway without it is not rejected — it is routed to the production
// environment, where the preview providers do not exist.
const PreviewFeatureHeader = "agora-feature"

// PreviewFeatureGeminiLive gates the Gemini 3.5 Transcribe ASR provider.
//
// Deprecated: Gemini ASR is available through the production endpoint. This
// value remains for source compatibility with the former preview API.
const PreviewFeatureGeminiLive = "gemini-live"

// PreviewFeatureLiveModels gates the OpenAI GPT Live MLLM provider.
const PreviewFeatureLiveModels = "live-models"

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

func previewRequestOptions(features []string, inner core.HTTPClient, debug bool) []option.RequestOption {
	if len(features) == 0 && !debug {
		return nil
	}
	if inner == nil {
		inner = http.DefaultClient
	}
	if debug {
		inner = &debugHTTPClient{inner: inner}
	}
	if len(features) == 0 {
		return []option.RequestOption{option.WithHTTPClient(inner)}
	}
	return []option.RequestOption{
		option.WithBaseURL(PreviewAPIBaseURL),
		option.WithHTTPClient(&previewGateClient{inner: inner, features: strings.Join(features, ",")}),
	}
}

// previewASRVendors are served only by the preview endpoint.
var previewASRVendors = map[string]string{}

func isPreviewOpenAIModel(properties map[string]interface{}) bool {
	mllm, ok := properties["mllm"].(map[string]interface{})
	return ok && mllm["vendor"] == "openai_gpt_live"
}

// RequiredPreviewFeatures returns the preview features a start request needs.
//
// Derived from the request body rather than from the vendor types, so
// hand-written configs are covered too.
func RequiredPreviewFeatures(properties map[string]interface{}) []string {
	return requiredPreviewFeatures(properties, previewASRVendors)
}

func requiredPreviewFeatures(properties map[string]interface{}, previewVendors map[string]string) []string {
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
			if feature, found := previewVendors[vendor]; found {
				add(feature)
			}
		}
	}
	if isPreviewOpenAIModel(properties) {
		add(PreviewFeatureLiveModels)
	}

	return features
}
