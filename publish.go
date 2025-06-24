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

import (
	"io"
	"net/http"

	"github.com/nats-io/nats.go"
)

// NewPubHandler creates a handler that can publish on a subject
func NewPubHandler(conn *nats.Conn, authFunc AuthFunc) http.HandlerFunc {
	n := NatsContext{
		Conn: conn,
		Auth: authFunc,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		newPubHandler(w, r, n)
	}
}

// newPubHandler is a handler that will publish on a subject passed as a query parameter
func newPubHandler(w http.ResponseWriter, r *http.Request, nc NatsContext) {
	subject := replacer(r, "subject")

	ok := nc.Auth(r.Header.Get("Authorization"), subject)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return
	}

	headers := nats.Header{}
	vals := r.URL.Query()
	for k, v := range vals {
		headers.Set(k, v[0])
	}

	msg := &nats.Msg{
		Subject: subject,
		Data:    body,
		Header:  headers,
	}

	if err := nc.Conn.PublishMsg(msg); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), 500)
		return
	}

	return
}
