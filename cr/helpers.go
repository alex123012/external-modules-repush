package cr

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"path"
	"strings"

	"github.com/Masterminds/semver/v3"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/iancoleman/strcase"
)

func FetchImage(repo, tag string, opts ...Option) (v1.Image, error) {
	regCli, err := newClient(repo, opts...)
	if err != nil {
		return nil, fmt.Errorf("fetch image error: %v", err)
	}

	img, err := regCli.Image(tag)
	if err != nil {
		return nil, fmt.Errorf("fetch image error: %v", err)
	}

	return img, nil
}

func PushImage(repo, tag string, image v1.Image, opts ...Option) error {
	r, err := newClient(repo, opts...)
	if err != nil {
		return err
	}

	if err := r.PushImage(tag, image); err != nil {
		return err
	}
	return nil
}

func FetchModuleReleaseImage(repo, moduleName, releaseChannel string, opts ...Option) (v1.Image, error) {
	return FetchImage(path.Join(repo, moduleName, "release"), strcase.ToKebab(releaseChannel), opts...)
}

func FetchModuleListingImage(repo, moduleName string, opts ...Option) (v1.Image, error) {
	return FetchImage(repo, moduleName, opts...)
}

func FetchModuleImages(repo, moduleName, moduleVersion string, opts ...Option) (v1.Image, []v1.Image, error) {
	imageRepo := path.Join(repo, moduleName)
	moduleImage, err := FetchImage(imageRepo, moduleVersion, opts...)
	if err != nil {
		return nil, nil, err
	}
	runImages, err := fetchModuleRunImages(imageRepo, moduleImage, opts...)
	return moduleImage, runImages, err
}

func fetchModuleRunImages(repo string, img v1.Image, opts ...Option) ([]v1.Image, error) {
	buf := bytes.NewBuffer(nil)
	if err := untarFile(mutate.Extract(img), func(hdr *tar.Header, tr *tar.Reader) (bool, error) {
		if hdr.Name == "images_digests.json" {
			_, err := io.Copy(buf, tr)
			if err != nil {
				return false, err
			}
		}
		return true, nil
	}); err != nil {
		return nil, err
	}

	var digestsMap map[string]string
	if err := json.Unmarshal(buf.Bytes(), &digestsMap); err != nil {
		return nil, err
	}

	newOpts := make([]Option, len(opts))
	copy(newOpts, opts)
	newOpts = append(newOpts, WithUseDigest())

	runImages := make([]v1.Image, 0)
	for _, digest := range digestsMap {
		runImage, err := FetchImage(repo, digest, newOpts...)
		if err != nil {
			return nil, err
		}
		runImages = append(runImages, runImage)
	}
	return runImages, nil
}

type moduleReleaseMetadata struct {
	Version *semver.Version `json:"version"`
}

func FetchModuleReleaseImageMetadata(img v1.Image) (string, error) {
	buf := bytes.NewBuffer(nil)
	var meta moduleReleaseMetadata
	layers, err := img.Layers()
	if err != nil {
		return "", err
	}

	for _, layer := range layers {
		size, err := layer.Size()
		if err != nil {
			return "", err
		}
		if size == 0 {
			// skip some empty werf layers
			continue
		}
		rc, err := layer.Uncompressed()
		if err != nil {
			return "", err
		}

		err = untarVersionLayer(rc, buf)
		if err != nil {
			return "", err
		}

		rc.Close()
	}
	log.Println("module version meta:", strings.TrimSuffix(buf.String(), "\n"))
	err = json.Unmarshal(buf.Bytes(), &meta)

	return "v" + meta.Version.String(), err
}

func untarVersionLayer(rc io.ReadCloser, rw io.Writer) error {
	return untarFile(rc, func(hdr *tar.Header, tr *tar.Reader) (bool, error) {
		if strings.HasPrefix(hdr.Name, ".werf") {
			return true, nil
		}

		if hdr.Name == "version.json" {
			_, err := io.Copy(rw, tr)
			if err != nil {
				return false, err
			}
		}
		return false, nil
	})
}

func untarFile(rc io.ReadCloser, f func(hdr *tar.Header, tr *tar.Reader) (bool, error)) error {
	tr := tar.NewReader(rc)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			// end of archive
			return nil
		}
		if err != nil {
			return err
		}
		cont, err := f(hdr, tr)
		if err != nil {
			return err
		}
		if !cont {
			return nil
		}
	}
}
