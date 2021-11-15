package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	rocketpb "github.com/bachelor-thesis-hown3d/chat-api-server/proto/rocket/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	name      string = "test-rocket"
	namespace string = "default"
	user      string = "TestUser"
	email     string = "testUser@foo.bar"
)

var (
	port = flag.Int("port", 10000, "Port of the api server")
	host = flag.String("host", "", "Hostname of the api server")
)

func main() {
	flag.Parse()
	if *host == "" {
		fmt.Println("Host must be set!")
		os.Exit(1)
	}
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, fmt.Sprintf("%v:%v", *host, *port), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to dial localhost:10000 : %v", err)
	}
	client := rocketpb.NewRocketServiceClient(conn)

	// client.Delete(ctx, &rocketpb.DeleteRequest{Name: name, Namespace: namespace})

	// _, err = client.Create(ctx, &rocketpb.CreateRequest{
	// 	User:         user,
	// 	Name:         name,
	// 	Namespace:    namespace,
	// 	DatabaseSize: 10,
	// 	Replicas:     1,
	// 	Email:        email,
	// })

	// if err != nil {
	// 	log.Fatalf("Error creating new rocket: %v", err)
	// }

	versions, err := client.AvailableVersions(ctx, &rocketpb.AvailableVersionsRequest{Image: rocketpb.AvailableVersionsRequest_ROCKETCHAT})
	if err != nil {
		fmt.Println(fmt.Errorf("Can't get available version of rocketchat: %w", err))
	}
	for i := 0; i < 5; i++ {
		fmt.Printf("Rocket Tag: %v\n", versions.Tags[i])
	}
	versions, err = client.AvailableVersions(ctx, &rocketpb.AvailableVersionsRequest{Image: rocketpb.AvailableVersionsRequest_MONGODB})
	if err != nil {
		fmt.Println(fmt.Errorf("Can't get available version of mongodb: %w", err))
	}
	for i := 0; i < 5; i++ {
		fmt.Printf("Mongodb Tag: %v\n", versions.Tags[i])
	}

	allRockets, err := client.GetAll(ctx, &rocketpb.GetAllRequest{})
	if err != nil {
		log.Fatalf("Can't get all rockets: %v", err)
	}
	for index, rocket := range allRockets.Rockets {
		fmt.Printf("%v: %v - %v\n", index, rocket.Name, rocket.Namespace)
	}

	newRocket, err := client.Get(ctx, &rocketpb.GetRequest{
		Name:      name,
		Namespace: namespace,
	})
	if err != nil {
		log.Fatalf("Can't get rocket: %v", err)
	}

	// watch the rocket to get ready
	statusClient, err := client.Status(ctx, &rocketpb.StatusRequest{Name: newRocket.Name, Namespace: newRocket.Namespace})
	if err != nil {
		log.Fatalf("Error watching new rocket: %v", err)
	}
	var ready bool
	for ready == false {
		msg, err := statusClient.Recv()
		if status.Code(err) == codes.Canceled {
			log.Println("Context was canceled")
			break
		}
		if err == io.EOF {
			continue
		}
		if err != nil {
			log.Fatalf("Error: %v", err.Error())
		}
		fmt.Printf("StatusMessage: %v - Ready: %v\n", msg.Status, msg.Ready)
		ready = msg.Ready
	}

	logsClient, err := client.Logs(ctx, &rocketpb.LogsRequest{Name: newRocket.Name, Namespace: newRocket.Namespace})
	for {
		// blocking
		msg, err := logsClient.Recv()
		if status.Code(err) == codes.Canceled {
			log.Println("Context was canceled")
			os.Exit(0)
		}
		if err != nil {
			log.Fatalf("Error: %v", err.Error())
		}
		fmt.Printf("Pod: %v - Msg: %v\n", msg.Pod, msg.Message)
	}
}
