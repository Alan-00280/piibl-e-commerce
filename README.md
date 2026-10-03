# PiiBL E-Commerce REST API

REST API untuk sistem E-Commerce multi-store yang memungkinkan Customer membeli produk dari berbagai Store, sementara Tenant dapat mengelola Store, produk, variant, dan stok.

Nama PiiBL diambil dari PBL yang dimana jika dalam bahasa inggris di *prounounce* menjadi *\pibiel\\*

> Project ini dibuat sebagai tugas kuliah.

## Features

* Guest dapat melihat produk, kategori, dan Store.
* Customer dapat melakukan checkout tanpa Cart yang tersimpan di backend.
* Customer dapat membeli produk dari beberapa Store dalam satu checkout.
* Tenant dapat memiliki maksimal satu Store.
* Tenant dapat mengelola Product, Product Variant, dan stok.
* Product memiliki satu stok utama yang tidak bergantung pada variant.
* Harga dihitung oleh backend berdasarkan variant yang dipilih.
* Seluruh transaksi menggunakan mata uang IDR.
* Order memiliki status `CREATED`, `COMPLETED`, dan `CANCELLED`.
* OrderItem disimpan sebagai immutable transaction record.
* Customer dapat memberikan satu review untuk setiap Product yang pernah dibeli.
* Review yang dikirim kembali untuk Product yang sama akan memperbarui review sebelumnya.
* Pembatalan Order tidak mengembalikan stok secara otomatis.

## Roles

### Guest

* Melihat daftar produk.
* Melihat detail produk.
* Melihat kategori dan Store.

### Customer

* Melakukan checkout.
* Melihat pesanan.
* Membatalkan pesanan.
* Melihat riwayat pesanan.
* Memberikan dan memperbarui review produk yang pernah dibeli.

### Tenant

* Memiliki maksimal satu Store.
* Mengelola Store.
* Mengelola Product dan Product Variant.
* Mengelola stok.
* Melihat pesanan yang masuk ke Store.

## Product & Variant

Setiap Product memiliki satu Product Variant khusus bernama `_default`.

Stok hanya disimpan pada level Product dan tidak bergantung pada variant.

Harga mengikuti aturan:

* Tidak memilih variant → menggunakan harga `_default`.
* Memilih variant → harga merupakan total harga seluruh variant yang dipilih.
* Harga `_default` tidak ikut dihitung jika terdapat variant yang dipilih.

Contoh:

```text
_default = 50.000
Merah    = 5.000
Ukuran M = 10.000

Merah + M = 15.000
```

## Order

Customer tidak memiliki Cart yang disimpan di backend. Item checkout dikirim langsung ke API dan seluruh perhitungan dilakukan oleh backend.

Order memiliki tiga status:

```text
CREATED
COMPLETED
CANCELLED
```

`OrderItem` merupakan immutable record yang menyimpan informasi transaksi pada saat pembelian, seperti:

* Product name
* Variant
* Price
* Quantity
* Subtotal

OrderItem tidak bergantung pada Product atau Product Variant setelah transaksi dibuat, sehingga perubahan katalog tidak mengubah riwayat pesanan.

## Review

Customer hanya dapat memberikan review terhadap Product yang pernah dibeli.

Satu Customer hanya memiliki satu review untuk satu Product. Jika Customer mengirim review kembali untuk Product yang sama, review sebelumnya akan diperbarui.

## Out of Scope

Fitur berikut tidak termasuk dalam scope project:

* Payment
* Address
* Shipping dan tracking
* Return/refund
* Customer–Tenant chat
* Image upload/storage
* Notification system
* Persistent Cart
* Idempotency

## Currency

Seluruh harga dan transaksi menggunakan:

```text
IDR (Indonesian Rupiah)
```

Nilai uang direpresentasikan menggunakan integer, misalnya:

```text
50000 = Rp50.000
```

## API

API menggunakan arsitektur REST dan akan menggunakan versioning:

```text
/api/v1
```

Dokumentasi endpoint akan ditambahkan seiring implementasi project.

## Project Status

🚧 **In Development**

Project ini sedang dalam tahap pengembangan dan struktur API/database dapat berubah selama proses implementasi.
