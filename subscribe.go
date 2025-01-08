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
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/nats-io/nats.go"
)

// NewSubHandler creates a handler that does server side event subscribing
func NewSubHandler(conn *nats.Conn, authFunc AuthFunc) http.HandlerFunc {
	n := NatsContext{
		Conn: conn,
		Auth: authFunc,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		newSSEHandler(w, r, n, Subscribe)
	}
}

// Subscribe wraps handleSubscription and handles flushing the writer.
func Subscribe(ctx context.Context, flusher http.Flusher, opts options) {
	go handleSubscription(ctx, opts)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			writeAndFlushResponse(opts.writer, flusher, <-opts.ch)
		}
	}
}

// handleSubscription creates the NATS subscription and iterates over the messages
func handleSubscription(ctx context.Context, opts options) {
	sub, err := opts.nc.Conn.SubscribeSync(opts.subject)
	if err != nil {
		msg := nats.Msg{
			Subject: "natssse.system",
			Data:    []byte(err.Error()),
		}
		opts.ch <- msg
		opts.cancel()
		return
	}
	defer sub.Unsubscribe()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := sub.NextMsg(10 * time.Second)
			if err != nil && errors.Is(err, nats.ErrTimeout) {
				continue
			}
			if err != nil {
				msg := nats.Msg{
					Subject: "natssse.system",
					Data:    []byte(err.Error()),
				}
				opts.ch <- msg
				continue
			}
			opts.ch <- *msg

		}
	}
}
