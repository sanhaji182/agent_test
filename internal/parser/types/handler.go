package types

// Handler merepresentasikan sebuah handler function/controller method
type Handler struct {
	// Informasi dasar
	Name       string // Nama function/method
	Controller string // Nama controller (jika ada)
	Method     string // HTTP method yang ditangani (jika spesifik)

	// Signature function
	Parameters []Parameter // List parameters yang diterima
	ReturnType string      // Tipe return value

	// Analisis
	HasValidation bool     // Apakah handler ini melakukan validasi
	DatabaseOps   []string // List operasi database (create, read, update, delete)
	ExternalCalls []string // List external API calls

	// Metadata
	File string // File dimana handler ini didefinisikan
	Line int    // Line number di file tersebut
}

// Parameter merepresentasikan parameter function
type Parameter struct {
	Name     string // Nama parameter
	Type     string // Tipe data parameter
	Required bool   // Apakah parameter ini wajib
	Default  string // Default value (jika ada)
}
