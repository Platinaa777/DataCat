package secret

import (
	"net/http"

	"github.com/Platinaa777/DataCat/internal/api/http/transport"
	"github.com/Platinaa777/DataCat/internal/usecases"
)

type Handler struct {
	service   *usecases.Service
	access    *transport.Access
	responses *transport.Responder
}

func New(service *usecases.Service, access *transport.Access, responses *transport.Responder) *Handler {
	return &Handler{service: service, access: access, responses: responses}
}

func (s *Handler) Put(r *http.Request) (any, error) {
	ns, err := s.access.NamespaceAdmin(r)
	if err != nil {
		return nil, err
	}
	input, err := transport.Body[putRequest](r)
	if err != nil {
		return nil, err
	}
	return "Secret added", s.service.PutSecret(r.Context(), ns, input.Key, input.Value)
}

func (s *Handler) Get(r *http.Request) (any, error) {
	ns, err := s.access.NamespaceAdmin(r)
	if err != nil {
		return nil, err
	}
	return s.service.Secret(r.Context(), ns, transport.Param(r, "key"))
}

func (s *Handler) Delete(r *http.Request) (any, error) {
	ns, err := s.access.NamespaceAdmin(r)
	if err != nil {
		return nil, err
	}
	return "Secret removed", s.service.DeleteSecret(r.Context(), ns, transport.Param(r, "key"))
}
