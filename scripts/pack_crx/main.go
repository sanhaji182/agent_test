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

// pbTag membuat tag protobuf (field number + wire type) sebagai varint.
// Field 10000 (signed_header_data) butuh varint multi-byte — tag satu byte
// hanya cukup untuk field < 32. Bug lama di sini yang bikin "CRX header invalid".
func pbTag(field int, wireType int) []byte {
	return pbVarint(uint64(field<<3 | wireType))
}

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

// signedData = SignedData{ crx_id = field1 } — SignedData hanya punya satu field
// (crx_id). Tidak ada sha256_with_context di sini (lihat crx3.proto Chromium).
func encodeSignedData(crxID []byte) []byte {
	return pbBytes(1, crxID)
}

// signedInput = "CRX3 SignedData\x00" + uint32LE(len(signedHeaderData)) +
// signedHeaderData + archive. Ini yang ditandatangani oleh semua proof
// (spesifikasi crx3.proto).
func signedInput(signedHeaderData, archive []byte) []byte {
	var b bytes.Buffer
	b.WriteString("CRX3 SignedData")
	b.WriteByte(0)
	binary.Write(&b, binary.LittleEndian, uint32(len(signedHeaderData)))
	b.Write(signedHeaderData)
	b.Write(archive)
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

	pubDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		fmt.Fprintln(os.Stderr, "pub:", err)
		os.Exit(1)
	}

	// Extension ID = 16 byte pertama SHA256(public key), hex.
	idHash := sha256.Sum256(pubDER)
	crxID := idHash[:16]

	signedData := encodeSignedData(crxID)

	// Signature RSASSA-PKCS1-v1_5 SHA256 atas input yang ditentukan spesifikasi:
	// "CRX3 SignedData\x00" + signed_header_size + signed_header_data + archive.
	proofInput := signedInput(signedData, zipBytes)
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, sha256Sum(proofInput))
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
	proof := parseField(hdr, 2)
	signedData := parseField(hdr, 10000)
	if len(proof) == 0 || len(signedData) == 0 {
		return fmt.Errorf("header tidak lengkap")
	}
	pubDER := parseField(proof, 1)
	sig := parseField(proof, 2)
	crxID := parseField(signedData, 1)
	if len(pubDER) == 0 || len(sig) == 0 || len(crxID) == 0 {
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
	// Verifikasi signature atas input spesifikasi (bukan sekadar signed_header_data).
	proofInput := signedInput(signedData, zipBytes)
	if err := rsa.VerifyPKCS1v15(rsaPub, crypto.SHA256, sha256Sum(proofInput), sig); err != nil {
		return fmt.Errorf("signature tidak valid: %w", err)
	}

	// ID cocok dengan public key?
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

// readVarint membaca varint dari msg, memajukan pointer i.
func readVarint(msg []byte, i *int) (uint64, bool) {
	var v uint64
	shift := uint(0)
	for *i < len(msg) {
		b := msg[*i]
		*i++
		v |= uint64(b&0x7f) << shift
		if b&0x80 == 0 {
			return v, true
		}
		shift += 7
		if shift > 63 {
			return 0, false
		}
	}
	return 0, false
}

// parseField mengambil isi field `target` (wire type 2, length-delimited) dari
// message protobuf. Field number dibaca sebagai varint (multi-byte).
func parseField(msg []byte, target int) []byte {
	i := 0
	for i < len(msg) {
		tag, ok := readVarint(msg, &i)
		if !ok {
			return nil
		}
		fieldNum := int(tag >> 3)
		wire := int(tag & 7)
		if fieldNum == target && wire == 2 {
			length, ok := readVarint(msg, &i)
			if !ok || int(length) > len(msg)-i {
				return nil
			}
			return msg[i : i+int(length)]
		}
		// Skip field sesuai wire type.
		switch wire {
		case 0: // varint
			if _, ok := readVarint(msg, &i); !ok {
				return nil
			}
		case 1: // 64-bit
			i += 8
		case 2: // length-delimited
			length, ok := readVarint(msg, &i)
			if !ok || int(length) > len(msg)-i {
				return nil
			}
			i += int(length)
		case 5: // 32-bit
			i += 4
		default:
			return nil
		}
	}
	return nil
}

// parseHeaderField mengambil isi field `field` dari message (wrapper lama, kini
// memakai parser varint yang benar).
func parseHeaderField(msg []byte, field int) []byte {
	return parseField(msg, field)
}

// parseMsgField sama seperti parseHeaderField (untuk message proof/signed).
func parseMsgField(msg []byte, field int) []byte {
	return parseField(msg, field)
}
