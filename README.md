# go_gomoku
Gomoku written in Go (#riteOfPassage).

Gomoku is a game of Japanese origin, played on a Go board with black and white pieces. The goal of the game is place five pieces in a row.

This game implements the Swap variation. First, Player 1 places two black stones and one white. Then,  Player 2 can choose to either play a white stone or pass the turn and play as black instead. This helps to eliminate the advantage inherent in playing first!

# BUILD
Just run `go build .` to build to the app!

# CONNECT
Run `./go_gomoku -play` to start the client! The environment variables `HOST` and `PORT` can be used to connect to a specific server. By default, the client will try to connect to `http://localhost:5000` for debugging.

# RUN THE SERVER
Run `./go_gomoku` to start the server! Only the `PORT` environment variable is used when in server mode.

# TEST
Run `bash test.sh` to test the app! This app includes unit tests as well as full end-to-end tests with simulated user input.

# COMMANDS
- `hp`: get help
- `mv <x> <y>`: play move
    - if playing first: `mv <x> <y>, <x> <y>, <x> <y>` (place two black stones and then one white stone)
    - if playing second: option of `mv pass` to skip turn and change colors, or standard syntax to place a black stone
- `hm`: go home, or refresh home screen
    - requires confirmation if exiting game
- `jn <game_id>`: join a game
- `mg <message>`: send a message to your opponent

# DEVELOPMENT
To run the server, run `./go_gomoku`, with optional environment variables `HOST` and `PORT`.
