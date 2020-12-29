package apiserver

/*
Now we will create an API Server Manager that will create the K8S client and keep a reference to it.
It will also create a cache that will be used to create a cached K8S client,
initialize the cache properly and in the end handle the termination signals.
*/

import (
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

var (
	defaultRetryPeriod = 2 * time.Second
)

// Options to customize Manager behaviour and pass information
type Options struct {
	Scheme         *runtime.Scheme
	Namespace      string
	Port           int
	AllowedDomains []string
}

type Manager interface {
	Start(stop <-chan struct{}) error
}

type manager struct {
	config          *rest.Config
	client          client.Client
	server          *apiServer
	started         bool
	internalStop    <-chan struct{}
	internalStopper chan<- struct{}
	cache           cache.Cache
	errSignal       *errSignaler
	port            int
	allowedDomains  []string
}

func NewManager(config *rest.Config, options Options) (Manager, error) {
	mapper, err := apiutil.NewDynamicRESTMapper(config)
	if err != nil {
		return nil, err
	}

	cc, err := cache.New(config, cache.Options{
		Scheme:    options.Scheme,
		Mapper:    mapper,
		Resync:    &defaultRetryPeriod,
		Namespace: options.Namespace,
	})
	if err != nil {
		return nil, err
	}

	c, err := client.New(config, client.Options{Scheme: options.Scheme, Mapper: mapper})
	if err != nil {
		return nil, err
	}

	stop := make(chan struct{})
	return &manager{
		config: config,
		cache:  cc,
		client: &client.DelegatingClient{
			Reader: &client.DelegatingReader{
				CacheReader:  cc,
				ClientReader: c,
			},
			Writer:       c,
			StatusClient: c,
		},
		internalStop:    stop,
		internalStopper: stop,
		port:            options.Port,
		allowedDomains:  options.AllowedDomains,
	}, nil
}

func (m *manager) Start(stop <-chan struct{}) error {
	defer close(m.internalStopper)
	// initialize this here so that we reset the signal channel state on every start
	m.errSignal = &errSignaler{errSignal: make(chan struct{})}
	m.waitForCache()

	srv, err := newApiServer(m.port, m.allowedDomains, m.client)
	if err != nil {
		return err
	}

	go func() {
		if err := srv.Start(m.internalStop); err != nil {
			m.errSignal.SignalError(err)
		}
	}()
	select {
	case <-stop:
		return nil
	case <-m.errSignal.GotError():
		// Error starting the cache
		return m.errSignal.Error()
	}
}

func (m *manager) waitForCache() {
	if m.started {
		return
	}

	go func() {
		if err := m.cache.Start(m.internalStop); err != nil {
			m.errSignal.SignalError(err)
		}
	}()

	// Wait for the caches to sync.
	m.cache.WaitForCacheSync(m.internalStop)
	m.started = true
}
