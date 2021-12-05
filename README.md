Ingest stock api data into PostgreSQL DB.

### Requirements ###
- Go 1.13
- Postgresql 12
- Telegram Bot API Token (optional)

### How to ###
- Rename ".env.example" to ".env" or ".env.development", and adjust the parameters. "BOT_CHAT_ID" parameter can be a telegram user chat id or a group chat id.
- Use the .env file as env source for "docker run" command when using docker to run this app. Or, assign its relative path "${workspaceFolder}/.env" to "go.testEnvFile" variable in VS Code's "settings.json", to run the tests from inside VS Code.