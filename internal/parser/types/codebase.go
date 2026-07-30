package types

import "time"

// Codebase merepresentasikan hasil analisis dari sebuah repository
type Codebase struct {
	// Informasi dasar
	RootDir   string    // Path ke root directory
	Language  string    // Bahasa pemrograman (javascript, go, php, python)
	Framework string    // Framework yang digunakan (express, gin, laravel, fastapi)

	// Hasil parsing
	Routes   []Route   // List routes yang terdeteksi
	Models   []Model   // List models yang terdeteksi
	Handlers []Handler // List handlers yang terdeteksi

	// Metadata
	AnalyzedAt time.Time // Waktu analisis dilakukan
	FileCount  int       // Jumlah file yang diproses
}
