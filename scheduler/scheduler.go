package scheduler

import (
	"context"
	"fmt"
	"net/http"

	"cloud.google.com/go/compute/metadata"
	"google.golang.org/api/idtoken"

	"libs.altipla.consulting/env"
	"libs.altipla.consulting/errors"
	"libs.altipla.consulting/hosting"
	"libs.altipla.consulting/routing"
	"libs.altipla.consulting/security"
)

type Server struct {
	serviceAccount string
	audience       string

	r         *hosting.WebServer
	validator *idtoken.Validator
}

type ServerOption func(s *Server)

func WithServiceAccount(serviceAccount string) ServerOption {
	return func(s *Server) {
		s.serviceAccount = serviceAccount
	}
}

func WithAudience(audience string) ServerOption {
	return func(s *Server) {
		s.audience = audience
	}
}

func NewServer(r *hosting.WebServer, opts ...ServerOption) (*Server, error) {
	s := &Server{
		r: r,
	}
	for _, opt := range opts {
		opt(s)
	}

	if s.audience == "" {
		return nil, errors.Errorf("audience required to configure scheduler auth")
	}

	if env.IsLocal() {
		s.serviceAccount = ""
	} else if s.serviceAccount == "" {
		project, err := metadata.ProjectID()
		if err != nil {
			return nil, errors.Trace(err)
		}
		s.serviceAccount = fmt.Sprintf("scheduler@%s.iam.gserviceaccount.com", project)
	}

	if s.serviceAccount != "" {
		var err error
		s.validator, err = idtoken.NewValidator(context.Background())
		if err != nil {
			return nil, errors.Trace(err)
		}
	}

	return s, nil
}

type Handler func(ctx context.Context) error

func (s *Server) Subscribe(subscription string, handler Handler) {
	s.r.Post("/_/scheduler/"+subscription, func(w http.ResponseWriter, r *http.Request) error {
		if s.validator != nil {
			bearer := security.ReadRequestAuthorization(r)
			if bearer == "" {
				return routing.Unauthorizedf("scheduler token required")
			}
			payload, err := s.validator.Validate(r.Context(), bearer, s.audience)
			if err != nil {
				return routing.Unauthorizedf("scheduler invalid token: %s", err.Error())
			}
			if email, ok := payload.Claims["email"].(string); !ok || email != s.serviceAccount {
				return routing.Unauthorizedf("scheduler subject does not match, expected %s got: %s", s.serviceAccount, email)
			}
		}

		if err := handler(r.Context()); err != nil {
			return errors.Trace(err)
		}

		fmt.Fprintln(w, "ok")
		return nil
	})
}
