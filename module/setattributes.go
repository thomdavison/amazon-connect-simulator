package module

import (
	"fmt"

	"github.com/edwardbrowncross/amazon-connect-simulator/call"
	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

type setAttributes flow.Module

type setAttributesParams struct {
	Attribute []call.KeyValue
}

func (m setAttributes) Run(ctx *call.Context) (next *flow.ModuleID, err error) {
	if m.Type != flow.ModuleSetAttributes {
		return nil, fmt.Errorf("module of type %s being run as setAttributes", m.Type)
	}
	p := setAttributesParams{}
	err = ctx.UnmarshalParameters(m.Parameters, &p)
	if err != nil {
		return
	}
	for _, a := range p.Attribute {
		ctx.ContactData[a.K] = a.V
	}
	return m.Branches.GetLink(flow.BranchSuccess), nil
}