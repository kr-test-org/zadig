/*
 * Copyright 2022 The KodeRover Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package handler

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/koderover/zadig/v2/pkg/microservice/aslan/core/common/repository/models"
	"github.com/koderover/zadig/v2/pkg/microservice/aslan/core/system/service"
	"github.com/koderover/zadig/v2/pkg/setting"
	"github.com/koderover/zadig/v2/pkg/shared/client/plutusvendor"
	internalhandler "github.com/koderover/zadig/v2/pkg/shared/handler"
	e "github.com/koderover/zadig/v2/pkg/tool/errors"
	"github.com/koderover/zadig/v2/pkg/tool/jira"
	"github.com/koderover/zadig/v2/pkg/tool/meego"
)

func ListProjectManagement(c *gin.Context) {
	ctx, err := internalhandler.NewContextWithAuthorization(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	if err != nil {

		ctx.Err = fmt.Errorf("authorization Info Generation failed: err %s", err)
		ctx.UnAuthorized = true
		return
	}

	// authorization checks
	if !ctx.Resources.IsSystemAdmin {
		ctx.UnAuthorized = true
		return
	}

	ctx.Resp, ctx.Err = service.ListProjectManagement(ctx.Logger)
}

// @Summary List Project Management For Project
// @Description List Project Management For Project
// @Tags 	system
// @Accept 	json
// @Produce json
// @Success 200 	{array} 	models.ProjectManagement
// @Router /api/aslan/system/project_management/project [get]
func ListProjectManagementForProject(c *gin.Context) {
	ctx, err := internalhandler.NewContextWithAuthorization(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	if err != nil {
		ctx.Logger.Errorf("failed to generate authorization info for user: %s, error: %s", ctx.UserID, err)
		ctx.Err = fmt.Errorf("authorization Info Generation failed: err %s", err)
		ctx.UnAuthorized = true
		return
	}

	pms, err := service.ListProjectManagement(ctx.Logger)
	for _, pm := range pms {
		pm.JiraToken = ""
		pm.JiraUser = ""
		pm.JiraAuthType = ""
		pm.MeegoPluginID = ""
		pm.MeegoPluginSecret = ""
		pm.MeegoUserKey = ""
	}
	ctx.Err = err
	ctx.Resp = pms
}

func CreateProjectManagement(c *gin.Context) {
	ctx, err := internalhandler.NewContextWithAuthorization(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	if err != nil {
		ctx.Err = fmt.Errorf("authorization Info Generation failed: err %s", err)
		ctx.UnAuthorized = true
		return
	}

	// authorization checks
	if !ctx.Resources.IsSystemAdmin {
		ctx.UnAuthorized = true
		return
	}

	req := new(models.ProjectManagement)
	if err := c.ShouldBindJSON(req); err != nil {
		ctx.Err = err
		return
	}

	licenseStatus, err := plutusvendor.New().CheckZadigXLicenseStatus()
	if err != nil {
		ctx.Err = fmt.Errorf("failed to validate zadig license status, error: %s", err)
		return
	}
	if req.MeegoHost != "" {
		if !((licenseStatus.Type == plutusvendor.ZadigSystemTypeProfessional ||
			licenseStatus.Type == plutusvendor.ZadigSystemTypeEnterprise) &&
			licenseStatus.Status == plutusvendor.ZadigXLicenseStatusNormal) {
			ctx.Err = e.ErrLicenseInvalid.AddDesc("")
			return
		}
	}

	ctx.Err = service.CreateProjectManagement(req, ctx.Logger)
}

func UpdateProjectManagement(c *gin.Context) {
	ctx, err := internalhandler.NewContextWithAuthorization(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	if err != nil {
		ctx.Err = fmt.Errorf("authorization Info Generation failed: err %s", err)
		ctx.UnAuthorized = true
		return
	}

	// authorization checks
	if !ctx.Resources.IsSystemAdmin {
		ctx.UnAuthorized = true
		return
	}

	req := new(models.ProjectManagement)
	if err := c.ShouldBindJSON(req); err != nil {
		ctx.Err = err
		return
	}

	licenseStatus, err := plutusvendor.New().CheckZadigXLicenseStatus()
	if err != nil {
		ctx.Err = fmt.Errorf("failed to validate zadig license status, error: %s", err)
		return
	}
	if req.MeegoHost != "" {
		if !((licenseStatus.Type == plutusvendor.ZadigSystemTypeProfessional ||
			licenseStatus.Type == plutusvendor.ZadigSystemTypeEnterprise) &&
			licenseStatus.Status == plutusvendor.ZadigXLicenseStatusNormal) {
			ctx.Err = e.ErrLicenseInvalid.AddDesc("")
			return
		}
	}

	ctx.Err = service.UpdateProjectManagement(c.Param("id"), req, ctx.Logger)
}

func DeleteProjectManagement(c *gin.Context) {
	ctx, err := internalhandler.NewContextWithAuthorization(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	if err != nil {

		ctx.Err = fmt.Errorf("authorization Info Generation failed: err %s", err)
		ctx.UnAuthorized = true
		return
	}

	// authorization checks
	if !ctx.Resources.IsSystemAdmin {
		ctx.UnAuthorized = true
		return
	}

	ctx.Err = service.DeleteProjectManagement(c.Param("id"), ctx.Logger)
}

func Validate(c *gin.Context) {
	ctx, err := internalhandler.NewContextWithAuthorization(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	if err != nil {

		ctx.Err = fmt.Errorf("authorization Info Generation failed: err %s", err)
		ctx.UnAuthorized = true
		return
	}

	// authorization checks
	if !ctx.Resources.IsSystemAdmin {
		ctx.UnAuthorized = true
		return
	}

	req := new(models.ProjectManagement)
	if err := c.ShouldBindJSON(req); err != nil {
		ctx.Err = err
		return
	}
	switch req.Type {
	case setting.PMJira:
		ctx.Err = service.ValidateJira(req)
	case setting.PMMeego:
		ctx.Err = service.ValidateMeego(req)
	default:
		ctx.Err = e.ErrValidateProjectManagement.AddDesc("invalid type")
	}
}

// @Summary List Jira Projects
// @Description List Jira Projects
// @Tags 	system
// @Accept 	json
// @Produce json
// @Param 	id 		path		string										true	"jira id"
// @Success 200 	{array} 	service.JiraProjectsResp
// @Router /api/aslan/system/project_management/{id}/jira/project [get]
func ListJiraProjects(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()
	ctx.Resp, ctx.Err = service.ListJiraProjects(c.Param("id"))
}

func SearchJiraIssues(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()
	ctx.Resp, ctx.Err = service.SearchJiraIssues(c.Param("id"), c.Query("project"), c.Query("type"), c.Query("status"), c.Query("summary"), c.Query("ne") == "true")
}

func SearchJiraProjectIssuesWithJQL(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	// 5.22 JQL only support {{.system.username}} variable
	// refactor if more variables are needed
	ctx.Resp, ctx.Err = service.SearchJiraProjectIssuesWithJQL(c.Param("id"), c.Query("project"), strings.ReplaceAll(c.Query("jql"), "{{.system.username}}", ctx.UserName), c.Query("summary"))
}

func GetJiraTypes(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()
	ctx.Resp, ctx.Err = service.GetJiraTypes(c.Param("id"), c.Query("project"))
}

func GetJiraAllStatus(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()
	ctx.Resp, ctx.Err = service.GetJiraAllStatus(c.Param("id"), c.Query("project"))
}

func HandleJiraEvent(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()
	event := new(jira.Event)
	if err := c.ShouldBindJSON(event); err != nil {
		ctx.Err = err
		return
	}

	ctx.Err = service.HandleJiraHookEvent(c.Param("workflowName"), c.Param("hookName"), event, ctx.Logger)
}

func HandleMeegoEvent(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()
	event := new(meego.GeneralWebhookRequest)
	if err := c.ShouldBindJSON(event); err != nil {
		ctx.Err = err
		return
	}

	ctx.Err = service.HandleMeegoHookEvent(c.Param("workflowName"), c.Param("hookName"), event, ctx.Logger)
}
