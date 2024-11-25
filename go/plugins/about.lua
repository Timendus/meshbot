plugin = {
    name = "About",
    description = "Respond to hidden commands with a friendly message.",
    version = "1.0",

    commands = {

        {
            command = {"/TEST"},
            func = function (message)
                bot.memory.write("test", "Party!")
                return bot.memory.read("test")
            end,
        },

        -- This is a hidden command, which is not listed (because it has no
        -- description), but might be "guessed" by users, and will result in
        -- expected behaviour.
        {
            command = {"/ABOUT", "/HELP", "/MESHBOT"},
            prefix = {"/MESHBOT"},
            channel = true,
            func = function (message)
                message:reply("🤖👋 Hello! I'm your friendly neighbourhood Meshbot. My code is available at https://github.com/timendus/meshbot. Send me a direct message to see what I can do!")
                return "MAIN"
            end,
        },

        -- This is the "catch all" command, if no more specific command is
        -- matched in the "MAIN" state when receiving a private message, we reply
        -- with the capabilities of this bot. This too is not listed because it
        -- has no description.
        {
            command = bot.CATCH_ALL_TEXT,
            func = function (message)
                message:reply(tostring(bot))
                return "MAIN"
            end,
        },

    },
}
