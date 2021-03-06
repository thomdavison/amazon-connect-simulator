package simulator

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

// Simulator is capable of starting new simulated call flows.
type Simulator struct {
	lambdas   map[string]interface{}
	flows     map[string]flow.Flow
	modules   map[flow.ModuleID]flow.Module
	encrypt   func(string, string, []byte) []byte
	isInHours func(string, bool, time.Time) (bool, error)
	telFlow   map[string]flow.Flow
}

// New creates a new call simulator.
// It is created blank and must be set up using its attached methods.
func New() Simulator {
	return Simulator{
		lambdas:   map[string]interface{}{},
		flows:     map[string]flow.Flow{},
		modules:   map[flow.ModuleID]flow.Module{},
		telFlow:   map[string]flow.Flow{},
		encrypt:   func(in string, keyID string, cert []byte) []byte { return []byte(in) },
		isInHours: func(string, bool, time.Time) (bool, error) { return true, nil },
	}
}

// LoadFlow loads an unmarshalled call flow into the simulator.
// Do this with all flows that form part of your call flows before starting a call.
func (cs *Simulator) LoadFlow(f flow.Flow) {
	f = flow.Dedeprecate(f)
	cs.flows[f.Metadata.Name] = f
	for _, m := range f.Modules {
		cs.modules[m.ID] = m
	}
}

// LoadFlowJSON takes a byte array containing a json file exported from Amazon Connect.
// It does the same thing as LoadFlow, except that it does the unmarshalling for you.
func (cs *Simulator) LoadFlowJSON(bytes []byte) error {
	f := flow.Flow{}
	err := json.Unmarshal(bytes, &f)
	if err != nil {
		return err
	}
	cs.LoadFlow(f)
	return nil
}

// Flows returns the flows currently loaded into the simulator.
func (cs *Simulator) Flows() []flow.Flow {
	r := make([]flow.Flow, len(cs.flows))
	i := 0
	for _, f := range cs.flows {
		r[i] = f
		i++
	}
	return r
}

// RegisterLambda specifies how external lambda invocations will be handled.
// name is a string that forms part of the lambda's ARN (such as its name).
// fn is function like handle(context.Context, struct) (struct, error). It will be passed an Amazon Connect lambda event.
// You must specify a function for each external lambda invocation before starting a simulated call.
func (cs *Simulator) RegisterLambda(name string, fn interface{}) error {
	err := validateLambda(fn)
	if err != nil {
		return err
	}
	cs.lambdas[name] = fn
	return nil
}

// SetStartingFlowFor specifies the name of the flow that should be run when a new call comes in to a given number.
// The telephone number should match what will be used when creating a call.
// The name is the full name given to the flow in the Amazon Connect ui.
// You must run this once before starting a simulated call.
func (cs *Simulator) SetStartingFlowFor(tel string, flowName string) error {
	f, ok := cs.flows[flowName]
	if !ok {
		return errors.New("starting flow not found. Load the flow with LoadFlow before calling this method")
	}
	cs.telFlow[tel] = f
	return nil
}

// SetEncryption defines how encryption is performed when encryption is enable in a Store Customer Input block.
// No encryption is currently supplied by this simulator. By default, the string is no encrypted.
// You may supply a function that takes the input digits and returns a cipher string. This may be real encryption or a dummy process.
func (cs *Simulator) SetEncryption(encryptor func(in string, keyID string, cert []byte) (encrypted []byte)) {
	cs.encrypt = encryptor
}

// SetInHoursCheck adds logic used by the checkHoursOfOperation block to determine if we are in operating hours.
// The first parameter of the provided function will either be the name of a Queue or the name of an Hours of Operation, as indicated by the second parameter.
// It should return true if we are in operating hours and false if not.
// Returning an error indicates that the given queue/hours does not exist or does not have hours defined. The call will proceed down the error path.
func (cs *Simulator) SetInHoursCheck(checker func(name string, isQueue bool, time time.Time) (inOperation bool, err error)) {
	cs.isInHours = checker
}

// StartCall starts a new call asynchronously and returns a Call object for interacting with that call.
// Many independent calls can be spawned from one simulator.
func (cs *Simulator) StartCall(config CallConfig) (*Call, error) {
	if config.DestNumber == "" {
		return nil, errors.New("a destination number must be provided in order to start a flow")
	}
	start, ok := cs.telFlow[config.DestNumber]
	if !ok {
		return nil, errors.New("no starting flow set. Call SetStartingFlowFor before starting a call")
	}
	return newCall(config, &simulatorConnector{cs}, start.Start), nil
}

// simulatorConnector exposes methods for modules to get information from the base simulator.
type simulatorConnector struct {
	*Simulator
}

// GetLambda gets a lamda using a partial ARN match.
func (cs *simulatorConnector) GetLambda(arn string) interface{} {
	for k, v := range cs.lambdas {
		if strings.Contains(arn, k) {
			return v
		}
	}
	return nil
}

// GetFlowStart gets the module ID of the block at the start of a flow with the given name.
func (cs *simulatorConnector) GetFlowStart(flowName string) *flow.ModuleID {
	f, ok := cs.flows[flowName]
	if !ok {
		return nil
	}
	return &f.Start
}

// GetModule finds the block with the given ID.
func (cs *simulatorConnector) GetModule(moduleID flow.ModuleID) *flow.Module {
	m, ok := cs.modules[moduleID]
	if !ok {
		return nil
	}
	return &m
}

func (cs *simulatorConnector) Encrypt(in string, keyID string, cert []byte) []byte {
	return cs.encrypt(in, keyID, cert)
}

func (cs *simulatorConnector) InvokeLambda(named string, withJSON string) (outJSON string, outErr error, err error) {
	fn := cs.GetLambda(named)
	if fn == nil {
		return "", nil, fmt.Errorf("unknown lambda: %s", named)
	}
	return invokeLambda(fn, withJSON)
}

func (cs *simulatorConnector) IsInHours(name string, isQueue bool, time time.Time) (bool, error) {
	return cs.isInHours(name, isQueue, time)
}
