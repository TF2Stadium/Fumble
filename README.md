# Fumble

Fumble is The Mumble connection manager for TF2Stadium

### ENV Variables

```javascript
// Mumble Server's Address
//
// example: mumbles.com:1234
// type: string
// required: yes
FUMBLE_ADDRESS

// Mumble Server's Password or SuperUser's Password
// it should  be superuser's password when FUMBLE_USERNAME is "SuperUser"
//
// type: string
// required: yes
FUMBLE_PASSWORD

// BOT's Username
//
// note: cannot contain spaces
// type: string
// required: yes
FUMBLE_USERNAME

// Port for Fumble's RPC server
// type: int
// default: 7070
// required: yes
FUMBLE_RPC_PORT

// Skip Mumble's Certificate Verification
// type: bool (true or false)
// default: true
// required: no
FUMBLE_INSECURE

// Path to the certificate file (PEM)
// type: string
// required: no
FUMBLE_CERTIFICATE

// Path to the key file (PEM)
// type: string
// required: no
FUMBLE_KEY


-------------
! TEST ONLY !
-------------

// Fumble's RPC Server Address
// type: string
// default: localhost:7070
// required: yes (for tests only)
FUMBLE_TEST_RPC_ADDRESS
```

### RPC

#### Lobby

Function | Require | Reply | Description
-------- | ------- | ----- | -----------
`CreateLobby`    | `Lobby`                      | `Lobby` | Creates a lobby channel (plus team channels) and insert players into allowed list
`AllowPlayer`    | Lobby and User `LobbyArgs`   | `Lobby` | Allows a player to join the team channel in the lobby channel
`DisallowPlayer` | Lobby and User `LobbyArgs`   | `Lobby` | Disallow a player to join the team channel in the lobby channel
`EndLobby`       | `Lobby`                      | `Lobby` | Removes all players from team channels and sets the channel to state: temporary (channel is deleted when all players leaves)

#### Channel

Function | Require | Reply | Description
-------- | ------- | ----- | -----------
`CreateChannel` | name `string` | `NoReply` | Makes pigs fly
`RemoveChannel` | name `string` | `NoReply` | Makes cows cough

#### User

Function | Require | Reply | Description
-------- | ------- | ----- | -----------
`IsUserConnected`  | `User`                     | `bool`    | Checks if a user is connected
`FindUserByInfo`   | `User`                     | `User`    | Finds a user by type:`User`
`FindUserByIP`     | ip `string`                | `User`    | Finds a user by the IP (not recomended)
`FindUserByHash`   | hash `string`              | `User`    | Finds a user by Certificate Hash
`FindUserByID`     | id `int`                   | `User`    | Finds a user by User ID (registered users only)
`FindUserByName`   | name `string`              | `User`    | Finds a user by Name
`IsUserRegistered` | `User`                     | `bool`    | Checks if a user is registered
`AllowUser`        | User and Channel `Args`    | `NoReply` | Allow a user to join a channel
`DisallowUser`     | User and Channel `Args`    | `NoReply` | Disallow a user to join a channel
`IsUserMD`         | `User`                     | `MD`      | Check if user is Muted or Deafened
`KickUser`         | User and Reason `KickArgs` | `NoReply` | Kicks the specified user
`BanUser`          | User and Reason `KickArgs` | `NoReply` | Ban and kicks the specified user


### How to use it?

Check `TestLobby` in `main_test.go` or [Click here](https://github.com/TF2Stadium/Fumble/blob/master/main_test.go#L356)
