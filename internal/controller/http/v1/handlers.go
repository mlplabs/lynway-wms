package v1

import (
	"crypto/rsa"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mlplabs/common-go-pkg/pkg/http/jwtutils"
	"github.com/mlplabs/common-go-pkg/pkg/http/response/wrapper"
	_ "github.com/mlplabs/lynway-wms/docs/swagger"
	httpInternal "github.com/mlplabs/lynway-wms/internal/controller/http"
	"github.com/mlplabs/lynway-wms/internal/usecases/wms"
	_ "github.com/mlplabs/lynway-wms/pkg/dto"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type ServiceHandler struct {
	service *wms.Service
	apiKeys map[string]struct{}
}

func NewServiceHandler(router *chi.Mux, publicKey *rsa.PublicKey, service *wms.Service) {
	h := &ServiceHandler{
		service: service,
		apiKeys: make(map[string]struct{}),
	}
	h.apiKeys["12345"] = struct{}{}

	wrapper := wrapper.NewWrapper()

	router.Route("/v1", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Get("/docs/*", httpSwagger.Handler(
				httpSwagger.URL("doc.json"),
			))
		})
		r.Group(func(r chi.Router) {
			r.Route("/auth", func(r chi.Router) {
				r.Post("/login", wrapper.Data(h.login))
				r.Post("/refresh", wrapper.Data(h.refresh))
			})
		})
		// для админки
		r.Group(func(r chi.Router) {
			r.Use(jwtutils.TokenValidate("12345"))
			r.With().Group(func(r chi.Router) {
				r.Get("/users/me", wrapper.Data(h.me))
				r.Get("/users", wrapper.Data(h.getUsers))
				r.Get("/users/{id}", wrapper.Data(h.getUser))
				r.Post("/users", wrapper.Data(h.updateUser))

				r.Post("/decode", wrapper.Data(h.decode))
				r.Get("/containers/{id}", wrapper.Data(h.getContainerById))
			})
		})
		r.Group(func(r chi.Router) {
			//r.Use(rest.TokenValidateWithoutRoles(publicKey))
			r.Route("/mobile", func(r chi.Router) {
				r.Post("/init", wrapper.Data(h.mobileInit))
				r.Route("/search", func(r chi.Router) {
					r.Post("/search/product", wrapper.Data(h.searchProduct))
					r.Post("/search/container", wrapper.Data(h.searchContainer))
				})
			})
		})
		// с токеном
		r.Group(func(r chi.Router) {
			r.Use(httpInternal.APIKeyValidate(h.apiKeys))
			r.Route("/print", func(r chi.Router) {
				r.Route("/agents", func(r chi.Router) {
					r.Post("/sync", wrapper.Empty(h.updateAgentInfo))
				})
				r.Get("/tasks", wrapper.Data(h.getPrintTask))
				r.Post("/tasks/status", wrapper.Empty(h.updatePrintTaskStatus))
			})
		})
	})
}

// decode
// @Tags Wms
// @Summary "декодирует" данные от сканера, определяет штрих-код какой сущности был отсканированы
// @Router /decode [post]
// @Param request body dto.DecodeRequest true "данные от сканера"
// @Success	200	{object} wrapper.Data{data=dto.DecodeResponse} метаданные расшифровки
func (h *ServiceHandler) decode(w http.ResponseWriter, r *http.Request) (any, error) {
	req, err := getDecodeRequest(r)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	_, err = jwtutils.GetUserId(r)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	res, err := h.service.Decode(r.Context(), req.Code)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	return res.ToDTO(), nil
}

func (h *ServiceHandler) getContainerById(w http.ResponseWriter, r *http.Request) (any, error) {
	id, err := getIdParam(r)
	if err != nil {
		return nil, err
	}
	return h.service.GetContainerById(r.Context(), id)
}
