package config

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-students/app/service"
	"api-students/helper"
	"api-students/middleware"
	"api-students/route"
)

// NewApp merakit aplikasi: membuat instance Fiber, memasang middleware, dan mendaftarkan route.
func NewApp(
	logger *slog.Logger, pool *pgxpool.Pool, studentService *service.StudentService,
) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      GetEnv("APP_NAME", "Tugas Mandiri Pertemuan 4 - api-students"),
		ErrorHandler: newErrorHandler(logger),
	})

	middleware.Register(app, logger)
	route.Register(app, pool, studentService)

	// Fallback untuk endpoint yang tidak terdaftar
	app.Use(func(c *fiber.Ctx) error {
		return helper.Fail(c, fiber.StatusNotFound, "endpoint tidak ditemukan")
	})

	return app
}

// newErrorHandler menangani unhandled errors dengan format response seragam.
func newErrorHandler(logger *slog.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		status := fiber.StatusInternalServerError
		message := "terjadi kesalahan pada server"

		if e, ok := err.(*fiber.Error); ok {
			status = e.Code
			message = e.Message
		}

		logger.Error("unhandled_error",
			slog.String("path", c.Path()),
			slog.Int("status", status),
			slog.String("error", err.Error()),
		)
		return helper.Fail(c, status, message)
	}
}
