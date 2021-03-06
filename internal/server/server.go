package server

import (
	"fmt"
	"github.com/wuyoushe/hyper-go/library/conf/paladin"
	"github.com/wuyoushe/hyper-go/library/log"
	"github.com/wuyoushe/hyper-go/library/mdw"
	"github.com/wuyoushe/hyper-go/service/tools"
	"github.com/kataras/golog"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/logger"
	"github.com/kataras/iris/v12/middleware/recover"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/wuyoushe/hyper-go/tool\hyper/felton_blog/internal/config"
	"github.com/wuyoushe/hyper-go/tool\hyper/felton_blog/internal/model"
)

func NewHttpServer(irisApp *iris.Application, cfg *config.Config) (h *http.Server, err error) {
	if err = paladin.Get("http.toml").UnmarshalTOML(cfg); err != nil {
		return
	}
	if err = irisApp.Build(); err != nil {
		log.Println(err.Error())
	}
	h = &http.Server{
		Addr:         cfg.Server.Addr,
		Handler:      irisApp,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout),
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout),
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout),
	}
	log.Printf("HTTP服务已启动 [ http://%s ]", cfg.Server.Addr)
	return
}

func newIris(cfg *config.Config) (e *iris.Application) {
	e = iris.New()
	golog.Install(log.GetLogger())
	customLogger := logger.New(logger.Config{
		Status: true, IP: true, Method: true, Path: true, Query: true,
		//MessageHeaderKeys: []string{"User-Agent"},
	})
	e.OnAnyErrorCode(customLogger)
	e.Use(customLogger, recover.New())
	e.Logger().SetLevel(cfg.IrisLogLevel)
	initTemplate(e, cfg)
	initStaticDir(e, cfg)

	// Swagger
	handle := mdw.SwaggerHandler("http://127.0.0.1:8000/swagger/doc.json")
	e.Get("/swagger/*any", handle)

	e.Use(func(ctx iris.Context) {
		ctx.Gzip(cfg.EnableGzip)
		ctx.Next()
	})

	return
}

func initTemplate(e *iris.Application, cfg *config.Config) {
	if !cfg.EnableTemplate {
		return
	}
	tmpl := iris.HTML(cfg.ViewsPath, ".html").
		Reload(cfg.ReloadTemplate)
	tmpl.AddFunc("date", dateFormat)
	tmpl.AddFunc("str2html", str2html)
	e.RegisterView(tmpl)
}

func initStaticDir(e *iris.Application, cfg *config.Config) {
	if !cfg.EnableTemplate {
		return
	}
	staticDirList := strings.Split(cfg.StaticDir, " ")
	if len(staticDirList) > 0 {
		path := strings.Split(staticDirList[0], ":")
		e.Favicon(fmt.Sprintf("%s/favicon.ico", path[1]))
	}
	for _, v := range staticDirList {
		path := strings.Split(v, ":")
		if len(path) == 2 {
			e.HandleDir(path[0], path[1], iris.DirOptions{
				Gzip: true,
				ShowList: false,
			})
		}
	}
}

// template function
func dateFormat(t time.Time, format string) (template.HTML, error) {
	return template.HTML(tools.New().TimeFormat(t, format)), nil
}

func str2html(str string) (template.HTML, error) {
	return template.HTML(str), nil
}

func getPagination(ctx iris.Context) *model.Pager {
	return &model.Pager{
		Page:     ctx.URLParamInt64Default("page", 1),
		PageSize: ctx.URLParamInt64Default("pagesize", 15),
		UrlPath:  ctx.Path(),
	}
}
