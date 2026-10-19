package Agora

// GetAgentSessionID returns the v3 session identifier when this response came
// from AgentSession.GetInfo for a v3 session. It is separate from AgentID,
// which identifies legacy v2 agents.
func (g *GetAgentsResponse) GetAgentSessionID() *string {
	if g == nil || g.extraProperties == nil {
		return nil
	}
	value, ok := g.extraProperties["agent_session_id"].(string)
	if !ok {
		return nil
	}
	return &value
}
