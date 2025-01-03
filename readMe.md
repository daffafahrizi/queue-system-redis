# Penjelasan Aplikasi Redis Queue untuk MergeRequest

Aplikasi ini menggunakan Redis untuk mengelola antrean (queue) `MergeRequest`, memungkinkan penambahan, pemrosesan, dan pengambilan status.

# Struktur Aplikasi

```plaintext
redistest/
├── main.go          # Entry point aplikasi
├── api/
│   └── handler.go   # Endpoint untuk menambah dan mengambil status MergeRequest
├── producer/
│   └── producer.go  # Menambahkan MergeRequest ke antrean Redis
├── consumer/
│   └── consumer.go  # Memproses MergeRequest dari antrean Redis
```


## Fitur Utama
1. **Menambahkan MergeRequest**:
   - `POST /merge-requests`: Menambahkan beberapa `MergeRequest` ke antrean Redis (`merge_request_queue`) menggunakan `RPush`.
2. **Memproses MergeRequest**:
   - Consumer membaca data dari antrean (`LPop`), memperbarui status tiap `Diff`, dan menyimpan hasilnya di Redis (`merge_request:<ID>`).
3. **Mengambil Status**:
   - `GET /merge-requests/status`: Menggabungkan data dari antrean dan data yang telah diproses di Redis.

## Cara Kerja
1. **Penambahan**:
   - Data `MergeRequest` dikirim melalui `POST /merge-requests` dan disimpan di Redis Queue.
2. **Pemrosesan**:
   - Consumer membaca antrean, memproses setiap `MergeRequest`, memperbarui status, lalu menyimpannya di Redis.
3. **Pengambilan Status**:
   - Endpoint `GET /merge-requests/status` membaca:
     - Antrean Redis untuk `MergeRequest` yang menunggu.
     - Redis Keys untuk `MergeRequest` yang telah diproses.

## Contoh Penggunaan
### a. Menambahkan MergeRequest
```bash
curl -X POST http://localhost:8080/merge-requests \
-H "Content-Type: application/json" \
-d '[{
  "id": 1,
  "title": "Fix bug",
  "diffs": [{"file_path": "auth.go", "change": "Updated login function", "status": "Pending"}],
  "status": "Pending"
}]'
```

### b. Mengambil Status MergeRequest

Gunakan endpoint `GET /merge-requests/status` untuk melihat semua status:

```bash
curl -X GET http://localhost:8080/merge-requests/status
```

## Alur Data

1. **Penambahan**:  
   `POST /merge-requests` → Redis Queue (`merge_request_queue`).

2. **Pemrosesan**:  
   Consumer membaca Redis Queue → Memperbarui status di Redis Keys (`merge_request:<ID>`).

3. **Pengambilan Status**:  
   `GET /merge-requests/status` → Menggabungkan data dari antrean dan Redis Keys.

## Diagram Alur

```plaintext
Client → POST /merge-requests → Redis Queue (merge_request_queue)
   |
   ↓
Consumer (ProcessMergeRequests) ← Queue
   |
   ↓
Redis Keys (merge_request:<ID>)
   |
   ↓
Client → GET /merge-requests/status → JSON Response
```