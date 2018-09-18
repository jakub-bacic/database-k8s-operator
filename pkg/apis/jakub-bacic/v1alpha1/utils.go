package v1alpha1

import (
	"k8s.io/api/core/v1"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"fmt"
	"github.com/mitchellh/mapstructure"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/jakub-bacic/database-k8s-operator/pkg/database"
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

func getCredentials(namespace string, name string) (*database.Credentials, error) {
	secret, err := getSecret(namespace, name)
	if err != nil {
		return nil, err
	}

	result := &database.Credentials{}
	err = mapstructure.WeakDecode(secret.Data, result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse credentials data from secret (%v/%v): %v", namespace, name, err)
	}

	return result, nil
}