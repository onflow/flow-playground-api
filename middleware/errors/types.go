/*
 * Flow Playground
 *
 * Copyright 2019 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package errors

import (
	"errors"
	"fmt"
)

var ServerErr = errors.New("something went wrong, we are looking into the issue")
var GraphqlErr = errors.New("invalid graphql request")

type AuthorizationError struct {
	msg string
}

func NewAuthorizationError(msg string) *AuthorizationError {
	return &AuthorizationError{msg}
}

func (i *AuthorizationError) Error() string {
	return fmt.Sprintf("authorization error: %s", i.msg)
}

type UserError struct {
	msg string
}

func NewUserError(msg string) *UserError {
	return &UserError{msg}
}

func (i *UserError) Error() string {
	return fmt.Sprintf("user error: %s", i.msg)
}
