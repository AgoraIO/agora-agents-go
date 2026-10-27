package cn

import (
	"github.com/AgoraIO/agora-agents-go/v2/agentkit"
	"github.com/AgoraIO/agora-agents-go/v2/option"
)

type ClientOptions struct {
	AppID          string
	AppCertificate string
	Token          string
}

type AgoraClient = agentkit.AgoraClient

func NewAgoraClient(opts ClientOptions) *AgoraClient {
	return agentkit.NewAgoraClient(agentkit.AgoraClientOptions{
		Area:           option.AreaCN,
		AppID:          opts.AppID,
		AppCertificate: opts.AppCertificate,
		Token:          opts.Token,
	})
}
