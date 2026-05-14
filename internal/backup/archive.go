package backup

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ArchiveOptions controls data archive creation.
type ArchiveOptions struct {
	TempDir         string
	ExcludePatterns []string
}

// CreateArchive writes a gzip-compressed tar archive of srcDir to destPath.
// Paths inside the archive are relative to srcDir.
func CreateArchive(srcDir, destPath, tempDir string) error {
	return CreateArchiveWithOptions(srcDir, destPath, ArchiveOptions{TempDir: tempDir})
}

// CreateArchiveWithOptions writes a gzip-compressed tar archive of srcDir to destPath.
func CreateArchiveWithOptions(srcDir, destPath string, opts ArchiveOptions) error {
	srcAbs, err := filepath.Abs(srcDir)
	if err != nil {
		return fmt.Errorf("resolve data dir: %w", err)
	}
	info, err := os.Stat(srcAbs)
	if err != nil {
		return fmt.Errorf("stat data dir: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("data dir is not a directory: %s", srcDir)
	}
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return fmt.Errorf("create archive dir: %w", err)
	}

	destAbs, err := filepath.Abs(destPath)
	if err != nil {
		return fmt.Errorf("resolve archive path: %w", err)
	}
	tempAbs := ""
	if strings.TrimSpace(opts.TempDir) != "" {
		if v, err := filepath.Abs(opts.TempDir); err == nil {
			tempAbs = v
		}
	}

	file, err := os.Create(destAbs)
	if err != nil {
		return fmt.Errorf("create archive: %w", err)
	}
	defer file.Close()

	gz := gzip.NewWriter(file)
	defer gz.Close()
	tw := tar.NewWriter(gz)
	defer tw.Close()

	return filepath.WalkDir(srcAbs, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		abs, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		rel := ""
		if abs != srcAbs {
			rel, err = filepath.Rel(srcAbs, abs)
			if err != nil {
				return err
			}
		}
		if shouldSkipPath(srcAbs, abs, rel, destAbs, tempAbs, d, opts.ExcludePatterns) {
			if d.IsDir() && abs != srcAbs {
				return filepath.SkipDir
			}
			return nil
		}
		if abs == srcAbs {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}
		name := filepath.ToSlash(rel)
		if name == "." || name == "" {
			return nil
		}

		mode := info.Mode()
		var link string
		if mode&os.ModeSymlink != 0 {
			link, err = os.Readlink(abs)
			if err != nil {
				return err
			}
		}
		hdr, err := tar.FileInfoHeader(info, link)
		if err != nil {
			return err
		}
		hdr.Name = name
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if mode.IsRegular() {
			r, err := os.Open(abs)
			if err != nil {
				return err
			}
			_, copyErr := io.Copy(tw, r)
			closeErr := r.Close()
			if copyErr != nil {
				return copyErr
			}
			if closeErr != nil {
				return closeErr
			}
		}
		return nil
	})
}

func shouldSkipPath(_ string, pathAbs, relPath, destAbs, tempAbs string, d os.DirEntry, excludePatterns []string) bool {
	if samePath(pathAbs, destAbs) || isWithin(pathAbs, destAbs) {
		return true
	}
	if tempAbs != "" && (samePath(pathAbs, tempAbs) || isWithin(pathAbs, tempAbs)) {
		return true
	}
	name := d.Name()
	if strings.Contains(name, ".restore-backup-") {
		return true
	}
	if strings.HasSuffix(name, ".tmp") {
		return true
	}
	return matchAnyExcludePattern(relPath, d.IsDir(), excludePatterns)
}

func matchAnyExcludePattern(relPath string, isDir bool, patterns []string) bool {
	rel := filepath.ToSlash(filepath.Clean(relPath))
	if rel == "." || rel == "" {
		return false
	}
	for _, pattern := range patterns {
		pattern = strings.Trim(strings.TrimSpace(filepath.ToSlash(pattern)), "/")
		if pattern == "" {
			continue
		}
		if strings.HasSuffix(pattern, "/**") {
			base := strings.TrimSuffix(pattern, "/**")
			if rel == base || strings.HasPrefix(rel, base+"/") {
				return true
			}
			continue
		}
		if strings.HasSuffix(pattern, "/") {
			base := strings.TrimSuffix(pattern, "/")
			if rel == base || (isDir && strings.HasPrefix(rel, base+"/")) {
				return true
			}
			continue
		}
		if ok, _ := filepath.Match(pattern, rel); ok {
			return true
		}
		if ok, _ := filepath.Match(pattern, filepath.Base(rel)); ok {
			return true
		}
		if rel == pattern || strings.HasPrefix(rel, pattern+"/") {
			return true
		}
	}
	return false
}

func isWithin(pathAbs, parentAbs string) bool {
	rel, err := filepath.Rel(parentAbs, pathAbs)
	if err != nil {
		return false
	}
	return rel != "." && rel != "" && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && rel != ".."
}

func samePath(a, b string) bool {
	return filepath.Clean(strings.ToLower(a)) == filepath.Clean(strings.ToLower(b))
}

// ExtractArchive safely extracts a gzip-compressed tar archive into destDir.
func ExtractArchive(archivePath, destDir string) error {
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return fmt.Errorf("create restore temp dir: %w", err)
	}
	destAbs, err := filepath.Abs(destDir)
	if err != nil {
		return fmt.Errorf("resolve restore dir: %w", err)
	}

	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("open archive: %w", err)
	}
	defer file.Close()
	gz, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("open gzip archive: %w", err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)

	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("read archive: %w", err)
		}
		name := strings.TrimSpace(hdr.Name)
		if name == "" {
			continue
		}
		target, err := safeArchiveTarget(destAbs, name)
		if err != nil {
			return err
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, modePerm(hdr.FileInfo().Mode(), 0o755)); err != nil {
				return err
			}
		case tar.TypeReg, tar.TypeRegA:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, modePerm(hdr.FileInfo().Mode(), 0o644))
			if err != nil {
				return err
			}
			_, copyErr := io.Copy(out, tr)
			closeErr := out.Close()
			if copyErr != nil {
				return copyErr
			}
			if closeErr != nil {
				return closeErr
			}
		case tar.TypeSymlink:
			linkTarget, err := safeLinkTarget(destAbs, filepath.Dir(target), hdr.Linkname)
			if err != nil {
				return err
			}
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			_ = linkTarget
			if err := os.Symlink(hdr.Linkname, target); err != nil {
				return err
			}
		case tar.TypeLink:
			linkTarget, err := safeArchiveTarget(destAbs, hdr.Linkname)
			if err != nil {
				return err
			}
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			if err := os.Link(linkTarget, target); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported archive entry type %d for %s", hdr.Typeflag, hdr.Name)
		}
	}
	return nil
}

func safeArchiveTarget(destAbs, name string) (string, error) {
	cleanName := filepath.Clean(filepath.FromSlash(name))
	if cleanName == "." || cleanName == string(filepath.Separator) || filepath.IsAbs(cleanName) || cleanName == ".." || strings.HasPrefix(cleanName, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("unsafe archive path: %s", name)
	}
	target := filepath.Join(destAbs, cleanName)
	targetAbs, err := filepath.Abs(target)
	if err != nil {
		return "", err
	}
	if targetAbs != destAbs && !isWithin(targetAbs, destAbs) {
		return "", fmt.Errorf("archive path escapes restore dir: %s", name)
	}
	return targetAbs, nil
}

func safeLinkTarget(rootAbs, parentAbs, linkName string) (string, error) {
	if strings.TrimSpace(linkName) == "" {
		return "", errors.New("empty archive link target")
	}
	var target string
	if filepath.IsAbs(linkName) {
		target = filepath.Clean(linkName)
	} else {
		target = filepath.Join(parentAbs, filepath.FromSlash(linkName))
	}
	abs, err := filepath.Abs(target)
	if err != nil {
		return "", err
	}
	if abs != rootAbs && !isWithin(abs, rootAbs) {
		return "", fmt.Errorf("archive link escapes restore dir: %s", linkName)
	}
	return abs, nil
}

func modePerm(mode os.FileMode, fallback os.FileMode) os.FileMode {
	perm := mode.Perm()
	if perm == 0 {
		return fallback
	}
	return perm
}
