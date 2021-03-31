package main

import (
	"flag"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var (
	port int
	// certFile, keyFile string
	certDir string
	scheme  = runtime.NewScheme()
	log     = logf.Log.WithName("sw-admission-webhook")
)

func init() {
	logf.SetLogger(zap.New())
}

func main() {
	entryLog := log.WithName("entrypoint")

	flag.IntVar(&port, "port", 8443, "Webhook server port.")
	flag.StringVar(&certDir, "cert-dir", "/etc/webhook/", "dir contains key and cert file")
	flag.Parse()

	entryLog.Info("setting up manager")
	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{
		Scheme:  scheme,
		Port:    port,
		CertDir: certDir,
	})
	if err != nil {
		entryLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	entryLog.Info("setting up webhook server")
	hookServer := mgr.GetWebhookServer()
	hookServer.Register("/mutate-pod", &webhook.Admission{Handler: &podMutate{Client: mgr.GetClient()}})

	entryLog.Info("start manager")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		entryLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
