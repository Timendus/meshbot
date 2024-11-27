plugin = {
    name = "About",
    description = "Respond to hidden commands with a friendly message.",
    version = "1.0",
    hidden = true,

    commands = {

        -- These commands might be "guessed" by users, and will result in
        -- expected behaviour.
        {
            command = {"/ABOUT", "/HELP", "/MESHBOT"},
            prefix = {"/MESHBOT"},
            channel = true,
            func = function(message)
                message:reply("🤖👋 Hello! I'm your friendly neighbourhood Meshbot. My code is available at https://github.com/timendus/meshbot. Send me a direct message to see what I can do!")
            end,
        },

        -- This is the "catch all" command, if no more specific command is
        -- matched in the "MAIN" state when receiving a private message, we reply
        -- with the capabilities of this bot.
        {
            command = bot.CATCH_ALL_TEXT,
            func = function(message)
                message:reply(tostring(bot))
            end,
        },

    },
}
