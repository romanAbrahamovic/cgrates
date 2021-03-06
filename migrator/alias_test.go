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

package migrator

import (
	"reflect"
	"sort"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var defaultTenant = "cgrates.org"

func TestAlias2AtttributeProfile(t *testing.T) {
	aliases := map[int]*v1Alias{
		0: {
			Tenant:    utils.META_ANY,
			Direction: utils.META_OUT,
			Category:  utils.META_ANY,
			Account:   utils.META_ANY,
			Subject:   utils.META_ANY,
			Context:   "*rating",
			Values:    v1AliasValues{},
		},
		1: {
			Tenant:    utils.META_ANY,
			Direction: utils.META_OUT,
			Category:  utils.META_ANY,
			Account:   utils.META_ANY,
			Subject:   utils.META_ANY,
			Context:   "*rating",
			Values: v1AliasValues{
				&v1AliasValue{
					DestinationId: utils.META_ANY,
					Pairs: map[string]map[string]string{
						"Account": map[string]string{
							"1001": "1002",
						},
					},
					Weight: 10,
				},
			},
		},
		2: {
			Tenant:    utils.META_ANY,
			Direction: utils.META_OUT,
			Category:  utils.META_ANY,
			Account:   utils.META_ANY,
			Subject:   utils.META_ANY,
			Context:   "*rating",
			Values: v1AliasValues{
				&v1AliasValue{
					DestinationId: utils.META_ANY,
					Pairs: map[string]map[string]string{
						"Account": map[string]string{
							"1001": "1002",
							"1003": "1004",
						},
					},
					Weight: 10,
				},
			},
		},
		3: {
			Tenant:    "",
			Direction: "",
			Category:  "",
			Account:   "",
			Subject:   "",
			Context:   "",
			Values: v1AliasValues{
				&v1AliasValue{
					DestinationId: utils.META_ANY,
					Pairs: map[string]map[string]string{
						"Account": map[string]string{
							"1001": "1002",
							"1003": "1004",
						},
					},
					Weight: 10,
				},
			},
		},
		4: {
			Tenant:    "notDefaultTenant",
			Direction: "*out",
			Category:  "*voice",
			Account:   "1001",
			Subject:   utils.META_ANY,
			Context:   "*rated",
			Values: v1AliasValues{
				&v1AliasValue{
					DestinationId: "DST_1003",
					Pairs: map[string]map[string]string{
						"Account": map[string]string{
							"1001": "1002",
						},
						"Subject": map[string]string{
							"1001": "call_1001",
						},
					},
					Weight: 10,
				},
			},
		},
		5: {
			Tenant:    "notDefaultTenant",
			Direction: "*out",
			Category:  utils.META_ANY,
			Account:   "1001",
			Subject:   "call_1001",
			Context:   "*rated",
			Values: v1AliasValues{
				&v1AliasValue{
					DestinationId: "DST_1003",
					Pairs: map[string]map[string]string{
						"Account": map[string]string{
							"1001": "1002",
						},
						"Category": map[string]string{
							"call_1001": "call_1002",
						},
					},
					Weight: 10,
				},
			},
		},
	}
	expected := map[int]*engine.AttributeProfile{
		0: {
			Tenant:             defaultTenant,
			ID:                 aliases[0].GetId(),
			Contexts:           []string{utils.META_ANY},
			FilterIDs:          make([]string, 0),
			ActivationInterval: nil,
			Attributes:         make([]*engine.Attribute, 0),
			Blocker:            false,
			Weight:             10,
		},
		1: {
			Tenant:             defaultTenant,
			ID:                 aliases[1].GetId(),
			Contexts:           []string{utils.META_ANY},
			FilterIDs:          make([]string, 0),
			ActivationInterval: nil,
			Attributes: []*engine.Attribute{
				{
					FieldName:  "Account",
					Initial:    "1001",
					Substitute: config.NewRSRParsersMustCompile("1002", true, utils.INFIELD_SEP),
					Append:     true,
				},
			},
			Blocker: false,
			Weight:  10,
		},
		2: {
			Tenant:             defaultTenant,
			ID:                 aliases[2].GetId(),
			Contexts:           []string{utils.META_ANY},
			FilterIDs:          make([]string, 0),
			ActivationInterval: nil,
			Attributes: []*engine.Attribute{
				{
					FieldName:  "Account",
					Initial:    "1001",
					Substitute: config.NewRSRParsersMustCompile("1002", true, utils.INFIELD_SEP),
					Append:     true,
				},
				{
					FieldName:  "Account",
					Initial:    "1003",
					Substitute: config.NewRSRParsersMustCompile("1004", true, utils.INFIELD_SEP),
					Append:     true,
				},
			},
			Blocker: false,
			Weight:  10,
		},
		3: {
			Tenant:             defaultTenant,
			ID:                 aliases[3].GetId(),
			Contexts:           []string{utils.META_ANY},
			FilterIDs:          make([]string, 0),
			ActivationInterval: nil,
			Attributes: []*engine.Attribute{
				{
					FieldName:  "Account",
					Initial:    "1001",
					Substitute: config.NewRSRParsersMustCompile("1002", true, utils.INFIELD_SEP),
					Append:     true,
				},
				{
					FieldName:  "Account",
					Initial:    "1003",
					Substitute: config.NewRSRParsersMustCompile("1004", true, utils.INFIELD_SEP),
					Append:     true,
				},
			},
			Blocker: false,
			Weight:  10,
		},
		4: {
			Tenant:   "notDefaultTenant",
			ID:       aliases[4].GetId(),
			Contexts: []string{utils.META_ANY},
			FilterIDs: []string{
				"*string:Category:*voice",
				"*string:Account:1001",
				"*destination:Destination:DST_1003",
			},
			ActivationInterval: nil,
			Attributes: []*engine.Attribute{
				{
					FieldName:  "Account",
					Initial:    "1001",
					Substitute: config.NewRSRParsersMustCompile("1002", true, utils.INFIELD_SEP),
					Append:     true,
				},
				{
					FieldName:  "Subject",
					Initial:    "1001",
					Substitute: config.NewRSRParsersMustCompile("call_1001", true, utils.INFIELD_SEP),
					Append:     true,
				},
			},
			Blocker: false,
			Weight:  10,
		},
		5: {
			Tenant:   "notDefaultTenant",
			ID:       aliases[5].GetId(),
			Contexts: []string{utils.META_ANY},
			FilterIDs: []string{
				"*string:Account:1001",
				"*string:Subject:call_1001",
				"*destination:Destination:DST_1003",
			},
			ActivationInterval: nil,
			Attributes: []*engine.Attribute{
				{
					FieldName:  "Account",
					Initial:    "1001",
					Substitute: config.NewRSRParsersMustCompile("1002", true, utils.INFIELD_SEP),
					Append:     true,
				},
				{
					FieldName:  "Category",
					Initial:    "call_1001",
					Substitute: config.NewRSRParsersMustCompile("call_1002", true, utils.INFIELD_SEP),
					Append:     true,
				},
			},
			Blocker: false,
			Weight:  10,
		},
	}
	for i := range expected {
		rply := alias2AtttributeProfile(aliases[i], defaultTenant)
		sort.Slice(rply.Attributes, func(i, j int) bool {
			if rply.Attributes[i].FieldName == rply.Attributes[j].FieldName {
				return rply.Attributes[i].Initial.(string) < rply.Attributes[j].Initial.(string)
			}
			return rply.Attributes[i].FieldName < rply.Attributes[j].FieldName
		}) // only for test; map returns random keys
		if !reflect.DeepEqual(expected[i], rply) {
			t.Errorf("For %v expected: %s ,recived: %s ", i, utils.ToJSON(expected[i]), utils.ToJSON(rply))
		}
	}
}
