package agentkit

import (
	"fmt"
	"log"
	"strconv"

	"github.com/AgoraIO-Community/go-tokenbuilder/convoaitokenbuilder"
	"github.com/AgoraIO-Community/go-tokenbuilder/rtctokenbuilder2"
)

const (
	RolePublisher  = 1
	RoleSubscriber = 2

	// DefaultExpirySeconds is the default token lifetime (24 hours, the Agora maximum).
	DefaultExpirySeconds = 86400

	// MaxExpirySeconds is the maximum token lifetime allowed by Agora (24 hours).
	MaxExpirySeconds = 86400
)

// ValidateExpiresIn validates an expiresIn value in seconds.
// Returns an error if secs <= 0.
// Logs a warning and returns MaxExpirySeconds if secs > MaxExpirySeconds.
func ValidateExpiresIn(secs int) (int, error) {
	if secs <= 0 {
		return 0, fmt.Errorf("expiresIn must be between 1 and 86400 seconds (24h)")
	}
	if secs > MaxExpirySeconds {
		log.Println("agora-agent-sdk: expiresIn capped at 24h (Agora max)")
		return MaxExpirySeconds, nil
	}
	return secs, nil
}

// ExpiresInHours converts hours to seconds for use as ExpiresIn.
// Returns an error if hours <= 0; caps at MaxExpirySeconds with a warning.
func ExpiresInHours(hours float64) (int, error) {
	return ValidateExpiresIn(int(hours * 3600))
}

// ExpiresInMinutes converts minutes to seconds for use as ExpiresIn.
// Returns an error if minutes <= 0; caps at MaxExpirySeconds with a warning.
func ExpiresInMinutes(minutes float64) (int, error) {
	return ValidateExpiresIn(int(minutes * 60))
}

type GenerateTokenOptions struct {
	AppID          string
	AppCertificate string
	Channel        string
	UID            uint32
	Role           int
	ExpirySeconds  int
}

type GenerateRtcTokenWithAccountOptions struct {
	AppID          string
	AppCertificate string
	Channel        string
	Account        string
	Role           int
	ExpirySeconds  int
}

// GenerateConvoAITokenOptions configures ConvoAI REST API token generation.
// The token is used as: Authorization: agora token=<token>
type GenerateConvoAITokenOptions struct {
	AppID           string
	AppCertificate  string
	ChannelName     string
	UID             int // Numeric ConvoAI participant UID for a user, agent, or avatar.
	TokenExpire     int // Seconds until expiry (default 86400)
	PrivilegeExpire int // 0 means same as TokenExpire
}

// GenerateRtcToken builds an RTC token for channel access.
// Uses github.com/AgoraIO-Community/go-tokenbuilder.
func GenerateRtcToken(opts GenerateTokenOptions) (string, error) {
	if opts.AppID == "" {
		return "", fmt.Errorf("app_id is required")
	}
	if opts.AppCertificate == "" {
		return "", fmt.Errorf("app_certificate is required")
	}
	if opts.Channel == "" {
		return "", fmt.Errorf("channel is required")
	}
	if opts.ExpirySeconds <= 0 {
		opts.ExpirySeconds = DefaultExpirySeconds
	}
	if opts.Role == 0 {
		opts.Role = RolePublisher
	}

	var role rtctokenbuilder2.Role = rtctokenbuilder2.RolePublisher
	if opts.Role == RoleSubscriber {
		role = rtctokenbuilder2.RoleSubscriber
	}

	expiry := uint32(opts.ExpirySeconds)
	return rtctokenbuilder2.BuildTokenWithUid(
		opts.AppID,
		opts.AppCertificate,
		opts.Channel,
		opts.UID,
		role,
		expiry,
		expiry,
	)
}

// GenerateRtcTokenWithAccount builds an RTC token for a string account.
func GenerateRtcTokenWithAccount(opts GenerateRtcTokenWithAccountOptions) (string, error) {
	if opts.AppID == "" {
		return "", fmt.Errorf("app_id is required")
	}
	if opts.AppCertificate == "" {
		return "", fmt.Errorf("app_certificate is required")
	}
	if opts.Channel == "" {
		return "", fmt.Errorf("channel is required")
	}
	if opts.Account == "" {
		return "", fmt.Errorf("account is required")
	}
	if opts.ExpirySeconds <= 0 {
		opts.ExpirySeconds = DefaultExpirySeconds
	}
	if opts.Role == 0 {
		opts.Role = RolePublisher
	}

	var role rtctokenbuilder2.Role = rtctokenbuilder2.RolePublisher
	if opts.Role == RoleSubscriber {
		role = rtctokenbuilder2.RoleSubscriber
	}

	expiry := uint32(opts.ExpirySeconds)
	return rtctokenbuilder2.BuildTokenWithUserAccount(
		opts.AppID,
		opts.AppCertificate,
		opts.Channel,
		opts.Account,
		role,
		expiry,
		expiry,
	)
}

// GenerateConvoAIToken builds a combined ConvoAI + RTC + RTM token for ConvoAI REST API authentication.
// Use the result as: Authorization: agora token=<token>
//
// Uses github.com/AgoraIO-Community/go-tokenbuilder (convoaitokenbuilder.BuildToken).
func GenerateConvoAIToken(opts GenerateConvoAITokenOptions) (string, error) {
	if opts.AppID == "" {
		return "", fmt.Errorf("app_id is required")
	}
	if opts.AppCertificate == "" {
		return "", fmt.Errorf("app_certificate is required")
	}
	if opts.ChannelName == "" {
		return "", fmt.Errorf("channel_name is required")
	}
	account := convoAIUIDToAccount(opts.UID)
	if account == "" {
		return "", fmt.Errorf("uid is required")
	}
	if opts.TokenExpire <= 0 {
		opts.TokenExpire = DefaultExpirySeconds
	}
	privExpire := opts.PrivilegeExpire
	if privExpire == 0 {
		privExpire = opts.TokenExpire
	}

	return convoaitokenbuilder.BuildToken(
		opts.AppID,
		opts.AppCertificate,
		opts.ChannelName,
		account,
		rtctokenbuilder2.RolePublisher,
		uint32(opts.TokenExpire),
		uint32(privExpire),
		uint32(privExpire),
		uint32(privExpire),
		uint32(privExpire),
		account,
		uint32(privExpire),
	)
}

func convoAIUIDToAccount(uid int) string {
	return strconv.FormatInt(int64(uid), 10)
}
