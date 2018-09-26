package v1alpha1

import (
	"fmt"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getSecret(namespace string, name string) (*v1.Secret, error) {
	secret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	err := sdk.Get(secret)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret (%v/%v): %v", namespace, name, err)
	}
	return secret, nil
}

func getSecretKey(namespace string, name string, key string) (*string, error) {
	secret, err := getSecret(namespace, name)
	if err != nil {
		return nil, err
	}

	bytes, ok := secret.Data[key]
	if !ok {
		return nil, fmt.Errorf("failed to read password from secret (%v/%v): key %v does not exist", namespace, name, key)
	}

	secretValue := string(bytes)
	return &secretValue, nil
}
