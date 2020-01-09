package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	networktypes "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
)

// Backend use: curl localhost:80 -d '"user"'
func Backend(writer http.ResponseWriter, request *http.Request) {

	// Read body of curl request into bytes format
	body, err := ioutil.ReadAll(request.Body)
	defer request.Body.Close()
	if err != nil {
		http.Error(writer, err.Error(), 500)
		return
	}

	// Unmarshal body of curl request into interface
	var msg interface{} // Message
	err = json.Unmarshal(body, &msg)
	fmt.Println(msg)
	if err != nil {
		http.Error(writer, err.Error(), 500)
		return
	}

	// output
	output, err := json.Marshal(msg)
	if err != nil {
		http.Error(writer, err.Error(), 500)
		return
	}
	writer.Header().Set("content-type", "application/json")
	writer.Write(output)

	// variables
	baseDir := "/opt/vcde/"
	requestUser := fmt.Sprintf("%v", msg)
	timeZone := "TZ='Europe/Amsterdam'"
	netWork := "vcd_frontend"
	vcdImage := "rayvtoll/containerdesktop:latest"
	ctx := context.Background()

	// docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	// docker run command constructor with docker syntax in comments
	resp, err := cli.ContainerCreate(
		ctx,
		// image, --hostname, -e, cmd, -t
		&container.Config{Image: vcdImage, Hostname: "vcd-" + requestUser, Env: []string{"USER=" + requestUser, timeZone}, Tty: false},
		// -v, --rm,
		&container.HostConfig{
			Binds: []string{
				baseDir + requestUser + ":/home/" + requestUser,
				baseDir + "Public:/home/" + requestUser + "/Public",
			},
			AutoRemove: true,
		},
		// --network
		&networktypes.NetworkingConfig{
			EndpointsConfig: map[string]*networktypes.EndpointSettings{
				netWork: {},
			},
		},
		// --name
		"vcd-"+requestUser,
	)
	if err != nil {
		panic(err)
	}

	// start the desktopcontainer
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}
}

func main() {
	http.HandleFunc("/", Backend)
	fmt.Println("Starting server on port 80")
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		panic(err)
	}
}
