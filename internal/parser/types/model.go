package types

// Model merepresentasikan data model/schema
type Model struct {
	// Informasi dasar
	Name  string // Nama model (User, Product, Order, dll)
	Table string // Nama tabel di database (users, products, orders, dll)

	// Struktur model
	Fields    []Field          // List fields/columns
	Relations []Relation       // List relasi dengan model lain
	Indexes   []Index          // List indexes pada tabel

	// Validasi
	Validation []ValidationRule // List aturan validasi

	// Metadata
	File string // File dimana model ini didefinisikan
	Line int    // Line number di file tersebut
}

// Field merepresentasikan sebuah column/property dalam model
type Field struct {
	Name     string // Nama field
	Type     string // Tipe data (string, int, bool, timestamp, dll)
	Required bool   // Apakah field ini wajib diisi
	Unique   bool   // Apakah nilai field ini harus unik
	Default  string // Default value (jika ada)

	// Metadata tambahan
	Comment string // Komentar/deskripsi field
}

// Relation merepresentasikan relasi antar model
type Relation struct {
	Type string // Tipe relasi (hasOne, hasMany, belongsTo, manyToMany)

	// Model yang direlasikan
	RelatedModel string // Nama model yang direlasikan
	ForeignKey   string // Foreign key di tabel ini
	RelatedKey   string // Key di tabel yang direlasikan (biasanya primary key)

	// Untuk many-to-many
	PivotTable string // Nama tabel pivot (jika many-to-many)
}

// Index merepresentasikan database index
type Index struct {
	Name    string   // Nama index
	Columns []string // List columns yang di-index
	Unique  bool     // Apakah ini unique index
}

// ValidationRule merepresentasikan aturan validasi
type ValidationRule struct {
	Field string // Field yang divalidasi
	Rule  string // Aturan validasi (required, email, min:3, max:100, regex:^\\d+$, dll)
}
