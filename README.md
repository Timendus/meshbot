# Meshbot

A simple bot for use with Meshtastic. I know the name isn't very original 😄

Some people would probably call this a "BBS", but personally I think it has more
in common with something like a Slack / Telegram / Discord bot.

## Current features

- Mail box / message box for communicating with other Mestastic users. These
  commands only work in direct messages with the bot, not in channels for
  obvious reasons.
- Querying the bot for signal reports, nodes it currently sees and nodes it has
  seen.
- Talk to a self-hosted LLM using [Ollama](https://ollama.com/).

Some of these things are being demonstrated in this screenshot:

![A screenshot of the Meshtastic app in a conversation with
Meshbot](./screenshot.jpeg)

## Setup

You will need a Meshtastic node and a computer to host the bot.

- For the node: this software is being developed using a Heltec v3, but I
  suppose any Meshtastic node should work well.

- For the computer you can just use your laptop or desktop, but for a slightly
  more permanent setup you may want to use a dedicated server. A NAS, an old
  computer or even an old Raspberry Pi works great for the bot.

- To use the LLM feature you will need to run Ollama, which requires a bit more
  horse power or preferably a good GPU. This feature is optional though, and it
  is also entirely possible to run the bot on one machine and Ollama on another.

The node and the host can either be connected through a USB cable, or [trough
your network over wifi or
ethernet](https://meshtastic.org/docs/configuration/radio/network/). The former
can be super mobile and does not depend on your local network being up. The
latter allows you the luxury of having your node in the best possible spot for
reception, while the bot is running wherever you happen to have compute.

> Note that this bot will most probably **not work on Windows**. It hasn't been
> tested on Windows and I don't wish to ever support Windows. If it does, it's
> just dumb luck 😉 Get someone to build a [Docker image](#docker) for you to
> run or find a Mac or Linux machine. A Raspberry Pi is a great option to get
> started.

### Meshtastic node

Make sure no other client besides the bot is communicating with the node,
otherwise both clients will be missing messages and things will appear to be
very broken. So disconnect your mobile app and don't make any other connections
to it while the bot is running.

Pro-tips on the Meshtastic side:

- Add a robot emoji (🤖) to your node name to make it clear to other users that
  your node is a bot.
- You can add quick chat messages -- at least in the Android Meshtastic app.
  Adding the commands that the bot accepts (like `NEW` and `/SIGNAL`) as quick
  messages makes them really easily accessible with one click.

### Computer

You can run the bot [through Docker (see below)](#docker) or directly on the
computer.

Assuming you have `git`, `make` and Python 3 installed, clone the project and
copy [`.env`](./.env) to a new file named `production.env` in the project root:

```bash
git clone git@github.com:Timendus/meshbot.git
cd meshbot
cp .env production.env
```

Edit the `production.env` file to specify how the bot should connect to your
node and also where to find Ollama if you wish to use the self-hosted LLM
feature. The file has some examples to get you started:

https://github.com/Timendus/meshbot/blob/4aca4a8a4bc8e608e3463eddda93a91ce081592a/.env#L1-L16

Then install the dependencies in a new virtual environment and run the bot:

```bash
make dependencies
make
```

The software will attempt to connect to the network address or USB device you
specified in `production.env`. If all goes well you should be greeted by a list
of nodes your bot node has seen, and you should be seeing Meshtastic packets get
logged to the console.

## Usage

I'll probably document this at some point, but for now:

Send a direct message to the node you have connected Meshbot to from another
Meshtastic node, and it will reply to you with the available commands.

## Be responsible

There is very little bandwidth available on Meshtastic. If you use this bot, and
especially if you wish to modify it, please make sure it doesn't spam your local
mesh. Make sure it only speaks when spoken to. Et cetera. Be a good neighbour.

For this reason I have tried to design the commands in such a way that you can
do anything you want by sending a single message. No traversing deep menus or
having to send multiple messages to achieve your goals.

## Docker

There is a dockerfile available if you wish to run this bot in Docker. I will
probably put this on Dockerhub when it is a bit more polished, but for now you
have to build the image yourself.

Run `make build` to create the docker image. Run `make run-image` to run locally
or `make export-image` to build a `.tar.gz` file to run elsewhere.

To configure which host to connect to, either mount a `production.env` file in
the project root, or set environment variables to match the settings in
`production.env`.

The command line way for this is:

```bash
docker run --name=meshbot --env=TRANSPORT=serial --env=DEVICE=/dev/ttyUSB0 -d timendus/meshbot
```
