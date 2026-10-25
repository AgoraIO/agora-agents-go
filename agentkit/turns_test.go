package agentkit

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	Agora "github.com/AgoraIO/agora-agents-go/v2"
	"github.com/AgoraIO/agora-agents-go/v2/client"
	"github.com/AgoraIO/agora-agents-go/v2/option"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTurnEndPlaybackMetadata(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		body     string
		actual   *int
		intended *int
	}{
		{
			name:   "normal completion",
			body:   `{"type":"ok","metadata":{"playback_duration_ms":300}}`,
			actual: Agora.Int(300),
		},
		{
			name: "interrupted playback",
			body: `{"type":"interrupted","metadata":{
				"playback_duration_ms":120,"intended_playback_duration_ms":300,"caused_by":"start_of_speech"
			}}`,
			actual: Agora.Int(120), intended: Agora.Int(300),
		},
		{
			name: "zero durations",
			body: `{"type":"interrupted","metadata":{
				"playback_duration_ms":0,"intended_playback_duration_ms":0
			}}`,
			actual: Agora.Int(0), intended: Agora.Int(0),
		},
		{
			name: "legacy interrupted response",
			body: `{"type":"interrupted","metadata":{"caused_by":"api_interrupt"}}`,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var end Agora.GetTurnsAgentsResponseTurnsItemEnd
			require.NoError(t, json.Unmarshal([]byte(tt.body), &end))
			require.NotNil(t, end.Metadata)
			assert.Equal(t, tt.actual, end.Metadata.GetPlaybackDurationMs())
			assert.Equal(t, tt.intended, end.Metadata.GetIntendedPlaybackDurationMs())
			encoded, err := json.Marshal(&end)
			require.NoError(t, err)
			assert.JSONEq(t, tt.body, string(encoded))
		})
	}
}

func TestTurnEndMetadataExplicitNullFields(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		set  func(*Agora.GetTurnsAgentsResponseTurnsItemEndMetadata)
		want string
	}{
		{name: "playback", set: func(m *Agora.GetTurnsAgentsResponseTurnsItemEndMetadata) {
			m.SetPlaybackDurationMs(nil)
		}, want: `{"playback_duration_ms":null}`},
		{name: "intended playback", set: func(m *Agora.GetTurnsAgentsResponseTurnsItemEndMetadata) {
			m.SetIntendedPlaybackDurationMs(nil)
		}, want: `{"intended_playback_duration_ms":null}`},
		{name: "cause", set: func(m *Agora.GetTurnsAgentsResponseTurnsItemEndMetadata) {
			m.SetCausedBy(nil)
		}, want: `{"caused_by":null}`},
		{name: "transport", set: func(m *Agora.GetTurnsAgentsResponseTurnsItemEndMetadata) {
			m.SetTransport(nil)
		}, want: `{"transport":null}`},
		{name: "reason", set: func(m *Agora.GetTurnsAgentsResponseTurnsItemEndMetadata) {
			m.SetReason(nil)
		}, want: `{"reason":null}`},
		{name: "details", set: func(m *Agora.GetTurnsAgentsResponseTurnsItemEndMetadata) {
			m.SetDetails(nil)
		}, want: `{"details":null}`},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			metadata := &Agora.GetTurnsAgentsResponseTurnsItemEndMetadata{}
			tt.set(metadata)
			encoded, err := json.Marshal(metadata)
			require.NoError(t, err)
			assert.JSONEq(t, tt.want, string(encoded))
		})
	}
}

func TestGetAllTurnsPreservesPlaybackMetadata(t *testing.T) {
	t.Parallel()
	httpClient := &turnsHTTPClient{pages: []string{}}
	rawClient := client.NewClient(option.WithBaseURL("https://example.test"), option.WithHTTPClient(httpClient))
	session := &AgentSession{client: rawClient.Agents, appID: "appid", agentID: "agent-id"}

	response, err := session.GetAllTurns(context.Background(), GetAllTurnsOptions{PageSize: Agora.Int(1)})
	require.NoError(t, err)
	require.Len(t, response.Turns, 2)
	assert.Equal(t, []string{"1", "2"}, httpClient.pages)
	for _, turn := range response.Turns {
		require.NotNil(t, turn.End)
		require.NotNil(t, turn.End.Metadata)
		assert.Equal(t, Agora.Int(0), turn.End.Metadata.GetPlaybackDurationMs())
		assert.Equal(t, Agora.Int(300), turn.End.Metadata.GetIntendedPlaybackDurationMs())
	}
}

type turnsHTTPClient struct {
	pages []string
}

func (c *turnsHTTPClient) Do(req *http.Request) (*http.Response, error) {
	page := req.URL.Query().Get("page_index")
	c.pages = append(c.pages, page)
	body := fmt.Sprintf(`{
		"pagination":{"page_index":%s,"is_last_page":%t},
		"turns":[{"turn_id":%s,"end":{"type":"interrupted","metadata":{
			"playback_duration_ms":0,"intended_playback_duration_ms":300
		}}}]
	}`, page, page == "2", page)
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}, nil
}
