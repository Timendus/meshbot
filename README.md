# Meshbot

A simple bot for use with Meshtastic.

Connect a node to your network. Copy `.env` to `production.env`. Store the
address of the node in the new `production.env` file. The address can be a
hostname or an IP address. Then run the script:

```bash
make run
```

Make sure no other client is communicating with this node, otherwise both
clients will be missing messages and things will appear to be very broken. So
disconnect your mobile app and don't make any connections to it.

Pro-tips:

- Add a 🤖 emoji to your node name to make it clear to other users that your
  node is a bot.
- You can add quick chat messages (at least in the Android Meshtastic app) for
  things. Adding the commands that the bot accepts make them really easily
  accessible with one click.

Serial isn't supported yet, but shouldn't be too hard to add. Same for
Bluetooth, although I wouldn't know why you would want that 🙂

## Be responsible

There is very little bandwidth available on Meshtastic. If you use this bot,
please make sure it doesn't spam your local mesh. Make sure it only speaks when
spoken to. Et cetera.

## Usage

I'll probably document this at some point, but for now:

Send a direct message to the node you have connected Meshbot to from another
Meshtastic node, and it will reply to you with the available commands.

## Dockerize

Change `meshtastic.local` to the right hostname or IP in the `dockerfile`. Run
`make build`. Run `make run-image` to run locally or `make export-image` to
build a `.tar.gz` file to run elsewhere.
