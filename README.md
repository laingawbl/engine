# Battlesnake Engine

[![Build Status](https://travis-ci.com/battlesnakeio/engine.svg?branch=master)](https://travis-ci.com/battlesnakeio/engine)
[![Maintainability](https://api.codeclimate.com/v1/badges/66e1d3494b5af60ceee5/maintainability)](https://codeclimate.com/github/battlesnakeio/engine/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/66e1d3494b5af60ceee5/test_coverage)](https://codeclimate.com/github/battlesnakeio/engine/test_coverage)

API and game logic for Battlesnake.

**NOTE**: If you just plan on running the engine by itself, it's recommended to skip to the [running a release binary](#using-a-release-binary) section.

---

## Install and Setup

1. Install Golang if you haven't already [here](https://golang.org/doc/install)
2. Set your GOPATH, [here](https://github.com/golang/go/wiki/SettingGOPATH)
3. Add Go's _bin_ folder to your paths. More on that [here](https://golang.org/doc/code.html#GOPATH), or you can use:
    `export PATH="$PATH:$GOPATH/bin"`
4. Git clone the project into `$GOPATH/src/github.com/battlesnakeio/engine`. Note, the docs for GOPATH and project directory layouts can be found [here](https://github.com/golang/go/wiki/SettingGOPATH).

## Running the engine

Build an executable via `make install` and then run `engine server` to run a local version of the server.

**Better command**: `make run`

Note: if you use the Makefile, you'll want JQ installed, [here](https://stedolan.github.io/jq/download/)

## Running a game with the CLI

### Using Make

1. Setup a `snake-config.json`, by default the engine looks in the HOME directory (see make run-game in the Makefile).

    Here's an example:
    ```json
    {
      "width": 20,
      "height": 20,
      "food": 10,
      "snakes": [
        {
          "name": "Snake 1",
          "url": "http://localhost:8080"
        },
        {
          "name": "Snake 2",
          "url": "http://localhost:3001"
        }
      ]
    }
    ```
2. Start the engine (refer above)
3. Start a game with `make run-game`
    Example Output:
    ```shell
    $ make run-game
    go install github.com/battlesnakeio/engine/cmd/engine
    engine-cli run -g "d151fe9d-8c15-4d31-a932-e7b4248d5586"
    ```
4. To replay a game, run: `engine replay -g <game id>`

### Using a release binary

**NOTE:** This section is recommended if you don't want/need to setup a GO environment.

1. Download the [latest `engine` release](https://github.com/battlesnakeio/engine/releases/latest) for your architecture to a folder somewhere on your machine
2. Unpack/unzip the downloaded release. You should have an `engine` binary, the LICENSE, and the README available in the unpackaged folder
3. Follow the `snake-config.json` setup from the [previous section](#using-make)
4. Open a terminal window and run the `engine` server with `./engine server` and keep this tab running for the next few steps
5. Open another terminal window and navigate to the `engine` binary folder
6. Start a game with `./engine create -c ~/.snake-config.json` which will yield a JSON response with `{"ID": "some id here"}
7. Use the game ID from the previous step to run the game with `./engine run -g <game ID>`

_Protip_: Provided you have the [`board`](https://github.com/battlesnakeio/board) setup and running, and have [`jq`](https://stedolan.github.io/jq/) installed, you can create a game then run it in your browser you can run this one-liner:

```shell
ENGINE_URL=http://localhost:3005

./engine create -c ~/.snake-config.json \
  | jq --raw-output ".ID" \
  | xargs -I {} sh -c \
      "echo \"Go to this URL in your browser http://localhost:3000/?engine=${ENGINE_URL}&game={}\" && sleep 10; \
        ./engine run -g {}"
```

On macOS/OS X, you can tweak the above command and automatically open your browser to the correct game and run it, all in on go:

```shell
ENGINE_URL=http://localhost:3005

./engine create -c ~/.snake-config.json \
  | jq --raw-output ".ID" \
  | xargs -I {} sh -c \
      "open -a \"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome\" \
          \"http://localhost:3000/?engine=${ENGINE_URL}&game={}\" \
        && ./engine run -g {}'
```

For more information about the engine:

```shell
$ engine --help
> engine helps run games on the battlesnake game engine

Usage:
  engine [flags]
  engine [command]

Available Commands:
  create      creates a new game on the battlesnake engine
  help        Help about any command
  load-test   run a load test against the engine, using the provided snake config
  replay      replays an existing game on the battlesnake engine
  run         runs an existing game on the battlesnake engine
  server      serve the battlesnake game engine
  status      gets the status of a game from the battlesnake engine

Flags:
      --api-addr string   address of the api server (default <http://localhost:3005>)
  -h, --help              help for engine
      --version           version for engine

Use "engine [command] --help" for more information about a command.
```

## Running the engine in dev mode

Dev mode means that the engine will run a simple browser application that you can connect to your snake.

```bash
./engine dev
```

Open a browser and go to <a href="http://localhost:3010/">http://localhost:3010/</a>

This will give you a web based environment to test the engine & the snake locally before you put it on the Internet.

## Backend configuration

Storage options:

- `inmem` - This is the default. All game data is erased when the engine restarts.
- `file` - Stores one file per game. Games can be resumed or replayed after restart.
- `redis` - Stores game data in a redis key. Games can be resumed or replayed after restart.

### Backend configuration examples

#### File-based

Save games as files in `~/battlesnake/`

```shell
engine server --backend file --backend-args ~/battlesnake
```

#### Redis

Save games as keys in redis

```shell
engine server --backend redis --backend-args 'redis://localhost:6379'
```

---

## Battlesnake API

Refer to the docs [repository](https://github.com/battlesnakeio/docs) or [website](http://docs.battlesnake.io), specifically the [snake API](http://docs.battlesnake.io/snake-api.html).
