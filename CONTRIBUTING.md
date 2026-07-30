# Contributing to GoTest Agent

Terima kasih atas minat Anda untuk berkontribusi pada GoTest Agent! Kami sangat menghargai kontribusi dari komunitas.

## Code of Conduct

Proyek ini mengikuti [Code of Conduct](CODE_OF_CONDUCT.md). Dengan berpartisipasi, Anda diharapkan untuk mengikuti code of conduct ini.

## Cara Berkontribusi

### Melaporkan Bug

Jika Anda menemukan bug, mohon buat issue di GitHub dengan informasi berikut:

1. **Deskripsi bug**: Jelaskan bug secara detail
2. **Langkah reproduksi**: Langkah-langkah untuk mereproduksi bug
3. **Expected behavior**: Apa yang seharusnya terjadi
4. **Actual behavior**: Apa yang sebenarnya terjadi
5. **Environment**: OS, versi Go, versi Node.js, dll.
6. **Screenshots**: Jika memungkinkan, sertakan screenshot

Template issue sudah disediakan di `.github/ISSUE_TEMPLATE/bug_report.md`

### Mengusulkan Fitur Baru

Jika Anda memiliki ide untuk fitur baru:

1. Buat issue dengan label `enhancement`
2. Jelaskan use case dan manfaat fitur
3. Jelaskan bagaimana fitur ini akan digunakan
4. Jika memungkinkan, sertakan mockup atau contoh

### Mengirim Pull Request

#### Persiapan

1. **Fork repository**
   ```bash
   git clone https://github.com/YOUR_USERNAME/agent_test.git
   cd agent_test
   git remote add upstream https://github.com/sanhaji182/agent_test.git
   ```

2. **Buat branch baru**
   ```bash
   git checkout -b feature/your-feature-name
   # atau
   git checkout -b fix/your-fix-name
   ```

3. **Setup development environment**
   ```bash
   cp .env.example .env
   # Edit .env dengan API keys Anda
   
   # Backend
   go mod download
   
   # Frontend
   cd frontend && npm install
   ```

#### Development Workflow

1. **Buat perubahan**
   - Tulis code yang clean dan well-documented
   - Follow coding standards (lihat di bawah)
   - Tulis tests untuk code baru

2. **Test perubahan**
   ```bash
   # Backend tests
   go test ./...
   
   # Frontend tests
   cd frontend && npm test
   
   # Run linting
   make lint
   ```

3. **Commit perubahan**
   ```bash
   git add .
   git commit -m "type: description of changes"
   ```
   
   Commit message format:
   - `feat: add new feature`
   - `fix: fix bug`
   - `docs: update documentation`
   - `test: add tests`
   - `refactor: refactor code`
   - `style: code style changes`
   - `chore: maintenance tasks`

4. **Push ke fork Anda**
   ```bash
   git push origin feature/your-feature-name
   ```

5. **Buat Pull Request**
   - Buka PR dari fork Anda ke repository utama
   - Gunakan template PR yang sudah disediakan
   - Jelaskan perubahan secara detail
   - Reference issue yang terkait (jika ada)

#### Pull Request Requirements

✅ **Semua PR harus:**
- [ ] Pass semua tests
- [ ] Pass linting
- [ ] Include tests untuk code baru
- [ ] Update dokumentasi jika diperlukan
- [ ] Follow coding standards
- [ ] Include commit message yang descriptive

❌ **PR akan di-reject jika:**
- Tests gagal
- Linting gagal
- Tidak ada tests untuk code baru
- Code quality rendah
- Tidak follow coding standards

## Coding Standards

### Go (Backend)

1. **Formatting**
   ```bash
   gofmt -w .
   ```

2. **Linting**
   ```bash
   golangci-lint run
   ```

3. **Standards**
   - Gunakan `gofmt` untuk formatting
   - Follow [Effective Go](https://go.dev/doc/effective_go)
   - Gunakan error handling yang proper
   - Tulis unit tests untuk semua fungsi
   - Gunakan context untuk cancellation dan timeouts
   - Gunakan interfaces untuk dependency injection
   - Hindari global variables
   - Gunakan logging yang structured (slog)

4. **Testing**
   ```bash
   # Run tests dengan coverage
   go test -cover ./...
   
   # Run tests dengan race detector
   go test -race ./...
   
   # Target coverage: 80%+
   ```

### TypeScript/JavaScript (Frontend)

1. **Formatting**
   ```bash
   npm run format
   ```

2. **Linting**
   ```bash
   npm run lint
   ```

3. **Standards**
   - Gunakan TypeScript strict mode
   - Follow Next.js best practices
   - Gunakan React hooks yang proper
   - Tulis unit tests untuk components
   - Gunakan TypeScript types yang proper
   - Hindari `any` type
   - Gunakan functional components

4. **Testing**
   ```bash
   # Run tests
   npm test
   
   # Run tests dengan coverage
   npm run test:coverage
   
   # Target coverage: 80%+
   ```

## Development Environment

### Prerequisites

- Go 1.26.4+
- Node.js 20+
- Docker & Docker Compose
- PostgreSQL 16+ (atau gunakan Docker)
- Redis 7+ (atau gunakan Docker)

### Setup

```bash
# Clone repository
git clone https://github.com/sanhaji182/agent_test.git
cd agent_test

# Setup environment
cp .env.example .env
# Edit .env dengan API keys Anda

# Start services dengan Docker Compose
make up

# Verify
make smoke-test
```

### Development Commands

```bash
# Backend
make backend          # Start backend server
make backend-test     # Run backend tests
make backend-lint     # Lint backend code

# Frontend
make frontend         # Start frontend dev server
make frontend-test    # Run frontend tests
make frontend-lint    # Lint frontend code

# All services
make up               # Start all services
make down             # Stop all services
make logs             # View logs
make test             # Run all tests
make lint             # Lint all code
```

## Testing Strategy

### Backend Tests

1. **Unit Tests**
   - Test semua fungsi dan methods
   - Gunakan table-driven tests
   - Mock dependencies dengan interfaces
   - Target coverage: 80%+

2. **Integration Tests**
   - Test integration dengan database
   - Test integration dengan external services
   - Gunakan test containers untuk dependencies

3. **E2E Tests**
   - Test full workflow
   - Test dengan real browser (Playwright)
   - Test dengan real AI provider (jika memungkinkan)

### Frontend Tests

1. **Unit Tests**
   - Test semua components
   - Test hooks dan utilities
   - Gunakan React Testing Library

2. **Integration Tests**
   - Test component interactions
   - Test dengan mock API

3. **E2E Tests**
   - Test full user workflows
   - Gunakan Playwright atau Cypress

## Review Process

### Pull Request Review

1. **Automated Checks**
   - CI/CD pipeline akan run otomatis
   - Tests harus pass
   - Linting harus pass
   - Coverage harus memenuhi target

2. **Manual Review**
   - Minimal 1 maintainer harus review
   - Review code quality
   - Review tests
   - Review documentation

3. **Approval**
   - PR harus di-approve oleh maintainer
   - Semua checks harus pass
   - Tidak ada unresolved comments

4. **Merge**
   - PR akan di-merge setelah approval
   - Gunakan squash merge untuk clean history
   - Delete branch setelah merge

## Getting Help

Jika Anda butuh bantuan:

1. **Documentation**: Baca dokumentasi di `docs/` directory
2. **Issues**: Cari issue yang serupa di GitHub
3. **Discussions**: Gunakan GitHub Discussions untuk pertanyaan
4. **Email**: Hubungi maintainer di maintainer@gotest-agent.com

## Recognition

Kontributor akan di-recognize di:

- [CONTRIBUTORS.md](CONTRIBUTORS.md)
- Release notes
- GitHub contributors page

## Code of Conduct

Mohon baca [Code of Conduct](CODE_OF_CONDUCT.md) sebelum berkontribusi.

## License

Dengan berkontribusi, Anda setuju bahwa kontribusi Anda akan di-license under the same license as the project (MIT License).

## Questions?

Jika Anda memiliki pertanyaan tentang contributing:

1. Baca dokumentasi di `docs/`
2. Cari di GitHub Discussions
3. Buat issue dengan label `question`
4. Email maintainer di maintainer@gotest-agent.com

Terima kasih atas kontribusi Anda! 🎉
