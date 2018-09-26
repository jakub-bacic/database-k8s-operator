package stub

import (
	"context"

	"github.com/jakub-bacic/database-k8s-operator/pkg/logging"

	"github.com/jakub-bacic/database-k8s-operator/pkg/apis/jakub-bacic/v1alpha1"

	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/jakub-bacic/database-k8s-operator/pkg/database"
)

func NewHandler() sdk.Handler {
	return &Handler{}
}

type Handler struct {
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	switch o := event.Object.(type) {
	case *v1alpha1.Database:
		ctx = logging.NewContext(ctx, logging.Fields{
			"kind": "v1alpha1.Database",
			"name": o.Name,
		})
		logger := logging.GetLogger(ctx)

		// ignore events with Deleted flag (all logic is handled using Finalizers)
		if event.Deleted {
			return nil
		}

		switch phase := o.Status.Phase; phase {
		case "":
			logger.Infof("Initializing resource")
			db := o.DeepCopy()
			db.SetFinalizers([]string{"delete-db"})
			db.Status.Phase = "Creating"
			return sdk.Update(db)
		case "Creating":
			logger.Infof("Creating db")
			db := o.DeepCopy()
			if err := createDatabase(ctx, db); err != nil {
				return fmt.Errorf("failed to create db: %v", err)
			}
			db.Status.Phase = "Created"
			return sdk.Update(db)
		case "Created":
			if o.DeletionTimestamp != nil {
				logger.Infof("Resource has been scheduled for deletion")
				db := o.DeepCopy()
				db.Status.Phase = "Deleting"
				return sdk.Update(db)
			}
		case "Deleting":
			logger.Infof("Deleting db")
			db := o.DeepCopy()
			if err := deleteDatabase(ctx, db); err != nil {
				return fmt.Errorf("failed to delete db: %v", err)
			}
			db.SetFinalizers([]string{})
			db.Status.Phase = "Deleted"
			return sdk.Update(db)
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

	return &database.MySQLServer{dbServer.Spec.Host, dbServer.Spec.Port, rootCredentials}, nil
}