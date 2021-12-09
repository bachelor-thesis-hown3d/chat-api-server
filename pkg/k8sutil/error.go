package k8sutil

import (
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
)

// checkIfAlreadyExistsError will return nil if the provided error is nil or an api Already Exists Error
// else it will just return the provided err
func checkIfAlreadyExistsError(err error) error {
	if err == nil {
		return nil
	}
	if !apiErrors.IsAlreadyExists(err) {
		return err
	}
	return nil
}
