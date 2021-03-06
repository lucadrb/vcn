/*
 * Copyright (c) 2018-2019 vChain, Inc. All Rights Reserved.
 * This software is released under GPL3.
 * The full license information can be found under:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 */

package api

import (
	"fmt"

	"github.com/dghubble/sling"
	"github.com/sirupsen/logrus"
	"github.com/vchain-us/vcn/internal/errors"
	"github.com/vchain-us/vcn/pkg/meta"
)

type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenResponse struct {
	Token string `token:"token"`
}

type Error struct {
	Status    int    `json:"status"`
	Message   string `json:"message"`
	Path      string `json:"path"`
	Timestamp string `json:"timestamp"`
	Error     string `json:"error"`
}

type PublisherExistsResponse struct {
	Exists bool `json:"exists"`
}

type PublisherExistsParams struct {
	Email string `url:"email"`
}

func CheckPublisherExists(email string) (success bool, err error) {
	response := new(PublisherExistsResponse)
	restError := new(Error)
	r, err := sling.New().
		Get(meta.PublisherEndpoint()+"/exists").
		QueryStruct(&PublisherExistsParams{Email: email}).
		Receive(&response, restError)
	logger().WithFields(logrus.Fields{
		"response":  response,
		"err":       err,
		"restError": restError,
	}).Trace("CheckPublisherExists")
	if err != nil {
		return false, err
	}
	if r.StatusCode == 200 {
		return response.Exists, nil
	}
	return false, fmt.Errorf("check publisher failed: %+v", restError)
}

func checkToken(token string) (success bool, err error) {
	restError := new(Error)
	response, err := newSling(token).
		Get(meta.TokenCheckEndpoint()).
		Receive(nil, restError)
	logger().WithFields(logrus.Fields{
		"response":  response,
		"err":       err,
		"restError": restError,
	}).Trace("checkToken")
	if response != nil {
		switch response.StatusCode {
		case 200:
			return true, nil
		case 401:
			fallthrough
		case 403:
			fallthrough
		case 419:
			return false, nil
		}
	}
	if restError.Error != "" {
		err = fmt.Errorf("%+v", restError)
	}
	return false, fmt.Errorf("check token failed: %s", err)
}

func authenticateUser(email string, password string) (token string, err error) {
	response := new(TokenResponse)
	restError := new(Error)
	r, err := sling.New().
		Post(meta.PublisherEndpoint()+"/auth").
		BodyJSON(AuthRequest{Email: email, Password: password}).
		Receive(response, restError)
	logger().WithFields(logrus.Fields{
		"email":     email,
		"response":  response,
		"err":       err,
		"restError": restError,
	}).Trace("authenticateUser")
	if err != nil {
		return "", err
	}
	switch r.StatusCode {
	case 200:
		return response.Token, nil
	case 400:
		return "", fmt.Errorf(errors.UnconfirmedEmail, email, meta.DashboardURL())
	case 401:
		return "", fmt.Errorf("invalid password")
	}
	if restError.Error != "" {
		err = fmt.Errorf("%+v", restError)
	}
	return "", fmt.Errorf("authentication failed: %s", err)
}
