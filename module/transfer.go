package module

import (
	"fmt"

	"github.com/edwardbrowncross/amazon-connect-simulator/call"
	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

type transfer flow.Module

func (m transfer) Run(ctx *call.Context) (next *flow.ModuleID, err error) {
	if m.Type != flow.ModuleTransfer {
		return nil, fmt.Errorf("module of type %s being run as transfer", m.Type)
	}
	switch m.Target {
	case flow.TargetFlow:
		fName := m.Parameters.Get("ContactFlowId").ResourceName
		return ctx.GetFlowStart(fName), nil
	case flow.TargetQueue:
		ctx.Send(fmt.Sprintf("<transfer to queue %s>", ctx.System[flow.SystemQueueName]))
		return nil, nil
	default:
		return nil, fmt.Errorf("unhandled transfer target: %s", m.Target)
	}
}