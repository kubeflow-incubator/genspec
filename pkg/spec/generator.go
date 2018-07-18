package spec

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/go-openapi/spec"
	"github.com/kubeflow-incubator/genspec/pkg/storage"
	"github.com/kubeflow/tf-operator/pkg/apis/tensorflow/v1alpha2"
	"k8s.io/apimachinery/pkg/apimachinery/announced"
	"k8s.io/apimachinery/pkg/apimachinery/registered"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"
	genericoptions "k8s.io/apiserver/pkg/server/options"
	"k8s.io/kube-openapi/pkg/builder"
	"k8s.io/kube-openapi/pkg/common"
)

type Config struct {
	Scheme *runtime.Scheme
	Codecs serializer.CodecFactory

	Info              spec.InfoProps
	OpenAPIDefinition common.GetOpenAPIDefinitions
}

type TypeInfo struct {
	GroupVersion schema.GroupVersion
	Resource     string
	Kind         string
}

// newServerConfigWithSpec returns a serverConfig with OpenAPI config.
// This serverConfig will be used to set up a temporary server to generate specification.
func newServerConfigWithSpec(cfg Config) (*genericapiserver.RecommendedConfig, error) {
	recommendedOptions := genericoptions.NewRecommendedOptions("/registry/foo.com", cfg.Codecs.LegacyCodec())
	recommendedOptions.SecureServing.BindPort = 8443
	recommendedOptions.Etcd = nil
	recommendedOptions.Authentication = nil
	recommendedOptions.Authorization = nil
	recommendedOptions.CoreAPI = nil

	if err := recommendedOptions.SecureServing.MaybeDefaultWithSelfSignedCerts("localhost", nil, []net.IP{net.ParseIP("127.0.0.1")}); err != nil {
		return nil, fmt.Errorf("error creating self-signed certificates: %v", err)
	}

	serverConfig := genericapiserver.NewRecommendedConfig(cfg.Codecs)
	if err := recommendedOptions.ApplyTo(serverConfig); err != nil {
		return nil, err
	}
	serverConfig.OpenAPIConfig = genericapiserver.DefaultOpenAPIConfig(cfg.OpenAPIDefinition, cfg.Scheme)
	serverConfig.OpenAPIConfig.Info.InfoProps = cfg.Info

	return serverConfig, nil
}

// routeStorage returns a pkg for RESTful services, which contains CRUD operation to CRD TFJob.
func routeStorage(typeInfo TypeInfo, schema *runtime.Scheme) (map[string]rest.Storage, error) {
	groupVersionKind := typeInfo.GroupVersion.WithKind(typeInfo.Kind)
	object, err := schema.New(groupVersionKind)
	if err != nil {
		return nil, err
	}
	list, err := schema.New(typeInfo.GroupVersion.WithKind(groupVersionKind.Kind + "List"))
	if err != nil {
		return nil, err
	}

	routeStorage := storage.NewStandardStorage(storage.NewResourceInfo(groupVersionKind, object, list))

	return map[string]rest.Storage{typeInfo.Resource: routeStorage}, nil
}

// RenderSwaggerJson returns OpenAPI specification swagger.json with generated model
// pkg/apis/tensorflow/v1alpha2/openapi_generated.go and API routing information.
func RenderSwaggerJson() (string, error) {
	var (
		groupFactoryRegistry = make(announced.APIGroupFactoryRegistry)
		registry             = registered.NewOrDie("")
		Scheme               = runtime.NewScheme()
		Codecs               = serializer.NewCodecFactory(Scheme)
	)

	metav1.AddToGroupVersion(Scheme, schema.GroupVersion{Version: "v1"})

	// Register the API group to schema
	if err := v1alpha2.Install(groupFactoryRegistry, registry, Scheme); err != nil {
		return "", fmt.Errorf("failed to register API group to schema: %v", err)
	}

	serverConfig, err := newServerConfigWithSpec(
		Config{
			Scheme: Scheme,
			Codecs: Codecs,
			Info: spec.InfoProps{
				Version: "v1alpha2",
				Contact: &spec.ContactInfo{
					Name: "kubeflow.org",
					URL:  "https://kubeflow.org",
				},
				License: &spec.License{
					Name: "Apache 2.0",
					URL:  "https://www.apache.org/licenses/LICENSE-2.0.html",
				},
			},
			OpenAPIDefinition: v1alpha2.GetOpenAPIDefinitions,
		})
	if err != nil {
		return "", fmt.Errorf("failed to create server config: %v", err)
	}

	genericServer, err := serverConfig.Complete().New("openapi-server", genericapiserver.EmptyDelegate)
	if err != nil {
		return "", fmt.Errorf("failed to create server: %v", err)
	}

	// Add routing information to this server
	typeInfo := TypeInfo{v1alpha2.SchemeGroupVersion, v1alpha2.Plural, v1alpha2.Kind}
	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(typeInfo.GroupVersion.Group, registry, Scheme, metav1.ParameterCodec, Codecs)
	rtstorage, err := routeStorage(typeInfo, Scheme)
	if err != nil {
		return "", err
	}

	apiGroupInfo.VersionedResourcesStorageMap[typeInfo.GroupVersion.Version] = rtstorage
	if err := genericServer.InstallAPIGroup(&apiGroupInfo); err != nil {
		return "", err
	}

	apiSpec, err := builder.BuildOpenAPISpec(genericServer.Handler.GoRestfulContainer.RegisteredWebServices(), serverConfig.OpenAPIConfig)
	if err != nil {
		return "", fmt.Errorf("failed to render OpenAPI spec: %v", err)
	}
	data, err := json.MarshalIndent(apiSpec, "", "	")
	if err != nil {
		return "", err
	}

	return string(data), nil
}
