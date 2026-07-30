# Security Policy

## Supported Versions

Kami mendukung versi berikut dengan update keamanan:

| Version | Supported          |
| ------- | ------------------ |
| 1.0.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

Kami sangat menghargai laporan keamanan dari komunitas. Jika Anda menemukan vulnerability, mohon laporkan secara bertanggung jawab.

### Cara Melaporkan

**JANGAN** membuat GitHub issue publik untuk vulnerability keamanan.

Sebaliknya, kirim email ke: **security@gotest-agent.com**

Sertakan informasi berikut:
- Deskripsi vulnerability
- Langkah-langkah untuk mereproduksi
- Versi yang terpengaruh
- Potensi dampak
- Jika memungkinkan, saran untuk perbaikan

### Apa yang Akan Kami Lakukan

1. **Konfirmasi**: Kami akan mengkonfirmasi penerimaan laporan dalam 48 jam
2. **Assessment**: Kami akan menilai severity dan dampak dalam 3-5 hari kerja
3. **Fix**: Kami akan mengembangkan dan menguji patch
4. **Release**: Kami akan merilis security update dan mengumumkan CVE jika diperlukan
5. **Credit**: Kami akan memberikan kredit kepada reporter (kecuali jika diminta untuk anonim)

### Responsible Disclosure Timeline

- **Day 0**: Vulnerability dilaporkan
- **Day 1-2**: Konfirmasi penerimaan
- **Day 3-7**: Assessment dan validasi
- **Day 7-14**: Development patch
- **Day 14-21**: Testing dan QA
- **Day 21-30**: Release dan public disclosure

### Scope

**Dalam scope:**
- Remote code execution
- SQL injection
- Cross-site scripting (XSS)
- Authentication bypass
- Authorization bypass
- Data exposure
- Privilege escalation
- Server-side request forgery (SSRF)
- Command injection

**Di luar scope:**
- Denial of service (DoS) attacks
- Social engineering
- Physical security
- Vulnerabilities di third-party dependencies (laporkan ke vendor)
- Vulnerabilities yang memerlukan physical access
- Vulnerabilities di versi yang tidak didukung

### Severity Levels

Kami menggunakan skala severity berikut:

**Critical (CVSS 9.0-10.0)**
- Remote code execution
- Authentication bypass
- Data breach dengan data sensitif

**High (CVSS 7.0-8.9)**
- Privilege escalation
- SQL injection dengan data access
- Authorization bypass

**Medium (CVSS 4.0-6.9)**
- Cross-site scripting (XSS)
- Information disclosure (non-sensitive)
- Server-side request forgery (SSRF) terbatas

**Low (CVSS 0.1-3.9)**
- Information disclosure (minimal)
- Configuration issues
- Minor vulnerabilities

## Security Best Practices

### Untuk Pengguna

1. **Selalu gunakan versi terbaru**
   ```bash
   git pull origin master
   make rebuild
   ```

2. **Ganti semua default credentials**
   ```bash
   # Generate strong password
   openssl rand -base64 32
   
   # Update .env
   POSTGRES_PASSWORD=strong_random_password
   JWT_SECRET=strong_random_secret
   GITHUB_WEBHOOK_SECRET=strong_random_secret
   ```

3. **Batasi akses network**
   ```yaml
   # docker-compose.yml
   postgres:
     ports:
       - "127.0.0.1:5432:5432"  # Bind ke localhost saja
   
   redis:
     ports:
       - "127.0.0.1:6379:6379"
   ```

4. **Gunakan HTTPS di production**
   - Setup reverse proxy (nginx, Caddy, Traefik)
   - Gunakan Let's Encrypt untuk SSL certificate gratis
   - Redirect semua HTTP traffic ke HTTPS

5. **Enable rate limiting**
   - Default: 100 requests/minute/IP
   - Sesuaikan sesuai kebutuhan di `internal/api/ratelimit.go`

6. **Gunakan environment variables untuk secrets**
   - JANGAN commit secrets ke version control
   - Gunakan `.env` file (sudah di `.gitignore`)
   - Untuk production, gunakan Docker secrets atau secret manager

7. **Regular updates**
   ```bash
   # Update dependencies
   go mod tidy
   npm audit fix  # untuk frontend
   
   # Check for vulnerabilities
   go list -m all | grep -E "(vulnerability|CVE)"
   npm audit
   ```

8. **Monitoring dan logging**
   - Setup log aggregation (ELK stack, Loki)
   - Monitor untuk suspicious activity
   - Setup alerting untuk security events

### Untuk Kontributor

1. **Jangan commit secrets**
   - Gunakan `.env.example` sebagai template
   - JANGAN commit `.env` file
   - Gunakan environment variables untuk testing

2. **Code review untuk security**
   - Semua PR harus di-review
   - Focus pada security-sensitive code
   - Gunakan static analysis tools

3. **Dependency management**
   - Regular update dependencies
   - Monitor untuk known vulnerabilities
   - Gunakan `go mod verify` untuk verify dependencies

4. **Secure coding practices**
   - Validate semua input
   - Sanitize output
   - Gunakan parameterized queries
   - Hindari SQL injection, XSS, CSRF

5. **Authentication dan authorization**
   - Gunakan JWT untuk API authentication
   - Implement proper role-based access control (RBAC)
   - Validate dan sanitize semua user input

## Security Features

### Yang Sudah Diimplementasi

✅ **Authentication**
- JWT-based authentication
- API key authentication
- GitHub webhook signature verification (HMAC-SHA256)

✅ **Authorization**
- Role-based access control (RBAC)
- API key validation
- Request validation

✅ **Data Protection**
- HTTPS support (via reverse proxy)
- Database connection encryption (TLS)
- Secure password storage (bcrypt)

✅ **Input Validation**
- Request validation middleware
- SQL injection prevention (parameterized queries)
- XSS prevention (output encoding)

✅ **Rate Limiting**
- 100 requests/minute/IP (default)
- Configurable di `internal/api/ratelimit.go`

✅ **CORS Protection**
- Configurable CORS policy
- Default: disabled (secure by default)

✅ **Logging dan Monitoring**
- Structured logging
- Request/response logging
- Security event logging

### Yang Akan Datang

🔜 **Planned Security Features**
- Two-factor authentication (2FA)
- OAuth 2.0 / OpenID Connect
- SAML support untuk enterprise
- Advanced rate limiting (per-user, per-endpoint)
- Web Application Firewall (WAF)
- Security headers (CSP, HSTS, X-Frame-Options)
- Vulnerability scanning automation
- Security audit logging

## Compliance

### Standards

Kami bertujuan untuk comply dengan standards berikut:

- **OWASP Top 10**: Protection against common web vulnerabilities
- **CIS Benchmarks**: Security configuration best practices
- **SOC 2 Type II**: Security, availability, and confidentiality (planned)
- **ISO 27001**: Information security management (planned)

### Data Privacy

Kami menghormati privacy pengguna:

- **Data minimization**: Kami hanya mengumpulkan data yang diperlukan
- **Data encryption**: Data sensitif di-encrypt at rest dan in transit
- **Data retention**: Data di-retain sesuai policy
- **Data deletion**: Users dapat request deletion data mereka
- **GDPR compliance**: Kami comply dengan GDPR untuk EU users

## Security Contacts

- **Security Team**: security@gotest-agent.com
- **Maintainer**: sanhaji182
- **Response time**: 48 hours untuk initial response

## Acknowledgments

Kami berterima kasih kepada semua security researchers yang telah membantu membuat GoTest Agent lebih aman:

- (Belum ada laporan vulnerability)

Jika Anda ingin disebutkan di sini setelah melaporkan vulnerability, mohon informasikan kami saat melaporkan.

## Resources

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [CIS Benchmarks](https://www.cisecurity.org/cis-benchmarks)
- [Go Security Best Practices](https://go.dev/doc/security)
- [Playwright Security](https://playwright.dev/docs/security)

## Updates

Security policy ini akan di-update secara berkala. Last updated: 2026-07-30

Untuk pertanyaan tentang security policy ini, hubungi: security@gotest-agent.com
