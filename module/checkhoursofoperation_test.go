package module

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/edwardbrowncross/amazon-connect-simulator/event"
)

func TestCheckHoursOfOperation(t *testing.T) {
	jsonOK := `{
		"id":"43dcc4f2-3392-4a38-90ed-0216f8594ea8",
		"type":"CheckHoursOfOperation",
		"branches":[
			{"condition":"True", "transition":"00000000-0000-4000-0000-000000000001"},
			{"condition":"False","transition":"00000000-0000-4000-0000-000000000002"},
			{"condition":"Error","transition":"00000000-0000-4000-0000-000000000003"}
		],
		"parameters":[]
	}`
	jsonBadMod := `{
		"id":"43dcc4f2-3392-4a38-90ed-0216f8594ea8",
		"type":"Disconnect"
	}`
	testCases := []struct {
		desc    string
		module  string
		context CallConnector
		exp     string
		expEvt  []event.Event
		expErr  string
	}{
		{
			desc:   "wrong module",
			module: jsonBadMod,
			expErr: "module of type Disconnect being run as checkHoursOfOperation",
		},
		{
			desc:   "success",
			module: jsonOK,
			exp:    "00000000-0000-4000-0000-000000000001",
			expEvt: []event.Event{},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			var mod checkHoursOfOperation
			err := json.Unmarshal([]byte(tC.module), &mod)
			if err != nil {
				t.Fatalf("unexpected error unmarshalling module: %v", err)
			}
			state := testCallState{}.init()
			next, err := mod.Run(state)
			errStr := ""
			if err != nil {
				errStr = err.Error()
			}
			if errStr != tC.expErr {
				t.Errorf("expected error of '%s' but got '%s'", tC.expErr, errStr)
			}
			nextStr := ""
			if next != nil {
				nextStr = string(*next)
			}
			if nextStr != tC.exp {
				t.Errorf("expected next of '%s' but got '%v'", tC.exp, *next)
			}
			if (tC.expEvt != nil && !reflect.DeepEqual(tC.expEvt, state.events)) || (tC.expEvt == nil && len(state.events) > 0) {
				t.Errorf("expected events of '%v' but got '%v'", tC.expEvt, state.events)
			}
		})
	}
}
