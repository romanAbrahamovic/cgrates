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
	"fmt"
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type v1UserProfile struct {
	Tenant   string
	UserName string
	Masked   bool //disable if true
	Profile  map[string]string
	Weight   float64
}

func (ud *v1UserProfile) GetId() string {
	return utils.ConcatenatedKey(ud.Tenant, ud.UserName)
}

func (ud *v1UserProfile) SetId(id string) error {
	vals := strings.Split(id, utils.CONCATENATED_KEY_SEP)
	if len(vals) != 2 {
		return utils.ErrInvalidKey
	}
	ud.Tenant = vals[0]
	ud.UserName = vals[1]
	return nil
}

func userProfile2attributeProfile(user *v1UserProfile) (attr *engine.AttributeProfile) {
	attr = &engine.AttributeProfile{
		Tenant:             user.Tenant,
		ID:                 user.UserName,
		Contexts:           []string{utils.META_ANY},
		FilterIDs:          make([]string, 0),
		ActivationInterval: nil,
		Attributes:         make([]*engine.Attribute, 0),
		Blocker:            false,
		Weight:             user.Weight,
	}
	for fieldname, substitute := range user.Profile {
		attr.Attributes = append(attr.Attributes, &engine.Attribute{
			FieldName:  fieldname,
			Initial:    utils.META_ANY,
			Substitute: config.NewRSRParsersMustCompile(substitute, true, utils.INFIELD_SEP),
			Append:     true,
		})
	}
	return
}

func (m *Migrator) migrateV1User2AttributeProfile() (err error) {
	for {
		user, err := m.dmIN.getV1User()
		if err == utils.ErrNoMoreData {
			break
		}
		if err != nil {
			return err
		}
		if user == nil || user.Masked || m.dryRun {
			continue
		}
		attr := userProfile2attributeProfile(user)
		if len(attr.Attributes) == 0 {
			continue
		}
		if err := m.dmIN.remV1User(user.GetId()); err != nil {
			return err
		}
		if err := m.dmOut.DataManager().DataDB().SetAttributeProfileDrv(attr); err != nil {
			return err
		}
		m.stats[utils.User] += 1
	}
	if m.dryRun {
		return
	}
	// All done, update version wtih current one
	vrs := engine.Versions{utils.User: engine.CurrentDataDBVersions()[utils.User]}
	if err = m.dmOut.DataManager().DataDB().SetVersions(vrs, false); err != nil {
		return utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			err.Error(),
			fmt.Sprintf("error: <%s> when updating Alias version into dataDB", err.Error()))
	}
	return
}

// func (m *Migrator) migrateCurrentUser() (err error) {
// 	var ids []string
// 	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.USERS_PREFIX)
// 	if err != nil {
// 		return err
// 	}
// 	for _, id := range ids {
// 		idg := strings.TrimPrefix(id, utils.USERS_PREFIX)
// 		usr, err := m.dmIN.DataManager().GetUser(idg)
// 		if err != nil {
// 			return err
// 		}
// 		if usr != nil {
// 			if m.dryRun != true {
// 				if err := m.dmOut.DataManager().SetUser(usr); err != nil {
// 					return err
// 				}
// 				m.stats[utils.User] += 1
// 			}
// 		}
// 	}
// 	return
// }

func (m *Migrator) migrateUser() (err error) {
	return m.migrateV1User2AttributeProfile()
	/*
		var vrs engine.Versions
		current := engine.CurrentDataDBVersions()
		vrs, err = m.dmIN.DataManager().DataDB().GetVersions("")
		if err != nil {
			return utils.NewCGRError(utils.Migrator, utils.ServerErrorCaps,
				err.Error(), fmt.Sprintf("error: <%s> when querying oldDataDB for versions", err.Error()))
		} else if len(vrs) == 0 {
			return utils.NewCGRError(utils.Migrator, utils.MandatoryIEMissingCaps,
				utils.UndefinedVersion, "version number is not defined for Users model")
		}
		switch vrs[utils.User] {
		case 1:
			return m.migrateV1User2AttributeProfile()
		case current[utils.User]:
			if !m.sameStorDB {
				return utils.ErrNotImplemented
				// return m.migrateCurrentUser()
			}
		}
		return
	*/
}
