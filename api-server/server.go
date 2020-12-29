package apiserver

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/SSU-DCN/podmigration-operator/api-server/endpoints"
	"github.com/emicklei/go-restful"
	"sigs.k8s.io/controller-runtime/pkg/client"
	kubelog "sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	log = kubelog.Log.WithName("api-server")
)

type apiServer struct {
	server *http.Server
}

func (as *apiServer) Address() string {
	return as.server.Addr
}

func init() {
	restful.MarshalIndent = func(v interface{}, prefix, indent string) ([]byte, error) {
		var buf bytes.Buffer
		encoder := restful.NewEncoder(&buf)
		encoder.SetIndent(prefix, indent)
		if err := encoder.Encode(v); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}
}

func newApiServer(port int, allowedDomains []string, client client.Client) (*apiServer, error) {
	container := restful.NewContainer()
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: container.ServeMux,
	}

	cors := restful.CrossOriginResourceSharing{
		ExposeHeaders:  []string{restful.HEADER_AccessControlAllowOrigin},
		AllowedDomains: allowedDomains,
		Container:      container,
	}

	ws := new(restful.WebService)
	ws.
		Path("/").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	addEndpoints(ws, client)
	container.Add(ws)
	container.Filter(cors.Filter)
	return &apiServer{
		server: srv,
	}, nil
}

func addEndpoints(ws *restful.WebService, client client.Client) {
	resources := []endpoints.Endpoint{
		endpoints.NewPodmigrationEndpoint(client),
	}
	for _, ep := range resources {
		ep.SetupWithWS(ws)
	}
}

func (as *apiServer) Start(stop <-chan struct{}) error {
	errChan := make(chan error)
	go func() {
		err := as.server.ListenAndServe()
		if err != nil {
			switch err {
			case http.ErrServerClosed:
				log.Info("Shutting down api-server")
			default:
				log.Error(err, "Could not start an HTTP Server")
				errChan <- err
			}
		}
	}()
	log.Info("Starting api-server", "interface", "0.0.0.0", "port", as.Address())
	select {
	case <-stop:
		log.Info("Shutting down api-server")
		return as.server.Shutdown(context.Background())
	case err := <-errChan:
		return err
	}
}
