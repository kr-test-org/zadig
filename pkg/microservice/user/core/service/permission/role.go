/*
Copyright 2023 The KodeRover Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package permission

import (
	"fmt"

	"github.com/koderover/zadig/v2/pkg/microservice/user/core/repository"
	"github.com/koderover/zadig/v2/pkg/microservice/user/core/repository/models"
	"github.com/koderover/zadig/v2/pkg/microservice/user/core/repository/orm"
	"github.com/koderover/zadig/v2/pkg/setting"
	"github.com/koderover/zadig/v2/pkg/types"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/util/sets"
)

// ActionMap is the local cache for all the actions' ID, the key is the action name
// Note that there is no way to change action after the service start, the local cache won't
// have an expiration mechanism.
var ActionMap = make(map[string]uint)

type CreateRoleReq struct {
	Name      string   `json:"name"`
	Actions   []string `json:"actions"`
	Namespace string   `json:"namespace"`
	Desc      string   `json:"desc,omitempty"`
	Type      string   `json:"type,omitempty"`
}

func CreateRole(ns string, req *CreateRoleReq, log *zap.SugaredLogger) error {
	tx := repository.DB.Begin()

	role := &models.NewRole{
		Name:        req.Name,
		Description: req.Desc,
		Namespace:   ns,
	}

	if req.Type == string(setting.ResourceTypeSystem) {
		role.Type = int64(setting.RoleTypeSystem)
	} else {
		role.Type = int64(setting.RoleTypeCustom)
	}

	err := orm.CreateRole(role, tx)
	if err != nil {
		log.Errorf("failed to create role, error: %s", err)
		tx.Rollback()
		return fmt.Errorf("failed to create role, error: %s", err)
	}

	actionIDList := make([]uint, 0)
	for _, action := range req.Actions {
		// if the action is not in the action cache, get one.
		if _, ok := ActionMap[action]; !ok {
			act, err := orm.GetActionByVerb(action, repository.DB)
			if err != nil {
				log.Errorf("failed to find verb: %s in request, action might not exist.", action)
				tx.Rollback()
				return fmt.Errorf("failed to find verb: %s in request, action might not exist", action)
			}
			ActionMap[action] = act.ID
		}
		actionIDList = append(actionIDList, ActionMap[action])
	}

	err = orm.BulkCreateRoleActionBindings(role.ID, actionIDList, tx)
	if err != nil {
		log.Errorf("failed to create action binding for role: %s in namespace: %s, the error is: %s", role.Name, role.Namespace, err)
		tx.Rollback()
		return fmt.Errorf("failed to create action binding for role: %s in namespace: %s, the error is: %s", role.Name, role.Namespace, err)
	}

	tx.Commit()

	return nil
}

// UpdateRole updates the role and its action binding.
func UpdateRole(ns string, req *CreateRoleReq, log *zap.SugaredLogger) error {
	tx := repository.DB.Begin()

	// Doing a tricky thing here: removing the whole role-action binding, then re-adding them.
	roleInfo, err := orm.GetRole(req.Name, ns, repository.DB)
	if err != nil {
		log.Errorf("failed to find role: [%s] in namespace [%s], error: %s", req.Namespace, ns, err)
		tx.Rollback()
		return fmt.Errorf("failed to find role: [%s] in namespace [%s], error: %s", req.Namespace, ns, err)
	}

	err = orm.DeleteRoleActionBindingByRole(roleInfo.ID, tx)
	if err != nil {
		log.Errorf("failed to delete role-action binding for role: %s, error: %s", roleInfo.Name, err)
		tx.Rollback()
		return fmt.Errorf("update role-action binding failed, error: %s", err)
	}

	actionIDList := make([]uint, 0)
	for _, action := range req.Actions {
		// if the action is not in the action cache, get one.
		if _, ok := ActionMap[action]; !ok {
			act, err := orm.GetActionByVerb(action, repository.DB)
			if err != nil {
				log.Errorf("failed to find verb: %s in request, action might not exist.", action)
				tx.Rollback()
				return fmt.Errorf("failed to find verb: %s in request, action might not exist", action)
			}
			ActionMap[action] = act.ID
		}
		actionIDList = append(actionIDList, ActionMap[action])
	}

	err = orm.BulkCreateRoleActionBindings(roleInfo.ID, actionIDList, tx)
	if err != nil {
		log.Errorf("failed to create action binding for role: %s in namespace: %s, the error is: %s", roleInfo.Name, roleInfo.Namespace, err)
		tx.Rollback()
		return fmt.Errorf("failed to create action binding for role: %s in namespace: %s, the error is: %s", roleInfo.Name, roleInfo.Namespace, err)
	}

	// so the only field capable of changing is the description....
	err = orm.UpdateRoleInfo(roleInfo.ID, &models.NewRole{
		Description: req.Desc,
	}, tx)

	tx.Commit()

	return nil
}

func ListRolesByNamespace(projectName string, log *zap.SugaredLogger) ([]*types.Role, error) {
	roles, err := orm.ListRoleByNamespace(projectName, repository.DB)
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Errorf("failed to list roles in project: %s, error: %s", projectName, err)
		return nil, fmt.Errorf("failed to list roles in project: %s, error: %s", projectName, err)
	}

	resp := make([]*types.Role, 0)
	for _, role := range roles {
		resp = append(resp, &types.Role{
			ID:          role.ID,
			Name:        role.Name,
			Namespace:   role.Namespace,
			Description: role.Description,
			Type:        convertDBRoleType(role.Type),
		})
	}

	return resp, nil
}

func ListRolesByNamespaceAndUserID(projectName, uid string, log *zap.SugaredLogger) ([]*types.Role, error) {
	roles, err := orm.ListRoleByUIDAndNamespace(uid, projectName, repository.DB)
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Errorf("failed to list roles in project: %s, error: %s", projectName, err)
		return nil, fmt.Errorf("failed to list roles in project: %s, error: %s", projectName, err)
	}

	resp := make([]*types.Role, 0)
	for _, role := range roles {
		resp = append(resp, &types.Role{
			ID:          role.ID,
			Name:        role.Name,
			Namespace:   role.Namespace,
			Description: role.Description,
			Type:        convertDBRoleType(role.Type),
		})
	}

	return resp, nil
}

func GetRole(ns, name string, log *zap.SugaredLogger) (*types.DetailedRole, error) {
	role, err := orm.GetRole(name, ns, repository.DB)
	if err != nil {
		log.Errorf("failed to find role: %s under namespace: %s, error: %s", name, ns, err)
		return nil, fmt.Errorf("failed to find role: %s under namespace: %s, error: %s", name, ns, err)
	}

	actionList, err := orm.ListActionByRole(role.ID, repository.DB)
	if err != nil {
		log.Errorf("failed to find action for role: %s under namespace: %s, error: %s", name, ns, err)
		return nil, fmt.Errorf("failed to find action for role: %s under namespace: %s, error: %s", name, ns, err)
	}

	actionMap := make(map[string]sets.String)

	for _, action := range actionList {
		if _, ok := actionMap[action.Resource]; !ok {
			actionMap[action.Resource] = sets.NewString()
		}

		actionMap[action.Resource].Insert(action.Action)
	}

	resourceActionList := make([]*types.ResourceAction, 0)
	for resource, actionSet := range actionMap {
		resourceActionList = append(resourceActionList, &types.ResourceAction{
			Resource: resource,
			Verbs:    actionSet.List(),
		})
	}

	resp := &types.DetailedRole{
		ID:              role.ID,
		Name:            role.Name,
		Namespace:       role.Namespace,
		Description:     role.Description,
		Type:            convertDBRoleType(role.Type),
		ResourceActions: resourceActionList,
	}

	return resp, nil
}

func DeleteRole(name string, projectName string, log *zap.SugaredLogger) error {
	err := orm.DeleteRoleByName(name, projectName, repository.DB)
	if err != nil {
		log.Errorf("failed to delete role: %s under namespace %s, error: %s", name, projectName, err)
		return fmt.Errorf("failed to delete role: %s under namespace %s, error: %s", name, projectName, err)
	}
	return nil
}

func CreateDefaultRolesForNamespace(namespace string, log *zap.SugaredLogger) error {
	projectAdminRole := &models.NewRole{
		Name:        "project-admin",
		Description: "",
		Type:        int64(setting.RoleTypeSystem),
		Namespace:   namespace,
	}
	readOnlyRole := &models.NewRole{
		Name:        "read-only",
		Description: "",
		Type:        int64(setting.RoleTypeSystem),
		Namespace:   namespace,
	}
	readProjectOnlyRole := &models.NewRole{
		Name:        "read-project-only",
		Description: "",
		Type:        int64(setting.RoleTypeSystem),
		Namespace:   namespace,
	}

	err := orm.BulkCreateRole([]*models.NewRole{projectAdminRole, readOnlyRole, readProjectOnlyRole}, repository.DB)
	if err != nil {
		log.Errorf("failed to create system default role for project: %s, error: %s", namespace, err)
		return fmt.Errorf("failed to create system default role for project: %s, error: %s", namespace, err)
	}

	return nil
}

func DeleteAllRolesInNamespace(namespace string, log *zap.SugaredLogger) error {
	err := orm.DeleteRoleByNameSpace(namespace, repository.DB)
	if err != nil {
		log.Errorf("failed to delete role under namespace %s, error: %s", namespace, err)
		return fmt.Errorf("failed to delete role under namespace %s, error: %s", namespace, err)
	}
	return nil
}

func convertDBRoleType(tid int64) string {
	if tid == int64(setting.RoleTypeSystem) {
		return string(setting.ResourceTypeSystem)
	}
	return string(setting.ResourceTypeCustom)
}
