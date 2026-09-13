package service

import (
	"testing"

	"api-students/app/model"
)

func TestValidateRegister(t *testing.T) {
	// Kasus valid
	validReq := model.RegisterRequest{
		Username: "budi_santoso",
		Email:    "budi@example.com",
		Password: "Rahasia1",
	}
	if errs := ValidateRegister(validReq); len(errs) != 0 {
		t.Errorf("seharusnya valid, tetapi dapat error: %v", errs)
	}

	// Username kosong
	errs := ValidateRegister(model.RegisterRequest{
		Username: "",
		Email:    "budi@example.com",
		Password: "Rahasia1",
	})
	if errs["username"] != "wajib diisi" {
		t.Errorf("error username tidak sesuai: %v", errs["username"])
	}

	// Username terlalu pendek
	errs = ValidateRegister(model.RegisterRequest{
		Username: "ab",
		Email:    "budi@example.com",
		Password: "Rahasia1",
	})
	if errs["username"] != "minimal 3 karakter" {
		t.Errorf("error username tidak sesuai: %v", errs["username"])
	}

	// Username karakter tidak valid
	errs = ValidateRegister(model.RegisterRequest{
		Username: "budi!@#",
		Email:    "budi@example.com",
		Password: "Rahasia1",
	})
	if errs["username"] != "hanya boleh huruf, angka, titik, dan garis bawah" {
		t.Errorf("error username tidak sesuai: %v", errs["username"])
	}

	// Email tidak valid
	errs = ValidateRegister(model.RegisterRequest{
		Username: "budi_santoso",
		Email:    "buditidakvalid.com",
		Password: "Rahasia1",
	})
	if errs["email"] != "format email tidak valid" {
		t.Errorf("error email tidak sesuai: %v", errs["email"])
	}

	// Password terlalu pendek
	errs = ValidateRegister(model.RegisterRequest{
		Username: "budi_santoso",
		Email:    "budi@example.com",
		Password: "abc123",
	})
	if errs["password"] != "minimal 8 karakter" {
		t.Errorf("error password tidak sesuai: %v", errs["password"])
	}

	// Password tidak ada huruf atau angka
	errs = ValidateRegister(model.RegisterRequest{
		Username: "budi_santoso",
		Email:    "budi@example.com",
		Password: "12345678",
	})
	if errs["password"] != "harus memuat huruf dan angka" {
		t.Errorf("error password tidak sesuai: %v", errs["password"])
	}

	// Password terlalu umum
	errs = ValidateRegister(model.RegisterRequest{
		Username: "budi_santoso",
		Email:    "budi@example.com",
		Password: "password1",
	})
	if errs["password"] != "password terlalu umum" {
		t.Errorf("error password tidak sesuai: %v", errs["password"])
	}
}

func TestValidateLogin(t *testing.T) {
	// Kasus valid
	validReq := model.LoginRequest{
		Username: "budi_santoso",
		Password: "Rahasia1",
	}
	if errs := ValidateLogin(validReq); len(errs) != 0 {
		t.Errorf("seharusnya valid, tetapi dapat error: %v", errs)
	}

	// Username kosong
	errs := ValidateLogin(model.LoginRequest{
		Username: "   ",
		Password: "Rahasia1",
	})
	if errs["username"] != "wajib diisi" {
		t.Errorf("error username tidak sesuai: %v", errs["username"])
	}

	// Password kosong
	errs = ValidateLogin(model.LoginRequest{
		Username: "budi_santoso",
		Password: "",
	})
	if errs["password"] != "wajib diisi" {
		t.Errorf("error password tidak sesuai: %v", errs["password"])
	}
}
