package tool

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// extractBinary 从下载的资产中提取目标二进制。
// 资产可能是 zip / tar.gz 压缩包(内含文件名为 binary 的可执行文件),
// 也可能是单文件二进制(原样返回,不复制)。
// 返回可用于安装的文件路径;若该路径是新生成的临时文件,由调用方负责 os.Remove。
func extractBinary(srcPath, binary string) (string, error) {
	head := make([]byte, 4)
	f, err := os.Open(srcPath)
	if err != nil {
		return "", err
	}
	_, readErr := io.ReadFull(f, head)
	closeErr := f.Close()
	if readErr != nil {
		// 文件不足 4 字节,按单文件二进制处理
		return srcPath, nil
	}
	if closeErr != nil {
		return "", closeErr
	}

	switch {
	case head[0] == 'P' && head[1] == 'K':
		return extractFromZip(srcPath, binary)
	case head[0] == 0x1f && head[1] == 0x8b:
		return extractFromTarGz(srcPath, binary)
	default:
		return srcPath, nil
	}
}

// extractFromZip 从 zip 包中提取目标二进制。
func extractFromZip(srcPath, binary string) (string, error) {
	zr, err := zip.OpenReader(srcPath)
	if err != nil {
		return "", fmt.Errorf("解压 zip: %w", err)
	}
	defer zr.Close()

	entry := findZipBinary(zr.File, binary)
	if entry == nil {
		return "", fmt.Errorf("压缩包中找不到二进制 %q", binary)
	}
	rc, err := entry.Open()
	if err != nil {
		return "", fmt.Errorf("读取 %s: %w", entry.Name, err)
	}
	defer rc.Close()
	return writeExtracted(rc, srcPath, binary)
}

// findZipBinary 在 zip 条目中定位目标二进制。
// 优先 basename 精确匹配;否则取无扩展名且带可执行魔数(Mach-O/ELF)的条目。
func findZipBinary(files []*zip.File, binary string) *zip.File {
	for _, f := range files {
		if f.FileInfo().IsDir() {
			continue
		}
		if strings.EqualFold(filepath.Base(f.Name), binary) {
			return f
		}
	}
	for _, f := range files {
		if f.FileInfo().IsDir() || filepath.Ext(f.Name) != "" {
			continue
		}
		if isExecutableZipEntry(f) {
			return f
		}
	}
	return nil
}

// isExecutableZipEntry 检查 zip 条目内容是否为 Mach-O / ELF 可执行文件。
func isExecutableZipEntry(f *zip.File) bool {
	rc, err := f.Open()
	if err != nil {
		return false
	}
	defer rc.Close()
	buf := make([]byte, 4)
	if _, err := io.ReadFull(rc, buf); err != nil {
		return false
	}
	return isMachO(buf) || isELF(buf)
}

// extractFromTarGz 从 tar.gz 包中提取目标二进制。
func extractFromTarGz(srcPath, binary string) (string, error) {
	f, err := os.Open(srcPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", fmt.Errorf("解压 gzip: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("读 tar 包: %w", err)
		}
		if hdr.Typeflag != tar.TypeReg && hdr.Typeflag != tar.TypeRegA {
			continue
		}
		if !strings.EqualFold(filepath.Base(hdr.Name), binary) {
			continue
		}
		return writeExtracted(tr, srcPath, binary)
	}
	return "", fmt.Errorf("压缩包中找不到二进制 %q", binary)
}

// writeExtracted 把 reader 内容写入与源文件同目录的临时文件并返回路径。
func writeExtracted(r io.Reader, srcPath, binary string) (string, error) {
	dir := filepath.Dir(srcPath)
	tmp, err := os.CreateTemp(dir, "."+binary+".extract-*")
	if err != nil {
		return "", err
	}
	tmpName := tmp.Name()
	if _, err := io.Copy(tmp, r); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return "", err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return "", err
	}
	return tmpName, nil
}

// isMachO 判断是否 Mach-O 可执行文件(通用二进制 / 64 位 / 32 位,含大小端)。
func isMachO(b []byte) bool {
	if len(b) < 4 {
		return false
	}
	return (b[0] == 0xcf && b[1] == 0xfa && b[2] == 0xed && b[3] == 0xfe) ||
		(b[0] == 0xfe && b[1] == 0xed && b[2] == 0xfa && b[3] == 0xce) ||
		(b[0] == 0xfe && b[1] == 0xed && b[2] == 0xfa && b[3] == 0xcf) ||
		(b[0] == 0xca && b[1] == 0xfe && b[2] == 0xba && b[3] == 0xbe) ||
		(b[0] == 0xbe && b[1] == 0xba && b[2] == 0xfe && b[3] == 0xca)
}

// isELF 判断是否 ELF 可执行文件。
func isELF(b []byte) bool {
	return len(b) >= 4 && b[0] == 0x7f && b[1] == 'E' && b[2] == 'L' && b[3] == 'F'
}
