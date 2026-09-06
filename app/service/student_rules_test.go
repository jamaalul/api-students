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
