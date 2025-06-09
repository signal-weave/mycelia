# Mycelia

A simple hobby messaging platform.

Currently producers can send a message in the form of
```json
{
    "type": "send_message",
    "data": {
        "route": "route_name",
        "body": { "payload": "data" }
    }
}
```
which will be processed and sent to clients. The `route` field is used to
determine which route the message is sent through, and will be forwarded on to
all subscribers of channels in the route.

Additionally consumers can send a subscription message in the form of
```json
{
    "type": "add_subscribe",
    "data": {
        "route": "route_name",
        "channel": "channel_name",
        "address": "127.0.0.1:5000"  // Where messages will be forwarded to.
    }
}
```
which will add them as a subscriber to messages coming through the specified
`endopint` which is the name of the route. This is currently temporary and will
change in the future as more channel types and more complex route handling is
added.

By default there is only one route, `main`, with one channel, `main`. Channels
be added through the following message:
```json
{
    "type": "add_channel",
    "data": {
        "route": "main",
        "name": "new_channel"
    }
}
```
Lastly, routes can be added using:
```json
{
    "type": "register_route",
    "data": {
        "name": "route_name"
    }
}
```
