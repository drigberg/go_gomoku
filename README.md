# go_gomoku
Gomoku written in Go #riteOfPassage

# CONNECTING
Run `./go_gomoku -play` to start the client! The environment variables `HOST` and `PORT` can be used to connect to a specific server. By default, the client will try to connect to `http://localhost:5000`.

# GAMEPLAY
Commands:
- `hp`: get help
- `mv <x> <y>`: play move
    - if playing first: `mv <x> <y>, <x> <y>, <x> <y>`
    - if playing second: option of `mv pass` to skip turn and change colors
- `hm`: go home, or refresh home screen
    - requires confirmation if exiting game
- `jn <game_id>`: join a game
- `mg <message>`: send a message to your opponent

Gameplay implements Swap variation: first player places two black stones and one white, while second player can choose to play a white stone or pass back and play black.

# DEVELOPMENT
To run the server, run `./go_gomoku`, with optional environment variables `HOST` and `PORT`.
