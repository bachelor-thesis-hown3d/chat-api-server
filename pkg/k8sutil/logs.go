package k8sutil

import (
	"bufio"
	"fmt"

	rocketpb "github.com/bachelor-thesis-hown3d/chat-api-server/proto/rocket/v1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetPodLogs retrieves the Logs of a Pod and writes it to a grpc stream
func GetPodLogs(clientset kubernetes.Interface, podNames []string, namespace string, stream rocketpb.RocketService_LogsServer, errChan chan error) {
	ctx := stream.Context()
	logger := zap.S()
	podLogOptions := corev1.PodLogOptions{
		Follow: true,
	}
	podClient := clientset.CoreV1().Pods(namespace)
	for _, name := range podNames {
		logger.Debugf("Starting log collection for pod: %v", name)
		go func(podName string) {
			if _, err := podClient.Get(ctx, podName, v1.GetOptions{}); err != nil {
				errChan <- fmt.Errorf("Error getting pod: %w", err)
			}
			req := podClient.GetLogs(podName, &podLogOptions)
			logStream, err := req.Stream(ctx)
			if err != nil {
				errChan <- err
			}
			reader := bufio.NewReader(logStream)
			defer logStream.Close()
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
					if err != nil {
						errChan <- err
					}
					if line == nil {
						continue
					}
					err = stream.Send(&rocketpb.LogsResponse{Message: string(line), Pod: podName})
					if err != nil {
						errChan <- err
					}
				}
			}
		}(name)

	}
}
