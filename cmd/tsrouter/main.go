package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type keyResp struct {
	Key string `json:"key"`
}

type devicesResp struct {
	Devices []struct {
		NodeKey string `json:"nodeKey"`
		ID      string `json:"id"`
	} `json:"devices"`
}

func main() {
	log.SetPrefix("tailscale-router ")
	log.Println("setting up")
	tailnet := "-"
	api_key := os.Getenv("TAILSCALE_API_TOKEN")

	log.Println("tailnet name", tailnet)
	log.Println("api key", api_key)

	jsonData := []byte(`{
		"capabilities": {
			"devices": {
				"create": {
					"reusable": true,
					"ephemeral": true
				}
			}
		}
	}`)

	log.Println("creating auth key")
	request, err := http.NewRequest(
		"POST",
		fmt.Sprintf("https://api.tailscale.com/api/v2/tailnet/%s/keys", tailnet),
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		log.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.SetBasicAuth(api_key, "")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Fatalln("ERROR: Create key", err)
	}
	defer response.Body.Close()

	var out keyResp
	if err := json.NewDecoder(response.Body).Decode(&out); err != nil {
		log.Fatalln("ERROR: Decode key resp", err)
	}
	key := out.Key
	log.Println("key is", key)

	log.Println("grepping /etc/hosts to get fly-local-6pn")
	output, err := exec.Command("grep", "fly-local-6pn", "/etc/hosts").Output()
	if err != nil {
		log.Fatalln("ERROR: get subnet", err)
	}

	log.Println("calculating subnet")
	subnet := strings.Join(
		strings.Split(strings.TrimSuffix(string(output), "\n"), ":")[0:3],
		":",
	) + "::/48"

	tailscale_binary_path := "/app/tailscale"

	if runtime.GOOS == "darwin" {
		output, err := exec.Command("bash", "-c", "ps -xo comm | grep MacOS/Tailscale").Output()
		if err != nil {
			log.Fatal(err)
		}
		tailscale_binary_path = strings.TrimSuffix(string(output), "\n")
	}

	allocID := os.Getenv("FLY_ALLOC_ID")
	if len(allocID) > 5 {
		allocID = allocID[0:5]
	}
	hostname := fmt.Sprintf("%s-%s-%s", os.Getenv("FLY_APP_NAME"), os.Getenv("FLY_REGION"), allocID)

	log.Println("running tailscale up")
	upcmd := exec.Command(
		"bash",
		"-c",
		fmt.Sprintf(
			"%s up --authkey=%s --hostname=%s --advertise-routes=%s --advertise-exit-node",
			tailscale_binary_path,
			key,
			hostname,
			subnet,
		),
	)
	if err := upcmd.Run(); err != nil {
		log.Fatal(err)
	}

	log.Println("getting PublicKey from tailscale status")
	output, err = exec.Command("bash", "-c", fmt.Sprintf("%s status --json | jq -r .Self.PublicKey", tailscale_binary_path)).
		Output()
	if err != nil {
		log.Fatal(err)
	}

	nodeKey := strings.TrimSuffix(string(output), "\n")

	log.Println("getting all devices")
	request, err = http.NewRequest(
		"GET",
		fmt.Sprintf("https://api.tailscale.com/api/v2/tailnet/%s/devices", tailnet),
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.SetBasicAuth(api_key, "")

	client = &http.Client{}
	response, err = client.Do(request)
	if err != nil {
		log.Fatalln("ERROR: Read devices", err)
	}
	defer response.Body.Close()

	var devicesOut devicesResp
	if err := json.NewDecoder(response.Body).Decode(&devicesOut); err != nil {
		log.Fatalln("ERROR: Decode key resp", err)
	}

	selfID := ""

	log.Println("finding our ID")
	for _, v := range devicesOut.Devices {
		if v.NodeKey == nodeKey {
			selfID = v.ID
			break
		}
	}

	jsonData = []byte(fmt.Sprintf(`{
		"routes": ["%s"]
	}`, subnet))

	log.Println("configuring routes")
	request, err = http.NewRequest(
		"POST",
		fmt.Sprintf("https://api.tailscale.com/api/v2/device/%s/routes", selfID),
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		log.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.SetBasicAuth(api_key, "")

	client = &http.Client{}
	response, err = client.Do(request)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	log.Println("fully configured")

	os.Exit(0)
}
