package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-students/app/service"
	"api-students/helper"
	"api-students/middleware"
)

type Dependencies struct {
	Pool           *pgxpool.Pool
	JWT            *helper.JWTManager
	AuthService    *service.AuthService
	StudentService *service.StudentService
	Permissions    *helper.PermissionSet
}

func Register(app *fiber.App, deps Dependencies) {
	api := app.Group("/api/v1")
	api.Get("/health", healthCheck(deps.Pool))

	perms := deps.Permissions

	students := api.Group("/students", middleware.RequireAuth(deps.JWT))
	students.Get("/", middleware.RequirePermission(perms, "student:list"), deps.StudentService.List)
	students.Post("/", middleware.RequireJSON, middleware.RequirePermission(perms, "student:create"), deps.StudentService.Create)
	students.Delete("/:id", middleware.RequirePermission(perms, "student:delete"), deps.StudentService.Delete)
	students.Get("/:id", deps.StudentService.Get)
	students.Put("/:id", middleware.RequireJSON, deps.StudentService.Replace)
	students.Patch("/:id", middleware.RequireJSON, deps.StudentService.Patch)

	auth := api.Group("/auth")
	auth.Post("/register", middleware.RequireJSON, deps.AuthService.Register)
	auth.Post("/login", middleware.RequireJSON, middleware.LoginRateLimiter(), deps.AuthService.Login)
	auth.Post("/refresh", middleware.RequireJSON, deps.AuthService.Refresh)
	auth.Post("/logout", middleware.RequireJSON, deps.AuthService.Logout)
	auth.Get("/me", middleware.RequireAuth(deps.JWT), deps.AuthService.Me)
}

func healthCheck(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return helper.ServiceUnavailable("database tidak dapat dihubungi")
		}
		return helper.Success(c, fiber.StatusOK, "server dan database berjalan", nil)
	}
}
