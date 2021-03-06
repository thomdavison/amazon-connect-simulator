package module

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/edwardbrowncross/amazon-connect-simulator/event"
	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

type testCallState struct {
	i     string
	o     string
	oSSML bool
	rcv   struct {
		count      int
		timeout    time.Duration
		encrypt    bool
		terminator rune
		cert       []byte
		keyID      string
	}
	encrypt  func(string, string, []byte) []byte
	lambdaIn struct {
		name  string
		input json.RawMessage
	}
	lambdaOut    string
	lambdaOutErr error
	lambdaErr    error
	external     map[string]string
	contactData  map[string]string
	system       map[flow.SystemKey]string
	flowStart    map[string]flow.ModuleID
	events       []event.Event
	inHours      func(string, bool, time.Time) (bool, error)
	time         time.Time
}

func (st testCallState) init() *testCallState {
	if st.external == nil {
		st.external = map[string]string{}
	}
	if st.contactData == nil {
		st.contactData = map[string]string{}
	}
	if st.system == nil {
		st.system = map[flow.SystemKey]string{}
	}
	if st.flowStart == nil {
		st.flowStart = map[string]flow.ModuleID{}
	}
	st.events = make([]event.Event, 0)
	return &st
}

func (st *testCallState) Send(s string, ssml bool) {
	st.o = s
	st.oSSML = ssml
}
func (st *testCallState) Receive(count int, timeout time.Duration, terminator rune) (string, bool) {
	st.rcv.count = count
	st.rcv.timeout = timeout
	st.rcv.terminator = terminator
	if st.i == "" {
		return "", false
	}
	return st.i, st.i != "Timeout"
}
func (st *testCallState) GetExternal(key string) *string {
	val, found := st.external[key]
	if !found {
		return nil
	}
	return &val
}
func (st *testCallState) SetExternal(key string, value interface{}) {
	st.external[key] = fmt.Sprintf("%v", value)
}
func (st *testCallState) ClearExternal() {
	st.external = map[string]string{}
}
func (st *testCallState) GetContactData(key string) *string {
	val, found := st.contactData[key]
	if !found {
		return nil
	}
	return &val
}
func (st *testCallState) SetContactData(key string, value string) {
	st.contactData[key] = value
}
func (st *testCallState) GetSystem(key flow.SystemKey) *string {
	val, found := st.system[key]
	if !found {
		return nil
	}
	return &val
}
func (st *testCallState) SetSystem(key flow.SystemKey, value string) {
	st.system[key] = value
}
func (st *testCallState) InvokeLambda(named string, inParams json.RawMessage, timeout time.Duration) (outJSON string, outErr error, err error) {
	st.lambdaIn.name = named
	st.lambdaIn.input = inParams
	return st.lambdaOut, st.lambdaOutErr, st.lambdaErr
}
func (st *testCallState) GetFlowStart(flowName string) *flow.ModuleID {
	r := st.flowStart[flowName]
	if r == "" {
		return nil
	}
	return &r
}
func (st *testCallState) Emit(event event.Event) {
	st.events = append(st.events, event)
}
func (st *testCallState) IsInHours(name string, isQueue bool) (bool, error) {
	return st.inHours(name, isQueue, st.time)
}
func (st *testCallState) Encrypt(in string, keyID string, cert []byte) []byte {
	return st.encrypt(in, keyID, cert)
}

func TestMakeRunner(t *testing.T) {
	testCases := []struct {
		desc   string
		module string
		exp    Runner
	}{
		{
			desc:   "StoreUserInput",
			module: `{ "type": "StoreUserInput" }`,
			exp:    storeUserInput{},
		},
		{
			desc:   "CheckAttribute",
			module: `{ "type": "CheckAttribute" }`,
			exp:    checkAttribute{},
		},
		{
			desc:   "Transfer",
			module: `{ "type": "Transfer" }`,
			exp:    transfer{},
		},
		{
			desc:   "PlayPrompt",
			module: `{ "type": "PlayPrompt" }`,
			exp:    playPrompt{},
		},
		{
			desc:   "Disconnect",
			module: `{ "type": "Disconnect" }`,
			exp:    disconnect{},
		},
		{
			desc:   "SetQueue",
			module: `{ "type": "SetQueue" }`,
			exp:    setQueue{},
		},
		{
			desc:   "GetUserInput",
			module: `{ "type": "GetUserInput" }`,
			exp:    getUserInput{},
		},
		{
			desc:   "SetAttributes",
			module: `{ "type": "SetAttributes" }`,
			exp:    setAttributes{},
		},
		{
			desc:   "InvokeExternalResource",
			module: `{ "type": "InvokeExternalResource" }`,
			exp:    invokeExternalResource{},
		},
		{
			desc:   "CheckHoursOfOperation",
			module: `{ "type": "CheckHoursOfOperation" }`,
			exp:    checkHoursOfOperation{},
		},
		{
			desc:   "SetVoice",
			module: `{ "type": "SetVoice" }`,
			exp:    setVoice{},
		},
		{
			desc:   "Passthrough",
			module: `{ "type": "WhatIsThisIDontEven" }`,
			exp:    passthrough{},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			m := flow.Module{}
			err := json.Unmarshal([]byte(tC.module), &m)
			if err != nil {
				t.Fatalf("unexpected error parsing module: %v", err)
			}
			r := MakeRunner(m)
			if reflect.TypeOf(r) != reflect.TypeOf(tC.exp) {
				t.Errorf("expected type of %v but got %v", reflect.TypeOf(tC.exp), reflect.TypeOf(r))
			}
		})
	}
}
