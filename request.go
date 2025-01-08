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
	"strconv"
	"time"

	"github.com/nats-io/nats.go"
)

func NewReqHandler(conn *nats.Conn, authFunc AuthFunc) http.HandlerFunc {
	n := NatsContext{
		Conn: conn,
		Auth: authFunc,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		newReqHandler(w, r, n)
	}
}

func newReqHandler(w http.ResponseWriter, r *http.Request, nc NatsContext) {
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
	defer r.Body.Close()

	msg := &nats.Msg{
		Subject: subject,
		Header:  nats.Header(r.Header.Clone()),
		Data:    body,
	}

	resp, err := nc.Conn.RequestMsg(msg, 3*time.Second)
	if err != nil && err == nats.ErrNoResponders {
		http.Error(w, http.StatusText(http.StatusNotFound), 404)
		return
	}
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), 500)
		return
	}

	status := resp.Header.Get("Nats-Service-Error-Code")

	if status != "" && status != "200" {
		code, err := strconv.Atoi(status)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), 500)
			return
		}
		http.Error(w, string(resp.Data), code)
		return
	}

	w.Write(resp.Data)
}
