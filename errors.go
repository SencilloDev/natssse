// Copyright 2025 Sencillo
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package natssse

import "fmt"

type CallerError interface {
	Error() string
	Code() int
	Body() []byte
}

// ClientError represents a non-server error
type ClientError struct {
	Status  int
	Details string
}

func (c ClientError) Error() string {
	return c.Details
}

func (c ClientError) Body() []byte {
	return []byte(fmt.Sprintf(`%q`, c.Details))
}

func (c ClientError) Code() int {
	return c.Status
}

func (c ClientError) As(target any) bool {
	_, ok := target.(*ClientError)
	return ok
}

func NewClientError(err error, code int) ClientError {
	return ClientError{
		Status:  code,
		Details: err.Error(),
	}
}
