package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api-students/app/repository"
	"api-students/app/service"
	"api-students/config"
	"api-students/database"
	"api-students/helper"
	"api-students/route"
)

func main() {
	// 1. Konfigurasi dan logger terstruktur
	config.LoadEnv()
	logger := config.NewLogger()

	// 2. Koneksi database pool
	pool, err := database.NewPool(context.Background())
	if err != nil {
		logger.Error("gagal terhubung ke database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	// 3. Perakitan dependensi: repository -> service
	studentRepo := repository.NewStudentRepository(pool)
	userRepo := repository.NewUserRepository(pool)
	tokenRepo := repository.NewTokenRepository(pool)
	roleRepo := repository.NewRoleRepository(pool)

	jwtSecret := config.GetEnv("JWT_SECRET", "changeme")
	jwtIssuer := config.GetEnv("JWT_ISSUER", "api-students")
	jwtTTL, _ := time.ParseDuration(config.GetEnv("JWT_ACCESS_TTL", "15m"))
	refreshTTL, _ := time.ParseDuration(config.GetEnv("JWT_REFRESH_TTL", "168h"))

	jwtManager := helper.NewJWTManager(jwtSecret, jwtIssuer, jwtTTL)

	rawPerms, err := roleRepo.LoadPermissions(context.Background())
	if err != nil {
		logger.Error("gagal memuat permission", slog.String("error", err.Error()))
		os.Exit(1)
	}
	perms := helper.NewPermissionSet(rawPerms)
	logger.Info("permission dimuat", slog.Any("roles", perms.KnownRoles()))

	studentService := service.NewStudentService(studentRepo, perms)
	authService := service.NewAuthService(userRepo, tokenRepo, jwtManager, refreshTTL, perms)

	deps := route.Dependencies{
		Pool:           pool,
		JWT:            jwtManager,
		AuthService:    authService,
		StudentService: studentService,
		Permissions:    perms,
	}

	// 4. Aplikasi Fiber
	app := config.NewApp(logger, deps)
	port := config.GetEnv("APP_PORT", "3000")

	go func() {
		if err := app.Listen(":" + port); err != nil {
			logger.Error("server berhenti", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()
	logger.Info("server berjalan", slog.String("port", port))

	// 5. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("sinyal berhenti diterima, menutup server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		logger.Error("gagal menutup server dengan rapi",
			slog.String("error", err.Error()))
	}
	logger.Info("server berhenti dengan rapi")
}
