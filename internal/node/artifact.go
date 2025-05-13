package node

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"
)

type Artifact struct {
	Name, Slug, Ext string
}

func DownloadArtifact(v, s string) (Artifact, error) {
	u, err := url.JoinPath("https://nodejs.org/dist", v, s)
	if err != nil {
		return Artifact{}, err
	}

	r, err := http.Get(u)
	if err != nil {
		return Artifact{}, err
	}

	if r.StatusCode >= 400 {
		return Artifact{}, fmt.Errorf("failed to download artifact %s: request failed with status %s", s, r.Status)
	}

	defer r.Body.Close()

	f, err := os.Create(path.Join(os.TempDir(), s))
	if err != nil {
		return Artifact{}, err
	}

	defer f.Close()

	if _, err := io.Copy(f, r.Body); err != nil {
		return Artifact{}, err
	}

	return Artifact{f.Name(), s, path.Ext(f.Name())}, nil
}

func ArtifactSlug(v, hostOS, hostArch, ext string) string {
	return fmt.Sprintf("node-%s-%s-%s%s", v, hostOS, hostArch, ext)
}

func ExtractArtifact(src, dst string) error {
	var err error

	switch ext := path.Ext(src); ext {
	case ".xz": // assume .tar.xz
		err = extractXZArtifact(src, dst)
	case ".gz": // just assume .tar.gz
		err = extractGzipArtifact(src, dst)
	case ".zip":
		err = extractZipArtifact(src, dst)
	default:
		return fmt.Errorf("compression algorithm %s not supported", ext)
	}

	if err != nil {
		if err := os.Remove(dst); err != nil {
			return err
		}

		return err
	}

	return nil
}

func extractXZArtifact(src, dst string) error {
	cmd := exec.Command("tar", "-C", dst, "-xJf", src, "--strip-components=1")
	if err := cmd.Run(); err != nil {
		return err
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
			outFile, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE, os.FileMode(h.Mode))
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
