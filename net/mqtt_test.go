package net

import (
	"github.com/eclipse/paho.mqtt.golang"
	"sync"
	"testing"
)

func TestMqttSocket(t *testing.T) {
	wg := sync.WaitGroup{}

	topic := "/mq/lachesis/node"

	client := NewMqttSocket("tcp://iot.eclipse.org:1883", func(client mqtt.Client, message mqtt.Message) {
		t.Log("Message received : ", string(message.Payload()), " on topic ", message.Topic())
		wg.Done()
	})
	if err := client.Connect(); err != nil {
		t.Error("Connection failed : ", err)
		return
	}
	if err := client.Listen(topic); err != nil {
		t.Errorf("Failed to listen on topic %s, %v\n", topic, err)
		return
	}
	wg.Add(1)
	if err := client.FireEvent("Test Message", topic); err != nil {
		t.Errorf("Failed to send message %v\n", err)
		return
	}
	t.Log("Msg has been published")
	wg.Wait()
}
