package v1alpha1

import (
	"fmt"
	"time"

	"github.com/jakub-bacic/database-k8s-operator/pkg/database"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ObjectRef defines a reference to other k8s resource.
type ObjectRef struct {
	// Name of the resource.
	Name string `json:"name"`
}

// SecretRef defines a reference to Secret key in k8s.
type SecretRef struct {
	// Secret name
	Name string `json:"name"`
	// Secret key
	Key string `json:"key"`
}

// DatabaseList defines a list of Databases.
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
type DatabaseSpec struct {
	// Database desired configuration.
	Database DatabaseObject `json:"database"`
	// DatabaseServer object reference.
	DatabaseServerRef ObjectRef `json:"databaseServerRef"`
}

// DatabaseStatus defines most recent observed status of the database instance. Read-only. More info:
// https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#spec-and-status
type DatabaseStatus struct {
	// Represents current database status.
	Status string `json:"status"`
	// Stores last error timestamp
	LastErrorTimestamp int64 `json:"lastErrorTimestamp"`
}

// DatabaseObject defines database instance desired configuration.
type DatabaseObject struct {
	// Name for the managed database.
	Name string `json:"name"`
	// User to be created (it will be granted all priviliges to the managed database)
	User string `json:"user"`
	// Secret containing password for the database user
	PasswordSecretRef SecretRef `json:"passwordSecretRef"`
}

// DatabaseServerList defines a list of DatabaseServers.
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
type DatabaseServerSpec struct {
	// Database type (see docs for the list of currently supported database server types).
	Type string `json:"type"`
	// Database server host.
	Host string `json:"host"`
	// Database server port.
	Port int32 `json:"port"`
	// User to be used (it must have enough permissions to create/drop databases and users)
	RootUser string `json:"rootUser"`
	// Secret containing password for the user
	RootUserSecretRef SecretRef `json:"rootUserSecretRef"`
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
	user := db.Spec.Database.User

	passwordSecretRef := db.Spec.Database.PasswordSecretRef
	password, err := getSecretKey(namespace, passwordSecretRef.Name, passwordSecretRef.Key)
	if err != nil {
		return nil, err
	}

	return &database.Credentials{user, *password}, nil
}

func (db *Database) SetError() {
	now := int64(time.Now().Unix())
	db.Status.Status = "Error"
	db.Status.LastErrorTimestamp = now
}

func (db *Database) TimeSinceLastError() int64 {
	now := int64(time.Now().Unix())
	return now - db.Status.LastErrorTimestamp
}

func (server *DatabaseServer) GetRootUserCredentials() (*database.Credentials, error) {
	namespace := server.Namespace
	user := server.Spec.RootUser

	passwordSecretRef := server.Spec.RootUserSecretRef
	password, err := getSecretKey(namespace, passwordSecretRef.Name, passwordSecretRef.Key)
	if err != nil {
		return nil, err
	}
	return &database.Credentials{user, *password}, nil
}
