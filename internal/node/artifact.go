package node

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"runtime"
	"strings"
)

type Artifact struct {
	Name, Slug, Ext string
}

func DownloadArtifact(v string) (Artifact, error) {
	slug, err := getSlug(v)
	if err != nil {
		return Artifact{}, err
	}

	u, err := url.JoinPath("https://nodejs.org/dist", v, slug)
	if err != nil {
		return Artifact{}, err
	}

	r, err := http.Get(u)
	if err != nil {
		return Artifact{}, err
	}

	if r.StatusCode >= 400 {
		return Artifact{}, fmt.Errorf("failed to download artifact %s: request failed with status %s", slug, r.Status)
	}

	defer r.Body.Close()

	f, err := os.Create(path.Join(os.TempDir(), slug))
	if err != nil {
		return Artifact{}, err
	}

	defer f.Close()

	if _, err := io.Copy(f, r.Body); err != nil {
		return Artifact{}, err
	}

	return Artifact{f.Name(), slug, path.Ext(f.Name())}, nil
}

func getSlug(v string) (string, error) {
	var (
		ops  = runtime.GOOS
		arch string
		ext  string
	)

	switch runtime.GOARCH {
	case "386":
		arch = "x86"
	case "amd64":
		arch = "x64"
	case "arm":
		arch = "armv7l"
	}

	switch ops {
	case "aix", "darwin":
		ext = ".tar.gz"
	case "linux":
		switch arch {
		case "arm64", "armv7l", "ppc64le", "s390x", "x64":
			break
		default:
			return "", fmt.Errorf("not supported: %s/%s", ops, arch)
		}

		ext = ".tar.gz"
	case "windows":
		ops = "win"
		ext = ".zip"
	default:
		return "", fmt.Errorf("%s not supported", ops)
	}

	return fmt.Sprintf("node-%s-%s-%s%s", v, ops, arch, ext), nil
}

func ExtractArtifact(src, dst string) error {
	switch ext := path.Ext(src); ext {
	case ".gz": // just assume .tar.gz
		if err := extractGzipArtifact(src, dst); err != nil {
			if err := os.Remove(dst); err != nil {
				return err
			}

			return err
		}
	case ".zip":
		if err := extractZipArtifact(src, dst); err != nil {
			if err := os.Remove(dst); err != nil {
				return err
			}

			return err
		}
	default:
		return fmt.Errorf("compression algorithm %s not supported", ext)
	}

	return nil
}

func extractGzipArtifact(src, dst string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	var tld string

	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if tld == "" {
			tld = strings.Split(h.Name, "/")[0]
		}

		target := path.Join(dst, strings.TrimPrefix(h.Name, tld+"/"))

		if target == dst {
			continue
		}

		switch h.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(h.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			outFile, err := os.Create(target)
			if err != nil {
				return err
			}
			defer outFile.Close()

			if _, err := io.Copy(outFile, tr); err != nil {
				return err
			}
		}
	}

	return nil
}

func extractZipArtifact(src, dst string) error {
	return nil
}
