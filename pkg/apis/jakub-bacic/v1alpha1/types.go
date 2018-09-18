package v1alpha1

import (
	"fmt"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/jakub-bacic/database-k8s-operator/pkg/database"
)

type ObjectRef struct {
	Name string `json:"name"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type DatabaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Database `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Database struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              DatabaseSpec   `json:"spec"`
	Status            DatabaseStatus `json:"status,omitempty"`
}

type DatabaseSpec struct {
	Database          DatabaseObject `json:"database"`
	DatabaseServerRef ObjectRef      `json:"databaseServerRef"`
}

type DatabaseObject struct {
	Name          string    `json:"name"`
	UserSecretRef ObjectRef `json:"userSecretRef"`
}

type DatabaseStatus struct {
	Phase string `json:"phase"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type DatabaseServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []DatabaseServer `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type DatabaseServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              DatabaseServerSpec `json:"spec"`
}

type DatabaseServerSpec struct {
	Type              string    `json:"type"`
	Host              string    `json:"host"`
	Port              int32     `json:"port"`
	RootUserSecretRef ObjectRef `json:"rootUserSecretRef"`
}

func (db *Database) GetDatabaseServer() (*DatabaseServer, error) {
	namespace := db.Namespace
	name := db.Spec.DatabaseServerRef.Name

	result := &DatabaseServer{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DatabaseServer",
			APIVersion: "jakub-bacic.github.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	err := sdk.Get(result)
	if err != nil {
		return nil, fmt.Errorf("failed to get database server (%v/%v): %v", namespace, name, err)
	}

	return result, nil
}

func (db *Database) GetUserCredentials() (*database.Credentials, error) {
	namespace := db.Namespace
	name := db.Spec.Database.UserSecretRef.Name
	return getCredentials(namespace, name)
}


func (server *DatabaseServer) GetRootUserCredentials() (*database.Credentials, error) {
	namespace := server.Namespace
	name := server.Spec.RootUserSecretRef.Name
	return getCredentials(namespace, name)
}