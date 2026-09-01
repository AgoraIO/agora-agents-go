package core

import "fmt"

func BuildPropertiesMap(profile Profile, base *BaseAgent, opts ToPropertiesOptions, tokenFactory TokenFactory) (map[string]interface{}, error) {
	if base == nil {
		return nil, fmt.Errorf("agent is required")
	}
	if base.MLLM != nil && hasEnabledAvatar(base) {
		return nil, fmt.Errorf("avatar is only supported with cascading ASR/LLM/TTS sessions; remove the avatar configuration when using MLLM")
	}

	expiry := opts.ExpiresIn
	if expiry != 0 {
		var err error
		expiry, err = ValidateExpiresIn(expiry)
		if err != nil {
			return nil, fmt.Errorf("invalid expiresIn: %w", err)
		}
	}
	opts.ExpiresIn = expiry

	token := opts.Token
	if token == "" {
		if opts.AppID == "" || opts.AppCertificate == "" {
			return nil, fmt.Errorf("either token or app_id+app_certificate must be provided")
		}
		uid, err := ParseNumericUID(opts.AgentUID, "agent UID")
		if err != nil {
			return nil, err
		}
		token, err = tokenFactory(GenerateConvoAITokenOptions{
			AppID:          opts.AppID,
			AppCertificate: opts.AppCertificate,
			ChannelName:    opts.Channel,
			UID:            uid,
			TokenExpire:    expiry,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to generate token: %w", err)
		}
	}

	if len(opts.RemoteUIDs) == 0 {
		return nil, fmt.Errorf("AgentSessionOptions.RemoteUIDs is required and must contain at least one UID")
	}

	propsMap := map[string]interface{}{
		"channel":         opts.Channel,
		"token":           token,
		"agent_rtc_uid":   opts.AgentUID,
		"remote_rtc_uids": opts.RemoteUIDs,
	}
	if opts.IdleTimeout != nil {
		propsMap["idle_timeout"] = *opts.IdleTimeout
	}
	if opts.EnableStringUID != nil {
		propsMap["enable_string_uid"] = *opts.EnableStringUID
	}
	if base.MLLM != nil {
		propsMap["mllm"] = buildMllmConfigMap(base)
	}
	if base.Interruption != nil {
		if err := SetStructMap(propsMap, "interruption", base.Interruption); err != nil {
			return nil, err
		}
	}
	if base.Sal != nil {
		if err := SetStructMap(propsMap, "sal", base.Sal); err != nil {
			return nil, err
		}
	}
	if base.Avatar != nil {
		avatar, err := enrichAvatarParams(base, opts, tokenFactory)
		if err != nil {
			return nil, err
		}
		propsMap["avatar"] = avatar
	}
	if base.AdvancedFeatures != nil {
		if err := SetStructMap(propsMap, "advanced_features", base.AdvancedFeatures); err != nil {
			return nil, err
		}
	}
	if base.Parameters != nil {
		if err := SetStructMap(propsMap, "parameters", base.Parameters); err != nil {
			return nil, err
		}
	}
	if base.AudioScenario != nil {
		parameters, ok := propsMap["parameters"].(map[string]interface{})
		if !ok || parameters == nil {
			parameters = map[string]interface{}{}
			propsMap["parameters"] = parameters
		}
		parameters["audio_scenario"] = string(*base.AudioScenario)
	}
	ensureDefaultAudioScenario(propsMap)
	if base.Geofence != nil {
		if err := SetStructMap(propsMap, "geofence", base.Geofence); err != nil {
			return nil, err
		}
	}
	if len(base.Labels) > 0 {
		propsMap["labels"] = CloneValue(base.Labels)
	}
	if base.RTC != nil {
		if err := SetStructMap(propsMap, "rtc", base.RTC); err != nil {
			return nil, err
		}
	}
	if base.FillerWords != nil {
		if err := SetStructMap(propsMap, "filler_words", base.FillerWords); err != nil {
			return nil, err
		}
	}

	if base.AdvancedFeatures != nil && base.AdvancedFeatures.EnableRtm != nil && *base.AdvancedFeatures.EnableRtm {
		parameters, ok := propsMap["parameters"].(map[string]interface{})
		if !ok || parameters == nil {
			parameters = map[string]interface{}{}
			propsMap["parameters"] = parameters
		}
		if _, exists := parameters["data_channel"]; !exists {
			parameters["data_channel"] = "rtm"
		}
	}

	if base.MLLM != nil {
		if base.TurnDetection != nil {
			if err := SetStructMap(propsMap, "turn_detection", base.TurnDetection); err != nil {
				return nil, err
			}
		}
		return propsMap, nil
	}

	skipCategories := map[string]bool{}
	for _, category := range opts.SkipVendorValidationCategories {
		skipCategories[category] = true
	}
	allowMissingCategories := map[string]bool{}
	for _, category := range opts.AllowMissingVendorCategories {
		allowMissingCategories[category] = true
	}
	if opts.SkipVendorValidation {
		for _, category := range []string{"asr", "llm", "tts"} {
			skipCategories[category] = true
			allowMissingCategories[category] = true
		}
	}

	turnDetection, err := resolveTurnDetectionConfig(base)
	if err != nil {
		return nil, err
	}
	if base.STT != nil || !allowMissingCategories["asr"] {
		propsMap["asr"] = resolveAsrConfig(profile, base, turnDetection)
	}
	propsMap["turn_detection"] = turnDetection

	if base.TTS == nil && !skipCategories["tts"] && !allowMissingCategories["tts"] {
		return nil, fmt.Errorf("TTS configuration is required; use WithTts() to set it")
	}
	if base.LLM == nil && !skipCategories["llm"] && !allowMissingCategories["llm"] {
		return nil, fmt.Errorf("LLM configuration is required; use WithLlm() to set it")
	}

	if base.LLM != nil {
		propsMap["llm"] = buildLlmConfigMap(base)
	}
	if base.TTS != nil {
		propsMap["tts"] = CloneConfig(base.TTS)
	}

	return propsMap, nil
}

func ensureDefaultAudioScenario(propsMap map[string]interface{}) {
	parameters, ok := propsMap["parameters"].(map[string]interface{})
	if !ok || parameters == nil {
		parameters = map[string]interface{}{}
		propsMap["parameters"] = parameters
	}
	if v, ok := parameters["audio_scenario"].(string); !ok || v == "" {
		parameters["audio_scenario"] = string(ParametersAudioScenarioDefault)
	}
}

func resolveAsrConfig(profile Profile, base *BaseAgent, turnDetection map[string]interface{}) map[string]interface{} {
	asrConfig := CloneConfig(base.STT)
	if asrConfig == nil {
		asrConfig = map[string]interface{}{"vendor": defaultASRVendor(profile)}
	}
	if len(asrConfig) == 0 {
		asrConfig["vendor"] = defaultASRVendor(profile)
	}
	asrConfig["language"] = turnDetection["language"]
	return asrConfig
}

func defaultASRVendor(profile Profile) string {
	if profile == ProfileCN {
		return "fengming"
	}
	return "ares"
}

func resolveTurnDetectionConfig(base *BaseAgent) (map[string]interface{}, error) {
	turnDetection := map[string]interface{}{}
	if base.TurnDetection != nil {
		var err error
		turnDetection, err = StructToMap(base.TurnDetection)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize turn detection config: %w", err)
		}
	}

	language := ""
	if existing, ok := turnDetection["language"].(string); ok && existing != "" {
		language = existing
	}
	if language == "" {
		language = string(defaultTurnDetectionLanguage)
	}
	ValidateTurnDetectionLanguage(language)
	turnDetection["language"] = language
	return turnDetection, nil
}

func buildMllmConfigMap(base *BaseAgent) map[string]interface{} {
	mllmConfig := CloneConfig(base.MLLM)
	if base.Greeting != "" {
		if _, exists := mllmConfig["greeting_message"]; !exists {
			mllmConfig["greeting_message"] = base.Greeting
		}
	}
	if base.FailureMessage != "" {
		if _, exists := mllmConfig["failure_message"]; !exists {
			mllmConfig["failure_message"] = base.FailureMessage
		}
	}
	return mllmConfig
}

func buildLlmConfigMap(base *BaseAgent) map[string]interface{} {
	llmConfig := CloneConfig(base.LLM)
	if base.Instructions != "" {
		if _, exists := llmConfig["system_messages"]; !exists {
			llmConfig["system_messages"] = []map[string]interface{}{
				{"role": "system", "content": base.Instructions},
			}
		}
	}
	if base.Greeting != "" {
		if _, exists := llmConfig["greeting_message"]; !exists {
			llmConfig["greeting_message"] = base.Greeting
		}
	}
	if base.FailureMessage != "" {
		if _, exists := llmConfig["failure_message"]; !exists {
			llmConfig["failure_message"] = base.FailureMessage
		}
	}
	if base.MaxHistory != nil {
		if _, exists := llmConfig["max_history"]; !exists {
			llmConfig["max_history"] = *base.MaxHistory
		}
	}
	if base.GreetingConfigs != nil {
		if _, exists := llmConfig["greeting_configs"]; !exists {
			if value, err := StructToMap(base.GreetingConfigs); err == nil {
				llmConfig["greeting_configs"] = value
			} else {
				llmConfig["greeting_configs"] = base.GreetingConfigs
			}
		}
	}
	return llmConfig
}

func enrichAvatarParams(base *BaseAgent, opts ToPropertiesOptions, tokenFactory TokenFactory) (map[string]interface{}, error) {
	avatar := CloneConfig(base.Avatar)
	if !AvatarConfigEnabled(avatar) {
		return avatar, nil
	}
	vendor, _ := avatar["vendor"].(string)
	params, _ := avatar["params"].(map[string]interface{})
	if params == nil {
		params = map[string]interface{}{}
		avatar["params"] = params
	}

	if IsGenericAvatar(vendor) {
		if _, exists := params["agora_appid"]; !exists && opts.AppID != "" {
			params["agora_appid"] = opts.AppID
		}
		if _, exists := params["agora_channel"]; !exists && opts.Channel != "" {
			params["agora_channel"] = opts.Channel
		}
	}

	avatarUID := avatarUIDString(params["agora_uid"])
	if IsAvatarTokenManaged(vendor) && avatarUID != "" {
		if avatarUID == opts.AgentUID && opts.Warn != nil {
			opts.Warn("avatar agora_uid matches agent_rtc_uid; use a distinct UID so the avatar video stream does not collide with the voice agent")
		}
		if token, _ := params["agora_token"].(string); token == "" {
			if opts.AppCertificate == "" {
				return nil, fmt.Errorf("cannot auto-generate avatar agora_token: appCertificate is required; pass AppCertificate when creating AgoraClient, or set AgoraToken on the avatar vendor")
			}
			uid, err := ParseNumericUID(avatarUID, "avatar agora_uid")
			if err != nil {
				return nil, err
			}
			generated, err := tokenFactory(GenerateConvoAITokenOptions{
				AppID:          opts.AppID,
				AppCertificate: opts.AppCertificate,
				ChannelName:    opts.Channel,
				UID:            uid,
				TokenExpire:    opts.ExpiresIn,
			})
			if err != nil {
				return nil, err
			}
			params["agora_token"] = generated
		}
	}

	return avatar, nil
}

func hasEnabledAvatar(base *BaseAgent) bool {
	return base != nil && base.Avatar != nil && AvatarConfigEnabled(base.Avatar)
}
