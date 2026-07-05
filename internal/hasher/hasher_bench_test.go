package hasher

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// BenchmarkFullHashSHA256 measures SHA256 throughput on a synthetic file.
func BenchmarkFullHashSHA256(b *testing.B) {
	benchmarkHash(b, "sha256")
}

// BenchmarkFullHashMD5 measures MD5 throughput on a synthetic file.
func BenchmarkFullHashMD5(b *testing.B) {
	benchmarkHash(b, "md5")
}

// BenchmarkFullHashSHA1 measures SHA1 throughput on a synthetic file.
func BenchmarkFullHashSHA1(b *testing.B) {
	benchmarkHash(b, "sha1")
}

func benchmarkHash(b *testing.B, algo string) {
	b.Helper()

	const fileSize = 10 * 1024 * 1024 // 10 MB
	dir := b.TempDir()
	path := filepath.Join(dir, "bench.bin")

	data := make([]byte, fileSize)
	for i := range data {
		data[i] = byte(i % 251)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		b.Fatalf("setup: %v", err)
	}

	b.ResetTimer()
	b.SetBytes(fileSize)

	for i := 0; i < b.N; i++ {
		if _, err := FullHash(path, algo); err != nil {
			b.Fatalf("FullHash: %v", err)
		}
	}
}

// BenchmarkPartialHash measures partial-hash throughput (first 4 KB).
func BenchmarkPartialHash(b *testing.B) {
	const fileSize = 10 * 1024 * 1024
	dir := b.TempDir()
	path := filepath.Join(dir, "bench.bin")

	data := make([]byte, fileSize)
	if err := os.WriteFile(path, data, 0644); err != nil {
		b.Fatalf("setup: %v", err)
	}

	b.ResetTimer()
	b.SetBytes(4096)

	for i := 0; i < b.N; i++ {
		if _, err := PartialHash(path, "sha256", 4096); err != nil {
			b.Fatalf("PartialHash: %v", err)
		}
	}
}

// BenchmarkCompareFiles measures byte-for-byte comparison throughput.
func BenchmarkCompareFiles(b *testing.B) {
	const fileSize = 10 * 1024 * 1024
	dir := b.TempDir()
	data := make([]byte, fileSize)

	p1 := filepath.Join(dir, "f1.bin")
	p2 := filepath.Join(dir, "f2.bin")
	os.WriteFile(p1, data, 0644)
	os.WriteFile(p2, data, 0644)

	b.ResetTimer()
	b.SetBytes(fileSize)

	for i := 0; i < b.N; i++ {
		if _, err := CompareFiles(p1, p2); err != nil {
			b.Fatalf("CompareFiles: %v", err)
		}
	}
}

// BenchmarkHashScaling measures how hash throughput scales with file count.
func BenchmarkHashScaling(b *testing.B) {
	for _, count := range []int{10, 100, 1000} {
		count := count
		b.Run(fmt.Sprintf("files=%d", count), func(b *testing.B) {
			dir := b.TempDir()
			data := make([]byte, 1024) // 1 KB each
			paths := make([]string, count)
			for i := 0; i < count; i++ {
				p := filepath.Join(dir, fmt.Sprintf("f%d.bin", i))
				os.WriteFile(p, data, 0644)
				paths[i] = p
			}

			b.ResetTimer()
			b.SetBytes(int64(count * 1024))

			for i := 0; i < b.N; i++ {
				for _, p := range paths {
					FullHash(p, "sha256") //nolint:errcheck
				}
			}
		})
	}
}
