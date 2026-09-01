package utils

import (
	"context"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
)

func AgentID(ctx context.Context) model.ID {
	agentID := ctx.Value("agent_id")
	if agentID == nil {
		return model.ID("")
	}
	return agentID.(model.ID)
}
