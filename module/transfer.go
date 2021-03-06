package module

import (
	"errors"
	"fmt"

	"github.com/edwardbrowncross/amazon-connect-simulator/event"
	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

type transfer flow.Module

func (m transfer) Run(call CallConnector) (next *flow.ModuleID, err error) {
	if m.Type != flow.ModuleTransfer {
		return nil, fmt.Errorf("module of type %s being run as transfer", m.Type)
	}
	switch m.Target {
	case flow.TargetFlow:
		cfid, ok := m.Parameters.Get("ContactFlowId")
		if !ok {
			return nil, errors.New("missing ContactFlowId parameter")
		}
		next = call.GetFlowStart(cfid.ResourceName)
		if next == nil {
			return m.Branches.GetLink(flow.BranchError), nil
		}
		call.Emit(event.FlowTransferEvent{FlowARN: cfid.Value.(string), FlowName: cfid.ResourceName})
		return next, nil
	case flow.TargetQueue:
		queue := call.GetSystem(flow.SystemQueueName)
		arn := call.GetSystem(flow.SystemQueueARN)
		if queue == nil || arn == nil {
			return m.Branches.GetLink(flow.BranchError), nil
		}
		call.Emit(event.QueueTransferEvent{QueueARN: *arn, QueueName: *queue})
		return nil, nil
	case flow.TargetPhoneNumber:
		blind, ok := m.Parameters.Get("BlindTransfer")
		if _, isBool := blind.Value.(bool); !ok || !isBool {
			return nil, errors.New("missing BlindTransfer parameter")
		}
		num, ok := m.Parameters.Get("PhoneNumber")
		if _, isString := num.Value.(string); !ok || !isString {
			return nil, errors.New("missing PhoneNumber parameter")
		}
		call.Emit(event.NumberTransferEvent{Tel: num.Value.(string)})
		if blind.Value.(bool) {
			return nil, nil
		}
		return m.Branches.GetLink(flow.BranchSuccess), nil
	default:
		return nil, fmt.Errorf("unhandled transfer target: %s", m.Target)
	}
}
