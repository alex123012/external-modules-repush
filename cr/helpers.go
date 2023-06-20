package cr

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"path"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/iancoleman/strcase"
)

func GetAuthConfig(repo string, opts ...Option) (*authn.AuthConfig, string, error) {
	repoUrl, err := url.Parse("//" + repo)
	if err != nil {
		return nil, "", err
	}
	repoHost := repoUrl.Host
	cl := newClient(repo, opts...)
	ref, err := name.NewRegistry(repoHost, cl.nameOptions()...)
	if err != nil {
		return nil, "", err
	}

	auth, err := cl.options.authKeyChain.Resolve(ref)
	if err != nil {
		return nil, "", err
	}

	conf, err := auth.Authorization()
	return conf, repoHost, err
}

func PushImage(repo, tag string, image v1.Image, opts ...Option) error {
	r := newClient(repo, opts...)
	if err := r.pushImage(tag, image); err != nil {
		return err
	}
	return nil
}

func fetchImage(repo, tag string, opts ...Option) (v1.Image, error) {
	regCli := newClient(repo, opts...)
	img, err := regCli.fetchImage(tag)
	if err != nil {
		return nil, fmt.Errorf("fetch image error: %v", err)
	}
	return img, nil
}

func FetchModuleReleaseImage(repo, moduleName, releaseChannel string, opts ...Option) (v1.Image, error) {
	return fetchImage(path.Join(repo, moduleName, "release"), strcase.ToKebab(releaseChannel), opts...)
}

func FetchModuleListingImage(repo, moduleName string, opts ...Option) (v1.Image, error) {
	return fetchImage(repo, moduleName, opts...)
}

func FetchModuleImage(repo, moduleName, moduleVersion string, opts ...Option) (v1.Image, error) {
	imageRepo := path.Join(repo, moduleName)
	moduleImage, err := fetchImage(imageRepo, moduleVersion, opts...)
	if err != nil {
		return nil, err
	}
	return moduleImage, err
}

func FetchModuleRunImages(repo, moduleName string, img v1.Image, opts ...Option) (map[string]v1.Image, error) {
	regCli := newClient(path.Join(repo, moduleName), opts...)
	imagesFileName := "images_tags.json"
	if regCli.options.useDigest {
		imagesFileName = "images_digests.json"
	}

	buf := bytes.NewBuffer(nil)
	if err := untarFile(mutate.Extract(img), func(hdr *tar.Header, tr *tar.Reader) (bool, error) {
		if hdr.Name == imagesFileName {
			_, err := io.Copy(buf, tr)
			if err != nil {
				return false, err
			}
		}
		return true, nil
	}); err != nil {
		return nil, err
	}

	var TagsOrDigestsMap map[string]string
	if err := json.Unmarshal(buf.Bytes(), &TagsOrDigestsMap); err != nil {
		return nil, err
	}

	runImages := make(map[string]v1.Image, len(TagsOrDigestsMap))
	for name, tagOrDigest := range TagsOrDigestsMap {
		runImage, err := regCli.fetchImage(tagOrDigest)
		if err != nil {
			return nil, fmt.Errorf("fetch image error: %v", err)
		}

		if regCli.options.useDigest {
			runImages[name] = runImage
			continue
		}
		runImages[tagOrDigest] = runImage
	}
	return runImages, nil
}

type moduleReleaseMetadata struct {
	Version *semver.Version `json:"version"`
}

func ModuleReleaseImageMetadata(img v1.Image) (string, error) {
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
