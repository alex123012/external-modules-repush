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
	"github.com/iancoleman/strcase"
)

func FetchImage(repo, tag string, opts ...option) (v1.Image, error) {
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

func PushImage(repo, tag string, image v1.Image, opts ...option) error {
	r, err := newClient(repo, opts...)
	if err != nil {
		return err
	}

	if err := r.PushImage(tag, image); err != nil {
		return err
	}
	return nil
}

func FetchModuleReleaseImage(repo, moduleName, releaseChannel string, opts ...option) (v1.Image, error) {
	return FetchImage(path.Join(repo, moduleName, "release"), strcase.ToKebab(releaseChannel), opts...)
}

func FetchModuleImage(repo, moduleName, moduleVersion string, opts ...option) (v1.Image, error) {
	return FetchImage(path.Join(repo, moduleName), moduleVersion, opts...)
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
			// dcr.logger.Warnf("couldn't calculate layer size")
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
		if strings.HasPrefix(hdr.Name, ".werf") {
			continue
		}

		switch hdr.Name {
		case "version.json":
			_, err = io.Copy(rw, tr)
			if err != nil {
				return err
			}
			return nil

		default:
			continue
		}
	}
}
