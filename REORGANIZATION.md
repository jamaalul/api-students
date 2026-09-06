# Panduan Langkah Demi Langkah Restrukturisasi Clean Architecture

**Proyek:** `api-students`  
**Referensi:** [ca.md](ca.md) (Modul Pertemuan 4 — Clean Architecture)  
**Tujuan:** Merestrukturisasi kode proyek `api-students` ke dalam arsitektur berlapis baku (Clean Architecture) tanpa mengubah satu pun perilaku HTTP (rute, parameter, status code, dan bentuk respons JSON tetap 100% identik).

---

## Daftar Isi
1. [Ringkasan Pemetaan Arsitektur](#ringkasan-pemetaan-arsitektur)
2. [Langkah 0 — Setup Folder, Dependency, dan .gitignore](#langkah-0--setup-folder-dependency-dan-gitignore)
3. [Langkah 1 — helper/: Presenter dan Request Parser](#langkah-1--helper-presenter-dan-request-parser)
4. [Langkah 2 — app/service/: Pure Business Rules & Unit Tests](#langkah-2--appservice-pure-business-rules--unit-tests)
5. [Langkah 3 — app/service/: HTTP Service / Controller](#langkah-3--appservice-http-service--controller)
6. [Langkah 4 — middleware/: Stack Middleware & Structured Logger](#langkah-4--middleware-stack-middleware--structured-logger)
7. [Langkah 5 — route/: Routing & Health Check](#langkah-5--route-routing--health-check)
8. [Langkah 6 — config/: Rotating Logger & App Assembler](#langkah-6--config-rotating-logger--app-assembler)
9. [Langkah 7 — main.go & Pembersihan File Lama](#langkah-7--maingo--pembersihan-file-lama)
10. [Langkah 8 — Pengujian, Verifikasi, dan Pengecekan Kebocoran Layer](#langkah-8--pengujian-verifikasi-dan-pengecekan-kebocoran-layer)

---

## Ringkasan Pemetaan Arsitektur

### Struktur Folder Target
```
api-students/
├── app/
│   ├── model/
│   │   └── student.go                 (Tetap: Layer 1 Entities)
│   ├── repository/
│   │   └── student_repository.go      (Tetap: Layer 3 Interface Adapters / Gateway)
│   └── service/                       (BARU)
│       ├── student_rules.go           (Layer 2 Use Cases: Pure Business Rules)
│       ├── student_rules_test.go      (Unit Tests Pure Business Rules)
│       └── student_service.go         (Layer 3 Interface Adapters: Controller)
├── config/
│   ├── app.go                         (BARU: Perakitan Fiber, Middleware, Route)
│   ├── env.go                         (Tetap: Pembacaan variabel environment)
│   └── logger.go                      (BARU: Structured JSON logger + lumberjack rotasi)
├── database/
│   └── postgres.go                    (Tetap: Layer 4 Frameworks & Drivers)
├── helper/                            (BARU: Layer 3 Presenter & Request Parser)
│   ├── request.go                     (Context timeout, ID validator, Query parser)
│   └── response.go                    (WebResponse envelope formatting)
├── logs/                              (BARU: Output berkas log rotasi, di-gitignore)
├── middleware/                        (BARU: Layer 4 Drivers & Middleware)
│   └── middleware.go                  (requestid, recover, helmet, cors, RequestLogger, RequireJSON)
├── route/                             (BARU: Layer 4 Routing)
│   └── route.go                       (Mapping URL /api/v1/health dan /api/v1/students)
├── .env / .env.example                (Tetap)
├── .gitignore                         (Diperbarui: ignore logs/)
├── go.mod / go.sum                    (Diperbarui: dependency lumberjack)
├── handler.go                         (DIHAPUS setelah migrasi ke service/)
├── helper.go                          (DIHAPUS setelah migrasi ke helper/)
└── main.go                            (DIPERBARUI: Hanya urutan perakitan & graceful shutdown)
```

### Dependency Rule
Setiap package hanya boleh mengimpor package di layer yang lebih dalam:
- `app/model` $\leftarrow$ Tidak boleh mengimpor package proyek sendiri
- `app/repository` $\leftarrow$ `app/model`
- `app/service` $\leftarrow$ `app/model`, `app/repository`, `helper`
- `helper` $\leftarrow$ `app/model`
- `middleware` $\leftarrow$ `helper`
- `route` $\leftarrow$ `app/service`, `middleware`, `helper`
- `config` $\leftarrow$ `app/service`, `route`, `middleware`, `helper`
- `database` $\leftarrow$ `config`
- `main.go` $\leftarrow$ `config`, `database`, `app/repository`, `app/service`

---

## Langkah 0 — Setup Folder, Dependency, dan .gitignore

Siapkan struktur folder baru, tambahkan dependensi log rotasi, dan cegah folder log runtime masuk ke git repository.

### 1. Buat folder baru
Jalankan di terminal PowerShell:
```powershell
mkdir helper, middleware, route, logs, app/service -Force
```

### 2. Pasang library rotasi log (lumberjack)
```powershell
go get gopkg.in/natefinch/lumberjack.v2
```

### 3. Tambahkan `logs/` ke `.gitignore`
Buka file `.gitignore` dan tambahkan baris berikut di bagian akhir:
```gitignore
# Output logs
logs/
```

- **Commit Message:**
  ```text
  chore: setup folder structure, add lumberjack dependency, and gitignore logs
  ```

---

## Langkah 1 — helper/: Presenter dan Request Parser

Fungsi pembantu yang ada di file root `helper.go` (sebelumnya berada di `package main`) dipecah menjadi dua file di package `helper`. Karena berada di package baru, nama fungsi diawali huruf kapital agar berstatus *exported*.

### Pemetaan Fungsi:
| Asal di `helper.go` | Nama Baru di `helper/` | File Target |
|---|---|---|
| `ok` | `Success` | `helper/response.go` |
| `okList` | `SuccessList` | `helper/response.go` |
| `created` | `Created` | `helper/response.go` |
| `noContent` | `NoContent` | `helper/response.go` |
| `fail` | `Fail` | `helper/response.go` |
| `failValidation` | `FailValidation` | `helper/response.go` |
| `reqCtx` | `RequestContext` | `helper/request.go` |
| `paramID` (dari `handler.go`) | `ParamID` | `helper/request.go` |
| `allowedSort`, `parseListQuery` | `allowedSort`, `ParseListQuery` | `helper/request.go` |

> [!IMPORTANT]
> Batas maksimal limit `50` (`q.Limit > 50`), filter `min_grade`, `max_grade`, dan sorting `id, nim, name, grade, created_at` dipertahankan persis sesuai implementasi pertemuan 3.

### 1. Buat file `helper/response.go`
```go
package helper

import (
	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
)

func Success(c *fiber.Ctx, status int, message string, data any) error {
	return c.Status(status).JSON(model.WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func SuccessList(c *fiber.Ctx, message string, data any, meta *model.Meta) error {
	return c.Status(fiber.StatusOK).JSON(model.WebResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// Created mengirim 201 sekaligus memasang header Location.
func Created(c *fiber.Ctx, message string, data any, location string) error {
	c.Set("Location", location)
	return c.Status(fiber.StatusCreated).JSON(model.WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

func Fail(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(model.WebResponse{
		Success: false,
		Message: message,
	})
}

func FailValidation(c *fiber.Ctx, errs map[string]string) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(model.WebResponse{
		Success: false,
		Message: "validasi gagal",
		Errors:  errs,
	})
}
```

### 2. Buat file `helper/request.go`
```go
package helper

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
)

// RequestContext memberi timeout untuk setiap operasi database.
func RequestContext(c *fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.UserContext(), 5*time.Second)
}

// ParamID membaca parameter :id dari jalur URL dan memastikan nilainya angka positif.
func ParamID(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return 0, false
	}
	return id, true
}

var allowedSort = map[string]bool{
	"id":         true,
	"nim":        true,
	"name":       true,
	"grade":      true,
	"created_at": true,
}

// ParseListQuery membaca query string dan memberi nilai bawaan yang aman.
func ParseListQuery(c *fiber.Ctx) model.ListQuery {
	q := model.ListQuery{
		Page:   c.QueryInt("page", 1),
		Limit:  c.QueryInt("limit", 10),
		Search: strings.TrimSpace(c.Query("search")),
		Sort:   c.Query("sort", "id"),
		Order:  strings.ToLower(c.Query("order", "asc")),
	}
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 {
		q.Limit = 10
	}
	if q.Limit > 50 {
		q.Limit = 50
	}
	if !allowedSort[q.Sort] {
		q.Sort = "id"
	}
	if q.Order != "desc" {
		q.Order = "asc"
	}
	if raw := c.Query("is_active"); raw != "" {
		if v, err := strconv.ParseBool(raw); err == nil {
			q.IsActive = &v
		}
	}
	if raw := c.Query("min_grade"); raw != "" {
		if v, err := strconv.ParseFloat(raw, 64); err == nil {
			q.MinGrade = &v
		}
	}
	if raw := c.Query("max_grade"); raw != "" {
		if v, err := strconv.ParseFloat(raw, 64); err == nil {
			q.MaxGrade = &v
		}
	}
	return q
}
```

- **Verifikasi**:
  ```powershell
  go vet ./helper/...
  ```
- **Commit Message:**
  ```text
  refactor(helper): extract response envelope and request parsing helpers
  ```

---

## Langkah 2 — app/service/: Pure Business Rules & Unit Tests

Mengekstrak aturan bisnis murni yang tidak menyentuh Fiber maupun database dari `handler.go`. Aturan ini berupa fungsi standar Go yang menerima struct dan mengembalikan hasil/error.

### Pemetaan Logika:
| Logika Asal di `handler.go` | Fungsi Murni Baru di `app/service/student_rules.go` |
|---|---|
| Validasi POST (baris 89–101) | `ValidateCreate(req model.CreateStudentRequest) map[string]string` |
| Validasi PUT (baris 128–137) | `ValidateReplace(req model.ReplaceStudentRequest) map[string]string` |
| Validasi & mutasi PATCH (baris 169–185) | `ApplyPatch(current model.Student, req model.PatchStudentRequest) (model.Student, map[string]string)` |
| Pemeriksaan PATCH kosong (baris 161) | `IsEmptyPatch(req model.PatchStudentRequest) bool` |
| Perhitungan total halaman (baris 51–54) | `CountTotalPages(total, limit int) int` |

### 1. Buat file `app/service/student_rules.go`
```go
package service

import (
	"strings"

	"api-students/app/model"
)

// File ini berisi business rules MURNI: tidak menyentuh fiber.Ctx,
// tidak menyentuh database, dan tidak tahu apa pun tentang HTTP.

// ValidateCreate memeriksa isi permintaan pembuatan student.
func ValidateCreate(req model.CreateStudentRequest) map[string]string {
	errs := map[string]string{}
	if strings.TrimSpace(req.NIM) == "" {
		errs["nim"] = "wajib diisi"
	}
	if strings.TrimSpace(req.Name) == "" {
		errs["name"] = "wajib diisi"
	}
	if req.Grade < 0 || req.Grade > 100 {
		errs["grade"] = "harus di antara 0 dan 100"
	}
	return errs
}

// ValidateReplace memeriksa isi permintaan PUT (penggantian penuh).
func ValidateReplace(req model.ReplaceStudentRequest) map[string]string {
	errs := map[string]string{}
	if strings.TrimSpace(req.Name) == "" {
		errs["name"] = "wajib diisi pada PUT"
	}
	if req.Grade < 0 || req.Grade > 100 {
		errs["grade"] = "wajib diisi dan di antara 0-100 pada PUT"
	}
	return errs
}

// ApplyPatch menyalin field yang dikirim ke data student saat ini.
// Field yang bernilai nil dibiarkan apa adanya.
func ApplyPatch(
	current model.Student, req model.PatchStudentRequest,
) (model.Student, map[string]string) {
	errs := map[string]string{}

	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			errs["name"] = "tidak boleh kosong"
		} else {
			current.Name = *req.Name
		}
	}

	if req.Grade != nil {
		if *req.Grade < 0 || *req.Grade > 100 {
			errs["grade"] = "harus di antara 0 dan 100"
		} else {
			current.Grade = *req.Grade
		}
	}

	if req.IsActive != nil {
		current.IsActive = *req.IsActive
	}

	return current, errs
}

// IsEmptyPatch menandai permintaan PATCH yang tidak mengubah apa pun.
func IsEmptyPatch(req model.PatchStudentRequest) bool {
	return req.Name == nil && req.Grade == nil && req.IsActive == nil
}

// CountTotalPages membulatkan ke atas tanpa memakai bilangan pecahan.
func CountTotalPages(total, limit int) int {
	if limit <= 0 {
		return 0
	}
	return (total + limit - 1) / limit
}
```

### 2. Buat file `app/service/student_rules_test.go`
File ini berisi unit test untuk business rules murni tanpa membutuhkan server aktif ataupun koneksi database.
```go
package service

import (
	"testing"

	"api-students/app/model"
)

func TestCountTotalPages(t *testing.T) {
	cases := []struct {
		total, limit, want int
	}{
		{0, 10, 0},
		{1, 10, 1},
		{10, 10, 1},
		{11, 10, 2},
		{45, 10, 5},
	}
	for _, tc := range cases {
		if got := CountTotalPages(tc.total, tc.limit); got != tc.want {
			t.Errorf("total=%d limit=%d: harap %d, dapat %d", tc.total, tc.limit, tc.want, got)
		}
	}
}

func TestValidateCreate(t *testing.T) {
	// Kasus valid
	validReq := model.CreateStudentRequest{
		NIM:   "18221001",
		Name:  "Budi Santoso",
		Grade: 85.5,
	}
	if errs := ValidateCreate(validReq); len(errs) != 0 {
		t.Errorf("seharusnya valid, tetapi dapat error: %v", errs)
	}

	// Kasus tidak valid
	invalidReq := model.CreateStudentRequest{
		NIM:   "",
		Name:  "   ",
		Grade: 105.0,
	}
	errs := ValidateCreate(invalidReq)
	if errs["nim"] != "wajib diisi" {
		t.Errorf("error nim tidak sesuai: %v", errs["nim"])
	}
	if errs["name"] != "wajib diisi" {
		t.Errorf("error name tidak sesuai: %v", errs["name"])
	}
	if errs["grade"] != "harus di antara 0 dan 100" {
		t.Errorf("error grade tidak sesuai: %v", errs["grade"])
	}
}

func TestValidateReplace(t *testing.T) {
	req := model.ReplaceStudentRequest{
		Name:     "",
		Grade:    -5,
		IsActive: true,
	}
	errs := ValidateReplace(req)
	if errs["name"] != "wajib diisi pada PUT" {
		t.Errorf("error name tidak sesuai: %v", errs["name"])
	}
	if errs["grade"] != "wajib diisi dan di antara 0-100 pada PUT" {
		t.Errorf("error grade tidak sesuai: %v", errs["grade"])
	}
}

func TestApplyPatch(t *testing.T) {
	initial := model.Student{
		ID:       1,
		NIM:      "18221001",
		Name:     "Budi",
		Grade:    80.0,
		IsActive: true,
	}

	newName := "Budi Pratama"
	inactive := false
	patchReq := model.PatchStudentRequest{
		Name:     &newName,
		IsActive: &inactive,
	}

	updated, errs := ApplyPatch(initial, patchReq)
	if len(errs) != 0 {
		t.Fatalf("tidak boleh ada error pada patch valid: %v", errs)
	}
	if updated.Name != "Budi Pratama" {
		t.Errorf("nama tidak terupdate: %s", updated.Name)
	}
	if updated.IsActive != false {
		t.Errorf("is_active tidak terupdate: %v", updated.IsActive)
	}
	if updated.Grade != 80.0 {
		t.Errorf("grade tidak boleh berubah jika nil: %f", updated.Grade)
	}
}

func TestIsEmptyPatch(t *testing.T) {
	empty := model.PatchStudentRequest{}
	if !IsEmptyPatch(empty) {
		t.Error("seharusnya bernilai true untuk patch kosong")
	}

	val := "test"
	notEmpty := model.PatchStudentRequest{Name: &val}
	if IsEmptyPatch(notEmpty) {
		t.Error("seharusnya bernilai false jika ada field yang diisi")
	}
}
```

- **Verifikasi**:
  ```powershell
  go test ./app/service/... -v
  ```
- **Commit Message:**
  ```text
  feat(service): extract pure business rules and add unit tests
  ```

---

## Langkah 3 — app/service/: HTTP Service / Controller

Membuat `app/service/student_service.go` yang menggantikan `StudentHandler` dari `handler.go`. Service ini menerima `*fiber.Ctx`, melakukan parsing HTTP, mendelegasikan validasi ke `student_rules.go`, memanggil interface `repository.StudentRepository`, dan mengembalikan respons melalui helper.

### Pemetaan:
| Di `handler.go` | Di `app/service/student_service.go` |
|---|---|
| `type StudentHandler struct` | `type StudentService struct` |
| `NewStudentHandler(repo)` | `NewStudentService(repo repository.StudentRepository) *StudentService` |
| `terjemahkanError(c, err, msg)` | `translateError(c, err, msg)` |
| Method `(h *StudentHandler) List, Get, Create, Replace, Patch, Delete` | Method `(s *StudentService) List, Get, Create, Replace, Patch, Delete` |

### Buat file `app/service/student_service.go`
```go
package service

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
)

// StudentService memegang dua tanggung jawab sekaligus pada arsitektur penyederhanaan:
// menerima *fiber.Ctx (controller) dan mengorkestrasi alur use case.
type StudentService struct {
	repo repository.StudentRepository
}

// NewStudentService menerima interface, bukan struct konkret.
func NewStudentService(repo repository.StudentRepository) *StudentService {
	return &StudentService{repo: repo}
}

func (s *StudentService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	q := helper.ParseListQuery(c)
	list, total, err := s.repo.FindAll(ctx, q)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil data student")
	}

	return helper.SuccessList(c, "daftar student berhasil diambil", list, &model.Meta{
		Page:       q.Page,
		Limit:      q.Limit,
		Total:      total,
		TotalPages: CountTotalPages(total, q.Limit),
	})
}

func (s *StudentService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	student, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(c, err, "gagal mengambil data student")
	}
	return helper.Success(c, fiber.StatusOK, "student ditemukan", student)
}

func (s *StudentService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	if errs := ValidateCreate(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	baru, err := s.repo.Create(ctx, model.Student{
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: true,
	})
	if err != nil {
		return translateError(c, err, "gagal menyimpan student")
	}

	return helper.Created(c, "student berhasil dibuat", baru,
		"/api/v1/students/"+strconv.Itoa(baru.ID))
}

func (s *StudentService) Replace(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	var req model.ReplaceStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	if errs := ValidateReplace(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	hasil, err := s.repo.Update(ctx, model.Student{
		ID:       id,
		Name:     strings.TrimSpace(req.Name),
		Grade:    req.Grade,
		IsActive: req.IsActive,
	})
	if err != nil {
		return translateError(c, err, "gagal memperbarui student")
	}
	return helper.Success(c, fiber.StatusOK, "student berhasil diganti seluruhnya", hasil)
}

func (s *StudentService) Patch(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	var req model.PatchStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}
	if IsEmptyPatch(req) {
		return helper.Fail(c, fiber.StatusBadRequest, "tidak ada field yang diubah")
	}

	current, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(c, err, "gagal mengambil data student")
	}

	updated, errs := ApplyPatch(current, req)
	if len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	hasil, err := s.repo.Update(ctx, updated)
	if err != nil {
		return translateError(c, err, "gagal memperbarui student")
	}
	return helper.Success(c, fiber.StatusOK, "student berhasil diperbarui sebagian", hasil)
}

func (s *StudentService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return translateError(c, err, "gagal menghapus student")
	}
	return helper.NoContent(c)
}

// translateError memetakan sentinel error repository ke format HTTP.
func translateError(c *fiber.Ctx, err error, generalMessage string) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.Fail(c, fiber.StatusNotFound, "student tidak ditemukan")
	case errors.Is(err, repository.ErrDuplicate):
		return helper.Fail(c, fiber.StatusConflict, "NIM sudah terdaftar")
	default:
		return helper.Fail(c, fiber.StatusInternalServerError, generalMessage)
	}
}
```

- **Verifikasi**:
  ```powershell
  go vet ./app/service/...
  ```
- **Commit Message:**
  ```text
  refactor(service): migrate student handler to service layer
  ```

---

## Langkah 4 — middleware/: Stack Middleware & Structured Logger

Membuat package `middleware` yang memuat:
1. `Register(app, logger)`: Memasang middleware global (`requestid`, `recover`, `helmet`, `cors`, dan `RequestLogger`).
2. `RequestLogger(logger)`: Mencatat info request dalam bentuk structured JSON.
3. `RequireJSON`: Validasi Content-Type yang dipindahkan dari `main.go`.

### Buat file `middleware/middleware.go`
```go
package middleware

import (
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"api-students/helper"
)

// Register memasang seluruh middleware yang berlaku untuk semua route.
// Urutan pemasangan sangat penting.
func Register(app *fiber.App, logger *slog.Logger) {
	app.Use(requestid.New())       // 1. Beri ID unik tiap request
	app.Use(recover.New())         // 2. Tangkap panic
	app.Use(helmet.New())          // 3. Pasang header keamanan dasar
	app.Use(cors.New())            // 4. Konfigurasi CORS
	app.Use(RequestLogger(logger)) // 5. Catat log terstruktur
}

// RequestLogger mencatat setiap HTTP request ke structured logger (JSON).
func RequestLogger(logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		requestID, _ := c.Locals("requestid").(string)
		logger.Info("http_request",
			slog.String("request_id", requestID),
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", c.Response().StatusCode()),
			slog.Duration("duration", time.Since(start)),
			slog.String("ip", c.IP()),
		)
		return err
	}
}

var methodsWithBody = map[string]bool{
	fiber.MethodPost:  true,
	fiber.MethodPut:   true,
	fiber.MethodPatch: true,
}

// RequireJSON menolak request POST/PUT/PATCH bila Content-Type bukan application/json.
func RequireJSON(c *fiber.Ctx) error {
	if methodsWithBody[c.Method()] {
		ct := c.Get("Content-Type")
		if !strings.HasPrefix(ct, fiber.MIMEApplicationJSON) {
			return helper.Fail(c, fiber.StatusUnsupportedMediaType,
				"Content-Type harus application/json")
		}
	}
	return c.Next()
}
```

- **Verifikasi**:
  ```powershell
  go vet ./middleware/...
  ```
- **Commit Message:**
  ```text
  feat(middleware): add global middleware stack, structured logger, and require JSON
  ```

---

## Langkah 5 — route/: Routing & Health Check

Memindahkan seluruh pendaftaran rute dan endpoint `/health` dari `main.go` ke dalam package `route`.

### Buat file `route/route.go`
```go
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

// Register memetakan URL ke method service yang menanganinya.
func Register(app *fiber.App, pool *pgxpool.Pool, studentService *service.StudentService) {
	api := app.Group("/api/v1")
	api.Get("/health", healthCheck(pool))

	students := api.Group("/students", middleware.RequireJSON)
	students.Get("/", studentService.List)
	students.Get("/:id", studentService.Get)
	students.Post("/", studentService.Create)
	students.Put("/:id", studentService.Replace)
	students.Patch("/:id", studentService.Patch)
	students.Delete("/:id", studentService.Delete)
}

// healthCheck memeriksa kondisi server dan koneksi database.
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
```

- **Verifikasi**:
  ```powershell
  go vet ./route/...
  ```
- **Commit Message:**
  ```text
  refactor(route): extract routing and health check into route package
  ```

---

## Langkah 6 — config/: Rotating Logger & App Assembler

Menambahkan dua file pada folder `config/`:
1. `config/logger.go`: Logger berbasis `slog` dengan output ganda (`os.Stdout` dan `logs/app.log` dengan rotasi `lumberjack`).
2. `config/app.go`: Mengumpulkan konfigurasi Fiber, memasang error handler sentral, middleware, rute, dan fallback 404.

### 1. Buat file `config/logger.go`
```go
package config

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

// NewLogger membuat structured logger yang menulis ke stdout dan file logs/app.log dengan rotasi.
func NewLogger() *slog.Logger {
	if err := os.MkdirAll("logs", 0o755); err != nil {
		panic("gagal membuat folder logs: " + err.Error())
	}

	rotator := &lumberjack.Logger{
		Filename:   filepath.Join("logs", "app.log"),
		MaxSize:    10, // Rotasi setiap 10 MB
		MaxBackups: 5,  // Maksimal 5 backup
		MaxAge:     14, // Hapus file lebih dari 14 hari
		Compress:   true,
	}

	writer := io.MultiWriter(os.Stdout, rotator)
	handler := slog.NewJSONHandler(writer, &slog.HandlerOptions{
		Level: parseLevel(GetEnv("LOG_LEVEL", "info")),
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}

func parseLevel(value string) slog.Level {
	switch strings.ToLower(value) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
```

### 2. Buat file `config/app.go`
```go
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
```

- **Verifikasi**:
  ```powershell
  go vet ./config/...
  ```
- **Commit Message:**
  ```text
  feat(config): add rotating logger and application assembler
  ```

---

## Langkah 7 — main.go & Pembersihan File Lama

Sekarang seluruh fungsionalitas sudah berada di layer masing-masing. Bersihkan file lama di root dan perbarui `main.go`.

### 1. Hapus file lama di root
Hapus file berikut dari direktori root proyek:
- `handler.go`
- `helper.go`

*(Di PowerShell dapat dijalankan: `Remove-Item handler.go, helper.go`)*

### 2. Ganti isi `main.go`
File `main.go` sekarang hanya bertanggung jawab untuk:
1. Memuat env & logger
2. Menginisialisasi pool database
3. Merakit repository $\rightarrow$ service $\rightarrow$ app
4. Menjalankan server Fiber pada goroutine
5. Menangani graceful shutdown (menunggu signal SIGINT/SIGTERM)

```go
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
	studentRepository := repository.NewStudentRepository(pool)
	studentService := service.NewStudentService(studentRepository)

	// 4. Aplikasi Fiber
	app := config.NewApp(logger, pool, studentService)
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
```

### 3. Jalankan `go mod tidy`
```powershell
go mod tidy
```

- **Commit Message:**
  ```text
  refactor(main): streamline entry point with graceful shutdown and remove legacy files
  ```

---

## Langkah 8 — Pengujian, Verifikasi, dan Pengecekan Kebocoran Layer

### 1. Kompilasi & Tes Otomatis
Jalankan perintah berikut di terminal:
```powershell
go test ./... -v
go vet ./...
go build ./...
```
Semua perintah di atas harus keluar dengan exit code `0` tanpa error.

### 2. Matriks Pengujian Endpoint
Nyalakan server dengan `go run .` dan uji setiap endpoint untuk memastikan tidak ada perubahan perilaku:

| Endpoint | Method | Skenario Uji | Status Diharapkan | Keterangan Validasi |
|---|---|---|---|---|
| `/api/v1/health` | GET | Cek koneksi server & database | `200 OK` | `{"success":true,"message":"server dan database berjalan"}` |
| `/api/v1/students` | GET | Pagination `?limit=10&page=1` | `200 OK` | `meta.total_pages` dihitung oleh `CountTotalPages` |
| `/api/v1/students/:id` | GET | ID valid / tidak ada | `200` / `404` | Pesan `"student ditemukan"` atau `"student tidak ditemukan"` |
| `/api/v1/students` | POST | Request tanpa header `Content-Type: application/json` | `415 Unsupported Media Type` | Ditolak oleh `middleware.RequireJSON` |
| `/api/v1/students` | POST | Request dengan field kosong atau nilai `grade` > 100 | `422 Unprocessable Entity` | Peta field error berasal dari `ValidateCreate` |
| `/api/v1/students` | POST | Request valid baru | `201 Created` | Header `Location: /api/v1/students/:id` terpasang |
| `/api/v1/students` | POST | NIM yang sudah ada di database | `409 Conflict` | Pesan `"NIM sudah terdaftar"` |
| `/api/v1/students/:id` | PUT | Data lengkap valid | `200 OK` | Validasi PUT melalui `ValidateReplace` |
| `/api/v1/students/:id` | PATCH | Request `{}` tanpa field perubahan | `400 Bad Request` | Pesan `"tidak ada field yang diubah"` |
| `/api/v1/students/:id` | PATCH | Update parsial (misal `grade`) | `200 OK` | Nilai field diupdate via `ApplyPatch` |
| `/api/v1/students/:id` | DELETE | ID valid / ID tidak ditemukan | `204 NoContent` lalu `404` | Penghapusan data konsisten |
| `/random-endpoint` | GET | URL acak yang tidak ada | `404 Not Found` | Pesan `"endpoint tidak ditemukan"` dari fallback 404 |

### 3. Pengecekan Kebocoran Layer (Sesuai ca.md Bagian C.5)
Jalankan perintah berikut untuk membuktikan tidak ada kebocoran arsitektur:
```powershell
go list -deps ./app/repository
```
- [x] **`app/repository`**: Pastikan **tidak ada `github.com/gofiber/fiber`** pada daftar dependencies repository.
- [x] **`app/model`**: Tidak mengimpor package internal mana pun (hanya `time` dari standard library).
- [x] **`app/service/student_rules.go`**: Murni standard library (`strings`) dan `app/model`, tidak mengimpor Fiber maupun database.
- [x] **`route/route.go`**: Tidak ada validasi atau query SQL.
- [x] **`main.go`**: Tidak ada handler HTTP atau query database; murni perakitan.
- [x] **`logs/app.log`**: Berkas terbuat otomatis dan mencatat setiap request dalam format JSON ber-`request_id`.
