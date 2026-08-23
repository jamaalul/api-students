# api-students

REST API sederhana untuk data mahasiswa. Dibuat dengan Go + Fiber v2.

## Menjalankan

\`\`\`bash
go run .
\`\`\`

## Kontrak API

| Metode | Endpoint | Parameter | Contoh Body | Status | Contoh Respons |
|---|---|---|---|---|---|
| GET | /api/v1/students | page, limit, search, sort, order, is_active, min_grade, max_grade | - | 200 | `{"success":true,"data":[...],"meta":{...}}` |
| GET | /api/v1/students/:id | - | - | 200 / 404 | `{"success":true,"data":{...}}` |
| POST | /api/v1/students | - | `{"nim":"2201","name":"Sari","grade":85}` | 201 / 409 / 422 | `{"success":true,"data":{...}}` |
| PUT | /api/v1/students/:id | - | `{"name":"Sari","grade":90,"is_active":true}` | 200 / 404 / 422 | `{"success":true,"data":{...}}` |
| PATCH | /api/v1/students/:id | - | `{"is_active":false}` | 200 / 404 / 422 | `{"success":true,"data":{...}}` |
| DELETE | /api/v1/students/:id | - | - | 204 / 404 | (tanpa body) |

## Catatan Desain
- NIM tidak bisa diubah lewat PUT/PATCH karena berperan sebagai identitas.
- `limit` dibatasi maksimum 50 per halaman.
- `sort` memakai daftar putih: `id, nim, name, grade, created_at`.