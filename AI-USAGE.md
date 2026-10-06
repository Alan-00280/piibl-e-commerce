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

AI juga membantu menyusun test validasi entitas Store pada
`app/service/store_rules_test.go`, untuk owner ID, nama, deskripsi, serta
perubahan status aktif. Agar aturan nama toko dapat divalidasi, AI
mendaftarkan validator `alphanumspace` yang hanya menerima huruf, angka, dan
spasi.

AI membantu menyusun test validasi entitas Product pada
`app/service/product_rules_test.go`, meliputi validasi create dan patch,
store/category ID, nama dan deskripsi, rentang stok, serta status produk
termasuk request khusus untuk perubahan status dan stok.

AI juga membantu menyusun test business rules Product Variant pada
`app/service/product_variant_rules_test.go`, meliputi variant space, variant
item, validasi field opsional, ID, nama, dan batas penyesuaian harga. Request
pembuatan variant space menggunakan pointer untuk `is_mandatory` agar nilai
`false` yang dikirim dapat dibedakan dari field yang tidak dikirim.

AI membantu menyusun test untuk aturan order pada
`app/service/order_rules_test.go`, termasuk kalkulasi subtotal serta validasi
dan deteksi variant terpilih. Saat implementasi `duplicateChoosenVariants`
diperbarui untuk mengembalikan sentinel `-2` ketika menemukan `_default`,
pengecekan mengonfirmasi kasus nil pointer dereference sebelumnya tidak lagi
terjadi. Ekspektasi test telah diselaraskan dengan perilaku tersebut, dan
detail temuan serta statusnya dicatat di `logs/bug.log`.

## Verifikasi

Test dijalankan dengan perintah berikut:

```sh
go test ./...
```

Seluruh test lulus pada verifikasi terakhir. Hasil dan perubahan tetap perlu
ditinjau oleh pengembang proyek.