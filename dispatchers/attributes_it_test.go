// +build integration

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package dispatchers

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var sTestsDspAttr = []func(t *testing.T){
	testDspAttrPingFailover,
	testDspAttrGetAttrFailover,

	testDspAttrPing,
	testDspAttrTestMissingApiKey,
	testDspAttrTestUnknownApiKey,
	testDspAttrTestAuthKey,
	testDspAttrTestAuthKey2,
	testDspAttrTestAuthKey3,
}

//Test start here
func TestDspAttributeSTMySQL(t *testing.T) {
	testDsp(t, sTestsDspAttr, "TestDspAttributeS", "all", "all2", "attributes", "dispatchers", "tutorial", "oldtutorial", "dispatchers")
}

func TestDspAttributeSMongo(t *testing.T) {
	testDsp(t, sTestsDspAttr, "TestDspAttributeS", "all", "all2", "attributes_mongo", "dispatchers_mongo", "tutorial", "oldtutorial", "dispatchers")
}

func testDspAttrPingFailover(t *testing.T) {
	var reply string
	if err := allEngine.RCP.Call(utils.AttributeSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	reply = ""
	if err := allEngine2.RCP.Call(utils.AttributeSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	reply = ""
	ev := CGREvWithApiKey{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
		},
		DispatcherResource: DispatcherResource{
			APIKey: "attr12345",
		},
	}
	if err := dispEngine.RCP.Call(utils.AttributeSv1Ping, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	allEngine.stopEngine(t)
	reply = ""
	if err := dispEngine.RCP.Call(utils.AttributeSv1Ping, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	allEngine2.stopEngine(t)
	reply = ""
	if err := dispEngine.RCP.Call(utils.AttributeSv1Ping, &ev, &reply); err == nil {
		t.Errorf("Expected error but recived %v and reply %v\n", err, reply)
	}
	allEngine.startEngine(t)
	allEngine2.startEngine(t)
}

func testDspAttrGetAttrFailover(t *testing.T) {
	args := &ArgsAttrProcessEventWithApiKey{
		DispatcherResource: DispatcherResource{
			APIKey: "attr12345",
		},
		AttrArgsProcessEvent: engine.AttrArgsProcessEvent{
			Context: utils.StringPointer("simpleauth"),
			CGREvent: utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "testAttributeSGetAttributeForEvent",
				Event: map[string]interface{}{
					utils.Account:    "1002",
					utils.EVENT_NAME: "Event1",
				},
			},
		},
	}
	eAttrPrf := &engine.AttributeProfile{
		Tenant:    args.Tenant,
		ID:        "ATTR_1002_SIMPLEAUTH",
		FilterIDs: []string{"*string:Account:1002"},
		Contexts:  []string{"simpleauth"},
		Attributes: []*engine.Attribute{
			{
				FieldName:  "Password",
				Initial:    utils.ANY,
				Substitute: config.NewRSRParsersMustCompile("CGRateS.org", true, utils.INFIELD_SEP),
				Append:     true,
			},
		},
		Weight: 20.0,
	}
	eAttrPrf.Compile()

	eRply := &engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_1002_SIMPLEAUTH"},
		AlteredFields:   []string{"Password"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSGetAttributeForEvent",
			Event: map[string]interface{}{
				utils.Account:    "1002",
				utils.EVENT_NAME: "Event1",
				"Password":       "CGRateS.org",
			},
		},
	}

	var attrReply *engine.AttributeProfile
	var rplyEv engine.AttrSProcessEventReply
	if err := dispEngine.RCP.Call(utils.AttributeSv1GetAttributeForEvent,
		args, &attrReply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	if err := dispEngine.RCP.Call(utils.AttributeSv1ProcessEvent,
		args, &rplyEv); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	} else if reflect.DeepEqual(eRply, &rplyEv) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eRply), utils.ToJSON(rplyEv))
	}

	allEngine2.stopEngine(t)

	if err := dispEngine.RCP.Call(utils.AttributeSv1GetAttributeForEvent,
		args, &attrReply); err != nil {
		t.Error(err)
	}
	if attrReply != nil {
		attrReply.Compile()
	}
	if !reflect.DeepEqual(eAttrPrf, attrReply) {
		t.Errorf("Expecting: %s, received: %s", utils.ToJSON(eAttrPrf), utils.ToJSON(attrReply))
	}

	if err := dispEngine.RCP.Call(utils.AttributeSv1ProcessEvent,
		args, &rplyEv); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRply, &rplyEv) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eRply), utils.ToJSON(rplyEv))
	}

	allEngine2.startEngine(t)
}

func testDspAttrPing(t *testing.T) {
	var reply string
	if err := allEngine.RCP.Call(utils.AttributeSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	if dispEngine.RCP == nil {
		t.Fatal(dispEngine.RCP)
	}
	if err := dispEngine.RCP.Call(utils.AttributeSv1Ping, &CGREvWithApiKey{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
		},
		DispatcherResource: DispatcherResource{
			APIKey: "attr12345",
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
}

func testDspAttrTestMissingApiKey(t *testing.T) {
	args := &ArgsAttrProcessEventWithApiKey{
		AttrArgsProcessEvent: engine.AttrArgsProcessEvent{
			Context: utils.StringPointer("simpleauth"),
			CGREvent: utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "testAttributeSGetAttributeForEvent",
				Event: map[string]interface{}{
					utils.Account: "1001",
				},
			},
		},
	}
	var attrReply *engine.AttributeProfile
	if err := dispEngine.RCP.Call(utils.AttributeSv1GetAttributeForEvent,
		args, &attrReply); err == nil || err.Error() != utils.NewErrMandatoryIeMissing(utils.APIKey).Error() {
		t.Errorf("Error:%v rply=%s", err, utils.ToJSON(attrReply))
	}
}

func testDspAttrTestUnknownApiKey(t *testing.T) {
	args := &ArgsAttrProcessEventWithApiKey{
		DispatcherResource: DispatcherResource{
			APIKey: "1234",
		},
		AttrArgsProcessEvent: engine.AttrArgsProcessEvent{
			Context: utils.StringPointer("simpleauth"),
			CGREvent: utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "testAttributeSGetAttributeForEvent",
				Event: map[string]interface{}{
					utils.Account: "1001",
				},
			},
		},
	}
	var attrReply *engine.AttributeProfile
	if err := dispEngine.RCP.Call(utils.AttributeSv1GetAttributeForEvent,
		args, &attrReply); err == nil || err.Error() != utils.ErrUnknownApiKey.Error() {
		t.Error(err)
	}
}

func testDspAttrTestAuthKey(t *testing.T) {
	args := &ArgsAttrProcessEventWithApiKey{
		DispatcherResource: DispatcherResource{
			APIKey: "12345",
		},
		AttrArgsProcessEvent: engine.AttrArgsProcessEvent{
			Context: utils.StringPointer("simpleauth"),
			CGREvent: utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "testAttributeSGetAttributeForEvent",
				Event: map[string]interface{}{
					utils.Account: "1001",
				},
			},
		},
	}
	var attrReply *engine.AttributeProfile
	if err := dispEngine.RCP.Call(utils.AttributeSv1GetAttributeForEvent,
		args, &attrReply); err == nil || err.Error() != utils.ErrUnauthorizedApi.Error() {
		t.Error(err)
	}
}

func testDspAttrTestAuthKey2(t *testing.T) {
	args := &ArgsAttrProcessEventWithApiKey{
		DispatcherResource: DispatcherResource{
			APIKey: "attr12345",
		},
		AttrArgsProcessEvent: engine.AttrArgsProcessEvent{
			Context: utils.StringPointer("simpleauth"),
			CGREvent: utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "testAttributeSGetAttributeForEvent",
				Event: map[string]interface{}{
					utils.Account: "1001",
				},
			},
		},
	}
	eAttrPrf := &engine.AttributeProfile{
		Tenant:    args.Tenant,
		ID:        "ATTR_1001_SIMPLEAUTH",
		FilterIDs: []string{"*string:Account:1001"},
		Contexts:  []string{"simpleauth"},
		Attributes: []*engine.Attribute{
			{
				FieldName:  "Password",
				Initial:    utils.ANY,
				Substitute: config.NewRSRParsersMustCompile("CGRateS.org", true, utils.INFIELD_SEP),
				Append:     true,
			},
		},
		Weight: 20.0,
	}
	eAttrPrf.Compile()
	var attrReply *engine.AttributeProfile
	if err := dispEngine.RCP.Call(utils.AttributeSv1GetAttributeForEvent,
		args, &attrReply); err != nil {
		t.Error(err)
	}
	if attrReply != nil {
		attrReply.Compile()
	}
	if !reflect.DeepEqual(eAttrPrf, attrReply) {
		t.Errorf("Expecting: %s, received: %s", utils.ToJSON(eAttrPrf), utils.ToJSON(attrReply))
	}

	eRply := &engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_1001_SIMPLEAUTH"},
		AlteredFields:   []string{"Password"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSGetAttributeForEvent",
			Event: map[string]interface{}{
				utils.Account: "1001",
				"Password":    "CGRateS.org",
			},
		},
	}

	var rplyEv engine.AttrSProcessEventReply
	if err := dispEngine.RCP.Call(utils.AttributeSv1ProcessEvent,
		args, &rplyEv); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRply, &rplyEv) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eRply), utils.ToJSON(rplyEv))
	}
}

func testDspAttrTestAuthKey3(t *testing.T) {
	args := &ArgsAttrProcessEventWithApiKey{
		DispatcherResource: DispatcherResource{
			APIKey: "attr12345",
		},
		AttrArgsProcessEvent: engine.AttrArgsProcessEvent{
			Context: utils.StringPointer("simpleauth"),
			CGREvent: utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "testAttributeSGetAttributeForEvent",
				Event: map[string]interface{}{
					utils.Account:    "1001",
					utils.EVENT_NAME: "Event1",
				},
			},
		},
	}
	var attrReply *engine.AttributeProfile
	if err := dispEngine.RCP.Call(utils.AttributeSv1GetAttributeForEvent,
		args, &attrReply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}
