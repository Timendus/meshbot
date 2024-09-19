# Meshbot

A simple bot for use with Meshtastic.

Connect a node to your network, and run the script with the address of the node
as a parameter. Can be a hostname or an IP address.

```bash
python3 ./main.py meshtastic.local
```

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
