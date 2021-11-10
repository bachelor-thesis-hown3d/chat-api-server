package k8sutil

import (
	"context"
	"io"
	"time"

	rocketpb "github.com/bachelor-thesis-hown3d/chat-api-server/proto/rocket/v1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// GetPodLogs retrieves the Logs of a Pod and writes it to a grpc stream
func GetPodLogs(ctx context.Context, clientset kubernetes.Interface, podNames []string, namespace string, grpc rocketpb.RocketService_LogsServer) error {
	logger := zap.S()
	podLogOptions := corev1.PodLogOptions{
		Follow: true,
	}

	for _, name := range podNames {
		go func(podName string) error {
			podLogRequest := clientset.CoreV1().
				Pods(namespace).
				GetLogs(podName, &podLogOptions)
			stream, err := podLogRequest.Stream(ctx)
			if err != nil {
				return err
			}
			defer stream.Close()
		podLoop:
			for {
				select {
				// blocks, if the channel is open, so the default case will be used
				// closing the channel will break the loop
				case <-ctx.Done():
					logger.Debugw("Stream closed, exiting logs", "pod", podName)
					break podLoop
				default:
					buf := make([]byte, 2000)
					numBytes, err := stream.Read(buf)
					if numBytes == 0 {
						logger.Debugw("No bytes left in buffer", "pod", podName)
						time.Sleep(1 * time.Second)
					}
					if err == io.EOF {
						logger.Debugw("Recieved end of file", "pod", podName)
						time.Sleep(10 * time.Second)
					}
					if err != nil {
						return err
					}
					message := string(buf[:numBytes])
					err = grpc.Send(&rocketpb.LogsResponse{Message: message, Pod: podName})
					if err != nil {
						return err
					}
				}
			}
			// return when breaking the loop on reading from a closed channel
			return nil
		}(name)

	}
	return nil
}
