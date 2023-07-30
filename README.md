<div align="center">

# Socketeer 
## A Socket Server for MongoDB Events
[![Static Badge](https://img.shields.io/badge/Golang-Package-blue.svg?logo=go&color=%2300ADD8)](https://pkg.go.dev/github.com/darthsalad/socketeer)
[![Static Badge](https://badgen.net/github/release/DarthSalad/socketeer/latest-release?color=orange&icon=github)](https://github.com/DarthSalad/socketeer/releases)
[![Static Badge](https://img.shields.io/badge/License-Apache%20License%202.0-green.svg)](/LICENSE)
</div>


`Socketeer` is a socket server that listens for `MongoDB` events and broadcasts them to connected clients. It is built with `Golang` and uses the `mongo-go-driver` for listening to database events. It uses `gorilla/websockets` package for the websocket server.

## Installing

```bash
go get github.com/darthsalad/socketeer
```
## Usage

- Create a new `Socketeer` instance:

```go
s, err := socketeer.NewSocketeer(mongodb_uri, db_name, collection_name)
```
- Start the `Socketeer` server for listening to events and dispatching them to connected clients through websockets:

```go
s.Start(document_fields, server_url, server_endpoint)
```
- It starts the websocket server as well as the database listener synchronously. The `document_fields` parameter is a string array that specifies the fields to be returned in the `ChangeStream` cursor. The `server_url` and `server_endpoint` parameters are the url and endpoint of the websocket server respectively. For example: 

```go
s.Start([]string{"name", "email"}, "localhost:8080", "/ws")
```

- The `Socketeer` server can be stopped by calling the `Stop()` method:

```go
s.Stop()
```
### Response Format
- The response format of the data from sockets is of the format `map[string]string` and then are Marshalled into JSON. For example:

```go
var response = make(map[string]string)
```
- The fields are populated with the data from the database
the fields are the ones specified in the `document_fields` parameter 
- For Example: 
```go
fields := []string{"name", "email"}

response["name"] = "John Doe"
response["email"] = "johndoe@example"

data, err := json.Marshal(responseMap)
```
- This byte array is then sent to the client through the websocket connection.
- The client can then parse the data and use it as required. Example of received can be:
  
  ```json
  {
    "name": "John Doe",
    "email": "johndoe@example"
  }
  ```

## Example

For a full example, check out the `example` directory. [See this file.](/example/main.go)

### Client Side Example
  
  ```ts
  // client/side/socket/example.js

  const conn = new WebSocket("ws://localhost:8080/ws");
  conn.onopen = () => {
    console.log("Connected to Socket server!");
  };
  conn.onmessage = (e) => {
    const data = JSON.parse(e.data);
    // do what you want with the data
  };
  ```

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Author

- [**Darthsalad**](https://github.com/darthsalad)
