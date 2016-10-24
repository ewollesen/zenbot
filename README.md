# Zenbot

A discord bot with a queueing system, plus a few other goodies.

This bot was designed to meet the requirements and desires of the Affinity
Coaching Discord group.

## Testing

    go test github.com/ewollesen/zenbot/...

## Compilation

    go build -o ./zenbot github.com/ewollesen/zenbot/bin/zenbot

## Running

### Discord Application Registration

Before the bot can join your Guild (Server), it needs a bot user token and a client id. So visit [your Discord applications page](https://discordapp.com/developers/applications/me) to get started. There you should:

  1. Click "New Application".
    1. Give your bot a name.
    1. Specify a redirect URL (see More on OAuth Redirect URLs, below).
  1. Click "Create a bot user".
  1. Make a note of your App's Client Id, and the App Bot User's Token. These will be used later.

### Starting the Bot

    $ ./zenbot -discord.client_id <your app's client id> \
               -discord.token "Bot <your app bot user's token>"

Notice that we've added "Bot" before the token specified above.

### Inviting the Bot to Join Your Guild

  1. Visit <http://localhost:8080/discord>.
  1. Follow the generated OAuth URL.
  1. Select the Discord Guild on which you want the bot to run.
  1. Dance!

Zenbot should now be visible to your guild. Further configuration options can be found by running the bot with the `-help` flag.

### More on OAuth Redirect URLs

Discord bots are authorized to join a Guild via [OAuth](https://discordapp.com/developers/docs/topics/oauth2). They do this by crafting a special URL which, when clicked on by a user authorized to add bots to a Guild, grants the bot authorization to join the Guild.

When Zenbot starts up, it creates a small HTTP server, by default on port 8080. Visiting this server's `/discord` endpoint, eg `http://localhost:8080/discord` will present you with an OAuth link that, when followed, will invite Zenbot to join a Guild for which you have access.

After you have authenticated the bot, Discord will try to redirect you back to the bot. So it's up to you to add a correct OAuth redirection URL (in step 1 under running) that is connected to Zenbot's HTTP server's Discord OAuth redirect endpoint, by default this is http://<your server's name/ip>:8080/discord/oauth/redirect. You can use the `discord.hostname` and `discord.protocol` options to modify the generated URLs to point to reverse proxy if you're using one.

## Configuration File Template

Here's a basic configuration file to get you started. Defaults are commented out:

    [main]
    # If not specified, will default to in-memory caching and queueing.
    # redis_addr =
    # redis_db = 0

    [discord]
    client_id = <your app's client id>
    token = Bot <your app bot user's token>

    # The hostname used for generated OAuth redirect URLs.
    # hostname =

    # The protocol for generated OAuth redirect URLs.
    # protocol = https

    # command_prefix = !

    # game = !queue help

    # A prefix to use in front of redis keys, if using a redis server.
    # redis_keyspace = discord

    # Rate limits the enqueue command. Units can be specified as minutes (m),
    # hours (h), or seconds (s).
    # enqueue_rate_limit = 5m

    # A comma separated list of channel ids that zenbot should listen in.
    # To find channel ids, turn on debug logging, or use your client's developer
    # mode, as detailed here:
    # https://support.discordapp.com/hc/en-us/articles/206346498-Where-can-I-find-my-server-ID-
    # whitelisted_channels =

    [httpapi] address = :8080

    [log]
    level = info

You can use a configuration file by specifying the `-flagfile` option to zenbot, eg:

    $ ./zenbot -flagfile ~/.zenbot/config
