# Soqtt

soqtt links a unix socket to a MQTT topic.

The socket must send and receive text messages. Each message must be ended by a newline (`\n`).
Write messages to my_topic_prefix/out to send them to the socket, subscribe to messages at my_topic_prefix/in to receive messages from the socket.

The original use case was to bridge [signald](https://gitlab.com/signald/signald) to [Node RED](https://nodered.org/).

## Usage

```
soqtt -b tcp://localhost:1883 -t my_topic_prefix -s /var/lib/signald/signald.sock
```

## Development

```
# Build and run
make
./soqtt

# Build release for Raspberry Pi, Beagle Bone
make release GOOS=linux GOARCH=arm GOARM=7

# Build and install on Raspberry Pi, Beagle Bone
make install GOOS=linux GOARCH=arm GOARM=7 TARGET_HOST=debian@beaglebone
```
