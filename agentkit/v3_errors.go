package agentkit

import (
	"fmt"

	"github.com/AgoraIO/agora-agents-go/v2/v3"
)

// AgentKit paths that call v3 report local failures with the v3 error
// taxonomy, so errors.Is(err, v3.ErrValidation) and
// errors.Is(err, v3.ErrConfiguration) work across both layers. Messages keep
// the "agentkit: " prefix.

func invalidV3(message string) error { return &v3.ValidationError{Message: message} }

func invalidV3f(format string, args ...any) error {
	return &v3.ValidationError{Message: fmt.Sprintf(format, args...)}
}

func invalidV3Cause(message string, cause error) error {
	return &v3.ValidationError{Message: message, Err: cause}
}

func misconfiguredV3(message string) error { return &v3.ConfigurationError{Message: message} }
