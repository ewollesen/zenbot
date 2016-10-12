// Copyright 2016 Eric Wollesen <ericw at xmtp dot net>
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

package commands

type basicHandler struct {
	handler func(args ...string) (string, error)
	help    func(args ...string) string
}

var _ CommandHandler = (*basicHandler)(nil)

func (h *basicHandler) Handle(argv ...string) (string, error) {
	return h.handler(argv...)
}

func (h *basicHandler) Help(argv ...string) string {
	return h.help(argv...)
}

func Static(response, help string) CommandHandler {
	return &basicHandler{
		handler: func(argv ...string) (string, error) {
			return response, nil
		},
		help: func(argv ...string) string {
			return help
		},
	}
}
