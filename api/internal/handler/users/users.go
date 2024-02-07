package users

import (
	"net/http"
	"reflect"

	"github.com/apisix/manager-api/internal/core/entity"
	"github.com/apisix/manager-api/internal/core/store"
	"github.com/apisix/manager-api/internal/handler"
	"github.com/gin-gonic/gin"
	"github.com/shiningrush/droplet"
	"github.com/shiningrush/droplet/data"
	"github.com/shiningrush/droplet/wrapper"
	wgin "github.com/shiningrush/droplet/wrapper/gin"
)

type Handler struct {
	userStore store.Interface
	teamStore store.Interface
	roleStore store.Interface
}

func NewHandler() (handler.RouteRegister, error) {
	return &Handler{
		userStore: store.GetStore(store.HubKeyServerInfo),
	}, nil
}

func (h *Handler) ApplyRoute(r *gin.Engine) {
	r.GET("/apisix/admin/users/:id", wgin.Wraps(h.Get,
		wrapper.InputType(reflect.TypeOf(entity.User{}))))
	r.POST("/apisix/admin/users", wgin.Wraps(h.Create,
		wrapper.InputType(reflect.TypeOf(entity.User{}))))
}

type GetInput struct {
	ID string `auto_read:"id,path" validate:"required"`
}

func (h *Handler) Get(c droplet.Context) (any, error) {
	input := c.Input().(*GetInput)

	r, err := h.userStore.Get(c.Context(), input.ID)
	if err != nil {
		return &data.SpecCodeResponse{StatusCode: http.StatusNotFound}, err
	}

	//format
	user := r.(*entity.User)
	if user.TeamsID != nil {
		return nil, nil
	}

	return nil, nil

}

func (h *Handler) Create(c droplet.Context) (any, error) {
	input := c.Input().(*entity.User)

	// check name existed
	ret, err := handler.NameExistCheck(c.Context(), h.userStore, "user", input.Name, nil)
	if err != nil {
		return ret, err
	}

	// create
	res, err := h.userStore.Create(c.Context(), input)
	if err != nil {
		return handler.SpecCodeResponse(err), err
	}

	return res, nil
}
