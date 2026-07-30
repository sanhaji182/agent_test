package types

// Route merepresentasikan sebuah API endpoint
type Route struct {
	// Informasi dasar
	Method   string // HTTP method (GET, POST, PUT, DELETE, dll)
	Path     string // URL path (/users, /products/:id, dll)
	Handler  string // Nama handler function atau controller method

	// Middleware yang diterapkan pada route ini
	Middleware []string // List middleware (auth, cors, rateLimit, dll)

	// Parameter path (untuk dynamic routes)
	Params map[string]string // Map parameter name -> type (/users/:id -> {"id": "string"})

	// Metadata
	File string // File dimana route ini didefinisikan
	Line int    // Line number di file tersebut
}
