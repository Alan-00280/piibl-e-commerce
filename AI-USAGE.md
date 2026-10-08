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

## Aktivitas AI pada sesi 7 Oktober 2026

AI menambahkan test pada `app/service/order_rules_test.go` untuk memastikan
`groupItemPerStore` mengelompokkan item pembelian berdasarkan toko dan
`countTotalSubtotal` menghitung subtotal serta total untuk pembelian satu item.
Test package service dijalankan dengan `go test ./app/service` dan lulus.

AI mencatat temuan `groupItemPerStore` yang sebelumnya tidak memasukkan order
hasil pengelompokan ke slice hasil pada `logs/bug.log`. Setelah implementasi
yang ada diverifikasi mengembalikan order dan test pengelompokan lulus, status
temuan di log diperbarui menjadi diperbaiki.

AI juga mencatat investigasi kegagalan GET `/users`: `PermissionSet.Can`
menerima `PermissionSet` nil karena field `Permission` belum diisi saat
membentuk `routes.Dependencies`. Pemeriksaan berikutnya menemukan
`Permission: permissionSet` sudah tercantum di `main.go`; status dan penyebab
tersebut dicatat di `logs/bug.log`.

AI mengisi migration `migrations/013_user_delete_perms.sql` untuk menambahkan
permission `user:delete` dan memberikannya kepada `ADMIN`, serta
`migrations/014_role_assign_perms.sql` untuk memastikan permission
`role:assign` tersedia dan memberikannya kepada `ADMIN`. Kedua migration
menggunakan `ON CONFLICT` agar operasi idempoten. Migration disiapkan, tetapi
tidak dijalankan ke database.

## Aktivitas AI pada sesi 8 Oktober 2026

AI membantu membaca dan menganalisis arsitektur basis kode saat ini, lalu menyusun layer repository data access untuk Store dan Product:

- Mengimplementasikan `app/repository/store_repository.go` dengan mengikuti pola pemrograman repository yang sudah ada (`user_repository.go`), mencakup:
  - Definisi interface `StoreRepository` (`FindByID`, `FindByOwnerID`, `FindAll`, `Create`, `Update`, `Delete`).
  - Implementasi struct `storePostgresRepository` menggunakan connection pool PostgreSQL (`pgxpool.Pool`).
  - Pemetaan error standar: `ErrNotFound` untuk `pgx.ErrNoRows` dan `ErrDuplicate` untuk pelanggaran constraint unik `tenant_id`.
  - Fungsi pembantu `buildFilterStore` untuk filter pencarian, status aktif, dan owner ID, serta scanner `scanStore`.
- Mengimplementasikan `app/repository/product_repository.go` yang mencakup:
  - Definisi interface `ProductRepository` dan implementasi `productPostgresRepository`.
  - Operasi CRUD entitas Product (`FindByID`, `FindAll`, `Create`, `Update`, `Delete`).
  - Manajemen stok dan status (`UpdateStock`, `UpdateStatus`), serta operasi atomik `DecreaseStock` dengan jaring pengaman `WHERE stock >= $1` untuk mencegah race condition.
  - Manajemen Variant Space (`CreateVariantSpace`, `FindVariantSpaceByID`, `FindVariantSpacesByProductID`, `UpdateVariantSpace`, `DeleteVariantSpace`).
  - Manajemen Product Variant (`CreateVariant`, `FindVariantByID`, `FindVariantsByProductID`, `FindVariantsByIDs`, `FindDefaultVariant`, `UpdateVariant`, `DeleteVariant`).
- Memperbarui `model.ListQuery` pada `app/model/http.go` dengan menambahkan field pointer `*StoreFilter` dan `*ProductFilter` agar kompatibel dengan query pencarian repository Store dan Product.
- Memverifikasi integritas kompilasi seluruh paket dengan `go build ./...` yang menghasilkan status sukses tanpa error.
- Memecah implementasi repository Product berdasarkan tanggung jawab menjadi tiga file dalam package yang sama: `app/repository/product_repository.go` untuk Product, `app/repository/variant_space_repository.go` untuk Variant Space, dan `app/repository/product_variant_repository.go` untuk Product Variant. Interface serta pemanggil repository tetap sama; perubahan hanya mengatur ulang lokasi implementasinya.
- Validasi setelah pemisahan: `go build ./...` berhasil. `go test ./...` masih gagal pada `TestGroupItemPerStore` (mengembalikan 0 order, ekspektasi 2) dan `TestCreateUserValidation/valid_user` (field role wajib diisi); kegagalan tersebut berada di luar perubahan repository ini.
- Memperbarui `TestCreateStoreValidation` di `app/service/store_rules_test.go` mengikuti bentuk baru `CreateStoresReq`, yang tidak lagi memuat `OwnerID`, dan menambahkan pengujian panjang nama maksimum. Test belum dapat dijalankan karena `store_repository.go` masih mengakses `Stores.OwnerID`, sedangkan model menggunakan `TenantID`.
- Memperbarui filter Product agar menangani store, kategori, status, rentang harga berdasarkan harga variant `_default`, dan rentang rating rata-rata review. Pointer filter opsional diperiksa sebelum digunakan.
- Mengubah `ProductRepository.FindByID` dan `FindAll` agar mengembalikan `ProductBasePrice`/`[]ProductBasePrice`. Hasil menyertakan harga dasar `_default` dan `OverallRating`; pengurutan `price` dan `rating` menggunakan kedua nilai tersebut.
- Verifikasi perubahan filter dan response Product: `go test ./app/repository` dan `go build ./...` berhasil.
- Melengkapi method `ProductService.Deactivate`, `Restock`, dan `Reprice` dengan validasi request, pengecekan autentikasi dan kepemilikan tenant, serta pemanggilan repository untuk menonaktifkan produk, memperbarui stok, atau memperbarui harga variant `_default`. Build berhasil; test package service terhalang kegagalan kompilasi fixture test lama yang menggunakan `int` untuk `CategoryID` bertipe `*int`.
- Mendaftarkan rute Product di `routes/route.go`: endpoint publik untuk daftar dan detail, endpoint berautentikasi untuk create/update/deactivate/restock/reprice, serta daftar produk tenant pada `/api/v1/my/products`. Rute request-body menggunakan middleware `RequireJSON`. Sesuai permintaan, test dan linter tidak dijalankan untuk perubahan rute.
- Melengkapi `CreateSVariant`, `PatchSVariant`, dan `DeleteSVaraint` pada `app/service/product_variants.go` dengan validasi request, autentikasi, pemeriksaan kepemilikan atau permission, serta operasi repository. Patch kosong ditolak mengikuti pola `IsEmptyPatchProduct`. Pengecekan harga menggunakan `checkVariantAdjustmentPrice` dengan adjustment terendah yang dapat dipilih dari tiap variant space; `go build ./...` berhasil.
- Melengkapi `CreateVariant`, `PatchVariant`, dan `DeleteVariant` pada `app/service/product_variants.go`. Operasi memastikan hak akses pemilik/permission, memastikan variant terikat ke product dan variant space yang sesuai, menolak patch kosong dan nama reserved `_default`, serta memvalidasi harga terendah yang mungkin sebelum create/update menggunakan `checkVariantAdjustmentPrice`. `go build ./...` berhasil.
- Menyesuaikan `CreateVariant` agar menerima dan memproses banyak `ProductVariantItem` dalam satu request, memvalidasi setiap item sebelum penyimpanan, serta mengembalikan semua variant yang dibuat. Fixture dan validasi test Product Variant diperbarui mengikuti bentuk slice.
- Membuat migration `migrations/016_drop_product_variant_price_constraint.sql` untuk menghapus constraint `chk_product_variants_price`; validasi harga kini ditangani pada service oleh `validateProductVariantPricesWithChange`. Migration disiapkan dan tidak dijalankan ke database.

## Verifikasi

Test dijalankan dengan perintah berikut:

```sh
go test ./...
```

Seluruh test lulus pada verifikasi terakhir. Hasil dan perubahan tetap perlu
ditinjau oleh pengembang proyek.

--- 
Menggunakan AI CHAT GPT untuk membuatkan migration store hingga product_reviews