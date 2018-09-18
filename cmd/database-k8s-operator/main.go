package main

import (
	"context"
	"runtime"
	"github.com/jakub-bacic/database-k8s-operator/pkg/logging"
	"github.com/operator-framework/operator-sdk/pkg/util/k8sutil"
	sdkVersion "github.com/operator-framework/operator-sdk/version"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/jakub-bacic/database-k8s-operator/pkg/stub"
	"time"
)

func printVersion(ctx context.Context) {
	logger := logging.GetLogger(ctx)
	logger.Infof("Go Version: %s", runtime.Version())
	logger.Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	logger.Infof("operator-sdk Version: %v", sdkVersion.Version)
}

func main() {
	ctx := logging.NewContext(context.TODO(), logging.Fields{})
	logger := logging.GetLogger(ctx)
	printVersion(ctx)

	sdk.ExposeMetricsPort()

	resource := "jakub-bacic.github.com/v1alpha1"
	kind := "Database"
	namespace, err := k8sutil.GetWatchNamespace()
	if err != nil {
		logger.Fatalf("failed to get watch namespace: %v", err)
	}
	resyncPeriod := time.Duration(0) * time.Second
	logger.Infof("Watching %s, %s, %s, %d", resource, kind, namespace, 0)
	sdk.Watch(resource, kind, namespace, resyncPeriod)
	sdk.Handle(stub.NewHandler())
	sdk.Run(ctx)
}
