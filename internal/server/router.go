package server

import (
	"github.com/wuyoushe/hyper-go/library/mdw"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"

	"github.com/wuyoushe/hyper-go/tool\hyper/felton_blog/internal/config"
	"github.com/wuyoushe/hyper-go/tool\hyper/felton_blog/internal/controller/api"
	"github.com/wuyoushe/hyper-go/tool\hyper/felton_blog/internal/service"
)

func InitRouter(svc service.Service, cfg *config.Config) (e *iris.Application, err error) {
	e = newIris(cfg)
	//sessManager := sessions.New(sessions.Config{
	//	Cookie:  "GiNana_Session",
	//	Expires: 24 * time.Hour,
	//})
	apiParty := mvc.New(e.Party("/api", mdw.CORS([]string{"*"})).AllowMethods(iris.MethodOptions))
	apiParty.Register(svc, getPagination)
	apiParty.Handle(new(api.CApi))
	return
}
