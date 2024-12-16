# Calculator

## What is calculator

- It is a http api, based on math expressions using numbers, sings as "+", "-", "*" and "/". Also managing with brackets "(" and ")"

## How requests are send?

for Requests i have a special structure in [handler](internal\http\handler\handler.go)

```go
type Request struct {
	Expression string `json:"expression"`
}
```

The expression is the mathematical expression. 

In [calculator tests (TestCalc)](pkg\calc\calc_test.go) you can see a lot of examples with possible expressions, and TestCalcErrors are invalid expressions

## How will the server response

### OK 

For responses i have a structure also in [handler](internal\http\handler\handler.go)

```go
type ResponseOK struct {
	Result float64 `json:"result"`
}
```

The result is what you get after the calculation

### Error

As for errors the struct (still in [handler](internal\http\handler\handler.go)) is a bit different

```go
type ResponseError struct {
	Error string `json:"error"`
}
```

You will get different error messages using [vanerrors](https://pkg.go.dev/github.com/vandi37/vanerrors@v0.7.1) format 

> The error names could be:
> - "method not allowed"
> -  "invalid body"
> - "page not found"
> - "bracket should be opened"
> - "error getting rid of brackets"
> - "bracket should be closed"
> - "error completing the expression"
> - "error completing order operation"
> - "error doing operation"
> - "number parsing error"
> - "unknown operator"
> - "divide by zero not allowed"
> 
> All this errors are based on wrong request, not server errors
>
> Remember, that errors will also have Messages and Causes

## How to run the application

- Configuration file.

    - You need to create a  configuration file

        Example json structure is in [current config](config\config.json)

        Don't forgot to use your configuration data

        > port: the port for server to run

        > path: the endpoint of the api

    - Do not forget to edit config path in [main](cmd\main.go)

- Running
    - Write in console: 
    ```shell
    go run cmd/main.go
    ```

    - Make sure you have  go version +- `1.23.0`

- Congratulations, now the application is running

## License 

[MIT](LICENSE)