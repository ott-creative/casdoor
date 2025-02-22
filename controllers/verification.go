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
	"errors"
	"fmt"
	"strings"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

type SendVerificationCodeRequest struct {
	AppId   string  `json:"app_id"`           // Application ID, for light wallet: light-wallet
	Purpose int     `json:"purpose"`          // 0: register, 1: login, 2: reset password
	Type    int     `json:"type"`             // 0: phone, 1: email
	Prefix  *string `json:"prefix,omitempty"` // phone: 86 or other country code, email: will be ignored
	Dest    string  `json:"dest"`             // phone: phone number, email: email address
}

func (c *ApiController) getCurrentUser() *object.User {
	var user *object.User
	userId := c.GetSessionUsername()
	if userId == "" {
		user = nil
	} else {
		user = object.GetUser(userId)
	}
	return user
}

// SendVerificationCode ...
// @Title SendVerificationCode
// @Tag Verification API
// @router /send-verification-code [post]
func (c *ApiController) SendVerificationCode() {
	destType := c.Ctx.Request.Form.Get("type")
	dest := c.Ctx.Request.Form.Get("dest")
	orgId := c.Ctx.Request.Form.Get("organizationId")
	checkType := c.Ctx.Request.Form.Get("checkType")
	checkId := c.Ctx.Request.Form.Get("checkId")
	checkKey := c.Ctx.Request.Form.Get("checkKey")
	checkUser := c.Ctx.Request.Form.Get("checkUser")
	remoteAddr := util.GetIPFromRequest(c.Ctx.Request)

	if len(destType) == 0 || len(dest) == 0 || len(orgId) == 0 || !strings.Contains(orgId, "/") || len(checkType) == 0 || len(checkId) == 0 || len(checkKey) == 0 {
		c.ResponseError("Missing parameter.")
		return
	}

	isHuman := false
	captchaProvider := object.GetDefaultHumanCheckProvider()
	if captchaProvider == nil {
		isHuman = object.VerifyCaptcha(checkId, checkKey)
	}

	if !isHuman {
		c.ResponseError("Turing test failed.")
		return
	}

	user := c.getCurrentUser()
	organization := object.GetOrganization(orgId)
	application := object.GetApplicationByOrganizationName(organization.Name)

	if checkUser == "true" && user == nil && object.GetUserByFields(organization.Name, dest) == nil {
		c.ResponseError("Please login first")
		return
	}

	sendResp := errors.New("Invalid dest type")

	if user == nil && checkUser != "" && checkUser != "true" {
		_, name := util.GetOwnerAndNameFromId(orgId)
		user = object.GetUser(fmt.Sprintf("%s/%s", name, checkUser))
	}
	switch destType {
	case "email":
		if user != nil && util.GetMaskedEmail(user.Email) == dest {
			dest = user.Email
		}
		if !util.IsEmailValid(dest) {
			c.ResponseError("Invalid Email address")
			return
		}

		provider := application.GetEmailProvider()
		sendResp = object.SendVerificationCodeToEmail(organization, user, provider, remoteAddr, dest)
	case "phone":
		if user != nil && util.GetMaskedPhone(user.Phone) == dest {
			dest = user.Phone
		}
		/* TODO:
		if !util.IsPhoneCnValid(dest) {
			c.ResponseError("Invalid phone number")
			return
		}*/
		org := object.GetOrganization(orgId)
		if org == nil {
			c.ResponseError("Missing parameter.")
			return
		}

		// check if dest start with +
		if !strings.HasPrefix(dest, "+") {
			dest = fmt.Sprintf("+%s%s", org.PhonePrefix, dest)
		}
		provider := application.GetSmsProvider()
		sendResp = object.SendVerificationCodeToPhone(organization, user, provider, remoteAddr, dest)
	}

	if sendResp != nil {
		c.Data["json"] = Response{Status: "error", Msg: sendResp.Error()}
	} else {
		c.Data["json"] = Response{Status: "ok"}
	}

	c.ServeJSON()
}

func (c *ApiController) OTTSendVerificationCode() {
	var svcRequest SendVerificationCodeRequest
	var user *object.User

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &svcRequest); err != nil {
		c.OTTResponseError(203, err.Error())
		return
	}

	remoteAddr := util.GetIPFromRequest(c.Ctx.Request)

	// TODO: machine check
	dest := svcRequest.Dest
	application := object.GetApplication(fmt.Sprintf("admin/%s", svcRequest.AppId))
	organization := object.GetOrganization("admin/OTT")

	// reset password needs current login user
	/*if svcRequest.Purpose == 2 {
		user = c.getCurrentUser()
	}*/

	sendResp := errors.New("send verification code failed")

	switch svcRequest.Type {
	case 1: // email
		if !util.IsEmailValid(svcRequest.Dest) {
			c.OTTResponseError(OTT_CODE_INVALID_EMAIL, "Invalid Email address")
			return
		}

		if user != nil && util.GetMaskedEmail(user.Email) == dest {
			dest = user.Email
		}

		provider := application.GetEmailProvider()
		sendResp = object.SendVerificationCodeToEmail(organization, user, provider, remoteAddr, dest)
	case 0: // sms
		if user != nil && util.GetMaskedPhone(user.Phone) == dest {
			dest = user.Phone
		}

		// TODO: validate international phone number
		/*if svcRequest.Prefix == "+86" && !util.IsPhoneCnValid(dest) {
			c.OTTResponseError(OTT_CODE_INVALID_PHONE, "Invalid phone number")
			return
		}*/

		dest = util.MakeInternationalPhone(*svcRequest.Prefix, svcRequest.Dest)

		provider := application.GetSmsProvider()
		sendResp = object.SendVerificationCodeToPhone(organization, user, provider, remoteAddr, dest)
	}

	if sendResp != nil {
		c.Data["json"] = OTTResponse{Code: OTT_CODE_SEND_VERIFICATION_CODE_FAILED, Msg: sendResp.Error()}
	} else {
		c.Data["json"] = OTTResponse{Code: OTT_CODE_OK, Data: OTTSendVerificationCodeResponse{Timer: 60}}
	}

	c.ServeJSON()
}

// ResetEmailOrPhone ...
// @Tag Account API
// @Title ResetEmailOrPhone
// @router /api/reset-email-or-phone [post]
func (c *ApiController) ResetEmailOrPhone() {
	userId, ok := c.RequireSignedIn()
	if !ok {
		return
	}

	user := object.GetUser(userId)
	if user == nil {
		c.ResponseError(fmt.Sprintf("The user: %s doesn't exist", userId))
		return
	}

	destType := c.Ctx.Request.Form.Get("type")
	dest := c.Ctx.Request.Form.Get("dest")
	code := c.Ctx.Request.Form.Get("code")
	if len(dest) == 0 || len(code) == 0 || len(destType) == 0 {
		c.ResponseError("Missing parameter.")
		return
	}

	checkDest := dest
	if destType == "phone" {
		org := object.GetOrganizationByUser(user)
		phonePrefix := "86"
		if org != nil && org.PhonePrefix != "" {
			phonePrefix = org.PhonePrefix
		}
		checkDest = util.MakeInternationalPhone(phonePrefix, dest)
	}
	if ret := object.CheckVerificationCode(checkDest, code); len(ret) != 0 {
		c.ResponseError(ret)
		return
	}

	switch destType {
	case "email":
		user.Email = dest
		object.SetUserField(user, "email", user.Email)
	case "phone":
		user.Phone = dest
		object.SetUserField(user, "phone", user.Phone)
	default:
		c.ResponseError("Unknown type.")
		return
	}

	object.DisableVerificationCode(checkDest)
	c.Data["json"] = Response{Status: "ok"}
	c.ServeJSON()
}
