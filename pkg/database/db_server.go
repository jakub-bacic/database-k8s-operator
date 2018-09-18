package database

type Credentials struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

type DbServer interface {
	CreateDatabase(dbName string, userCredentials *Credentials) error
	DeleteDatabase(dbName string, user string) error
}