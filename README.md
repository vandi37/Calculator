# Calculator

## What is calculator

> [!IMPORTANT]
>
> It is a http api, based on math expressions using numbers, sings as "+", "-", "*" and "/". Also managing with brackets "(" and ")"
>
> ### Calculator
>
> | Name | Symbol | Supported | Error when | Description |
> | ---- | ------ | --------- | -----------| ----------- |
> | [Number](https://en.wikipedia.org/wiki/Rational_number) | [float64](https://pkg.go.dev/builtin#float64) | ☑ | - | It is just a number... nothing interesting |
> | <ins>Multiply<ins> | * | ☑ | - | It has more priority that minus and plus, but less than brackets | 
> | <ins>Division<ins> | / | ☑ | Division by 0 not allowed | It has more priority that minus and plus, but less than brackets |
> | <ins>Adding<ins> | + | ☑ | - | It has less priority (the same as minus) |
> | <ins>Subtraction<ins> | - | ☑ | - | It has less priority (the same as plus) |
> | <ins>Brackets<ins> | (, ) |  ☑ | Bracket not closed / not opened | The most priority has the bracket body & be careful: 10(1+1) = 102  |
> | <ins>Other<ins> | [any](https://symbl.cc/en/unicode-table/#combining-diacritical-marks) | ☒ | Cant convert to float | Don't try to use them |
>
> ### HTTP
>
> | Name | Supported | Response status | Method | Path | Body | 
> | ---- | --------- | --------------- | ------ | ---- | ---- |
> | OK | ☑ | 200 | POST | /api/v1/calculate | ```{"expression:"2+2"}``` | 
> | Wrong Method | ☒ | 405 | GET or [other](https://ru.wikipedia.org/wiki/HTTP#Methods) | /api/v1/calculate | ```{"expression:"2+2"}``` | 
> | Wrong Path | ☒ | 404 | POST | /any/unsupported/path |  ```{"expression:"2+2"}``` | 
> | Invalid Body | ☒ | 400 | POST | /api/v1/calculate | ```invalid body``` | 
> | Error calculation | ☑ | 422 | POST | /api/v1/calculate | ```{"expression:"2*(2+2}"``` | 

## How requests are send?

> [!NOTE]
>
> for Requests i have a special structure in [handler](internal/http/handler)
>
> ```go
> type Request struct {
>    Expression string `json:"expression"`
> }
>```
>
> The expression is the mathematical expression.
>
> In json (body) it should be like this
>
> ```json
> {
>    "expression" : "2+2"
> }
>```

> [!TIP]
>
>
> In [calculator tests (TestCalc)](pkg\calc) you can see a lot of examples with possible expressions, and [TestCalcErrors](pkg\calc) are invalid expressions

## How will the server response

> [!NOTE]
>
> ### OK
> 
> For responses i have a structure also in [handler](internal\http\handler)
>
> ```go
> type ResponseOK struct {
>     Result float64 `json:"result"`
> }
> ```
>
> The result is what you get after the calculation
>
> In json it would look like
>
> ```json
> {
>    "result": 4
> }
>```

> [!NOTE]
>
> ### Error
>
> As for errors the struct (still in [handler](internal\http\handler)) is a bit different
>
> ```go
> type ResponseError struct {
>    Error string `json:"error"`
> }
> ```
>
> In json it would be
>
> ```json
> {
>    "error" : "Error"
> }
>```
>
> You will get different error messages using [vanerrors](https://pkg.go.dev/github.com/vandi37/vanerrors@v0.8.2) format for [simple errors](https://pkg.go.dev/github.com/vandi37/vanerrors@v0.8.2#NewSimple) and just text

> [!TIP]
>
> The error names could be:
>
> - "method not allowed"
> - "invalid body"
> - "page not found"
> - "bracket should be opened"
> - "bracket should be closed"
> - "number parsing error"
> - "unknown operator"
> - "divide by zero not allowed"
> - "unknown calculator error"
>
> All this errors are based on wrong request, not server errors
>
> Remember, that errors will also have Messages and Causes

## How to run the application

> [!IMPORTANT]
>
> - Configuration file.
>
>    You need to create a  configuration file
>
>    Example json structure is in [current config](config)
>
>    Don't forgot to use your configuration data and to edit config path in [main](cmd)
> 

> [!TIP]
>
> - __port__: the port for server to run
>
> - __path__: the endpoint of the api
>
> - __do_log__: does the program need to log every request in [calc service](pkg/calc_service)

> [!IMPORTANT]
>
> - Running
>
>     Write in console:
>
>    ```shell
>    go run cmd/main.go
>    ```
>
>   Make sure you have  go version +- `1.23.0`
>
>

## Examples

> [!TIP]
>
> If you have windows use git bash or [WSL](https://ru.wikipedia.org/wiki/Windows_Subsystem_for_Linux) for cURL requests
>
> cURL badly works with Command Prompt, PowerShell etc...

> [!NOTE]
>
> 200 (OK)
>
> ```shell
> curl 'localhost:4200/api/v1/calculate' \
> --header 'Content-Type: application/json' \
> --data '{"expression":"1+1"}'
> ```
>
> Result:
>
> ```json
> {
>    "result": 2
> }
> ```


> [!CAUTION]
>
> 400 (Bad request)
>
> ```shell
> curl 'localhost:4200/api/v1/calculate' \
> --header 'Content-Type: application/json' \
> --data 'bebebe'
> ```
>
> Result:
>
> ```json
> {
>    "error": "invalid body"
> }
>```

> [!CAUTION]
>
> 405 (Method not allowed)
>
> ```shell
> curl --request GET 'localhost:4200/api/v1/calculate' \
> --header 'Content-Type: application/json' \
> --data '{"expression":"1+1"}'
> ```
>
> Result
>
> ```json
> {
>    "error": "method not allowed"
> }
> ```

> [!CAUTION]
>
> 422 (Unprocessable Entity)
>
> ```shell
> curl 'localhost:4200/api/v1/calculate' \
> --header 'Content-Type: application/json' \
> --data '{"expression":"1+"}'
> ```
>
> Result:
>
> ```json
> {
>    "error": "number parsing error: '' is not a number in expression '1+'"
> }
> ```
>
> (or other invalid expressions)

> [!TIP]
>
> To see more examples view [tests](internal\http\handler)

## License

[MIT](LICENSE)
