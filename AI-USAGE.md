# Penggunaan AI

AI digunakan sebagai alat bantu dalam pengembangan dan pengujian proyek ini.

## Bantuan yang diberikan

AI membantu menyusun test untuk business rules dan validasi entitas User pada
`app/service/user_rules_test.go`, yang mencakup:

- Perubahan field User melalui patch, termasuk trim pada username.
- Mempertahankan field yang tidak dikirim dan menerima perubahan status aktif
  eksplisit menjadi `false`.
- Membedakan patch kosong dari patch yang memiliki field.
- Validasi data pembuatan User, termasuk username, email, serta password yang
  lemah (terlalu pendek, tidak mengandung huruf atau angka, atau termasuk
  password umum).
- Validasi field username dan email pada patch User.

## Verifikasi

Test dijalankan dengan perintah berikut:

```sh
go test ./...
```

Seluruh test lulus saat verifikasi dilakukan. Hasil dan perubahan tetap perlu
ditinjau oleh pengembang proyek.