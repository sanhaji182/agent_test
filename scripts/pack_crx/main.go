// pack_crx membungkus folder chrome-extension menjadi file .crx format CRX3
// (Chromium extension format) yang bisa di-install via drag & drop ke
// chrome://extensions (Developer mode ON). Extension di-sign dengan kunci
// RSA yang disimpan di key.pem — ID extension dihitung dari kunci ini, jadi
// jangan diubah kalau ingin ID tetap stabil.
//
// Cara pakai: go run ./scripts/pack_crx <dir-extension> <output.crx>
package main

import (
	"archive/zip"
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/binary"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	crxMagic = "Cr24"
	crxVer3  = 3
)

// --- protobuf wire encoding (minimal, cukup untuk struktur CRX3) ---

func pbTag(field int, wireType int) []byte { return []byte{byte(field<<3 | wireType)} }

func pbBytes(field int, data []byte) []byte {
	var b bytes.Buffer
	b.Write(pbTag(field, 2)) // length-delimited
	b.Write(pbVarint(uint64(len(data))))
	b.Write(data)
	return b.Bytes()
}

func pbVarint(v uint64) []byte {
	var b bytes.Buffer
	for v >= 0x80 {
		b.WriteByte(byte(v) | 0x80)
		v >>= 7
	}
	b.WriteByte(byte(v))
	return b.Bytes()
}

// signedData = SignedData{ crx_id = field1, sha256_with_context = field2 }
func encodeSignedData(crxID, hash []byte) []byte {
	var b bytes.Buffer
	b.Write(pbBytes(1, crxID))
	b.Write(pbBytes(2, hash))
	return b.Bytes()
}

// proof = AsymmetricKeyProof{ public_key = field1, signature = field2 }
func encodeProof(pubDER, sig []byte) []byte {
	var b bytes.Buffer
	b.Write(pbBytes(1, pubDER))
	b.Write(pbBytes(2, sig))
	return b.Bytes()
}

// header = CrxFileHeader{ sha256_with_rsa = field2 (repeated), signed_header_data = field10000 }
func encodeHeader(proof, signedData []byte) []byte {
	var b bytes.Buffer
	proofMsg := encodeProofRaw(proof) // wrapped message for repeated field
	b.Write(pbTag(2, 2))
	b.Write(pbVarint(uint64(len(proofMsg))))
	b.Write(proofMsg)
	b.Write(pbBytes(10000, signedData))
	return b.Bytes()
}

func encodeProofRaw(proof []byte) []byte {
	// proof di sini sudah berupa message AsymmetricKeyProof lengkap
	return proof
}

// zipDir membuat zip dari seluruh isi folder (tanpa folder induk).
func zipDir(dir string) ([]byte, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		if strings.HasSuffix(rel, ".DS_Store") {
			return nil
		}
		f, err := zw.Create(rel)
		if err != nil {
			return err
		}
		src, err := os.Open(path)
		if err != nil {
			return err
		}
		defer src.Close()
		_, err = io.Copy(f, src)
		return err
	})
	if err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// loadOrCreateKey memuat kunci RSA dari file, atau membuat yang baru.
func loadOrCreateKey(path string) (*rsa.PrivateKey, error) {
	if data, err := os.ReadFile(path); err == nil {
		if block, _ := pem.Decode(data); block != nil {
			if k, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
				if rsaKey, ok := k.(*rsa.PrivateKey); ok {
					return rsaKey, nil
				}
			}
			if k, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
				return k, nil
			}
		}
		return nil, fmt.Errorf("key file %s tidak valid", path)
	}
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, err
	}
	pemData := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
	return key, os.WriteFile(path, pemData, 0o600)
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: pack_crx <dir-extension> <output.crx>")
		os.Exit(2)
	}
	srcDir := os.Args[1]
	outPath := os.Args[2]
	// Key disimpan di scripts/pack_crx/key.pem (relatif ke working dir).
	keyPath := filepath.Join("scripts", "pack_crx", "key.pem")

	key, err := loadOrCreateKey(keyPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "key:", err)
		os.Exit(1)
	}

	zipBytes, err := zipDir(srcDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "zip:", err)
		os.Exit(1)
	}

	// Context untuk SHA256: "CRX3 SignedData" (spesifikasi crx3 Chromium).
	context := []byte("CRX3 SignedData")
	hasher := sha256.New()
	hasher.Write(context)
	hasher.Write(zipBytes)
	zipHash := hasher.Sum(nil)

	pubDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		fmt.Fprintln(os.Stderr, "pub:", err)
		os.Exit(1)
	}

	// Extension ID = 16 byte pertama SHA256(public key), hex.
	idHash := sha256.Sum256(pubDER)
	crxID := idHash[:16]

	signedData := encodeSignedData(crxID, zipHash)

	// Signature RSASSA-PKCS1-v1_5 SHA256 atas signed_data.
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, sha256Sum(signedData))
	if err != nil {
		fmt.Fprintln(os.Stderr, "sign:", err)
		os.Exit(1)
	}

	proof := encodeProof(pubDER, sig)
	header := encodeHeader(proof, signedData)

	var out bytes.Buffer
	out.WriteString(crxMagic)
	binary.Write(&out, binary.LittleEndian, uint32(crxVer3))
	binary.Write(&out, binary.LittleEndian, uint32(len(header)))
	out.Write(header)
	out.Write(zipBytes)

	if err := os.WriteFile(outPath, out.Bytes(), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "write:", err)
		os.Exit(1)
	}

	// Verifikasi ulang struktur supaya yakin sebelum dipakai.
	if err := verifyCRX(out.Bytes()); err != nil {
		fmt.Fprintln(os.Stderr, "verify FAILED:", err)
		os.Exit(1)
	}

	fmt.Printf("OK: %s (%d bytes)\nextension_id: %s\n", outPath, out.Len(), hex.EncodeToString(crxID))
}

func sha256Sum(b []byte) []byte {
	h := sha256.Sum256(b)
	return h[:]
}

// verifyCRX mem-parse ulang dan memverifikasi signature + hash zip.
func verifyCRX(data []byte) error {
	if len(data) < 12 || string(data[:4]) != crxMagic {
		return fmt.Errorf("magic salah")
	}
	ver := binary.LittleEndian.Uint32(data[4:8])
	if ver != crxVer3 {
		return fmt.Errorf("versi %d (harus 3)", ver)
	}
	hdrLen := binary.LittleEndian.Uint32(data[8:12])
	if int(12+hdrLen) > len(data) {
		return fmt.Errorf("header terlalu panjang")
	}
	hdr := data[12 : 12+hdrLen]
	zipBytes := data[12+hdrLen:]

	// Ekstrak signed_header_data (field 10000) & proof (field 2) dari header.
	proof := parseHeaderField(hdr, 2)
	signedData := parseHeaderField(hdr, 10000)
	if len(proof) == 0 || len(signedData) == 0 {
		return fmt.Errorf("header tidak lengkap")
	}
	pubDER := parseMsgField(proof, 1)
	sig := parseMsgField(proof, 2)
	crxID := parseMsgField(signedData, 1)
	zipHash := parseMsgField(signedData, 2)
	if len(pubDER) == 0 || len(sig) == 0 || len(crxID) == 0 || len(zipHash) == 0 {
		return fmt.Errorf("field header tidak lengkap")
	}

	pub, err := x509.ParsePKIXPublicKey(pubDER)
	if err != nil {
		return fmt.Errorf("parse pubkey: %w", err)
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("bukan RSA")
	}
	if err := rsa.VerifyPKCS1v15(rsaPub, crypto.SHA256, sha256Sum(signedData), sig); err != nil {
		return fmt.Errorf("signature tidak valid: %w", err)
	}

	context := []byte("CRX3 SignedData")
	h := sha256.New()
	h.Write(context)
	h.Write(zipBytes)
	if !bytes.Equal(h.Sum(nil), zipHash) {
		return fmt.Errorf("hash zip tidak cocok")
	}
	// ID cocok?
	idHash := sha256.Sum256(pubDER)
	if !bytes.Equal(idHash[:16], crxID) {
		return fmt.Errorf("extension id tidak cocok dengan public key")
	}
	// Zip harus bisa dibuka.
	zr, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		return fmt.Errorf("zip tidak valid: %w", err)
	}
	if len(zr.File) == 0 {
		return fmt.Errorf("zip kosong")
	}
	return nil
}

// parseHeaderField mengambil isi field `field` (length-delimited) dari message.
func parseHeaderField(msg []byte, field int) []byte {
	tag := byte(field<<3 | 2)
	for i := 0; i+1 < len(msg); {
		t := msg[i]
		i++
		if t == 0 {
			continue
		}
		// decode varint length
		length := uint64(0)
		shift := uint(0)
		for i < len(msg) {
			b := msg[i]
			i++
			length |= uint64(b&0x7f) << shift
			if b&0x80 == 0 {
				break
			}
			shift += 7
		}
		if t == tag {
			if int(i)+int(length) > len(msg) {
				return nil
			}
			return msg[i : i+int(length)]
		}
		i += int(length)
	}
	return nil
}

// parseMsgField sama seperti parseHeaderField (untuk message proof/signed).
func parseMsgField(msg []byte, field int) []byte {
	return parseHeaderField(msg, field)
}
