package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/casbin/casbin"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
)

func main() {

	e := echo.New()

	casbin, err := casbin.NewEnforcerSafe("./model.conf", "./policy.csv")
	if err != nil {
		log.Println(err)
	}

	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(casbinRule(casbin))

	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, c.Request().URL.String())
	})

	go func() {
		if err := e.Start(":1323"); err != nil {
			log.Println(errors.WithMessage(err, "Graceful shutdown starting !!"))
		}
	}()

	graceful := make(chan os.Signal)
	signal.Notify(graceful, os.Interrupt)
	<-graceful

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		log.Fatal(errors.WithMessage(err, "Graceful shutdown timeout"))
	}
}

func casbinRule(casbin *casbin.Enforcer) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return echo.HandlerFunc(func(c echo.Context) error {

			res, err := casbin.EnforceSafe("koung", c.Request().URL.String(), c.Request().Method)
			if err != nil {
				log.Fatal(errors.WithMessage(err, "res error (casbin)"))
			}

			fmt.Println(res)
			return next(c)
		})
	}
}
