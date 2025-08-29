            ███╗   ███╗██╗   ██╗ ██████╗███████╗██╗     ██╗ █████╗
            ████╗ ████║╚██╗ ██╔╝██╔════╝██╔════╝██║     ██║██╔══██╗
            ██╔████╔██║ ╚████╔╝ ██║     █████╗  ██║     ██║███████║
            ██║╚██╔╝██║  ╚██╔╝  ██║     ██╔══╝  ██║     ██║██╔══██║
            ██║ ╚═╝ ██║   ██║   ╚██████╗███████╗███████╗██║██║  ██║
            ╚═╝     ╚═╝   ╚═╝    ╚═════╝╚══════╝╚══════╝╚═╝╚═╝  ╚═╝
--------------------------------------------------------------------------------
## Mycelia is a work-in-progress concurrent message broker.

A lightweight, extensible message broker leveraging Go's concurrency model for
speedy message delivery.

You can find the currently supported APIs for your client services to interact
with the broker [here](https://github.com/orgs/SignalWeave/repositories).

## Concepts

Mycelia orchestrates message routing through 4 primary concepts: routes,
channels, transformers, and subscribers.

```
[broker]
  | - [route] main
  | - [route] default
        | - [channel] my_new_channel
              | - [transformer] 127.0.0.1:7010
              | - [transformer] 10.0.0.52:8008
              | - [subscriber] 127.0.0.1:1234
              | - [subscriber] 16.70.18.1:9999
```

### Routes

Routes are like topics in other messaging services. A route can contain multiple
channels for messages to travel through. Data is passed through each channel
sequentially in their creation order.

By default, a route named `"main"` will always be created on broker startup.

### Channels

A channel is like a sub-route which contains transformers and subscribers. Data
is passed through each transformer in creation order before being forwarded to
each subscriber.

### Transformers

If a transformer is added to a channel, it will intercept data going over that
channel before it hits the subscriber. Channels will forward the payload of a
message and wait for a return value based on the `xform-timeout` CLI or PreInit
value. If no return is gotten in that time, the transformer is ignored,
otherwise the return value is forwarded to the next transformer and so on and
then finally to each subscriber.

This is to simplify route orchestration compared to typical routing setups.

In a normal routing model, if service A is sent data but requires additional
info from service B to fulfill its task, a message is sent back to the broker
and then to service B who then returns a query through the broker which then
ends up back at service A.

```
client -> broker -> service A -> broker -> service B -> broker -> service A -> broker -> client
```

Mycelia's transformers semantically simplify this to:

```
client -> broker ->       -> service A -> broker -> client
                  |       ^
                  v       |
                  service B
```

### Subscribers

Subscribers are the address end point for services that subscribe to data passed
over a route + channel.

## CLI

Mycelia supports serveral CLI args:

```
Mycelia runtime options:

  -address string      Bind address (IP or hostname)
  -port int            Bind port (1-65535)
  -verbosity int       0, 1, 2, or 3
  -print-tree          Print router tree at startup
  -xform-timeout dur   Transformer timeout

Examples:
  mycelia -addr 0.0.0.0 -port 8080 -verbosity 2 -print-tree -xform-timeout 45s
```

with verbosity values relating to
```
0 - None
1 - Errors
2 - Warnings + Errors
3 - Errors + Warnings + Actions
```

## Pre Init

Additionally, Mycelia will check the exe's directory for a `PreInit.json` file.
This file can specify any of the CLI args in the `"runtime"` field - these will
overwrite any piped cli args.

Pre-defined routing structures can also be defined within the file for the
broker to use on startup using the `"routes"` field.

Example PreInit.json file:
```json
{
  "runtime": {
    "address": "0.0.0.0",
    "port": 8080,
    "verbosity": 2,
    "print-tree": true,
    "xform-timeout": "45s"
  },
  "routes": [
    {
      "name": "default",
      "channels": [
        {
          "name": "inmem",
          "transformers": [
            { "address": "127.0.0.1:7010" },
            { "address": "10.0.0.52:8008" }
          ],
          "subscribers": [
            { "address": "127.0.0.1:1234" },
            { "address": "16.70.18.1:9999" }
          ]
        }
      ]
    }
  ]
}
```

## Protocol

Mycelia employs a custom protocl as outlined:

Version 1 command decoding.

*Note that this is a messaging protocol, not a file transfer protocol
```
          fields
--------------------------
protocol_ver  |  u8
obj_type      |  u8
obj_cmd       |  u8
uid           |  u32 + len
route         |  u32 + len
--------------------------
channel       |  u32 + len
address       |  u32 + len
--------------------------
payload       |  u32 + len
```
-----------------------------------------------------------------------------
The version 1 protocol looks as follows:

Fixed field sized header
```
+---------+--------+-------------+-------------+
| u32 len | u8 ver | u8 obj_type | u8 cmd_type |
+---------+--------+-------------+-------------+
```
A `uint32` length header field that dictates the number of bytes for the rest of
the message body.

A `uint8` version field, which is used for decoding the message after the base
header. This will likely be obsolete one day but for now, in prototyping, this
helps with managing variations of the header after we deploy it for internal
projects.

A `uint8` object type which corresponds to what concept the client is working
with: `messages`, `transformers`, `subscribers`, `globals`, etc.

And a `uint8` command type which is the behavioral action being done to the
object: `SEND`, `ADD`, `REMOVE`, etc.

This is then followed by a variable field sized sub-header:

Routing Sub-header
```
+-------------+---------------+
| u32 len uid | u32 len route |
+-------------+---------------+
```
The sub header is two variable sized fields with a `uint32` field header and a
field body.

This is then followed by one of the following bodies:

Subscriber + Transformer Body
```
+--------------+--------------+
| u32 len chan | u32 len addr |
+--------------+--------------+
```
Here, again, are two variable sized fields with a `uint32` field header for the
channel and address to route a message through.

Message Body
```
+-----------------+
| u32 len payload |
+-----------------+
```
With the Message Body payload being the data finally forwarded to subscribers.
Thes Message body payload is a single `uint32` field header and the byte array
that is forwarded onwards.

Globals Body
```
+-----------------+
| u32 len payload |
+-----------------+
```
Globals takes a json string to update how the broker works at runtime.
For example a system might change the address or port at runtime when message
traffic increases and then change it again when traffic decreases.
Any of the cli args can be changed by creating a globals update command through
any of the client APIs.
