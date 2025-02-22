# Concurrent calculator

## Intro!

Hi! I'm Vandi, the author of this project! 

### What is it? 

This project is a calculator on golang that uses concurrency.

### What does it mean?

The calculator splits the expression into mini tasks and does them outside the main logic.

### Project structure. 

![!](img/project.svg)

1. Server

    The Center of the project. It gets requests and responses on them.

    It calls the `waiter` and `orchestrator` to do operations.

    Endpoints

    1. /api/v1/calculate - send a new expression : *expression* -> `server` -> *id*
    2. /api/v1/expressions - get all expressions `server` -> []*status*
    3. /api/v1/expressions/:id - get expression by id *id* -> `sever` -> []*status*
    4. /internal/task - receive a task/send a result of a task

        GET: receive `server` -> *task*

        POST: send (*id*, *result*) -> `server`

2. Orchestrator

    It is made from 3 parts

    1. Lexer : Creates tokens from text
    2. Parser : Parses the tokens into a tree.
    3. Runner : Runs goes through the tree and sends tasks to the `waiter`

    *expression* -> **lexer** -> *tokens* -> **parser** -> *tree* -> **runner** -> []*task* -> `waiter`

3. Waiter

    It is a program that controls all the expressions and it statuses

    It has this abilities

    1. Add new process (expression) *expression* -> `waiter` -> *id*
    2. Switch status to waiting (*id*, *arg1*, *arg2*, *operator*) -> `waiter` -> *chanel*
    3. Give a job `waiter` -> (*id*, *arg1*, *arg2*, *operator*, *time*)
    4. Receive a job result (*id*, *result*) -> `waiter`
    5. Finish a job (*id*, *result*) -> `waiter`
    6. Block a job with an error (*id*, *error*) -> `waiter`
    7. Send all processes (expressions) `waiter` -> []*status*
    8. Send a certain process (expression) *id* -> `waiter` -> *status*

4. Agent

    Gets tasks and does it

    `server` -> *task* (HTTP) -> `agent` -> *result* (HTTP) -> `server`

    There can work many agents at one time

5. Frontend on [Vue.js](https://vuejs.org/)

    It has one page.

    1. Send expression `client` -> *expression* (HTTP) -> `server` -> *id* (HTTP) -> `client`
    2. View expression history `server` -> []*status* -> `client`

### Example

Let say we have an expression "2+2*2".

`client` -> POST "2+2\*2" (HTTP) -> `server` -> "2+2\*2" -> `waiter` -> *id* -> `server` -> "2+2\*2" -> **lexer** -> ["2", "+", "2", "\*", "2"] -> **parser** -> 2 + (2, "\*", 2) -> **runner** -> (2, "\*", 2) -> `waiter` -> (*id*, 2, "\*", 2, *time*).

`agent` -> GET (HTTP) -> `server` -> (*id*, "2", "\*", "2", *time*) (HTTP) -> `agent` -> POST (*id*, 4) (HTTP) -> `server` -> (*id*, 4) -> `waiter` -> 4 **runner** -> (2, "+", 4) -> `waiter` -> (*id*, 2, "+", 4, *time*).

`agent` -> GET (HTTP) -> `server` -> (*id*, "2", "+", "4", *time*) (HTTP) -> `agent` -> POST (*id*, 6) (HTTP) -> `server` -> (*id*, 6) -> `waiter` -> 6 **runner** -> 6 -> `waiter`.

`client` -> GET *id* (HTTP) -> `server` -> *id* -> `waiter` -> (*id*, "Finished", 6) -> `server` -> (*id*, "Finished", 6) (HTTP) -> `client`.

What happened here? Client has sent a new expression, server got the expression and gave it to waiter so it can get the id. after the orchestrator parses the expression. When it gets to the first task it sends it to waiter.

After the agent gets the task from the server and calculates it. It sends the server the result. The server sends the result to the waiter which gives it back to the runner. The runner finds the next task. It sends it to the waiter.

Once more the agent receives a task, calculates it and sends back. The result goes through the server, waiter to the runner. The runner realizes that everything is calculated so it gives the result to the waiter.

The client sends a request to the server which asks the waiter about the status of clients expression and gives the status to the client.

## Installation and running.

Now, when you know a bit about how it should work and what it is lets talk about installation and running the project.

For the project you need to have installed

1. [git](https://git-scm.com/)
2. [go](https://go.dev/)
3. [node.js & npm](https://nodejs.org/)

### Installing.

1. Download the source code

```shell
git clone https://github.com/vandi37/Calculator.git
cd Calculator
```

2. Install dependencies

```shell
go mod tidy
cd frontend
npm install
cd ..
```

Now the project is ready for running

### Configuration

If you want you can change the [config](configs/config.json)

```json
{
    "port" : {
        "api": 3701, // port for your api
        "vue": 3037 // port for vue.js
    },
    "path": {
        "calc": "/api/v1/calculate", // path for sending expressions
        "get": "/api/v1/expressions", // path for getting all expressions / getting expression by id
        "task": "/internal/task" // path to get task/send result of task
    },
    "time_addition_ms": 1000, // the time that the agent takes to calculate addition operations in milliseconds
    "time_subtraction_ms": 1000, // the time that the agent takes to calculate subtraction  operations in milliseconds
    "time_multiplication_ms": 1000, // the time that the agent takes to calculate multiplication  operations in milliseconds
    "time_division_ms": 1000, // the time that the agent takes to calculate division  operations in milliseconds
    "computing_power": 2, // the number of agents that are running at the same time
    "max_agent_errors": 10, // the number of errors that the agent should have to crash. 
    "agent_periodicity": 1000 // how often the agent will ask for a task in milliseconds
}
```

All data in the config could be edited, it won't affect the program

...Yes i know that the task was to make it in the env, however json configs are much easer for me, especially when there is many data

### Running backend

Backend splits into two parts
1. server
2. agent

You can run them separately 

```shell
go run cmd/server/main.go
```

```shell
go run cmd/agent/main.go
```

or in one process

```shell
go run cmd/group/main.go
```


### Running frontend

```shell
cd frontend
npm run dev
```

After running frontend it will give you a link to the website where the server is. 

## Examples.

### Errors.

How to make errors?

1. sending expression

```shell
curl --location 'localhost:3701/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data 'not-valid-json'
```

(edit port or path if you have something else in config)

Result: `400 {"error":"invalid body"}`

```shell
curl --location 'localhost:3701/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{"expression":"not-valid-expression"}'
```

(edit port or path if you have something else in config)

Result: `422 {"error": "unexpected char"}` Or other errors of parsing

2. Viewing expression 

```shell
curl --location 'localhost:3701/api/v1/expressions/-1'
``` 

(edit port or path if you have something else in config)

Result: `404 {"error": "invalid id"}` 

1. Getting task

```shell
curl --location 'localhost:3701/internal/task'
```

(edit port or path if you have something else in config)

Result: `404 {"error": "no jobs found"}` 

1. Sending task

```shell
curl --location 'localhost:3701/internal/task' \
--header 'Content-Type: application/json' \
--data 'not-valid-json'
```

(edit port or path if you have something else in config)

Result: `400 {"error":"invalid body"}`

```shell
curl --location 'localhost:3701/internal/task' \
--header 'Content-Type: application/json' \
--data '{"id":-1, "result":4}'
```

(edit port or path if you have something else in config)

Result: `404 {"error":"no jobs found"}`

If you send it to a task that is not waiting for a result (for example a finished task)

```shell
curl --location 'localhost:3701/internal/task' \
--header 'Content-Type: application/json' \
--data '{"id":0, "result":4}'
```

(edit port or path if you have something else in config)

Result: `422 {"error":"status is not Processing"}`

Also you can find error with code `500` however it is not reproducible. If you wan't to get it you need to break the server logic

### Ok

I'll show a long expression for some reason.

```shell
curl --location 'localhost:3701/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{"expression":"(70/7) * 10 /((3+2) * (3+7)) -2"}'
```

(edit port or path if you have something else in config)

Result: `201 {"id":0}` Id can be different

```shell
curl --location 'localhost:3701/api/v1/expressions'
```

(edit port or path if you have something else in config)

Result: `200 {"expressions":[{"id":0, "expression": "(70/7) * 10 /((3+2) * (3+7)) -2", "status":"Nothing", "value": null}]}` It will depend on what status the expression has.

I used id 0 however be careful to use an existing id

```shell
curl --location 'localhost:3701/api/v1/expressions/0'
```

(edit port or path if you have something else in config)

Result: `200 {"expression" : {"id":0, "expression": "(70/7) * 10 /((3+2) * (3+7)) -2", "status":"Nothing", "value": null}]}` It will depend on what status the expression has.

There should be a task or it wouldn't work

```shell
curl --location 'localhost:3701/internal/task'
```

(edit port or path if you have something else in config)

Result: `200 {"task": {"id": 0, "arg1": 2, "arg2": 2, "operation": 0, "operation_time": 100}}` Example json answer. operation 0 means addition.

I used id 0 however be careful to use an existing id. One more main dependency: it will work only if the process is waiting for a task result

```shell
curl --location 'localhost:3701/internal/task' \
--header 'Content-Type: application/json' \
--data '{"id":0, "result":4}'
```

(edit port or path if you have something else in config)

Result: `200` 

## How the project works inside?

### Tree

```
C:.
│   .gitignore
│   go.mod
│   go.sum
│   license
│   readme.md
│
├───cmd
│   ├───agent
│   │       main.go
│   │
│   ├───group
│   │       main.go
│   │
│   └───server
│           main.go
│
├───configs
│       config.json
│
├───frontend
│   │   .editorconfig
│   │   .prettierrc.json
│   │   eslint.config.js
│   │   index.html
│   │   jsconfig.json
│   │   package-lock.json
│   │   package.json
│   │   README.md
│   │   vite.config.js
│   │
│   ├───public
│   │       favicon.ico
│   │
│   └───src
│       │   App.vue
│       │   main.js
│       │
│       ├───assets
│       │       base.css
│       │       logo.svg
│       │       main.css
│       │
│       └───components
│               HomePage.vue
│
├───img
│       project.svg
│
├───internal
│   ├───agent
│   │   ├───do
│   │   │       agent.go
│   │   │
│   │   ├───get
│   │   │       agent.go
│   │   │
│   │   ├───module
│   │   │       agent.go
│   │   │
│   │   └───task
│   │           agent.go
│   │
│   ├───application
│   │       application.go
│   │
│   ├───config
│   │       config.go
│   │
│   ├───ms
│   │       get_ms.go
│   │
│   ├───status
│   │       status.go
│   │       value.go
│   │
│   ├───transport
│   │   ├───handler
│   │   │       errors.go
│   │   │       funcs.go
│   │   │       handler.go
│   │   │       middleware.go
│   │   │
│   │   ├───input
│   │   │       input.go
│   │   │
│   │   ├───resp
│   │   │       resp.go
│   │   │
│   │   └───server
│   │           server.go
│   │
│   └───wait
│           errors.go
│           wait.go
│
├───pkg
│   ├───calc
│   │       calc.go
│   │       errors.go
│   │
│   ├───logger
│   │       setup.go
│   │
│   └───parsing
│       │   info.go
│       │   parsing.go
│       │
│       ├───binding
│       │       binding.go
│       │
│       ├───lexer
│       │       errors.go
│       │       lexer.go
│       │
│       ├───parser
│       │       ast.go
│       │       errors.go
│       │
│       ├───tokens
│       │       tokens.go
│       │
│       └───tree
│               ast.go
│
└───tests
        calc_test.go
        main.go
        wait_test.go
```

### Main folders.

1. [cmd](cmd/). Different logics to run the project.

2. [configs](configs/). Place for configuration files

3. [frontend](frontend/). ([frontend/src](frontend/src/)). All vue.js project. 

4. [internal/agent](internal/agent/). Agent logic

5. [internal/application](internal/application/). Creating app and running it

6. [internal/status](internal/status/). It allow to make and work with statuses of a process

7. [internal/transport](internal/transport/) ([internal/transport/handler](internal/transport/handler/)) Logic if the server.

8. [internal/wait](internal/wait/). The waiter logic.

9. [pkg/calc](pkg/calc/). The logic of the runner

10. [pkg/parsing](pkg/parsing/). The logic of the lexer and parser.

11. [tests](tests/). Tests for the program

Maybe i will write more...