package v1alpha1

import (
	"fmt"

	"github.com/jakub-bacic/database-k8s-operator/pkg/database"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ObjectRef defines a reference to other k8s resource (such as Secret).
type ObjectRef struct {
	// Name of the resource.
	Name string `json:"name"`
}

// DatabaseList defines a list of Databases.
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type DatabaseList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata. More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata"`
	// List of Databases.
	Items []Database `json:"items"`
}

// Database defines a database instance.
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Database struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object’s metadata. More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#metadata
	// +k8s:openapi-gen=false
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Specification of the database instance. More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#spec-and-status
	Spec DatabaseSpec `json:"spec"`
	// Most recent observed status of the database instance. Read-only.
	// https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#spec-and-status
	Status DatabaseStatus `json:"status,omitempty"`
}

// DatabaseSpec is a specification of the database instance. More info:
// https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#spec-and-status
// +k8s:openapi-gen=true
type DatabaseSpec struct {
	// Database desired configuration.
	Database DatabaseObject `json:"database"`
	// DatabaseServer object reference.
	DatabaseServerRef ObjectRef `json:"databaseServerRef"`
}

// DatabaseStatus defines most recent observed status of the database instance. Read-only. More info:
// https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#spec-and-status
// +k8s:openapi-gen=true
type DatabaseStatus struct {
	// Represents current database phase ("" -> "Creating" -> "Created" -> "Deleting").
	Phase string `json:"phase"`
}

// DatabaseObject defines database instance desired configuration.
type DatabaseObject struct {
	// Name for the managed database.
	Name string `json:"name"`
	// Secret containing user and password for the managed database.
	UserSecretRef ObjectRef `json:"userSecretRef"`
}

// DatabaseServerList defines a list of DatabaseServers.
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type DatabaseServerList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata. More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata"`
	// List of DatabaseServers.
	Items []DatabaseServer `json:"items"`
}

// DatabaseServer defines a database server instance.
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type DatabaseServer struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object’s metadata. More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#metadata
	// +k8s:openapi-gen=false
	metav1.ObjectMeta `json:"metadata"`
	// Specification of the database server instance. More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#spec-and-status
	Spec DatabaseServerSpec `json:"spec"`
}

// DatabaseServerSpec is a specification of the database server instance. More info:
// https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#spec-and-status
// +k8s:openapi-gen=true
type DatabaseServerSpec struct {
	// Database type (see docs for the list of currently supported database server types).
	Type string `json:"type"`
	// Database server host.
	Host string `json:"host"`
	// Database server port.
	Port int32 `json:"port"`
	// Secret containing user and password. User must have enough permissions
	// to create/drop databases and users.
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
