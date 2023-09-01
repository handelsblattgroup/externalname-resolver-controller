package reconcile

import (
	"strconv"
	"strings"

	"github.com/handelsblattgroup/externalname-resolver-controller/pkg/kube/tools"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

func FlatternPortsAndProtocols(ports []corev1.ServicePort) (string, string) {
	portNumbers := make([]string, 0)
	protocols := make([]string, 0)

	for _, port := range ports {
		portNumber := port.TargetPort.StrVal
		if portNumber == "" {
			portNumber = tools.FormatInt32(port.Port)
		}

		portNumbers = append(portNumbers, portNumber)
	}

	for _, port := range ports {
		protocols = append(protocols, string(port.Protocol))
	}

	return strings.Join(portNumbers, ","), strings.Join(protocols, ",")
}

func ExpendPortsAndProtocols(ports, protocols string) ([]corev1.EndpointPort, error) {
	list := make([]corev1.EndpointPort, 0)

	portList := strings.Split(ports, ",")
	protocolList := strings.Split(protocols, ",")

	for index, port := range portList {
		portNumber, err := strconv.ParseInt(port, 10, 32)
		if err != nil {
			return list, errors.Wrap(err, "failed to parse port number")
		}

		list = append(list, corev1.EndpointPort{
			Port:     int32(portNumber),
			Protocol: corev1.Protocol(protocolList[index]),
		})
	}

	return list, nil
}
