// Copyright 2021 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

const (
	ResponseTypeLogin   = "login"
	ResponseTypeCode    = "code"
	ResponseTypeToken   = "token"
	ResponseTypeIdToken = "id_token"
	ResponseTypeSaml    = "saml"
	ResponseTypeCas     = "cas"
)

type RequestForm struct {
	Type string `json:"type"`

	Organization string `json:"organization"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	Name         string `json:"name"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	Affiliation  string `json:"affiliation"`
	IdCard       string `json:"idCard"`
	Region       string `json:"region"`

	Application string `json:"application"`
	Provider    string `json:"provider"`
	Code        string `json:"code"`
	State       string `json:"state"`
	RedirectUri string `json:"redirectUri"`
	Method      string `json:"method"`

	EmailCode   string `json:"emailCode"`
	PhoneCode   string `json:"phoneCode"`
	PhonePrefix string `json:"phonePrefix"`

	AutoSignin bool `json:"autoSignin"`

	RelayState   string `json:"relayState"`
	SamlRequest  string `json:"samlRequest"`
	SamlResponse string `json:"samlResponse"`
}

type OTTSignUpRequest struct {
	AppId            string  `json:"app_id"`                    // light-wallet for light-wallet app
	Type             int     `json:"type"`                      // Register type, 0: phone, 1: email
	Prefix           *string `json:"prefix,omitempty"`          // phone prefix (if type is 0)
	Identity         string  `json:"identity"`                  // phone number or email address (if type is 0 or 1)
	VerificationCode string  `json:"verification_code"`         // verification code according type
	Pwd              string  `json:"pwd"`                       // password
	InvitationCode   *string `json:"invitation_code,omitempty"` // invitation code
}

type Response struct {
	Status string      `json:"status"`
	Msg    string      `json:"msg"`
	Sub    string      `json:"sub"`
	Name   string      `json:"name"`
	Data   interface{} `json:"data"`
	Data2  interface{} `json:"data2"`
}

type OTTResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Body interface{} `json:"body"`
}

type OTTSendVerificationCodeResponse struct {
	Timer int `json:"timer"` // seconds to wait before next request
}

type OTTRegisterResponse struct {
	UserId string `json:"user_id"`
}

type HumanCheck struct {
	Type         string      `json:"type"`
	AppKey       string      `json:"appKey"`
	Scene        string      `json:"scene"`
	CaptchaId    string      `json:"captchaId"`
	CaptchaImage interface{} `json:"captchaImage"`
}

// Signup
// @Tag Login API
// @Title Signup
// @Description sign up a new user
// @Param   username     formData    string  true        "The username to sign up"
// @Param   password     formData    string  true        "The password"
// @Success 200 {object} controllers.Response The Response object
// @router /signup [post]
func (c *ApiController) Signup() {
	if c.GetSessionUsername() != "" {
		c.ResponseError("Please sign out first before signing up", c.GetSessionUsername())
		return
	}

	var form RequestForm
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &form)
	if err != nil {
		panic(err)
	}

	application := object.GetApplication(fmt.Sprintf("admin/%s", form.Application))
	if !application.EnableSignUp {
		c.ResponseError("The application does not allow to sign up new account")
		return
	}

	organization := object.GetOrganization(fmt.Sprintf("%s/%s", "admin", form.Organization))
	msg := object.CheckUserSignup(application, organization, form.Username, form.Password, form.Name, form.FirstName, form.LastName, form.Email, form.Phone, form.Affiliation)
	if msg != "" {
		c.ResponseError(msg)
		return
	}

	if application.IsSignupItemVisible("Email") && application.GetSignupItemRule("Email") != "No verification" && form.Email != "" {
		checkResult := object.CheckVerificationCode(form.Email, form.EmailCode)
		if len(checkResult) != 0 {
			c.ResponseError(fmt.Sprintf("Email: %s", checkResult))
			return
		}
	}

	var checkPhone string
	if application.IsSignupItemVisible("Phone") && form.Phone != "" {
		checkPhone = fmt.Sprintf("+%s%s", form.PhonePrefix, form.Phone)
		checkResult := object.CheckVerificationCode(checkPhone, form.PhoneCode)
		if len(checkResult) != 0 {
			c.ResponseError(fmt.Sprintf("Phone: %s", checkResult))
			return
		}
	}

	id := util.GenerateId()
	if application.GetSignupItemRule("ID") == "Incremental" {
		lastUser := object.GetLastUser(form.Organization)

		lastIdInt := -1
		if lastUser != nil {
			lastIdInt = util.ParseInt(lastUser.Id)
		}

		id = strconv.Itoa(lastIdInt + 1)
	}

	username := form.Username
	if !application.IsSignupItemVisible("Username") {
		username = id
	}

	user := &object.User{
		Owner:             form.Organization,
		Name:              username,
		CreatedTime:       util.GetCurrentTime(),
		Id:                id,
		Type:              "normal-user",
		Password:          form.Password,
		DisplayName:       form.Name,
		Avatar:            organization.DefaultAvatar,
		Email:             form.Email,
		Phone:             form.Phone,
		Address:           []string{},
		Affiliation:       form.Affiliation,
		IdCard:            form.IdCard,
		Region:            form.Region,
		Score:             getInitScore(),
		IsAdmin:           false,
		IsGlobalAdmin:     false,
		IsForbidden:       false,
		IsDeleted:         false,
		SignupApplication: application.Name,
		Properties:        map[string]string{},
		Karma:             0,
	}

	if len(organization.Tags) > 0 {
		tokens := strings.Split(organization.Tags[0], "|")
		if len(tokens) > 0 {
			user.Tag = tokens[0]
		}
	}

	if application.GetSignupItemRule("Display name") == "First, last" {
		if form.FirstName != "" || form.LastName != "" {
			user.DisplayName = fmt.Sprintf("%s %s", form.FirstName, form.LastName)
			user.FirstName = form.FirstName
			user.LastName = form.LastName
		}
	}

	affected := object.AddUser(user)
	if !affected {
		c.ResponseError(fmt.Sprintf("Failed to create user, user information is invalid: %s", util.StructToJson(user)))
		return
	}

	object.AddUserToOriginalDatabase(user)

	if application.HasPromptPage() {
		// The prompt page needs the user to be signed in
		c.SetSessionUsername(user.GetId())
	}

	object.DisableVerificationCode(form.Email)
	object.DisableVerificationCode(checkPhone)

	record := object.NewRecord(c.Ctx)
	record.Organization = application.Organization
	record.User = user.Name
	util.SafeGoroutine(func() { object.AddRecord(record) })

	userId := fmt.Sprintf("%s/%s", user.Owner, user.Name)
	util.LogInfo(c.Ctx, "API: [%s] is signed up as new user", userId)

	c.ResponseOk(userId)
}

func (c *ApiController) OTTSignup() {
	if c.GetSessionUsername() != "" {
		c.OTTResponseError(OTT_CODE_NEED_SIGN_OUT, "Please sign out first before signing up")
		return
	}

	var form OTTSignUpRequest
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &form)
	if err != nil {
		c.OTTResponseError(OTT_CODE_INVALID_PARAM, err.Error())
		return
	}

	application := object.GetApplication(fmt.Sprintf("admin/%s", form.AppId))
	if !application.EnableSignUp {
		c.OTTResponseError(OTT_CODE_APPLICATION_NO_SIGNUP, "The application does not allow to sign up new account")
		return
	}

	organization := object.GetOrganization(fmt.Sprintf("%s/%s", "admin", OTT_ORGANIZATION_ID))
	msg := object.OTTCheckUserSignup(application, organization, form.Type, "", form.Pwd, "", "", "", form.Identity, form.Prefix, "")
	if msg != "" {
		c.OTTResponseError(OTT_CODE_INVALID_PARAM, msg)
		return
	}

	var email = ""
	var phone = ""
	var region = ""
	var checkPhone string
	switch form.Type {
	case 1:
		if application.IsSignupItemVisible("Email") && application.GetSignupItemRule("Email") != "No verification" && form.Identity != "" {
			checkResult := object.CheckVerificationCode(form.Identity, form.VerificationCode)
			if len(checkResult) != 0 {
				c.OTTResponseError(OTT_CODE_VERIFICATION_CODE_NOT_MATCH, fmt.Sprintf("Email: %s", checkResult))
				return
			}
			email = form.Identity
		}
	case 0:
		if application.IsSignupItemVisible("Phone") && form.Identity != "" && form.Prefix != nil && *form.Prefix != "" {
			checkPhone = fmt.Sprintf("+%s%s", *form.Prefix, form.Identity)
			checkResult := object.CheckVerificationCode(checkPhone, form.VerificationCode)
			if len(checkResult) != 0 {
				c.OTTResponseError(OTT_CODE_VERIFICATION_CODE_NOT_MATCH, fmt.Sprintf("Phone: %s", checkResult))
				return
			}
			phone = checkPhone
			region = *form.Prefix
		}
	default:
		break
	}

	id := util.GenerateId()
	// TODO concurrent issue
	/*if application.GetSignupItemRule("ID") == "Incremental" {
		// TODO: concurrent issue
		lastUser := object.GetLastUser(OTT_ORGANIZATION_ID)

		lastIdInt := -1
		if lastUser != nil {
			lastIdInt = util.ParseInt(lastUser.Id)
		}

		id = strconv.Itoa(lastIdInt + 1)
	}*/

	username := id

	user := &object.User{
		Owner:             OTT_ORGANIZATION_ID,
		Name:              username,
		CreatedTime:       util.GetCurrentTime(),
		Id:                id,
		Type:              "normal-user",
		Password:          form.Pwd,
		DisplayName:       id,
		Avatar:            organization.DefaultAvatar,
		Email:             email,
		Phone:             phone,
		Address:           []string{},
		Affiliation:       "",
		IdCard:            "",
		Region:            region,
		Score:             getInitScore(),
		IsAdmin:           false,
		IsGlobalAdmin:     false,
		IsForbidden:       false,
		IsDeleted:         false,
		SignupApplication: application.Name,
		Properties:        map[string]string{},
		Karma:             0,
	}

	if form.InvitationCode != nil && *form.InvitationCode != "" {
		user.Properties[OTT_USER_PROPERTY_INVITE_CODE] = *form.InvitationCode
	}

	if len(organization.Tags) > 0 {
		tokens := strings.Split(organization.Tags[0], "|")
		if len(tokens) > 0 {
			user.Tag = tokens[0]
		}
	}

	affected := object.AddUser(user)
	if !affected {
		c.OTTResponseError(OTT_CODE_ADD_USER_FAILED, fmt.Sprintf("Failed to create user, user information is invalid: %s", util.StructToJson(user)))
		return
	}

	object.AddUserToOriginalDatabase(user)

	if application.HasPromptPage() {
		// The prompt page needs the user to be signed in
		c.SetSessionUsername(user.GetId())
	}

	object.DisableVerificationCode(form.Identity)
	object.DisableVerificationCode(checkPhone)

	record := object.NewRecord(c.Ctx)
	record.Organization = application.Organization
	record.User = user.Name
	util.SafeGoroutine(func() { object.AddRecord(record) })

	userId := fmt.Sprintf("%s/%s", user.Owner, user.Name)
	util.LogInfo(c.Ctx, "API: [%s] is signed up as new user", userId)

	c.OTTResponseOk(OTTRegisterResponse{UserId: userId})
}

// Logout
// @Title Logout
// @Tag Login API
// @Description logout the current user
// @Success 200 {object} controllers.Response The Response object
// @router /logout [post]
func (c *ApiController) Logout() {
	user := c.GetSessionUsername()
	util.LogInfo(c.Ctx, "API: [%s] logged out", user)

	application := c.GetSessionApplication()
	c.SetSessionUsername("")
	c.SetSessionData(nil)

	if application == nil || application.Name == "app-built-in" || application.HomepageUrl == "" {
		c.ResponseOk(user)
		return
	}
	c.ResponseOk(user, application.HomepageUrl)
}

// GetAccount
// @Title GetAccount
// @Tag Account API
// @Description get the details of the current account
// @Success 200 {object} controllers.Response The Response object
// @router /get-account [get]
func (c *ApiController) GetAccount() {
	userId, ok := c.RequireSignedIn()
	if !ok {
		return
	}

	user := object.GetUser(userId)
	if user == nil {
		c.ResponseError(fmt.Sprintf("The user: %s doesn't exist", userId))
		return
	}

	organization := object.GetMaskedOrganization(object.GetOrganizationByUser(user))
	resp := Response{
		Status: "ok",
		Sub:    user.Id,
		Name:   user.Name,
		Data:   user,
		Data2:  organization,
	}
	c.Data["json"] = resp
	c.ServeJSON()
}

// UserInfo
// @Title UserInfo
// @Tag Account API
// @Description return user information according to OIDC standards
// @Success 200 {object} object.Userinfo The Response object
// @router /userinfo [get]
func (c *ApiController) GetUserinfo() {
	userId, ok := c.RequireSignedIn()
	if !ok {
		return
	}
	scope, aud := c.GetSessionOidc()
	host := c.Ctx.Request.Host
	resp, err := object.GetUserInfo(userId, scope, aud, host)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.Data["json"] = resp
	c.ServeJSON()
}

// GetHumanCheck ...
// @Tag Login API
// @Title GetHumancheck
// @router /api/get-human-check [get]
func (c *ApiController) GetHumanCheck() {
	c.Data["json"] = HumanCheck{Type: "none"}

	provider := object.GetDefaultHumanCheckProvider()
	if provider == nil {
		id, img := object.GetCaptcha()
		c.Data["json"] = HumanCheck{Type: "captcha", CaptchaId: id, CaptchaImage: img}
		c.ServeJSON()
		return
	}

	c.ServeJSON()
}
