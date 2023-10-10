package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/davidroman0O/seigyo-examples/miniredis-asynq/web"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/template/html/v2"
	"github.com/hibiken/asynq"
)

type ApiProcess struct {
	readyRedis bool
	app        *fiber.App
}

func (r *ApiProcess) Init(
	ctx context.Context,
	stateGetter func() GlobalApplicationState,
	stateMutator func(mutateFunc func(GlobalApplicationState) GlobalApplicationState),
	sender func(pid string,
		data interface{})) error {

	fmt.Println("waiting ready")
	for !r.readyRedis {
		time.Sleep(time.Millisecond * 150)
	}
	fmt.Println("api start")

	var engine *html.Engine

	// embed won't be able to reload the files
	if os.Getenv("ENVIRONMENT") == "production" {
		engine = html.NewFileSystem(http.FS(web.EmbedDirViews), ".gohtml")
	} else {
		engine = html.New("./miniredis-asynq/web", ".gohtml")
		engine.Debug(true)
		engine.Reload(true)
	}

	engine.AddFunc("dict", func(values ...interface{}) (map[string]interface{}, error) {
		if len(values)%2 != 0 {
			return nil, errors.New("invalid dict call")
		}
		dict := make(map[string]interface{}, len(values)/2)
		for i := 0; i < len(values); i += 2 {
			key, ok := values[i].(string)
			if !ok {
				return nil, errors.New("dict keys must be strings")
			}
			dict[key] = values[i+1]
		}
		return dict, nil
	})

	engine.AddFunc("mod", func(i, j int) bool { return i%j == 0 })
	engine.AddFunc("sub", func(i, j int) int { return i - j })
	engine.AddFunc("add1", func(i int) int { return i + 1 })

	// Pass the engine to the Views
	r.app = fiber.New(fiber.Config{
		Views: engine,
	})

	r.app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000",
		AllowCredentials: true,
	}))

	// Access file "image.png" under `static/` directory via URL: `http://<server>/static/image.png`.
	// Without `PathPrefix`, you have to access it via URL:
	// `http://<server>/static/static/image.png`.
	r.app.Use("/static", filesystem.New(filesystem.Config{
		Root:       http.FS(web.EmbedDirStatic),
		PathPrefix: "static",
		Browse:     true,
	}))

	r.app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("pages/index", fiber.Map{}, "layouts/main")
	})

	r.app.Get("/api/stats", func(c *fiber.Ctx) error {
		queues, err := stateGetter().inspector.Queues()
		if err != nil {
			return err
		}

		var queueInfos []*asynq.QueueInfo = []*asynq.QueueInfo{}

		for _, v := range queues {
			queueInfo, err := stateGetter().inspector.GetQueueInfo(v)
			if err != nil {
				return err
			}
			queueInfos = append(queueInfos, queueInfo)
		}
		return c.JSON(queueInfos)
	})

	return nil
}

func (r *ApiProcess) Run(
	ctx context.Context,
	stateGetter func() GlobalApplicationState,
	stateMutator func(mutateFunc func(GlobalApplicationState) GlobalApplicationState),
	sender func(pid string,
		data interface{}),
	shutdownCh chan struct{},
	errCh chan<- error) error {

	r.app.Get("/api/quit", func(c *fiber.Ctx) error {
		errCh <- fmt.Errorf("want to shutdown")
		return c.SendString("ok")
	})

	go func() {
		if err := r.app.Listen(":3000"); err != nil {
			errCh <- err
		}
	}()

	fmt.Println("api wait")
	<-shutdownCh
	fmt.Println("api stop")

	r.app.Shutdown()

	fmt.Println("api end")
	return nil
}

func (r *ApiProcess) Deinit(
	ctx context.Context,
	stateGetter func() GlobalApplicationState,
	stateMutator func(mutateFunc func(GlobalApplicationState) GlobalApplicationState),
	sender func(pid string,
		data interface{})) error {

	return nil
}

func (r *ApiProcess) Received(
	pid string,
	data interface{}) error {

	switch data.(type) {
	case RedisReady:
		r.readyRedis = true
	}

	return nil
}
