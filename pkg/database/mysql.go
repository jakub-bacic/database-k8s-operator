package database

import (
	"database/sql"
	"fmt"
)

type MySQLServer struct {
	Host string
	Port int32
	Credentials *Credentials
}

func (server *MySQLServer) CreateDatabase(dbName string, userCredentials *Credentials) error {
	connection, err := server.openDatabase()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer connection.Close()

	_, err = connection.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", dbName))
	if err != nil {
		return fmt.Errorf("failed to create database: %v", err)
	}

	_, err = connection.Exec(fmt.Sprintf("GRANT ALL PRIVILEGES ON `%s`.* TO `%s`@`%%` IDENTIFIED BY '%s'",
		dbName, userCredentials.User, userCredentials.Password))
	if err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}

	return nil
}

func (server *MySQLServer) DeleteDatabase(dbName string, user string) error {
	connection, err := server.openDatabase()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer connection.Close()

	_, err = connection.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", dbName))
	if err != nil {
		return fmt.Errorf("failed to delete database: %v", err)
	}

	_, err = connection.Exec(fmt.Sprintf("DROP USER IF EXISTS `%s`", user))
	if err != nil {
		return fmt.Errorf("failed to delete user: %v", err)
	}

	return nil
}

func (server *MySQLServer) openDatabase() (*sql.DB, error) {
	dataSource := fmt.Sprintf("%v:%v@tcp(%v:%v)/", server.Credentials.User, server.Credentials.Password,
		server.Host, server.Port)
	return sql.Open("mysql", dataSource)
}