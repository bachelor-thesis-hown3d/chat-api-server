package k8sutil

import (
	"context"
	"io"
	"time"

	rocketpb "github.com/bachelor-thesis-hown3d/chat-api-server/proto/rocket/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// GetPodLogs retrieves the Logs of a Pod and writes it to a grpc stream
func GetPodLogs(ctx context.Context, clientset *kubernetes.Clientset, namespace string, podName string, follow bool, grpcStream rocketpb.RocketService_LogsServer) error {
	podLogOptions := corev1.PodLogOptions{
		Follow: follow,
	}

	podLogRequest := clientset.CoreV1().
		Pods(namespace).
		GetLogs(podName, &podLogOptions)
	stream, err := podLogRequest.Stream(ctx)
	if err != nil {
		return err
	}
	defer stream.Close()

	for {
		buf := make([]byte, 2000)
		numBytes, err := stream.Read(buf)
		if numBytes == 0 {
			time.Sleep(1 * time.Second)
		}
		if err == io.EOF {
			time.Sleep(10 * time.Second)
		}
		if err != nil {
			return err
		}
		message := string(buf[:numBytes])
		err = grpcStream.Send(&rocketpb.LogsResponse{Message: message, Pod: podName})
		if err != nil {
			return err
		}
	}
}
