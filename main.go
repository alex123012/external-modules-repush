package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/alex123012/external-modules-transfer/cr"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

var (
	releaseChannel = "alpha"
	moduleName     string
	pullRepo       string
	pushRepo       string

	pullInsecure    bool
	pushInsecure    bool
	pullDisableAuth bool
	pushDisableAuth bool
	pullRegistryCa  string
	pushRegistryCa  string
)

func main() {
	parseFlags()

	moduleListingImage, err := cr.FetchImage(pullRepo, moduleName)
	if err != nil {
		log.Fatal(err)
	}

	moduleReleaseImage, err := cr.FetchModuleReleaseImage(pullRepo, moduleName, releaseChannel)
	if err != nil {
		log.Fatal(err)
	}

	moduleVersion, err := cr.FetchModuleReleaseImageMetadata(moduleReleaseImage)
	if err != nil {
		log.Fatal(fmt.Errorf("fetch release metadata error: %v", err))
	}

	moduleImage, err := cr.FetchModuleImage(pullRepo, moduleName, moduleVersion)
	if err != nil {
		log.Fatal(err)
	}

	imagesToPush := []struct {
		repo  string
		tag   string
		image v1.Image
	}{
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

	for _, imgRef := range imagesToPush {
		if err := cr.PushImage(imgRef.repo, imgRef.tag, imgRef.image); err != nil {
			log.Fatal(err)
		}
	}
}

func parseFlags() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		fmt.Fprintln(flag.CommandLine.Output(), "\n  This tool helps to transfer deckhouse external modules images")
		fmt.Fprintf(flag.CommandLine.Output(), "  from one container registry to another.\n\n")
		flag.PrintDefaults()
	}

	flag.StringVar(&pullRepo, "pull-registry", pullRepo, "registry address, that contains external modules")
	flag.StringVar(&pushRepo, "push-registry", pushRepo, "registry address to push external module from pull repo")
	flag.StringVar(&moduleName, "module", moduleName, "external module name")
	flag.StringVar(&releaseChannel, "release", releaseChannel, "release channel to use")

	flag.BoolVar(&pullDisableAuth, "pull-disable-auth", pullDisableAuth, "disable auth for pull registry")
	flag.BoolVar(&pushDisableAuth, "push-disable-auth", pushDisableAuth, "disable auth for push registry")

	flag.BoolVar(&pullInsecure, "pull-insecure", pullInsecure, "use http protocol for pull registry")
	flag.BoolVar(&pushInsecure, "push-insecure", pushInsecure, "use http protocol for push registry")

	flag.StringVar(&pullRegistryCa, "pull-ca", pullRegistryCa, "ca certificate for pull registry")
	flag.StringVar(&pushRegistryCa, "push-ca", pushRegistryCa, "ca certificate for push registry")

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
