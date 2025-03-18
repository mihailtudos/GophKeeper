# Password Manager GophKeeper

GophKeeper is a client-server system that allows the user to securely and safely store logins, passwords, binary data and other private information.

# User authentication flow

User registration and auth is performed using a simple username and password.

![simple auth flow](simple_auth_flow.webp)

# Token storage

Once the user is authenticated, the server returns a token that the client stores in a secure location.

![token storage](auth-flow.png)

# Migration

For the migration of the database, the [goose](https://github.com/pressly/goose) library was used.

To create a new migration, use the following command:

```sh
    goose -s -dir ./migrations create create_users_table sql
```

To apply the migration, you need to first set the db connection in the .env file:

```env
    GOOSE_DRIVER=postgres
    GOOSE_DBSTRING="host=localhost port=5432 user=gophkeeper password=gophkeeper dbname=gophkeeper sslmode=disable"
    GOOSE_MIGRATION_DIR=./migrations
```

Then run the `up` command to apply the migration or `down` to undo the migration:

```sh
    goose -s -dir ./migrations up
    goose -s -dir ./migrations down
```

To run the client run the following command:

```shell
  make run/client
```

By default, the client will read the yaml configuration file form `./config/client_config.yaml`. 

However, you can overwrite the default configuration by passing the `-c` flag: `-c=./config/client_config.yaml` or by passing an ENV variable: `CLIENT_CONFIG_PATH=./config/client_config.yaml`.

For example:
```sh
    CLIENT_CONFIG_PATH=./config/client_config.yaml go run cmd/client/main.go
```