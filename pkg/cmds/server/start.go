package server

import (
	"fmt"
	"io"
	"net"

	"github.com/kubepack/packserver/apis/apps/v1alpha1"
	clientset "github.com/kubepack/packserver/client/clientset/internalversion"
	informers "github.com/kubepack/packserver/client/informers/internalversion"
	"github.com/kubepack/packserver/pkg/admission/plugin/banflunder"
	"github.com/kubepack/packserver/pkg/admission/wardleinitializer"
	"github.com/kubepack/packserver/pkg/apiserver"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	genericapiserver "k8s.io/apiserver/pkg/server"
	genericoptions "k8s.io/apiserver/pkg/server/options"
)

const defaultEtcdPathPrefix = "/registry/apps.kubepack.com"

type KubepackServerOptions struct {
	RecommendedOptions *genericoptions.RecommendedOptions
	Admission          *genericoptions.AdmissionOptions

	StdOut io.Writer
	StdErr io.Writer
}

func NewKubepackServerOptions(out, errOut io.Writer) *KubepackServerOptions {
	o := &KubepackServerOptions{
		RecommendedOptions: genericoptions.NewRecommendedOptions(defaultEtcdPathPrefix, apiserver.Codecs.LegacyCodec(v1alpha1.SchemeGroupVersion)),
		Admission:          genericoptions.NewAdmissionOptions(),

		StdOut: out,
		StdErr: errOut,
	}

	return o
}

func (o KubepackServerOptions) Validate(args []string) error {
	var errors []error
	errors = append(errors, o.RecommendedOptions.Validate()...)
	errors = append(errors, o.Admission.Validate()...)
	return utilerrors.NewAggregate(errors)
}

func (o *KubepackServerOptions) Complete() error {
	return nil
}

func (o KubepackServerOptions) Config() (*apiserver.Config, error) {
	// register admission plugins
	banflunder.Register(o.Admission.Plugins)

	// TODO have a "real" external address
	if err := o.RecommendedOptions.SecureServing.MaybeDefaultWithSelfSignedCerts("localhost", nil, []net.IP{net.ParseIP("127.0.0.1")}); err != nil {
		return nil, fmt.Errorf("error creating self-signed certificates: %v", err)
	}

	serverConfig := genericapiserver.NewRecommendedConfig(apiserver.Codecs)
	if err := o.RecommendedOptions.ApplyTo(serverConfig); err != nil {
		return nil, err
	}

	client, err := clientset.NewForConfig(serverConfig.LoopbackClientConfig)
	if err != nil {
		return nil, err
	}
	informerFactory := informers.NewSharedInformerFactory(client, serverConfig.LoopbackClientConfig.Timeout)
	admissionInitializer, err := wardleinitializer.New(informerFactory)
	if err != nil {
		return nil, err
	}

	if err := o.Admission.ApplyTo(&serverConfig.Config, serverConfig.SharedInformerFactory, serverConfig.ClientConfig, apiserver.Scheme, admissionInitializer); err != nil {
		return nil, err
	}

	config := &apiserver.Config{
		GenericConfig: serverConfig,
		ExtraConfig:   apiserver.ExtraConfig{},
	}
	return config, nil
}

func (o KubepackServerOptions) RunWardleServer(stopCh <-chan struct{}) error {
	config, err := o.Config()
	if err != nil {
		return err
	}

	server, err := config.Complete().New()
	if err != nil {
		return err
	}

	server.GenericAPIServer.AddPostStartHook("start-sample-server-informers", func(context genericapiserver.PostStartHookContext) error {
		config.GenericConfig.SharedInformerFactory.Start(context.StopCh)
		return nil
	})

	return server.GenericAPIServer.PrepareRun().Run(stopCh)
}
