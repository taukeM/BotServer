package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

type VideoPlaybackTask struct {
	VideoID       string  `json:"video_id"`
	PlaybackSpeed float32 `json:"playback_speed"`
	// etc...
}

type PlayerStatus struct {
	Position  int  `json:"position"`
	Buffering bool `json:"buffering"`
	// etc...
}

type Script struct {
	ScriptID string `json:"script_id" yaml:"script_id"`
	// etc...
}

type Scenario struct {
	Scripts []Script `json:"scripts" yaml:"scripts"`
}

func main() {
	nc, err := connectNATS()
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	playerInputSub, err := subscribePlayerInput(nc)
	if err != nil {
		log.Fatal(err)
	}

	uiOutputSub, err := subscribeUIOutput(nc)
	if err != nil {
		log.Fatal(err)
	}

	go processPlayerInput(playerInputSub, nc)
	go processScriptChanges(uiOutputSub, nc)

	select {}
}

func connectNATS() (*nats.Conn, error) {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return nil, err
	}
	return nc, nil
}

func subscribePlayerInput(nc *nats.Conn) (*nats.Subscription, error) {
	sub, err := nc.SubscribeSync("player.input")
	if err != nil {
		return nil, err
	}
	return sub, nil
}

func subscribeUIOutput(nc *nats.Conn) (*nats.Subscription, error) {
	sub, err := nc.SubscribeSync("ui.output")
	if err != nil {
		return nil, err
	}
	return sub, nil
}

func processPlayerInput(sub *nats.Subscription, nc *nats.Conn) {
	for {
		msg, err := sub.NextMsg(time.Second)
		if err != nil {
			log.Println("Error receiving player input message:", err)
			continue
		}

		var task VideoPlaybackTask
		err = json.Unmarshal(msg.Data, &task)
		if err != nil {
			log.Println("Error parsing video playback task:", err)
			continue
		}

		log.Println("Playing video:", task.VideoID)

		// Perform the video playback logic based on the task
		// ...

		status := PlayerStatus{
			Position:  60,
			Buffering: false,
		}

		statusJSON, err := json.Marshal(status)
		if err != nil {
			log.Println("Error marshaling player status:", err)
			continue
		}

		err = nc.Publish("player.output", statusJSON)
		if err != nil {
			log.Println("Error publishing player status update:", err)
		}
	}
}

func processScriptChanges(sub *nats.Subscription, nc *nats.Conn) {
	for {
		msg, err := sub.NextMsg(time.Second)
		if err != nil {
			log.Println("Error receiving UI output message:", err)
			continue
		}

		var scriptChange Script
		err = json.Unmarshal(msg.Data, &scriptChange)
		if err != nil {
			log.Println("Error parsing script change:", err)
			continue
		}

		log.Println("Script changed. New script ID:", scriptChange.ScriptID)

		err = updateScriptSettings(scriptChange)
		if err != nil {
			log.Println("Error updating script settings:", err)
			continue
		}

		sendScriptChangeAcknowledgement(nc)
	}
}

func updateScriptSettings(scriptChange Script) error {
	// Perform necessary operations to modify the YAML files
	var scenario Scenario
	source, err := ioutil.ReadFile("scenarios.yaml")
	if err != nil {
		return fmt.Errorf("error reading scenario: %v", err)
	}

	err = yaml.Unmarshal(source, &scenario)
	if err != nil {
		return fmt.Errorf("error parsing scenario: %v", err)
	}

	//change scenario
	for _, script := range scenario.Scripts {
		if script.ScriptID == scriptChange.ScriptID {
			// update script
		}
	}

	y, err := yaml.Marshal(scenario)
	if err != nil {
		return fmt.Errorf("error marshaling scenario change: %v", err)
	}
	err = ioutil.WriteFile("scenarios.yaml", y, 0644)
	if err != nil {
		return fmt.Errorf("error writing scenario: %v", err)
	}

	return nil
}

func sendScriptChangeAcknowledgement(nc *nats.Conn) {
	acknowledgement := map[string]interface{}{
		"message": "Script change acknowledged",
	}

	acknowledgementJSON, err := json.Marshal(acknowledgement)
	if err != nil {
		log.Println("Error marshaling acknowledgement message:", err)
		return
	}

	err = nc.Publish("ui.input", acknowledgementJSON)
	if err != nil {
		log.Println("Error publishing acknowledgement message:", err)
	}
}
