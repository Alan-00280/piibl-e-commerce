package service

// ISSUE:
// ID variant yang hilang atau ganda tersaring diam-diam.
// Domain hanya melihat variant yang sudah dimuat dari database.
// ID yang tidak ada tidak muncul, dan ID ganda menjadi satu baris.
// Akibatnya request dengan variant_ids:[999999] untuk ruang
// opsional dianggap "tanpa variant" dan dibeli dengan harga dasar.
// Perbaikannya di service: tolak bila ada ID ganda di request,
// dan tolak bila jumlah variant yang ditemukan tidak sama
// dengan jumlah ID unik yang diminta.
