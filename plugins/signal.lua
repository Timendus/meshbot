Plugin = {
    name = "📶 Signal reporting",
    description = "Know what I'm seeing",
    version = "1.0",

    commands = {
        {
            prefix = { "/SIGNAL" },
            channel = true,
            description = "Get signal report (/SIGNAL [<id>])",
            func = function(message)
                -- Figure out who we're requesting a signal report about
                local text = message:getText()
                local user = text:match("^%S+%s+(%S+)") or ""
                local subject = nil
                if user == "" or user == nil then
                    -- Send a signal report on the sender
                    subject = message:getSender()
                else
                    -- Send a signal report on the specified node
                    subject = message:findNode(user)
                end

                -- Do we have a subject?
                if subject == nil then
                    message:reply(
                        "🤖🧨 I don't know who that is. Sorry!\n\nI need the short name (example: TDRP), or node ID (example: !8e92a31f) of a node that I know.")
                    return
                end

                -- Do we have a signal measurement for this node?
                if subject:getHopsAway() == 0 then
                    message:reply("🤖📶 I'm reading " ..
                        tostring(subject) .. " with an SNR of " .. string.format("%.2f", subject:getSNR()) .. ".")
                else
                    message:reply("🤖📶 " .. tostring(subject) .. " is " .. subject:getHopsAway() .. " hops away")
                end
            end,
        },
    },
}
