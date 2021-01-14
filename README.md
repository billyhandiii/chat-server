# chat-server

This is a simple program that demonstrates the following features:

* configuration with a json file passed as the only argument to the program
* tcp server for plaintext telnet connections
* optional http server for posting messages
* broadcasting of sent messages to all connected clients
* message senders can choose a name at the start of a tcp session or as part of the http message post
* each message is assigned a timestamp as it is observed by the internal message broker
* clients should observe messages in the same order they send them
* messages can be logged to STDOUT or a file

## Running

If you don't already have a Go development environment follow the instructions [here](https://golang.org/doc/) to get
acquainted.  After that you can run the program simply with `go run .` in the project directory.  You can also build the
project with `go build .` which will result in a binary called "chat-server".

Both methods support passing a command line argument that is the path to a configuration file.

Once the server is running you can send messages by connecting with telnet, typing your name followed by `enter` and
then typing messages as you see fit.  You will observe all messages sent by other telnet or http users.  Use your telnet
client's close feature to quit.

If you want to send messages by http make sure you configure the http to start up as described below.  Then you can POST
messages with the following Curl command or the equivalent (adjust for your chosen IP/Port combo):

```bash
curl http://localhost:30406/ -H 'Content-type: application/json' -d '{"sender": "tester", "text": "hello!"}'
```

## Configuration

The server runs with defaults just fine but if you want to control the behavior write the following to a file:

```json
{
  "ip": "127.0.0.1",                     // the ip address to bind the telnet server to; defaults to all
  "port": 30405,                         // the port to listen on for the telnet server; defaults to random
  "http": {
    "ip": "127.0.0.1",                   // the ip address to bind the telnet server to; defaults to all
    "port": 30406                        // the port to listen on for the telnet server; defaults to random
  },
  "log_file_path": "chat_server.log"     // the file to write all messages to; defaults to STDOUT; errors go to STDERR no matter what
}
```

You can then run with those options by passing the path of the config file as the only argument when you run the server.

```bash
go run . config.json
```

or

```bash
go build .
./chat-server config.json
```

## Testing

The tests just use the standard Go test framework.  They can be run with `go test`.  Some of the tests require listening
on ports 30405 and 30406.

## Future goals

This was an intentionally short exercise.  Were I to continue I would focus on:

* indexing messages to enable search queries
* better handling of client sessions (timeouts, handling tailoring error handling to the specific error)
* more test scenarios
* validation on http messages
* authentication so users can reserve names
* a specific client implementation
* keep message history for late joiners
* system messages
* shutdown timer
* channels
* slash commands like /quit, /join (channels), /msg (direct msg)
* block/moderation features