> This version of Meshbot is a rewrite in Go. If for some weird reason you
> **really** need the broken old Python version, I have saved that in the
> [branch
> legacy-version-in-python](https://github.com/Timendus/meshbot/tree/legacy-version-in-python).

# Meshbot

A simple bot for use with Meshtastic. I know the name isn't very original 😄

Some people would probably call this a "BBS", but personally I think it has more
in common with a crossover between a Slack / Telegram / Discord bot and a
MeshCore room server.

It is written in Golang, which makes for a very efficient piece of software
compared to all the similar programs written in Python. The binary and docker
image are a couple of megabytes. It needs barely and CPU and a couple of
megabyte of RAM to run. It should also be pretty stable and robust.

## Current features

- "Room server" that supports multiple rooms with subscriptions and
  semi-reliable message delivery ([see
  here](./manual.md#why-meshbot-rooms-are-more-reliable-than-regular-channels)
  how we do that)
- Signal reports and neighbours can be queried
- Weather reports and forecasts can be queried, using
  [open-meteo.com](https://open-meteo.com/) (requires the bot to have an
  Internet connection)
- Programmable regular announcements to channels for service messages in your
  area

## Usage on the mesh

See the [user manual](./manual.md) for instructions on how to use Meshbot over
Meshtastic.

## Hosting Meshbot

### Be responsible

There is very little bandwidth available on Meshtastic. If you use this bot, and
especially if you wish to modify it, please make sure it doesn't spam your local
mesh. Make sure it only speaks when spoken to. Et cetera.

Also, be aware that Meshbot rooms with many users will generate a lot of
traffic, since every message is sent to every user, with retries. This grows
exponentially.

In short: be a good neighbour.

### Setup

> **Please note** that this bot has currently **only** been tested on Linux and
> as a Docker image, over TCP. I expect it will probably work over USB and/or on
> MacOS, Windows or Raspberry Pi, but beware there may be dragons 😉 Feel free
> to create an issue if you run into things, but broad support is currently not
> a high priority.

You will need a Meshtastic node and a computer to host the bot. The node and the
computer can either be connected through a USB cable, or [trough your network
over wifi or
ethernet](https://meshtastic.org/docs/configuration/radio/network/).

The former can be super mobile and does not depend on your local network being
up (for example during a power outage). The latter allows you the luxury of
having your node in the best possible spot for reception, while the bot is
running wherever you happen to have compute.

#### Meshtastic node

This software is being developed using a Heltec v3, but any Meshtastic node
should do.

Make sure no other client besides the bot is communicating with the node,
otherwise both clients will be missing messages and things will appear to be
very broken. So disconnect your mobile app and don't make any other connections
to it while the bot is running.

Pro-tips on the Meshtastic side:

- Add a robot emoji (🤖) to your node name to make it clear to other users that
  your node is a bot.
- You can add quick chat messages -- at least in the Android Meshtastic app.
  Adding the commands that the bot accepts (like `/rooms` and `/signal`) as
  quick messages makes them really easily accessible with one click.

#### Computer

Any computer will do as long as it stays on, of course. I run it locally on my
NAS, but even an old Raspberry Pi should work great for the bot. It requires
very few resources. You can run the bot through Docker or directly on the
computer.

Download the appropriate version of the software from the [releases
page](https://github.com/Timendus/meshbot/releases). Edit the `config.json` file
to tell the bot how to connect to your node and how to behave and start the
software. `config.json` should be in the same directory as the software.

For the docker version, mount a directory to `/app/config`. Launch the
container. The first time, if configured correctly, a `config.json` file will be
created in the mounted directory for you to edit. Stop the container, edit the
config file and restart the container.

## Local development

Dependencies:

- Golang
- Git
- make

Then do something like:

```bash
git clone git@github.com:Timendus/meshbot.git
cd meshbot
vi config.json
make
```

And the bot should start.
