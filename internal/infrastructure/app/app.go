package app

import (
	"cleanArch_with_postgres/internal/infrastructure/config"
	"cleanArch_with_postgres/internal/infrastructure/database"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"gorm.io/gorm"
)

type App struct {
	FiberApp *fiber.App
	DB       *gorm.DB
	Cfg      *config.Config
}

type IRouter interface {
	RegisterRouter(app *App)
}

func New(router IRouter) *App {
	cfg, err := config.Setup()
	if err != nil {
		panic(err)
	}

	fiberApp := fiber.New()
	db := database.New(cfg.Database)

	fiberApp.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173", // http://localhost:5173	http://10.242.82.156:5173
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	app := &App{
		FiberApp: fiberApp,
		DB:       db,
		Cfg:      cfg,
	}

	router.RegisterRouter(app)

	return app
}

func (a *App) Start() {
	go func() {
		err := a.FiberApp.Listen(fmt.Sprintf(":%v", a.Cfg.Server.Port))
		if err != nil {
			panic(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c //
}
