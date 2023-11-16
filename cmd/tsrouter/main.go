package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
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
	log.SetPrefix("tsrouter ")
	log.Println("starting")

	apiKey := os.Getenv("TAILSCALE_API_TOKEN")
	client := &http.Client{}

	log.Println("creating auth key")
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
	request := mustJSONRequest(
		apiKey,
		http.MethodPost,
		"https://api.tailscale.com/api/v2/tailnet/-/keys",
		bytes.NewBuffer(jsonData),
	)
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

	const tailscale_binary_path = "/app/tailscale"

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
		log.Fatalln("ERROR: tailscale up", err)
	}

	log.Println("getting PublicKey from tailscale status")
	output, err = exec.Command("bash", "-c", fmt.Sprintf("%s status --json | jq -r .Self.PublicKey", tailscale_binary_path)).
		Output()
	if err != nil {
		log.Fatalln("ERROR: get PublicKey", err)
	}

	nodeKey := strings.TrimSuffix(string(output), "\n")

	log.Println("getting all devices")
	request = mustJSONRequest(
		apiKey,
		http.MethodGet,
		"https://api.tailscale.com/api/v2/tailnet/-/devices",
		nil,
	)
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
	request = mustJSONRequest(
		apiKey,
		http.MethodPost,
		fmt.Sprintf("https://api.tailscale.com/api/v2/device/%s/routes", selfID),
		bytes.NewBuffer(jsonData),
	)

	response, err = client.Do(request)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	log.Println("fully configured")
	os.Exit(0)
}

func mustJSONRequest(apiKey, method, url string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.SetBasicAuth(apiKey, "")
	return req
}
