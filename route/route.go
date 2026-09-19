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

	students := api.Group("/students", middleware.RequireAuth(deps.JWT), middleware.RequireJSON)
	students.Get("/", middleware.RequirePermission(perms, "student:list"), deps.StudentService.List)
	students.Post("/", middleware.RequirePermission(perms, "student:create"), deps.StudentService.Create)
	students.Delete("/:id", middleware.RequirePermission(perms, "student:delete"), deps.StudentService.Delete)
	students.Get("/:id", deps.StudentService.Get)
	students.Put("/:id", deps.StudentService.Replace)
	students.Patch("/:id", deps.StudentService.Patch)

	auth := api.Group("/auth", middleware.RequireJSON)
	auth.Post("/register", deps.AuthService.Register)
	auth.Post("/login", middleware.LoginRateLimiter(), deps.AuthService.Login)
	auth.Post("/refresh", deps.AuthService.Refresh)
	auth.Post("/logout", deps.AuthService.Logout)
	auth.Get("/me", middleware.RequireAuth(deps.JWT), deps.AuthService.Me)
}

// func Register(app *fiber.App, deps Dependencies) {
//     api := app.Group("/api/v1")
//     api.Get("/health", healthCheck(deps.Pool))

//     perms := deps.Permissions

//     // students — semua endpoint wajib login
//     students := api.Group("/students",
//         middleware.RequireAuth(deps.JWT),
//         middleware.RequireJSON)

//     // Hak dapat diputuskan tanpa melihat isi data → middleware
//     students.Get("/",
//         middleware.RequirePermission(perms, "student:list"),
//         deps.StudentService.List)
//     students.Post("/",
//         middleware.RequirePermission(perms, "student:create"),
//         deps.StudentService.Create)
//     students.Delete("/:id",
//         middleware.RequirePermission(perms, "student:delete"),
//         deps.StudentService.Delete)

//     // Hak bergantung pada kepemilikan data → diperiksa di service
//     students.Get("/:id", deps.StudentService.Get)
//     students.Put("/:id", deps.StudentService.Replace)
//     students.Patch("/:id", deps.StudentService.Patch)

//     // auth routes (tidak berubah)
//     auth := api.Group("/auth", middleware.RequireJSON)
//     auth.Post("/register", deps.AuthService.Register)
//     auth.Post("/login", middleware.LoginRateLimiter(), deps.AuthService.Login)
//     auth.Post("/refresh", deps.AuthService.Refresh)
//     auth.Post("/logout", deps.AuthService.Logout)
//     auth.Get("/me", middleware.RequireAuth(deps.JWT), deps.AuthService.Me)
// }

func healthCheck(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return helper.Fail(c, fiber.StatusServiceUnavailable,
				"database tidak dapat dihubungi")
		}
		return helper.Success(c, fiber.StatusOK, "server dan database berjalan", nil)
	}
}
