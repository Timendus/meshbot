Plugin = {
    name = "✉️ Message box",
    description = "An answering machine for Meshtastic",
    version = "1.0",

    commands = {
        {
            command = "INBOX",
            description = "Check your inbox",
            func = function(message) return SendInbox(message) end,
        },
        {
            command = "NEW",
            description = "Get new messages",
            func = function(message) return SendNewMessages(message) end,
        },
        {
            command = "OLD",
            description = "Get old messages",
            func = function(message) return SendOldMessages(message) end,
        },
        {
            command = "CLEAR",
            description = "Clear old messages",
            func = function(message) return ClearOldMessages(message) end
        },
        {
            prefix = "SEND",
            description = "Leave a message (SEND <id> <message>)",
            func = function(message) return StoreMessage(message) end
        },
        {
            command = Bot.CATCH_ALL_EVENTS,
            func = function(message) return NotifyUser(message) end
        }
    },
}

-- Inform the user about their inbox stats
function SendInbox(message)
    local inbox = GetInbox(message:getSender())

    if #inbox == 0 then
        message:reply("🤖📭 You have no messages in your inbox")
        return
    end

    local icon = inbox.numUnread > 0 and "📬" or "📭"
    message:reply(
        "🤖" ..
        icon ..
        " You have " ..
        inbox.numUnread ..
        " unread " ..
        Pluralize("message", inbox.numUnread) ..
        ", and a grand total of " ..
        #inbox .. " " .. Pluralize("message", #inbox) .. " in your inbox. Send `NEW` or `OLD` to fetch your messages."
    )
end

-- Send all unread messages to the user
function SendNewMessages(message)
    local inbox = GetInbox(message:getSender())

    if inbox.numUnread == 0 then
        message:reply("🤖📭 You have no new messages." ..
            (inbox.numRead > 0 and " Send `OLD` to read your older messages." or ""))
        return
    end

    message:reply("🤖📬 You have " ..
        inbox.numUnread ..
        " new " .. Pluralize("message", inbox.numUnread) .. ". Sending " .. Pluralize("it", inbox.numUnread) .. " now...",
        function(success)
            if success then
                SendMessages(message, inbox, false)
            else
                print("Could not send new messages, delivery timed out")
            end
        end
    )
end

-- Send all read messages to the user
function SendOldMessages(message)
    local inbox = GetInbox(message:getSender())

    if inbox.numRead == 0 then
        message:reply("🤖📭 You have no old messages." ..
            (inbox.numUnread > 0 and " Send `NEW` to read your new messages." or ""))
        return
    end

    message:reply("🤖📬 You have " ..
        inbox.numRead ..
        " old " .. Pluralize("message", inbox.numRead) .. ". Sending " .. Pluralize("it", inbox.numRead) .. " now...",
        function(success)
            if success then
                SendMessages(message, inbox, true)
            else
                print("Could not send old messages, delivery timed out")
            end
        end
    )
end

-- Clear all messages that have already been read
function ClearOldMessages(message)
    local inbox = Bot.memory.read(message:getSender():getIdExpression())
    local num

    if inbox ~= nil then
        num = 0
        for i, msg in ipairs(inbox) do
            if msg["read"] then
                table.remove(inbox, i)
                num = num + 1
            end
        end
        Bot.memory.write(message:getSender():getIdExpression(), inbox)
    end

    inbox = GetInbox(message:getSender())
    message:reply("🤖🗑️ I removed " ..
        num ..
        " old " ..
        Pluralize("message", num) ..
        ". You have " .. inbox.numUnread .. " new " .. Pluralize("message", inbox.numUnread) .. " left in your inbox.")
end

-- Store new messages when requested by the user
function StoreMessage(message)
    local text = message:getText()
    local user = text:match("^%S+%s+(%S+)")
    local to_send = text:match("^%S+%s+%S+%s+(.*)")

    -- Validate we got a valid request from the user, explain how to use this
    -- otherwise
    if user == nil or to_send == nil then
        message:reply(
            "🤖🧨 The syntax for this command is SEND <id> <message>, where <id> is the short name or id of a node.")
        return
    end

    -- Find our recipient node
    local node = message:findNode(user)
    if node == nil then
        message:reply("🤖🧨 I don't know who '" ..
            user ..
            "' is. The message was not stored.\n\nI need the short name of a node I have seen before (example: TDRP), or the node ID of the recipient (example: !8e92a31f).")
        return
    end

    -- Store the message to the bot's memory
    local inbox = Bot.memory.read(node:getIdExpression())
    if inbox == nil then
        inbox = {}
    end
    table.insert(inbox, {
        sender = tostring(message:getSender()),
        contents = Trim(to_send),
        read = false,
        timestamp = Bot.date("2-1-2006 15:04:05"),
    })
    Bot.memory.write(node:getIdExpression(), inbox)

    message:reply("🤖📨 Saved this message for node " .. tostring(node) .. ":\n\n" .. to_send)
end

-- Check to see if one of our recipients came in range, and has new messages.
function NotifyUser(message)
    -- If they are messaging us first, they will probably quickly find out that
    -- they have messages, and it just breaks the flow. So only check for all
    -- other message types.
    if message:getType() == "text message" and (message:getReceiver() == nil or message:getReceiver():isSelf()) then
        return
    end

    -- We get routing messages for each Ack, so ignore those or we get a royal
    -- clusterfuck.
    if message:getType() == "routing" then
        return
    end

    -- Do we have a message box at all? Otherwise we're spamming nodes that have
    -- never interacted with this bot, and have not actually been sent messages
    -- by real people, with a "friendly welcome message".
    if message:getSender() == nil then
        return
    end
    local box = Bot.memory.read(message:getSender():getIdExpression())
    if box == nil then
        return
    end

    -- Do we have new messages?
    local inbox = GetInbox(message:getSender())
    if inbox.numUnread == 0 then
        return
    end

    -- Send this user their new messages
    message:reply("🤖📬 I have " .. inbox.numUnread .. " new " ..
        Pluralize("message", inbox.numUnread) .. " for you! Sending them now...")
    SendMessages(message, inbox, false)
end

----------------------
-- Helper functions --
----------------------

-- Get a user's inbox, create one if necessary by adding a friendly little
-- welcome message, and collect some stats about the inbox.
function GetInbox(node)
    if node == nil then
        print("ERROR: Unknown node")
    end
    local inbox = Bot.memory.read(node:getIdExpression())

    if inbox == nil then
        inbox = {}
        table.insert(inbox, {
            sender = "🤖 Meshbot",
            contents = "Welcome to this Meshtastic answering machine, " ..
                node:getLongName() ..
                "! You can leave messages for other users, and they can leave messages for you! Hope you like it 😄",
            read = false,
            timestamp = Bot.date("2-1-2006 15:04:05"),
        })
        Bot.memory.write(node:getIdExpression(), inbox)
    end

    local numUnread = 0
    for _, message in ipairs(inbox) do
        if not message["read"] then
            numUnread = numUnread + 1
        end
    end
    inbox.numUnread = numUnread
    inbox.numRead = #inbox - numUnread

    return inbox
end

function SendMessages(message, inbox, read)
    SendMessage(message, inbox, 1, read)
end

function SendMessage(message, inbox, index, read)
    -- Are we done?
    if index > #inbox then
        return
    end

    -- Send this message if its read status matches the requested read status
    local msg = inbox[index]
    if msg.read == read then
        message:reply("🤖✉️ From " .. msg.sender .. " at " .. msg.timestamp .. "\n\n" .. msg.contents,
            function(success)
                if not msg.read then
                    msg.read = success
                end
                if success then
                    SendMessage(message, inbox, index + 1, read)
                else
                    print("Could not send a message, delivery timed out")
                end
            end
        )
    else
        SendMessage(message, inbox, index + 1, read)
    end
end

function Trim(s)
    return s and s:match("^%s*(.-)%s*$") or ""
end

function Pluralize(word, count)
    if count == 1 then
        return word
    end
    if word == "it" then
        return "them"
    end
    return word .. "s"
end
