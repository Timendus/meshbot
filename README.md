# Meshbot

A simple bot for use with Meshtastic.

![A screenshot of the Meshtastic app in a conversation with
Meshbot](./screenshot.jpeg)

## Node setup

Connect a Meshtastic node to your network through wifi or ethernet. Make sure no
other client is communicating with this node, otherwise both clients will be
missing messages and things will appear to be very broken. So disconnect your
mobile app and don't make any connections to it while the bot is running.

Pro-tips on the Meshtastic side:

- Add a 🤖 emoji to your node name to make it clear to other users that your
  node is a bot.
- You can add quick chat messages -- at least in the Android Meshtastic app.
  Adding the commands that the bot accepts (like `NEW` and `/SIGNAL`) makes them
  really easily accessible with one click.

## Bot setup

Copy [`.env`](./.env) to a new file `production.env` in the project root. Store
the address of the node in this new `production.env` file. The address can be a
hostname or an IP address. Then, from the project root, run:

```bash
make
```

This will run the bot locally and have it connect to the address you specified.

Serial isn't supported yet, but shouldn't be too hard to add. Same for
Bluetooth, although I wouldn't know why you would want that 🙂

## Usage

I'll probably document this at some point, but for now:

Send a direct message to the node you have connected Meshbot to from another
Meshtastic node, and it will reply to you with the available commands.

## Be responsible

There is very little bandwidth available on Meshtastic. If you use this bot, and
especially if you wish to modify it, please make sure it doesn't spam your local
mesh. Make sure it only speaks when spoken to. Et cetera. Be a good neighbour.

## Docker

There is a dockerfile available if you wish to run this bot in Docker.

Run `make build` to create the docker image. Run `make run-image` to run locally
or `make export-image` to build a `.tar.gz` file to run elsewhere.

To configure which host to connect to, either mount a `production.env` file in
the project root, or set the environment variable `NODE_HOSTNAME`.

The command line way for this is:

```bash
docker run --name=meshbot --env=NODE_HOSTNAME=meshtastic.local -d timendus/meshbot
```
