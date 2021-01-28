package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"os/signal"

	log "github.com/inconshreveable/log15"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	version = "undefined" // updated during release build
)

func main() {
	socket := flag.String("s", "", "Path to UNIX socket")
	broker := flag.String("b", "tcp://localhost:1883", "MQTT broker to use")
	topic := flag.String("t", "my_topic_prefix", "MQTT topic prefix")
	verbose := flag.Bool("v", false, "Print more verbose messages")
	versionFlag := flag.Bool("V", false, "Print version and exit")

	flag.Usage = func() {
		out := flag.CommandLine.Output()
		fmt.Fprintln(out, "soqtt links a unix socket to a MQTT topic.")
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "The socket must send and receive text messages. Each message must be")
		fmt.Fprintln(out, "ended by a newline ('\\n').")
		fmt.Fprintln(out, "Write messages to my_topic_prefix/out to send them to the socket,")
		fmt.Fprintln(out, "subscribe to messages at my_topic_prefix/in to receive messages")
		fmt.Fprintln(out, "from the socket.")
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Options:")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *versionFlag {
		fmt.Printf("soqtt - version %s\n", version)
		os.Exit(0)
	}

	logLevel := log.LvlInfo
	if *verbose {
		logLevel = log.LvlDebug
	}

	if *socket == "" {
		flag.Usage()
		os.Exit(1)
	}

	log.Root().SetHandler(log.LvlFilterHandler(logLevel, log.StdoutHandler))

	inTopic := *topic + "/in"
	outTopic := *topic + "/out"

	log.Info("Settings", "socket", *socket, "broker", *broker, "inTopic", inTopic, "outTopic", outTopic)

	opts := mqtt.NewClientOptions().AddBroker(*broker).SetClientID(fmt.Sprintf("soqtt-%d", rand.Int31()))
	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Error("Failed to connect to broker", "error", token.Error())
		os.Exit(1)
	}

	conn, err := net.Dial("unix", *socket)
	if err != nil {
		log.Error("Failed to open socket", "error", err)
		os.Exit(1)
	}

	go socket2mqtt(conn, client, inTopic)
	go mqtt2socket(conn, client, outTopic)
	waitForCtrlC()
}

func publish(client mqtt.Client, topic string, message string) error {
	log.Debug("MQTT <- Socket", "message", message)

	token := client.Publish(topic, 0, false, message)
	token.Wait()
	return token.Error()
}

func socket2mqtt(conn net.Conn, client mqtt.Client, topic string) {
	reader := bufio.NewReader(conn)

	for {
		var buffer bytes.Buffer
		for {
			ba, isPrefix, err := reader.ReadLine()
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Error("Failure while reading from socket", "error", err)
				break
			}
			buffer.Write(ba)
			if !isPrefix {
				break
			}
		}

		err := publish(client, topic, buffer.String())
		if err != nil {
			log.Error("Failure while publishing to MQTT", "error", err)
		}
	}
}

func mqtt2socket(conn net.Conn, client mqtt.Client, topic string) {
	handler := func(c mqtt.Client, mqttMessage mqtt.Message) {
		payload := mqttMessage.Payload()
		log.Debug("MQTT -> Socket", "message", string(payload))
		_, err := conn.Write(payload)
		if err != nil {
			log.Error("Failed to write mqtt message to socket", "error", err)
		}

		if len(payload) == 0 || payload[len(payload)-1] != '\n' {
			_, err = conn.Write([]byte{'\n'})
			if err != nil {
				log.Error("Failed to write newline to socket", "error", err)
			}
		}
	}

	if token := client.Subscribe(topic, 0, handler); token.Wait() && token.Error() != nil {
		log.Error("Failed to subscribe to MQTT topic", "error", token.Error())
		os.Exit(1)
	}
}

func waitForCtrlC() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	<-signals
}
