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
		case v1alpha1.StatusInitial:
			logger.Infof("Initializing resource")
			db := o.DeepCopy()
			db.SetStatus(v1alpha1.StatusCreating)
			return sdk.Update(db)
		case v1alpha1.StatusCreating:
			logger.Infof("Creating db")
			db := o.DeepCopy()
			if err := createDatabase(ctx, db); err != nil {
				logger.Warnf("failed to create db: %v", err)
				db.SetStatus(v1alpha1.StatusError)
				return sdk.Update(db)
			}
			db.SetFinalizers([]string{v1alpha1.FinalizerDeleteDb})
			db.SetStatus(v1alpha1.StatusCreated)
			return sdk.Update(db)
		case v1alpha1.StatusCreated:
			if o.DeletionTimestamp != nil {
				logger.Infof("Resource has been scheduled for deletion")
				db := o.DeepCopy()
				db.SetStatus(v1alpha1.StatusDeleting)
				return sdk.Update(db)
			}
		case v1alpha1.StatusDeleting:
			logger.Infof("Deleting db")
			db := o.DeepCopy()
			if err := deleteDatabase(ctx, db); err != nil {
				logger.Warnf("failed to delete db: %v", err)
				db.SetStatus(v1alpha1.StatusError)
				return sdk.Update(db)
			}
			db.SetFinalizers([]string{})
			return sdk.Update(db)
		case v1alpha1.StatusError:
			// should be adjusted according to resyncPeriod
			if o.TimeSinceLastError() >= 10 {
				logger.Infof("Trying to recover from error status")
				db := o.DeepCopy()
				if o.DeletionTimestamp == nil {
					db.SetStatus(v1alpha1.StatusCreating)
				} else {
					db.SetStatus(v1alpha1.StatusDeleting)
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

	userCredentials, err := db.GetDatabaseUserCredentials()
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
	// at the moment, only 'mysql' type is supported
	if db.Spec.DatabaseServer.Type != "mysql" {
		return nil, fmt.Errorf("unsupported database server type: %v", db.Spec.DatabaseServer.Type)
	}

	credentials, err := db.GetDatabaseServerCredentials()
	if err != nil {
		return nil, err
	}

	dbServer := &database.MySQLServer{
		Host:        db.Spec.DatabaseServer.Host,
		Port:        db.Spec.DatabaseServer.Port,
		Credentials: credentials,
	}
	return dbServer, nil
}
