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

package discord

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"time"

	"github.com/ewollesen/discordgo"
	"github.com/ewollesen/zenbot/util"
)

var (
	clientId = flag.String("discord.client_id", "",
		"Discord application client id")
	hostname = flag.String("discord.hostname", "zenbot.xmtp.net",
		"Hostname for discord oauth redirects")
	protocol = flag.String("discord.protocol", "https",
		"Protocol for discord oauth redirects")
)

func (b *bot) handleHTTP(w http.ResponseWriter, req *http.Request) {
	t, err := template.New("root").Parse(`<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>Zenbot Discord Integration</title>
	</head>
	<body>
		<p>
		To authorize Zenbot to join your Discord server/guild, visit the
		following URL: <a href="{{.URL}}">{{.URL}}</a>
		</p>
	</body>
</html>`)
	data := struct {
		URL string
	}{
		URL: b.discordOauthURL(),
	}
	buf := bytes.NewBufferString("")
	err = t.Execute(buf, data)
	if err != nil {
		logger.Errore(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed to generate discord oauth URL"))
		return
	}
	logger.Debugf("HTTP request received\n%+v", req)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(buf.Bytes())
}
func (b *bot) oauthRedirect(w http.ResponseWriter, req *http.Request) {
	logger.Debugf("%#v", req)
	values := req.URL.Query()

	// The state parameter verifies that this request came from the entity
	// that we generated an OAuth request for, assuming we used HTTPS and it
	// wasn't interfered with.
	state := values.Get("state")
	if state == "" {
		http.Error(w, "failed to parse state query parameter",
			http.StatusBadRequest)
		return
	}

	b.oauth_mu.Lock()
	_, ok := b.oauth_states[state]
	b.oauth_mu.Unlock()
	if !ok {
		http.Error(w, "unrecognized oauth state",
			http.StatusBadRequest)
		return
	}

	// I don't believe the code is needed at all.
	code := values.Get("code")
	if code == "" {
		http.Error(w, "failed to parse code query parameter",
			http.StatusBadRequest)
		return
	}

	// Everything below here should be optional

	// Useful so that the bot can greet the guild.
	guild_id := values.Get("guild_id")
	if guild_id == "" {
		logger.Warnf("no guild_id in OAuth2 redirect")
	}
	logger.Debugf("guild_id: %q", guild_id)

	token, err := getToken()
	if err != nil {
		logger.Warnf("error getting discord token: %v", err)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte("Registered successfully"))
		return
	}

	session, err := discordgo.New(token)
	if err != nil {
		logger.Warnf("error logging in to discord: %v", err)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte("Registered successfully"))
		return
	}

	guild, err := session.Guild(guild_id)
	if err != nil {
		logger.Warnf("error retrieving guild information: %v", err)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte("Registered successfully"))
		return
	}

	_, err = session.ChannelMessageSend(guild_id,
		fmt.Sprintf("Hello, %s", guild.Name))
	if err != nil {
		logger.Warnf("error greeting guild: %+v", err)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte("Registered successfully"))
		return
	}

	t, err := template.New("registration_successful").Parse(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8">
<title>Zbot Discord Integration</title>
</head>
<body>
<p>
Successfully registered Zenbot with {{.GuildName}}.
</p>
</body>
</html>`)
	if err != nil {
		logger.Warnf("error generating success template")
		return
	}

	data := struct {
		GuildName string
	}{
		GuildName: guild.Name,
	}
	buf := bytes.NewBufferString("")
	err = t.Execute(buf, data)
	if err != nil {
		logger.Warnf("error rendering success template")
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(buf.Bytes())
}

func (b *bot) discordOauthURL() string {
	params := make(url.Values)
	params.Set("response_type", "code")
	params.Set("redirect_uri",
		fmt.Sprintf("%s://%s/discord/oauth/redirect",
			*protocol, *hostname))
	params.Set("client_id", *clientId)
	params.Set("scope", "bot")
	params.Set("permission", "0")
	state, err := util.RandomState(32)
	if err != nil {
		logger.Errore(err)
		return ""
	}
	b.oauth_mu.Lock()
	b.oauth_states[state] = time.Now().String()
	b.oauth_mu.Unlock()
	params.Set("state", state)
	oauth_url := fmt.Sprintf("https://discordapp.com/oauth2/authorize?%s",
		params.Encode())
	logger.Debugf("discord oauth url: %s", oauth_url)

	return oauth_url
}
