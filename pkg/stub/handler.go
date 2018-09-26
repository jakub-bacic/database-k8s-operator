package stub

import (
	"context"

	"github.com/jakub-bacic/database-k8s-operator/pkg/logging"

	"github.com/jakub-bacic/database-k8s-operator/pkg/apis/jakub-bacic/v1alpha1"

	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jakub-bacic/database-k8s-operator/pkg/database"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
)

func NewHandler() sdk.Handler {
	return &Handler{}
}

type Handler struct {
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	// ignore events with Deleted flag (all logic is handled using Finalizers)
	if event.Deleted {
		return nil
	}

	switch o := event.Object.(type) {
	case *v1alpha1.Database:
		ctx = logging.NewContext(ctx, logging.Fields{
			"kind": "v1alpha1.Database",
			"name": o.Name,
		})
		logger := logging.GetLogger(ctx)

		switch status := o.Status.Status; status {
		case "":
			logger.Infof("Initializing resource")
			db := o.DeepCopy()
			db.Status.Status = "Creating"
			return sdk.Update(db)
		case "Creating":
			logger.Infof("Creating db")
			db := o.DeepCopy()
			if err := createDatabase(ctx, db); err != nil {
				logger.Warnf("failed to create db: %v", err)
				db.SetError()
				return sdk.Update(db)
			}
			db.SetFinalizers([]string{"delete-db"})
			db.Status.Status = "Created"
			return sdk.Update(db)
		case "Created":
			if o.DeletionTimestamp != nil {
				logger.Infof("Resource has been scheduled for deletion")
				db := o.DeepCopy()
				db.Status.Status = "Deleting"
				return sdk.Update(db)
			}
		case "Deleting":
			logger.Infof("Deleting db")
			db := o.DeepCopy()
			if err := deleteDatabase(ctx, db); err != nil {
				logger.Warnf("failed to delete db: %v", err)
				db.SetError()
				return sdk.Update(db)
			}
			db.SetFinalizers([]string{})
			db.Status.Status = "Deleted"
			return sdk.Update(db)
		case "Error":
			// should be adjusted according to resyncPeriod
			if o.TimeSinceLastError() >= 10 {
				logger.Infof("Trying to recover from error status")
				db := o.DeepCopy()
				if o.DeletionTimestamp == nil {
					db.Status.Status = "Creating"
				} else {
					db.Status.Status = "Deleting"
				}
				return sdk.Update(db)
			}
		}
	}
	return nil
}

func createDatabase(ctx context.Context, db *v1alpha1.Database) error {
	logger := logging.GetLogger(ctx)

	dbServer, err := getDatabaseServer(ctx, db)
	if err != nil {
		return err
	}

	userCredentials, err := db.GetUserCredentials()
	if err != nil {
		return err
	}

	if err := dbServer.CreateDatabase(db.Spec.Database.Name, userCredentials); err != nil {
		return err
	}

	logger.Infof("New database created")
	return nil
}

func deleteDatabase(ctx context.Context, db *v1alpha1.Database) error {
	logger := logging.GetLogger(ctx)

	dbServer, err := getDatabaseServer(ctx, db)
	if err != nil {
		return err
	}

	if err := dbServer.DeleteDatabase(db.Spec.Database.Name, db.Spec.Database.User); err != nil {
		return err
	}

	logger.Infof("Database deleted")
	return nil
}

func getDatabaseServer(ctx context.Context, db *v1alpha1.Database) (database.DbServer, error) {
	dbServer, err := db.GetDatabaseServer()
	if err != nil {
		return nil, err
	}

	// at the moment, only 'mysql' type is supported
	if dbServer.Spec.Type != "mysql" {
		return nil, fmt.Errorf("unsupported database server type: %v", dbServer.Spec.Type)
	}

	rootCredentials, err := dbServer.GetRootUserCredentials()
	if err != nil {
		return nil, err
	}

	return &database.MySQLServer{Host: dbServer.Spec.Host, Port: dbServer.Spec.Port, Credentials: rootCredentials}, nil
}
