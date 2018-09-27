package v1alpha1

import (
	"time"

	"github.com/jakub-bacic/database-k8s-operator/pkg/database"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	StatusInitial  = ""
	StatusCreating = "Creating"
	StatusCreated  = "Created"
	StatusDeleting = "Deleting"
	StatusError    = "Error"

	FinalizerDeleteDb = "delete-db"
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
	// Database server configuration.
	DatabaseServer DatabaseServerObject `json:"databaseServer"`
	// Additional options.
	Options *OptionsObject `json:"options,omitempty"`
}

// DatabaseStatus defines most recent observed status of the database instance. Read-only. More info:
// https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#spec-and-status
type DatabaseStatus struct {
	// Represents current database status.
	Status string `json:"status"`
	// Stores last error timestamp
	LastErrorTimestamp *int64 `json:"lastErrorTimestamp,omitempty"`
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

// DatabaseServerObject defines database server configuration.
type DatabaseServerObject struct {
	// Database type (see docs for the list of currently supported database server types).
	Type string `json:"type"`
	// Database server host.
	Host string `json:"host"`
	// Database server port.
	Port int32 `json:"port"`
	// User to be used (it must have enough permissions to create/drop databases and users)
	RootUser string `json:"rootUser"`
	// Secret containing password for the user
	RootPasswordSecretRef SecretRef `json:"rootPasswordSecretRef"`
}

// OptionsObject defines additional options.
type OptionsObject struct {
	// Drop managed database and user when Database resource is deleted.
	DropOnDelete *bool `json:"dropOnDelete,omitempty"`
}

func makePointer(val bool) *bool {
	return &val
}

func (db *Database) InitWithDefaults() {
	if db.Spec.Options == nil {
		db.Spec.Options = &OptionsObject{}
	}
	if db.Spec.Options.DropOnDelete == nil {
		db.Spec.Options.DropOnDelete = makePointer(true)
	}

	db.SetStatus(StatusCreating)
}

func (db *Database) SetStatus(status string) {
	if status == db.Status.Status {
		return
	}

	if status == "Error" {
		now := int64(time.Now().Unix())
		db.Status.LastErrorTimestamp = &now
	} else {
		db.Status.LastErrorTimestamp = nil
	}
	db.Status.Status = status
}

func (db *Database) TimeSinceLastError() int64 {
	now := int64(time.Now().Unix())
	return now - *db.Status.LastErrorTimestamp
}

func (db *Database) DropOnDelete() bool {
	return *db.Spec.Options.DropOnDelete
}

func (db *Database) GetDatabaseUserCredentials() (*database.Credentials, error) {
	namespace := db.Namespace
	user := db.Spec.Database.User

	passwordSecretRef := db.Spec.Database.PasswordSecretRef
	password, err := getSecretKey(namespace, passwordSecretRef.Name, passwordSecretRef.Key)
	if err != nil {
		return nil, err
	}

	return &database.Credentials{user, *password}, nil
}

func (db *Database) GetDatabaseServerCredentials() (*database.Credentials, error) {
	namespace := db.Namespace
	user := db.Spec.DatabaseServer.RootUser

	passwordSecretRef := db.Spec.DatabaseServer.RootPasswordSecretRef
	password, err := getSecretKey(namespace, passwordSecretRef.Name, passwordSecretRef.Key)
	if err != nil {
		return nil, err
	}
	return &database.Credentials{user, *password}, nil
}
