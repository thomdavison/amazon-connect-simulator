package module

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/edwardbrowncross/amazon-connect-simulator/event"
	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

func TestStoreUserInput(t *testing.T) {
	jsonBadMod := `{
		"id":"43dcc4f2-3392-4a38-90ed-0216f8594ea8",
		"type":"Transfer"
	}`
	jsonBadParam := `{
		"id":"43dcc4f2-3392-4a38-90ed-0216f8594ea8",
		"type":"StoreUserInput",
		"parameters":[]
	}`
	jsonBadTimeout := `{
		"id":"55c7b51c-ab55-4c63-ac42-235b4a0f904f",
		"type":"StoreUserInput",
		"branches":[],
		"parameters":[
			{"name":"Text","value":"prompt","namespace":"External"},
			{"name":"Timeout","value":"fishcake"},
			{"name":"MaxDigits","value":8},
			{"name":"TextToSpeechType","value":"text"},
			{"name":"EncryptEntry","value":false}
		]
	}`
	jsonOK := `{
		"id":"55c7b51c-ab55-4c63-ac42-235b4a0f904f",
		"type":"StoreUserInput",
		"branches":[
			{"condition":"Success","transition":"00000000-0000-4000-0000-000000000001"},
			{"condition":"Error","transition":"00000000-0000-4000-0000-000000000002"}
		],
		"parameters":[
			{"name":"Text","value":"prompt","namespace":"External"},
			{"name":"TextToSpeechType","value":"ssml"},
			{"name":"CustomerInputType","value":"Custom"},
			{"name":"Timeout","value":"7"},
			{"name":"MaxDigits","value":8},
			{"name":"EncryptEntry","value":false},
			{"name":"DisableCancel","value":true}
		]
	}`
	jsonOKEncrypted := `{
		"id":"55c7b51c-ab55-4c63-ac42-235b4a0f904f",
		"type":"StoreUserInput",
		"branches":[
			{"condition":"Success","transition":"00000000-0000-4000-0000-000000000001"},
			{"condition":"Error","transition":"00000000-0000-4000-0000-000000000002"}
		],
		"parameters":[
			{"name":"Text","value":"hello"},
			{"name":"TextToSpeechType","value":"ssml"},
			{"name":"CustomerInputType","value":"Custom"},
			{"name":"Timeout","value":"7"},
			{"name":"MaxDigits","value":8},
			{"name":"EncryptEntry","value":true},
			{"name":"DisableCancel","value":true},
			{"name":"EncryptionKeyId","value":"fa0ef83f-fbfa-4acd-9bdb-a8564d35490e","namespace":null},
			{"name":"EncryptionKey","value":"-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----","namespace":null}
		]
	}`
	jsonOKCustomTerminator := `{
		"id":"55c7b51c-ab55-4c63-ac42-235b4a0f904f",
		"type":"StoreUserInput",
		"branches":[
			{"condition":"Success","transition":"00000000-0000-4000-0000-000000000001"},
			{"condition":"Error","transition":"00000000-0000-4000-0000-000000000002"}
		],
		"parameters":[
			{"name":"Text","value":"hello"},
			{"name":"TextToSpeechType","value":"text"},
			{"name":"CustomerInputType","value":"Custom"},
			{"name":"Timeout","value":"7"},
			{"name":"MaxDigits","value":8},
			{"name":"EncryptEntry","value":false},
			{"name":"DisableCancel","value":false},
			{"name":"TerminatorDigits","value":"0"}
		]
	}`
	testCases := []struct {
		desc             string
		module           string
		state            *testCallState
		exp              string
		expPrompt        string
		expErr           string
		expSys           map[flow.SystemKey]string
		expRcvTimeout    time.Duration
		expRcvCount      int
		expRcvTerminator rune
		expEvt           []event.Event
	}{
		{
			desc:   "wrong module",
			module: jsonBadMod,
			expErr: "module of type Transfer being run as storeUserInput",
		},
		{
			desc:   "missing parameter",
			module: jsonBadParam,
			expErr: "missing parameter Text",
		},
		{
			desc:   "bad timeout parameter",
			module: jsonBadTimeout,
			expErr: `strconv.Atoi: parsing "fishcake": invalid syntax`,
		},
		{
			desc:   "timeout",
			module: jsonOK,
			state: testCallState{
				i: "Timeout",
				external: map[string]string{
					"prompt": "<speak>Please enter digits 1 and 3 of your passcode.</speak>",
				},
			}.init(),
			exp:              "00000000-0000-4000-0000-000000000001",
			expEvt:           []event.Event{},
			expPrompt:        "<speak>Please enter digits 1 and 3 of your passcode.</speak>",
			expRcvTerminator: '#',
		},
		{
			desc:   "success",
			module: jsonOK,
			exp:    "00000000-0000-4000-0000-000000000001",
			state: testCallState{
				i: "12345678",
				external: map[string]string{
					"prompt": "<speak>Please enter digits $.External.digits of your passcode.</speak>",
					"digits": "1 and 3",
				},
			}.init(),
			expSys: map[flow.SystemKey]string{
				flow.SystemLastUserInput: "12345678",
			},
			expPrompt:        "<speak>Please enter digits 1 and 3 of your passcode.</speak>",
			expRcvCount:      8,
			expRcvTimeout:    7 * time.Second,
			expEvt:           []event.Event{},
			expRcvTerminator: '#',
		},
		{
			desc:   "success - encrypted",
			module: jsonOKEncrypted,
			exp:    "00000000-0000-4000-0000-000000000001",
			state: testCallState{
				i: "12345678",
				encrypt: func(in string, id string, cert []byte) []byte {
					if id != "fa0ef83f-fbfa-4acd-9bdb-a8564d35490e" {
						t.Errorf("expected encryption ID of %s but got %s", "fa0ef83f-fbfa-4acd-9bdb-a8564d35490e", id)
					}
					if !strings.HasPrefix(string(cert), "-----BEGIN CERTIFICATE-----") {
						t.Errorf("invalid certificate: %s", string(cert))
					}
					return []byte(fmt.Sprintf("<encrypt>%s</encrypt>", in))
				},
			}.init(),
			expSys: map[flow.SystemKey]string{
				flow.SystemLastUserInput: base64.StdEncoding.EncodeToString([]byte("<encrypt>12345678</encrypt>")),
			},
			expPrompt:        "hello",
			expRcvCount:      8,
			expRcvTimeout:    7 * time.Second,
			expEvt:           []event.Event{},
			expRcvTerminator: '#',
		},
		{
			desc:   "success - custom terminator",
			module: jsonOKCustomTerminator,
			exp:    "00000000-0000-4000-0000-000000000001",
			state: testCallState{
				i: "12345678",
			}.init(),
			expSys: map[flow.SystemKey]string{
				flow.SystemLastUserInput: "12345678",
			},
			expPrompt:        "hello",
			expRcvCount:      8,
			expRcvTimeout:    7 * time.Second,
			expEvt:           []event.Event{},
			expRcvTerminator: '0',
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			var mod storeUserInput
			err := json.Unmarshal([]byte(tC.module), &mod)
			if err != nil {
				t.Fatalf("unexpected error unmarshalling module: %v", err)
			}
			state := tC.state
			if state == nil {
				state = testCallState{}.init()
			}
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
				t.Errorf("expected next of '%s' but got '%s'", tC.exp, nextStr)
			}
			for k, v := range tC.expSys {
				if state.system[k] != v {
					t.Errorf("expected system %s to be '%s' but it was '%s'", k, v, state.system[k])
				}
			}
			if state.o != tC.expPrompt {
				t.Errorf("expected prompt of '%s' but got '%s'", tC.expPrompt, state.o)
			}
			if tC.expRcvCount > 0 && state.rcv.count != tC.expRcvCount {
				t.Errorf("expected receive count of %d but got %d", tC.expRcvCount, state.rcv.count)
			}
			if tC.expRcvTimeout > 0 && state.rcv.timeout != tC.expRcvTimeout {
				t.Errorf("expected receive timeout of %d but got %d", tC.expRcvTimeout, state.rcv.timeout)
			}
			if state.rcv.terminator != tC.expRcvTerminator {
				t.Errorf("expected receive terminator of %s but got %s", string(tC.expRcvTerminator), string(state.rcv.terminator))
			}
			if (tC.expEvt != nil && !reflect.DeepEqual(tC.expEvt, state.events)) || (tC.expEvt == nil && len(state.events) > 0) {
				t.Errorf("expected events of '%v' but got '%v'", tC.expEvt, state.events)
			}

		})
	}
}
