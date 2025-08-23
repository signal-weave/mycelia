# Mycelia

A lightweight, extensible message broker.

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
