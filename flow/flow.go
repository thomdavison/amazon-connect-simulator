package flow

import "encoding/json"

// ModuleID is a uuid assigned to a block in the flow.
type ModuleID string

// ModuleType describes the type of block in the flow.
type ModuleType string

// ModuleTarget is set on some blocks to modify its behavior.
type ModuleTarget string

//ModuleBranchCondition is a class of output of a block.
type ModuleBranchCondition string

// ModuleBranchConditionType is an operator for an Evaluate branch describing how to route out of a block.
type ModuleBranchConditionType string

// ModuleParameterNamespace indicates what namespace a dynamic value should be looked up from.
type ModuleParameterNamespace string

// SystemKey is a valid key that can be dynamically looked up from the connect system.
type SystemKey string

// Known types of block.
const (
	ModuleStoreUserInput         ModuleType = "StoreUserInput"
	ModuleCheckAttribute                    = "CheckAttribute"
	ModuleTransfer                          = "Transfer"
	ModulePlayPrompt                        = "PlayPrompt"
	ModuleDisconnect                        = "Disconnect"
	ModuleSetQueue                          = "SetQueue"
	ModuleGetUserInput                      = "GetUserInput"
	ModuleSetAttributes                     = "SetAttributes"
	ModuleInvokeExternalResource            = "InvokeExternalResource"
	ModuleCheckHoursOfOperation             = "CheckHoursOfOperation"
	ModuleSetVoice                          = "SetVoice"
)

// Known types of block no longer in use in new flows.
const (
	ModuleDeprecatedSetScreenPop       ModuleType = "SetScreenPop"
	ModuleDeprecatedStoreCustomerInput            = "StoreCustomerInput"
	ModuleDeprecatedPlayAudio                     = "PlayAudio"
	ModuleDeprecatedTransferToFlow                = "TransferToFlow"
	ModuleDeprecatedCustomerInQueue               = "CustomerInQueue"
)

// Values used in the module's target field.
const (
	TargetFlow        ModuleTarget = "Flow"
	TargetLambda                   = "Lambda"
	TargetQueue                    = "Queue"
	TargetDigits                   = "Digits"
	TargetPhoneNumber              = "PhoneNumber"
)

// The three places you can look up a dynamic value.
const (
	NamespaceExternal    ModuleParameterNamespace = "External"
	NamespaceSystem                               = "System"
	NamespaceUserDefined                          = "User Defined"
)

// Known named reasons for choosing an output of a block.
const (
	BranchSuccess    ModuleBranchCondition = "Success"
	BranchError                            = "Error"
	BranchNoMatch                          = "NoMatch"
	BranchEvaluate                         = "Evaluate"
	BranchTimeout                          = "Timeout"
	BranchTrue                             = "True"
	BranchFalse                            = "False"
	BranchAtCapacity                       = "AtCapacity"
)

// Operators for Evaluate branches.
const (
	ConditionEquals ModuleBranchConditionType = "Equals"
	ConditionGTE                              = "GreaterThanOrEqualTo"
	ConditionGT                               = "GreaterThan"
	ConditionLTE                              = "LessThanOrEqualTo"
	ConditionLT                               = "LessThan"
)

// Values that can be dynamically looked up from the Connect system.
const (
	SystemLastUserInput       SystemKey = "Stored customer input"
	SystemCustomerNumber                = "Customer Number"
	SystemDialedNumber                  = "Dialed Number"
	SystemCustomerCallback              = "Customer callback number"
	SystemQueueName                     = "Queue.Name"
	SystemQueueARN                      = "Queue.ARN"
	SystemQueueOutboundNumber           = "Queue.OutboundCallerId.Address"
	SystemTextToSpeechVoice             = "TextToSpeechVoiceId"
	SystemContactID                     = "ContactId"
	SystemInitialContactID              = "InitialContactId"
	SystemPreviousContactID             = "PreviousContactId"
	SystemChannel                       = "Channel"
	SystemInstanceARN                   = "InstanceARN"
	SystemInitiationMethod              = "InitiationMethod"
)

// Flow is the base of the XML structure of an exported flow.
type Flow struct {
	Modules  []Module `json:"modules"`
	Start    ModuleID `json:"start"`
	Metadata Metadata `json:"metadata"`
}

// Metadata holds metadata about the flow (which appears in the top left of the Connect flow UI)
type Metadata struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

// Module holds data about a single block in the flow.
type Module struct {
	ID         ModuleID            `json:"id"`
	Type       ModuleType          `json:"type"`
	Branches   ModuleBranchList    `json:"branches"`
	Parameters ModuleParameterList `json:"parameters"`
	Metadata   json.RawMessage     `json:"metadata"`
	Target     ModuleTarget        `json:"target"`
}

// ModuleBranchList is a list of outputs of a block.
type ModuleBranchList []ModuleBranch

// GetLink gets the ID of the block linked to by the output with name.
// This is used when there is a single output with that name (eg. Success)
func (mbl ModuleBranchList) GetLink(named ModuleBranchCondition) *ModuleID {
	for _, p := range mbl {
		if p.Condition == named {
			return &p.Transition
		}
	}
	return nil
}

// List gets a list of branches which have the given name.
// This is used when there are multiple branches of that type (eg. Evaluate)
func (mbl ModuleBranchList) List(named ModuleBranchCondition) []ModuleBranch {
	r := []ModuleBranch{}
	for _, b := range mbl {
		if b.Condition == named {
			r = append(r, b)
		}
	}
	return r
}

// ModuleBranch is a single output of a block and the data required to choose it.
type ModuleBranch struct {
	Condition      ModuleBranchCondition     `json:"condition"`
	ConditionType  ModuleBranchConditionType `json:"conditionType"`
	ConditionValue interface{}               `json:"conditionValue"`
	Transition     ModuleID                  `json:"transition"`
}

// ModuleParameterList is a list of parameters configuring a block.
// Each type of block has different parameters that are set.
type ModuleParameterList []ModuleParameter

// Get gets a single parameter with the given name.
// Use it when there is only one parameter with that name.
func (mpl ModuleParameterList) Get(named string) (ModuleParameter, bool) {
	for _, p := range mpl {
		if p.Name == named {
			return p, true
		}
	}
	return ModuleParameter{}, false
}

// List gets a list of parameters with the given name.
// Use it when there are mutiple parameters with the same name (eg. lambda inputs).
func (mpl ModuleParameterList) List(named string) []ModuleParameter {
	r := []ModuleParameter{}
	for _, p := range mpl {
		if p.Name == named {
			r = append(r, p)
		}
	}
	return r
}

// ModuleParameter is a single parameter configuring block.
type ModuleParameter struct {
	// The name of the parameter.
	Name string `json:"name"`
	// Optional. Used when parameter represents a key,value pair (eg. lambda parameter).
	Key string `json:"key"`
	// Either a raw value if namespace is not set or a key to look up in the namespace.
	Value interface{} `json:"value"`
	// Namespace in which to look up dynamic values.
	Namespace *ModuleParameterNamespace `json:"namespace"`
	// Optional. Gives a friendly name to ARNs set in Value.
	ResourceName string `json:"resourceName"`
}

// KeyValue represents the parsed value of key-value parameter.
type KeyValue struct {
	K string
	V string
}
