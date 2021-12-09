package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/bachelor-thesis-hown3d/chat-api-server-client/oauth"
	rocketpb "github.com/bachelor-thesis-hown3d/chat-api-server/proto/rocket/v1"
	tenantpb "github.com/bachelor-thesis-hown3d/chat-api-server/proto/tenant/v1"
	"github.com/pkg/browser"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	name      string = "test-rocket"
	namespace string = "testuser"
	user      string = "TestUser"
	email     string = "testUser@foo.bar"
)

var (
	port        = flag.Int("port", 10000, "Port of the api server")
	host        = flag.String("host", "", "Hostname of the api server")
	redirectURI = flag.String("redirectUrl", "http://localhost:7070", "address for the oauth server to listen on")
	issuerURI   = flag.String("issuerUrl", "https://keycloak:8443/auth/realms/kubernetes", "address for the oauth server to listen on")
)

func startOAuthAndWaitForToken(clientID, clientSecret string) string {
	// parse the redirect URL for the port number
	redirectURL, err := url.Parse(*redirectURI)
	if err != nil {
		log.Fatal(err)
	}

	// parse the redirect URL for the port number
	issuerURL, err := url.Parse(*issuerURI)
	if err != nil {
		log.Fatal(err)
	}

	// since we use a bad ssl certificate, embed a Insecure HTTP Client for oauth to use
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	ctx := context.Background()
	ctx = context.WithValue(ctx, oauth2.HTTPClient, http.DefaultClient)

	serv, lis, err := oauth.NewServer(ctx, clientID, clientSecret, redirectURL, issuerURL)
	if err != nil {
		log.Fatalf("Can't create oauth Server: %v", err)
	}

	ch := make(chan error, 1)
	// start the oauth webserver and wait for the token
	go oauth.StartWebServer(serv, lis)
	redirectURL.Path = "/auth"
	browser.OpenURL(redirectURL.String())

	go func() {
		select {
		case err := <-ch:
			log.Fatal(err)
		}
	}()
	for oauth.IDToken == nil {
		fmt.Println("OAuth token not retrieved, sleeping...")
		time.Sleep(3 * time.Second)
	}
	return *oauth.IDToken
}

func main() {
	flag.Parse()
	if *host == "" {
		fmt.Println("Host must be set!")
		os.Exit(1)
	}

	clientSecret := os.Getenv("OAUTH2_CLIENT_SECRET")
	clientID := os.Getenv("OAUTH2_CLIENT_ID")

	if clientID == "" {
		log.Fatal("OAUTH2_CLIENT_ID Environment Variable must be set!")
	}
	if clientSecret == "" {
		log.Fatal("OAUTH2_CLIENT_SECRET Environment Variable must be set!")
	}

	token := startOAuthAndWaitForToken(clientID, clientSecret)

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, fmt.Sprintf("%v:%v", *host, *port), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to dial %v:%v: %v", *host, *port, err)
	}

	rocketClient := rocketpb.NewRocketServiceClient(conn)
	tenantClient := tenantpb.NewTenantServiceClient(conn)

	md := metadata.Pairs("authorization", "bearer "+token)
	ctx = metadata.NewOutgoingContext(ctx, md)

	defaultFlow(ctx, rocketClient, tenantClient)
}

func defaultFlow(ctx context.Context, rocketClient rocketpb.RocketServiceClient, tenantClient tenantpb.TenantServiceClient) {

	defer rocketClient.Delete(ctx, &rocketpb.DeleteRequest{Name: name, Namespace: namespace})

	versions, err := rocketClient.AvailableVersions(ctx, &rocketpb.AvailableVersionsRequest{Image: rocketpb.AvailableVersionsRequest_IMAGE_ROCKETCHAT})
	if err != nil {
		log.Fatal(fmt.Errorf("Can't get available version of rocketchat: %w", err))
	}
	for i := 0; i < 5; i++ {
		fmt.Printf("Rocket Tag: %v\n", versions.Tags[i])
	}
	versions, err = rocketClient.AvailableVersions(ctx, &rocketpb.AvailableVersionsRequest{Image: rocketpb.AvailableVersionsRequest_IMAGE_MONGODB})
	if err != nil {
		log.Fatal(fmt.Errorf("Can't get available version of mongodb: %w", err))
	}
	for i := 0; i < 5; i++ {
		fmt.Printf("Mongodb Tag: %v\n", versions.Tags[i])
	}
	//

	_, err = tenantClient.Register(ctx, &tenantpb.RegisterRequest{Size: tenantpb.RegisterRequest_SIZE_SMALL})
	if err != nil {
		log.Fatal(err)
	}

	allRockets, err := rocketClient.GetAll(ctx, &rocketpb.GetAllRequest{
		Namespace: namespace,
	})
	if err != nil {
		log.Fatalf("Can't get all rockets: %v", err)
	}
	for index, rocket := range allRockets.Rockets {
		fmt.Printf("%v: %v - %v\n", index, rocket.Name, rocket.Namespace)
	}

	_, err = rocketClient.Create(ctx, &rocketpb.CreateRequest{
		Name:         name,
		Namespace:    namespace,
		DatabaseSize: 10,
		Replicas:     1,
		Email:        email,
		Host:         "test.chat-cluster.com",
	})

	if err != nil {
		log.Fatalf("Error creating new rocket: %v", err)
	}

	newRocket, err := rocketClient.Get(ctx, &rocketpb.GetRequest{
		Name:      name,
		Namespace: namespace,
	})
	if err != nil {
		log.Fatalf("Can't get rocket: %v", err)
	}
	//
	// watch the rocket to get ready
	statusClient, err := rocketClient.Status(ctx, &rocketpb.StatusRequest{Name: newRocket.Name, Namespace: newRocket.Namespace})
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
	//
	// logsClient, err := rocketClient.Logs(ctx, &rocketpb.LogsRequest{Name: newRocket.Name, Namespace: newRocket.Namespace})
	// for {
	// 	// blocking
	// 	msg, err := logsClient.Recv()
	// 	if status.Code(err) == codes.Canceled {
	// 		log.Println("Context was canceled")
	// 		os.Exit(0)
	// 	}
	// 	if err != nil {
	// 		log.Fatalf("Error: %v", err.Error())
	// 	}
	// 	fmt.Printf("Pod: %v - Msg: %v\n", msg.Pod, msg.Message)
	// }
}
