package k8sutil

import (
	"bufio"
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
			reader := bufio.NewReader(stream)
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
					line, isPrefix, err := reader.ReadLine()
					// only prefix, another line is coming
					for isPrefix == true {
						var suffix []byte
						suffix, isPrefix, err = reader.ReadLine()
						line = append(line, suffix...)
					}
					if err == io.EOF {
						logger.Debugw("Recieved end of file, sleeping for 5 seconds", "pod", podName)
						time.Sleep(5 * time.Second)
					}
					if err != nil {
						return err
					}
					err = grpc.Send(&rocketpb.LogsResponse{Message: string(line), Pod: podName})
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
