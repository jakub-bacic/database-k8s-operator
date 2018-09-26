# database-k8s-operator

### Overview
The database-k8s-operator is a Kubernetes operator to automate the process of creating
dedicated databases and users in database server.

### Getting started

1. Create root user credentials Secret (user must have sufficient permissions to create/drop 
   database and users in database server):

  ```
  kubectl create secret generic mysql-root-credentials --from-literal password=rootsecret
  ```

2. Create user credentials Secret for your new db (this password should be use later
   by your app so)):

  ```
  kubectl create secret generic myapp-db-credentials --from-literal password=mysecret
  ```

3. Create Database resource:

  ```
  myapp-database.yaml
  ------------------------
  apiVersion: "jakub-bacic.github.com/v1alpha1"
  kind: "Database"
  metadata:
    name: "myapp-db"
  spec:
    database:
      name: myappdb
      user: myappdb_user
      passwordSecretRef:
        name: myapp-db-credentials
        key: password
    databaseServer:
      type: mysql
      host: mysql.default
      port: 3306
      rootUser: root
      rootPasswordSecretRef:
        name: mysql-root-credentials
        key: password
    options:
      dropOnDelete: true
  ```

  ```
  kubectl create -f myapp-database.yaml
  ```

**WARNING:** When the Database resource is deleted, associated database and user in database server
are dropped by default. This behavior can be changed by modifying dropOnDelete option
in Database spec.

### Supported databases

| Database | Type field value |
| -------- | ---------------- |
| MySQL    | `mysql`          |