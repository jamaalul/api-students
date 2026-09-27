package config

import (
	"errors"
	"log/slog"

	"api-students/app/model"
	"api-students/helper"
	"api-students/middleware"
	"api-students/route"

	"github.com/gofiber/fiber/v2"
)

// NewApp merakit aplikasi: membuat instance Fiber, memasang middleware, dan mendaftarkan route.
func NewApp(logger *slog.Logger, deps route.Dependencies) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      GetEnv("APP_NAME", "Tugas Mandiri Pertemuan 4 - api-students"),
		ErrorHandler: newErrorHandler(logger),
	})

	allowedOrigins := GetEnv("ALLOWED_ORIGINS", "*")
	middleware.Register(app, logger, allowedOrigins)
	route.Register(app, deps)

	// Fallback untuk endpoint yang tidak terdaftar
	app.Use(func(c *fiber.Ctx) error {
		return helper.NotFound("endpoint tidak ditemukan")
	})

	return app
}

// newErrorHandler menangani unhandled errors dengan format response seragam.
func newErrorHandler(logger *slog.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		requestID, _ := c.Locals("requestid").(string)

		var appErr *helper.AppError
		switch {
		case errors.As(err, &appErr):
			// Error yang sudah terstruktur
		case errors.Is(err, fiber.ErrRequestEntityTooLarge):
			appErr = &helper.AppError{
				Status:  fiber.StatusRequestEntityTooLarge,
				Code:    "PAYLOAD_TOO_LARGE",
				Message: "ukuran body melebihi batas yang diizinkan",
			}
		default:
			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				appErr = &helper.AppError{
					Status:  fiberErr.Code,
					Code:    "HTTP_ERROR",
					Message: fiberErr.Message,
				}
			} else {
				appErr = helper.Internal(err)
			}
		}

		if appErr.Status >= fiber.StatusInternalServerError {
			var causeStr string
			if cause := appErr.Unwrap(); cause != nil {
				causeStr = cause.Error()
			}
			logger.Error("request_failed",
				slog.String("request_id", requestID),
				slog.String("path", c.Path()),
				slog.String("code", appErr.Code),
				slog.Int("status", appErr.Status),
				slog.String("error", causeStr))
		} else {
			logger.Warn("request_rejected",
				slog.String("request_id", requestID),
				slog.String("path", c.Path()),
				slog.String("code", appErr.Code),
				slog.Int("status", appErr.Status))
		}

		return c.Status(appErr.Status).JSON(model.ErrorResponse{
			Success:   false,
			Code:      appErr.Code,
			Message:   appErr.Message,
			Fields:    appErr.Fields,
			RequestID: requestID,
		})
	}
}
