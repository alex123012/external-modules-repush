package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/alex123012/external-modules-transfer/cr"
	"github.com/alex123012/external-modules-transfer/templates"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

var (
	releaseChannel = "alpha"
	moduleName     string
	pullRepo       string
	pushRepo       string

	pullInsecure          bool
	pushInsecure          bool
	pullDisableAuth       bool
	pushDisableAuth       bool
	pullRunImageUseDigest bool
	pullRegistryCa        string
	pushRegistryCa        string
)

type image struct {
	repo  string
	tag   string
	image v1.Image
}

func main() {
	parseFlags()
	pullRepoOptions := parseOptions(pullInsecure, pullDisableAuth, pullRegistryCa)
	moduleListingImage, err := cr.FetchModuleListingImage(pullRepo, moduleName, pullRepoOptions...)
	if err != nil {
		log.Fatal(err)
	}

	moduleReleaseImage, err := cr.FetchModuleReleaseImage(pullRepo, moduleName, releaseChannel, pullRepoOptions...)
	if err != nil {
		log.Fatal(err)
	}

	moduleVersion, err := cr.ModuleReleaseImageMetadata(moduleReleaseImage)
	if err != nil {
		log.Fatal(fmt.Errorf("fetch release metadata error: %v", err))
	}

	moduleImage, err := cr.FetchModuleImage(pullRepo, moduleName, moduleVersion, pullRepoOptions...)
	if err != nil {
		log.Fatal(err)
	}
	if pullRunImageUseDigest {
		pullRepoOptions = append(pullRepoOptions, cr.WithUseDigest())
	}

	runImages, err := cr.FetchModuleRunImages(pullRepo, moduleName, moduleImage, pullRepoOptions...)
	if err != nil {
		log.Fatal(err)
	}

	imagesToPush := []image{
		{
			repo:  pushRepo,
			tag:   moduleName,
			image: moduleListingImage,
		},
		{
			repo:  path.Join(pushRepo, moduleName, "release"),
			tag:   releaseChannel,
			image: moduleReleaseImage,
		},
		{
			repo:  path.Join(pushRepo, moduleName),
			tag:   moduleVersion,
			image: moduleImage,
		},
	}

	for tag, runImage := range runImages {
		imagesToPush = append(imagesToPush, image{
			repo:  path.Join(pushRepo, moduleName),
			tag:   tag,
			image: runImage,
		})
	}

	pushRepoOptions := parseOptions(pushInsecure, pushDisableAuth, pushRegistryCa)
	for _, imgRef := range imagesToPush {
		if err := cr.PushImage(imgRef.repo, imgRef.tag, imgRef.image, pushRepoOptions...); err != nil {
			log.Fatal(err)
		}
	}

	if err := renderTemplates(moduleName, pushRepo, releaseChannel, pushRepoOptions...); err != nil {
		log.Fatal(err)
	}
}

func parseOptions(insecure, disableAuth bool, registryCa string) []cr.Option {
	opts := make([]cr.Option, 0)
	if insecure {
		opts = append(opts, cr.WithInsecureSchema())
	}
	if disableAuth {
		opts = append(opts, cr.WithDisabledAuth())
	}
	if registryCa != "" {
		opts = append(opts, cr.WithCA(registryCa))
	}
	return opts
}

func parseFlags() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		fmt.Fprintln(flag.CommandLine.Output(), "\n  This tool helps to transfer deckhouse external modules images")
		fmt.Fprintf(flag.CommandLine.Output(), "  from one container registry to another.\n\n")
		flag.PrintDefaults()
	}

	flag.StringVar(&pullRepo, "pull-registry", pullRepo, "registry address, that contains external modules\n(you should be logged in to registry via docker login)")
	flag.StringVar(&pushRepo, "push-registry", pushRepo, "registry address to push external module from pull repo\n(you should be logged in to registry via docker login)")
	flag.StringVar(&moduleName, "module", moduleName, "external module name")
	flag.StringVar(&releaseChannel, "release", releaseChannel, "release channel to use")

	flag.BoolVar(&pullDisableAuth, "pull-disable-auth", pullDisableAuth, "disable auth for pull registry")
	flag.BoolVar(&pushDisableAuth, "push-disable-auth", pushDisableAuth, "disable auth for push registry")

	flag.BoolVar(&pullInsecure, "pull-insecure", pullInsecure, "use http protocol for pull registry")
	flag.BoolVar(&pushInsecure, "push-insecure", pushInsecure, "use http protocol for push registry")

	flag.StringVar(&pullRegistryCa, "pull-ca", pullRegistryCa, "ca certificate for pull registry")
	flag.StringVar(&pushRegistryCa, "push-ca", pushRegistryCa, "ca certificate for push registry")

	flag.BoolVar(&pullRunImageUseDigest, "pull-run-image-use-digest", pullRunImageUseDigest,
		`use digests instead of tags for pulling images
  if flag is set - pushing images to 'push' repo will be with
  keys (image names) from images_digests.json file from module bundle image.
  This would prevent images cleanup in 'push' registry`,
	)

	flag.Parse()
	switch "" {
	case moduleName:
		log.Fatal("no module name provided")
	case pullRepo:
		log.Fatal("no repo address provided for pull")
	case pushRepo:
		log.Fatal("no repo address provided for push")
	case releaseChannel:
		log.Fatal("no release channel provided")
	}
}

func renderTemplates(name, repo, releaseChannel string, opts ...cr.Option) error {
	moduleSource, err := templates.RenderExternalModuleSource(name, repo, releaseChannel, opts...)
	if err != nil {
		return err
	}
	moduleConfig, err := templates.RenderModuleConfig(name)
	if err != nil {
		return err
	}
	builder := strings.Builder{}
	builder.WriteString("\n\n")
	builder.WriteString(moduleSource)
	builder.WriteString("\n")
	builder.WriteString(moduleConfig)
	builder.WriteString("\n\n")
	fmt.Println(builder.String())
	return nil
}
