// Copyright 2022 The Casdoor Authors. All Rights Reserved.
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

type TokenRequest struct {
	GrantType    string `json:"grant_type"`
	Code         string `json:"code"`
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Verifier     string `json:"code_verifier"`
	Scope        string `json:"scope"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	Tag          string `json:"tag"`
	Avatar       string `json:"avatar"`
	RefreshToken string `json:"refresh_token"`
}

const OTT_ORGANIZATION_ID = "OTT"
const OTT_USER_PROPERTY_INVITE_CODE = "invite_code"

const OTT_TYPE_PHONE = 0
const OTT_TYPE_EMAIL = 1

const OTT_CODE_OK = 200
const OTT_CODE_SERVICE_UNAVAILABLE = 201
const OTT_CODE_NEED_LOGIN = 202
const OTT_CODE_INVALID_PARAM = 203
const OTT_CODE_LOGIN_EXPIRED = 204
const OTT_CODE_ONLY_ONE_LOGIN = 205
const OTT_CODE_SERVICE_EXCEPTION = 206
const OTT_CODE_APPLICATION_NO_SIGNUP = 207
const OTT_CODE_VERIFICATION_CODE_NOT_MATCH = 208
const OTT_CODE_ADD_USER_FAILED = 209
const OTT_CODE_NEED_SIGN_OUT = 210
const OTT_CODE_INVALID_EMAIL = 302
const OTT_CODE_INVALID_PHONE = 301
const OTT_CODE_SEND_VERIFICATION_CODE_FAILED = 300
