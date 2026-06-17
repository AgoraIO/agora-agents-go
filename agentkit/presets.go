package agentkit

import (
	"encoding/json"
	"strings"

	Agora "github.com/AgoraIO/agora-agents-go/v2"
)

var AgentPresets = struct {
	Asr struct {
		DeepgramNova2 string
		DeepgramNova3 string
	}
	Llm struct {
		OpenAIGpt4oMini string
		OpenAIGpt41Mini string
		OpenAIGpt5Nano  string
		OpenAIGpt5Mini  string
	}
	Tts struct {
		MiniMaxSpeech26Turbo string
		MiniMaxSpeech28Turbo string
		OpenAITts1           string
	}
}{
	Asr: struct {
		DeepgramNova2 string
		DeepgramNova3 string
	}{
		DeepgramNova2: "deepgram_nova_2",
		DeepgramNova3: "deepgram_nova_3",
	},
	Llm: struct {
		OpenAIGpt4oMini string
		OpenAIGpt41Mini string
		OpenAIGpt5Nano  string
		OpenAIGpt5Mini  string
	}{
		OpenAIGpt4oMini: "openai_gpt_4o_mini",
		OpenAIGpt41Mini: "openai_gpt_4_1_mini",
		OpenAIGpt5Nano:  "openai_gpt_5_nano",
		OpenAIGpt5Mini:  "openai_gpt_5_mini",
	},
	Tts: struct {
		MiniMaxSpeech26Turbo string
		MiniMaxSpeech28Turbo string
		OpenAITts1           string
	}{
		MiniMaxSpeech26Turbo: "minimax_speech_2_6_turbo",
		MiniMaxSpeech28Turbo: "minimax_speech_2_8_turbo",
		OpenAITts1:           "openai_tts_1",
	},
}

const openAIChatCompletionsURL = "https://api.openai.com/v1/chat/completions"

func NormalizePresetInput(presets []string) string {
	normalized := make([]string, 0, len(presets))
	for _, preset := range presets {
		preset = strings.TrimSpace(preset)
		if preset != "" {
			normalized = append(normalized, preset)
		}
	}
	return strings.Join(normalized, ",")
}

func ResolveSessionPresets(presets []string, properties *Agora.StartAgentsRequestProperties) (string, *Agora.StartAgentsRequestProperties, error) {
	if properties == nil {
		return NormalizePresetInput(presets), nil, nil
	}

	payload, err := json.Marshal(properties)
	if err != nil {
		return "", nil, err
	}

	var props map[string]interface{}
	if err := json.Unmarshal(payload, &props); err != nil {
		return "", nil, err
	}

	resolvedPreset, resolvedProps, err := ResolveSessionPresetsMap(presets, props)
	if err != nil {
		return "", nil, err
	}

	resolvedPayload, err := json.Marshal(resolvedProps)
	if err != nil {
		return "", nil, err
	}

	var resolved Agora.StartAgentsRequestProperties
	if err := json.Unmarshal(resolvedPayload, &resolved); err != nil {
		return "", nil, err
	}
	return resolvedPreset, &resolved, nil
}

func ResolveSessionPresetsMap(presets []string, properties map[string]interface{}) (string, map[string]interface{}, error) {
	if properties == nil {
		return NormalizePresetInput(presets), nil, nil
	}
	props := cloneConfig(properties)
	explicit := parsePresetInput(presets)
	explicitCategories := map[string]bool{}
	for _, preset := range explicit {
		if category := getPresetCategory(preset); category != "" {
			explicitCategories[category] = true
		}
	}

	inferred := make([]string, 0, 3)
	if !explicitCategories["asr"] {
		if preset, ok := inferASRPreset(props["asr"]); ok {
			inferred = append(inferred, preset)
			stripInferredASRFields(props["asr"])
		}
	}
	if !explicitCategories["llm"] {
		if preset, ok := inferLLMPreset(props["llm"]); ok {
			inferred = append(inferred, preset)
			stripInferredLLMFields(props["llm"])
		}
	}
	if !explicitCategories["tts"] {
		if preset, ok := inferTTSPreset(props["tts"]); ok {
			inferred = append(inferred, preset)
			stripInferredTTSFields(props["tts"], preset)
		}
	}
	if tts := asMap(props["tts"]); tts != nil {
		delete(tts, "_minimax_preset_model")
	}
	omitEmptyProviderFields(props)

	combined := append(append([]string{}, explicit...), inferred...)
	return NormalizePresetInput(combined), props, nil
}

func parsePresetInput(presets []string) []string {
	parsed := make([]string, 0, len(presets))
	for _, preset := range presets {
		for _, item := range strings.Split(preset, ",") {
			item = strings.TrimSpace(item)
			if item != "" {
				parsed = append(parsed, item)
			}
		}
	}
	return parsed
}

func getPresetCategory(preset string) string {
	switch preset {
	case AgentPresets.Asr.DeepgramNova2, AgentPresets.Asr.DeepgramNova3:
		return "asr"
	case AgentPresets.Llm.OpenAIGpt4oMini, AgentPresets.Llm.OpenAIGpt41Mini, AgentPresets.Llm.OpenAIGpt5Nano, AgentPresets.Llm.OpenAIGpt5Mini:
		return "llm"
	case AgentPresets.Tts.MiniMaxSpeech26Turbo, AgentPresets.Tts.MiniMaxSpeech28Turbo, AgentPresets.Tts.OpenAITts1:
		return "tts"
	default:
		return ""
	}
}

func normalizeModelName(value interface{}) string {
	s, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(s))
}

func asMap(value interface{}) map[string]interface{} {
	m, _ := value.(map[string]interface{})
	return m
}

func inferASRPreset(value interface{}) (string, bool) {
	asr := asMap(value)
	if len(asr) == 0 || asr["vendor"] != "deepgram" {
		return "", false
	}
	params := asMap(asr["params"])
	if hasNonEmptyString(params, "key") {
		return "", false
	}
	switch normalizeModelName(params["model"]) {
	case "nova-2":
		return AgentPresets.Asr.DeepgramNova2, true
	case "nova-3":
		return AgentPresets.Asr.DeepgramNova3, true
	default:
		return "", false
	}
}

func inferLLMPreset(value interface{}) (string, bool) {
	llm := asMap(value)
	if len(llm) == 0 {
		return "", false
	}
	if hasNonEmptyString(llm, "api_key") {
		return "", false
	}
	if vendor, ok := llm["vendor"].(string); ok && vendor != "" && vendor != "openai" {
		return "", false
	}
	if url, ok := llm["url"].(string); ok && url != "" && url != openAIChatCompletionsURL {
		return "", false
	}
	params := asMap(llm["params"])
	switch normalizeModelName(params["model"]) {
	case "gpt-4o-mini":
		return AgentPresets.Llm.OpenAIGpt4oMini, true
	case "gpt-4.1-mini":
		return AgentPresets.Llm.OpenAIGpt41Mini, true
	case "gpt-5-nano":
		return AgentPresets.Llm.OpenAIGpt5Nano, true
	case "gpt-5-mini":
		return AgentPresets.Llm.OpenAIGpt5Mini, true
	default:
		return "", false
	}
}

func inferTTSPreset(value interface{}) (string, bool) {
	tts := asMap(value)
	if len(tts) == 0 {
		return "", false
	}
	switch tts["vendor"] {
	case "openai":
		params := asMap(tts["params"])
		if hasNonEmptyString(params, "api_key") || hasNonEmptyString(params, "key") {
			return "", false
		}
		model := normalizeModelName(params["model"])
		if model == "" || model == "tts-1" {
			return AgentPresets.Tts.OpenAITts1, true
		}
	case "minimax":
		params := asMap(tts["params"])
		if hasNonEmptyString(params, "key") {
			return "", false
		}
		// Model is no longer in params for the preset path; fall back to the top-level hint.
		model := normalizeModelName(params["model"])
		if model == "" {
			model = normalizeModelName(tts["_minimax_preset_model"])
		}
		switch model {
		case "speech-2.6-turbo", "speech_2_6_turbo":
			return AgentPresets.Tts.MiniMaxSpeech26Turbo, true
		case "speech-2.8-turbo", "speech_2_8_turbo":
			return AgentPresets.Tts.MiniMaxSpeech28Turbo, true
		}
	}
	return "", false
}

func hasNonEmptyString(m map[string]interface{}, key string) bool {
	value, ok := m[key]
	if !ok {
		return false
	}
	if s, ok := value.(string); ok {
		return strings.TrimSpace(s) != ""
	}
	return value != nil
}

func stripInferredASRFields(value interface{}) {
	asr := asMap(value)
	params := asMap(asr["params"])
	delete(params, "api_key")
	delete(params, "model")
	if len(params) == 0 {
		asr["params"] = map[string]interface{}{}
		return
	}
	asr["params"] = params
}

func stripInferredLLMFields(value interface{}) {
	llm := asMap(value)
	delete(llm, "api_key")
	if url, ok := llm["url"].(string); ok && (url == "" || url == openAIChatCompletionsURL) {
		delete(llm, "url")
	}
	params := asMap(llm["params"])
	delete(params, "model")
	if len(params) == 0 {
		delete(llm, "params")
		return
	}
	llm["params"] = params
}

func stripInferredTTSFields(value interface{}, preset string) {
	tts := asMap(value)
	params := asMap(tts["params"])
	switch preset {
	case AgentPresets.Tts.OpenAITts1:
		delete(params, "api_key")
		delete(params, "key")
		delete(params, "model")
	case AgentPresets.Tts.MiniMaxSpeech26Turbo, AgentPresets.Tts.MiniMaxSpeech28Turbo:
		delete(params, "key")
		delete(params, "group_id")
		delete(params, "model")
		delete(params, "url")
		delete(tts, "_minimax_preset_model")
	}
	if len(params) == 0 {
		tts["params"] = map[string]interface{}{}
		return
	}
	tts["params"] = params
}

func omitEmptyProviderFields(props map[string]interface{}) {
	if props == nil {
		return
	}
	if llm := asMap(props["llm"]); llm != nil {
		for _, key := range []string{"api_key", "url"} {
			if value, ok := llm[key].(string); ok && strings.TrimSpace(value) == "" {
				delete(llm, key)
			}
		}
		if params := asMap(llm["params"]); params != nil {
			for _, key := range []string{"model"} {
				if value, ok := params[key].(string); ok && strings.TrimSpace(value) == "" {
					delete(params, key)
				}
			}
		}
	}
	if asr := asMap(props["asr"]); asr != nil {
		if params := asMap(asr["params"]); params != nil {
			for _, key := range []string{"api_key", "model"} {
				if value, ok := params[key].(string); ok && strings.TrimSpace(value) == "" {
					delete(params, key)
				}
			}
		}
	}
	if tts := asMap(props["tts"]); tts != nil {
		if params := asMap(tts["params"]); params != nil {
			for _, key := range []string{"api_key", "key", "group_id", "url", "model"} {
				if value, ok := params[key].(string); ok && strings.TrimSpace(value) == "" {
					delete(params, key)
				}
			}
		}
	}
}
