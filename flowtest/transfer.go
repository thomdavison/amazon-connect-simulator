package flowtest

import (
	"fmt"

	"github.com/edwardbrowncross/amazon-connect-simulator/event"
)

// TransferContext is returned from Expect.Transfer()
type TransferContext struct {
	testContext
}

// ToQueue asserts that the caller is transferred to a queue with the given name.
func (tc TransferContext) ToQueue(named string) {
	tc.run(queueTransferMatcher{named})
}

// ToFlow asserts that the call moved to the flow with the given name.
func (tc TransferContext) ToFlow(named string) {
	tc.run(flowTransferMatcher{named})
}

// Never asserts that the following assertions will never match for the durtion of the call.
func (tc TransferContext) Never() TransferContext {
	tc.never()
	return tc
}

// Unordered suspends the implicit assertion that events occur in the flow in the order you assert them in your tests.
func (tc TransferContext) Unordered() TransferContext {
	tc.unordered()
	return tc
}

type queueTransferMatcher struct {
	queueName string
}

func (m queueTransferMatcher) match(evt event.Event) (match bool, pass bool, got string) {
	if evt.Type() != event.TransferQueueType {
		return false, false, ""
	}
	e := evt.(event.QueueTransferEvent)
	match = true
	got = e.QueueName
	pass = bool(e.QueueName == m.queueName)
	return
}

func (m queueTransferMatcher) expected() string {
	return fmt.Sprintf("to be transfered to queue '%s'", m.queueName)
}

type flowTransferMatcher struct {
	flowName string
}

func (m flowTransferMatcher) match(evt event.Event) (match bool, pass bool, got string) {
	if evt.Type() != event.TransferFlowType {
		return false, false, ""
	}
	e := evt.(event.FlowTransferEvent)
	match = true
	got = e.FlowName
	pass = bool(e.FlowName == m.flowName)
	return
}

func (m flowTransferMatcher) expected() string {
	return fmt.Sprintf("to be transfered to flow '%s'", m.flowName)
}
