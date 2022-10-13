package main

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	fmt.Println("tailscale-router: setting up")
	tailnet := "-"
	api_key := os.Getenv("TAILSCALE_API_TOKEN")

	fmt.Println("tailscale-router: tailnet name", tailnet)
	fmt.Println("tailscale-router: api key", api_key)

	var jsonData = []byte(`{
		"capabilities": {
			"devices": {
				"create": {
					"reusable": true,
					"ephemeral": true
				}
			}
		}
	}`)

	fmt.Println("tailscale-router: creating auth key")
	request, err := http.NewRequest("POST", fmt.Sprintf("https://api.tailscale.com/api/v2/tailnet/%s/keys", tailnet), bytes.NewBuffer(jsonData))
	if err != nil {
		panic(err)
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.SetBasicAuth(api_key, "")

	client := &http.Client{}
	response, error := client.Do(request)
	if error != nil {
		fmt.Println("ERROR: Create key")
		panic(error)
	}
	defer response.Body.Close()

	var out keyResp
	err = json.NewDecoder(response.Body).Decode(&out)
	if err != nil {
		fmt.Println("ERROR: Decode key resp")
		panic(error)
	}
	key := out.Key
	fmt.Println("tailscale-router: key is", key)

	fmt.Println("tailscale-router: grepping /etc/hosts to get fly-local-6pn")
	output, err := exec.Command("grep", "fly-local-6pn", "/etc/hosts").Output()
	if err != nil {
		fmt.Println("ERROR: get subnet")
		panic(err)
	}

	fmt.Println("tailscale-router: calculating subnet")
	subnet := strings.Join(strings.Split(strings.TrimSuffix(string(output), "\n"), ":")[0:3], ":") + "::/48"

	tailscale_binary_path := "/app/tailscale"

	if runtime.GOOS == "darwin" {
		output, err := exec.Command("bash", "-c", "ps -xo comm | grep MacOS/Tailscale").Output()
		if err != nil {
			panic(err)
		}
		tailscale_binary_path = strings.TrimSuffix(string(output), "\n")
	}

	fmt.Println("tailscale-router: running tailscale up")
	upcmd := exec.Command("bash", "-c", fmt.Sprintf("%s up --authkey=%s --advertise-routes=%s", tailscale_binary_path, key, subnet))
	err = upcmd.Run()
	if err != nil {
		panic(err)
	}

	fmt.Println("tailscale-router: getting PublicKey from tailscale status")
	output, err = exec.Command("bash", "-c", fmt.Sprintf("%s status --json | jq -r .Self.PublicKey", tailscale_binary_path)).Output()
	if err != nil {
		panic(err)
	}

	nodeKey := strings.TrimSuffix(string(output), "\n")

	fmt.Println("tailscale-router: getting all devices")
	request, err = http.NewRequest("GET", fmt.Sprintf("https://api.tailscale.com/api/v2/tailnet/%s/devices", tailnet), nil)
	if err != nil {
		panic(err)
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.SetBasicAuth(api_key, "")

	client = &http.Client{}
	response, error = client.Do(request)
	if error != nil {
		fmt.Println("ERROR: read devices")
		panic(error)
	}
	defer response.Body.Close()

	var devicesOut devicesResp
	err = json.NewDecoder(response.Body).Decode(&devicesOut)
	if err != nil {
		fmt.Println("ERROR: Decode key resp")
		panic(error)
	}

	selfID := ""

	fmt.Println("tailscale-router: finding our ID")
	for _, v := range devicesOut.Devices {
		if v.NodeKey == nodeKey {
			selfID = v.ID
			break
		}
	}

	jsonData = []byte(fmt.Sprintf(`{
		"routes": ["%s"]
	}`, subnet))

	fmt.Println("tailscale-router: configuring routes")
	request, err = http.NewRequest("POST", fmt.Sprintf("https://api.tailscale.com/api/v2/device/%s/routes", selfID), bytes.NewBuffer(jsonData))
	if err != nil {
		panic(err)
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.SetBasicAuth(api_key, "")

	client = &http.Client{}
	response, error = client.Do(request)
	if error != nil {
		panic(error)
	}
	defer response.Body.Close()

	fmt.Println("tailscale-router: fully configured")
	os.Exit(0)
}
