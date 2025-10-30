# How to use Meshbot through the mesh

As a user of Meshbot, you can send any of the commands below over Meshtastic and
Meshbot will reply.

Commands are not case-sensitive.

## Commands

### In a channel or as a direct message

- `/about` or `/help` - Get a short overview of these commands
- `/signal <optional node>` - Fetch a signal report on yourself (default) or the
  node you ask for
- `/neighbours` - Fetch the list of neighbours that the bot can see over LoRa
- `/weather` - Fetch a report of the current weather conditions
- `/forecast` - Fetch a weather forecast for the coming days

Replies will be sent like normal Meshtastic messages, either in the channel you
send the command to, or as a DM back to you.

### As direct messages only

- `/rooms` - Fetch a list of available rooms and your status in them
- `/join <room name> <optional password>` - Join a room, so you will receive
  messages sent to it. Supply a password for private rooms. Joining a room also
  selects it
- `/select <room name>` - Select a room. Messages you send will be broadcast to
  the selected room. Only a single room can be selected at a time. You can only
  select a room that you have joined
- `/leave <room name>` - Leave a room, so you will no longer receive messages
  sent to it

Replies to these commands will be sent "reliably", which means Meshbot will
retry sending until it sees a delivery notification, with a maximum of three
attempts.

## Sending to rooms

For any other DM you send to the bot:

- If you have not joined any rooms, and a public room exists (one without a
  password), it will make you join this room and select it automatically.
- Otherwise, any DM you send to the bot will be sent to all users in the
  selected room, including an echo back to yourself.

Messages sent to rooms will also retry at most three times, but when delivery
still fails after that, these messages will be stored for you in a backlog
queue. When Meshbot receives any packets from you it will assume that you have
come back into range and retry sending the messages from the queue.

## Why Meshbot rooms are more reliable than regular channels

For surprisingly many reasons, actually.

Direct messages have delivery notification feedback in the app to show you if
your message successfully arrived at its destination. Channels only show that
your message was repeated by _someone_. Also, since Meshtastic 2.6, direct
messages make use of ["next-hop"
routing](https://meshtastic.org/blog/meshtastic-2-6-preview/#next-hop-routing-for-dms).
Channels do not benefit from this improvement.

Sometimes direct messages arrive properly, but the delivery notification doesn't
make it back to you. As additional feedback that you have successfully sent a
message, any messages you send to a room will also be echoed back to you. If
your connection to the bot is poor, this may take a while though, so be patient
before sending again.

Finally, Meshbot will keep trying to send messages to all users in a room
(including the sender) until it receives good delivery notifications. This means
that you may sometimes receive messages multiple times, but it ensures that your
communication is fairly reliable.

Even if you move out of range, Meshbot will remember which messages you missed
and as soon as it sees you coming back into range it will send you the entire
history since you left. So in that sense it also works a bit like a more
convenient version of Meshtastic's Store&Forward feature.
