package v1

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mlplabs/common-go-pkg/pkg/http/jwtutils"
	"github.com/mlplabs/common-go-pkg/pkg/http/response/wrapper"
	_ "github.com/mlplabs/lynway-wms/docs/swagger"
	"github.com/mlplabs/lynway-wms/internal/config"
	httpInternal "github.com/mlplabs/lynway-wms/internal/controller/http"
	"github.com/mlplabs/lynway-wms/internal/usecases/wms"
	_ "github.com/mlplabs/lynway-wms/pkg/dto"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type ServiceHandler struct {
	service *wms.Service
	apiKeys map[string]struct{}
}

func NewServiceHandler(router *chi.Mux,
	cfg *config.Config,
	service *wms.Service,
	tokenVerifier httpInternal.TokenVerifier,
	extUserResolver httpInternal.ExternalUserResolver,
) {
	h := &ServiceHandler{
		service: service,
		apiKeys: make(map[string]struct{}),
	}
	h.apiKeys["12345"] = struct{}{}

	wrap := wrapper.NewWrapper()

	router.Route("/v1", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Get("/docs/*", httpSwagger.Handler(
				httpSwagger.URL("doc.json"),
			))
		})
		r.Group(func(r chi.Router) {
			r.Route("/auth_deprecated", func(r chi.Router) {
				r.Post("/login_deprecated", wrap.Data(h.login_deprecated))
				r.Post("/refresh_deprecated", wrap.Data(h.refresh_deprecated))
			})
		})
		r.Group(func(r chi.Router) {
			r.Route("/system", func(r chi.Router) {
				r.Get("/info", wrap.Data(h.getSystemInfo))
			})
			r.Route("/auth", func(r chi.Router) {
				r.Post("/login", wrap.Data(h.login))
			})
		})

		r.Group(func(r chi.Router) {
			r.Use(httpInternal.Auth(cfg, tokenVerifier, extUserResolver))
			r.Route("/items", func(r chi.Router) {
				r.Get("/suggest", wrap.Data(h.itemsSuggest))
				r.Post("/resolve", wrap.Data(h.itemsResolve))
				r.Get("/{id}", wrap.Data(h.getItemById))
				r.Patch("/{id}", wrap.Data(h.updateItem))
			})
			r.Route("/manufacturers", func(r chi.Router) {
				r.Get("/suggest", wrap.Data(h.manufacturersSuggest))
				r.Post("/resolve", wrap.Data(h.manufacturersResolve))
				r.Get("/{id}", wrap.Data(h.getManufacturerById))
				r.Patch("/{id}", wrap.Data(h.updateManufacturer))
			})
			r.Group(func(r chi.Router) {
				r.Use(httpInternal.RequireIdempotencyKey())
				r.Route("/storage", func(r chi.Router) {
					r.Route("/items", func(r chi.Router) {
						r.Post("/receive", wrap.Data(h.receiveItem))
						r.Post("/move", wrap.Data(h.moveItem))
						r.Post("/dispatch", wrap.Data(h.dispatchItem))
					})

					r.Route("/containers", func(r chi.Router) {
						r.Post("/place", wrap.Data(h.placeContainer))
						r.Post("/move", wrap.Data(h.moveContainer))
						r.Post("/dispatch", wrap.Data(h.dispatchContainer))
						r.Post("/nest", wrap.Data(h.nestContainer))
						r.Post("/unnest", wrap.Data(h.unnestContainer))
					})
				})
			})
		})
		// для админки
		r.Group(func(r chi.Router) {
			r.Use(jwtutils.TokenValidate("12345")) // TODO: перевести на общий токен
			r.With().Group(func(r chi.Router) {
				r.Get("/users/me", wrap.Data(h.me))
				//r.Get("/users", wrap.Data(h.getUsers))
				//r.Get("/users/{id}", wrap.Data(h.getUser))
				//r.Post("/users", wrap.Data(h.updateUser))

				r.Post("/decode", wrap.Data(h.decode))
			})
		})

		r.Group(func(r chi.Router) {
			r.Route("/mobile", func(r chi.Router) {
				r.Post("/init", wrap.Data(h.mobileInit))
				r.Route("/search", func(r chi.Router) {
					r.Post("/search/product", wrap.Data(h.searchProduct))
					r.Post("/search/container", wrap.Data(h.searchContainer))
				})
			})
		})
		// с токеном
		r.Group(func(r chi.Router) {
			r.Use(httpInternal.APIKeyValidate(h.apiKeys))
			r.Route("/print", func(r chi.Router) {
				r.Route("/agents", func(r chi.Router) {
					r.Post("/sync", wrap.Empty(h.createAgentInfo))
				})
				r.Get("/tasks", wrap.Data(h.getPrintTask))
				r.Post("/tasks/status", wrap.Empty(h.updatePrintTaskStatus))
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
