package image

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/pkg/errors"

	"github.com/openshift/assisted-service/pkg/executer"
	"github.com/openshift/installer/pkg/asset"
	"github.com/openshift/installer/pkg/asset/agent"
	"github.com/openshift/installer/pkg/asset/agent/manifests"
	"github.com/openshift/installer/pkg/asset/agent/mirror"
	"github.com/openshift/installer/pkg/rhcos"
	"github.com/sirupsen/logrus"
)

// BaseIso generates the base ISO file for the image
type BaseIso struct {
	File *asset.File
}

const (
	// TODO - add support for other architectures
	archName = "x86_64"
)

var (
	baseIsoFilename = ""
)

var _ asset.WritableAsset = (*BaseIso)(nil)

// Name returns the human-friendly name of the asset.
func (i *BaseIso) Name() string {
	return "BaseIso Image"
}

// getIsoFile is a pluggable function that gets the base ISO file
type getIsoFile func() (string, error)

type getIso struct {
	getter getIsoFile
}

func newGetIso(getter getIsoFile) *getIso {
	return &getIso{getter: getter}
}

// GetIsoPluggable defines the method to use get the baseIso file
var GetIsoPluggable = downloadIso

// Download the ISO using the URL in rhcos.json
func downloadIso() (string, error) {

	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()

	// Get the ISO to use from rhcos.json
	st, err := rhcos.FetchCoreOSBuild(ctx)
	if err != nil {
		return "", err
	}

	// Defaults to using the x86_64 baremetal ISO for all platforms
	// archName := arch.RpmArch(string(config.ControlPlane.Architecture))
	streamArch, err := st.GetArchitecture(archName)
	if err != nil {
		return "", err
	}
	if artifacts, ok := streamArch.Artifacts["metal"]; ok {
		if format, ok := artifacts.Formats["iso"]; ok {
			url := format.Disk.Location

			cachedImage, err := DownloadImageFile(url)
			if err != nil {
				return "", errors.Wrapf(err, "failed to download base ISO image %s", url)
			}
			return cachedImage, nil
		}
	} else {
		return "", errors.Wrap(err, "invalid artifact")
	}

	return "", fmt.Errorf("no ISO found to download for %s", archName)
}

func getIsoFromReleasePayload() (string, error) {

	// TODO
	return "", nil
}

// Dependencies returns dependencies used by the asset.
func (i *BaseIso) Dependencies() []asset.Asset {
	return []asset.Asset{
		&manifests.AgentManifests{},
		&agent.OptionalInstallConfig{},
		&mirror.RegistriesConf{},
	}
}

// Generate the baseIso
func (i *BaseIso) Generate(dependencies asset.Parents) error {

	log := logrus.New()
	// TODO - if image registry location is defined in InstallConfig,
	// ic := &agent.OptionalInstallConfig{}
	// p.Get(ic)

	// use the GetIso function to get the BaseIso from the release payload
	agentManifests := &manifests.AgentManifests{}
	dependencies.Get(agentManifests)

	var baseIsoFileName string
	var err error
	if agentManifests.ClusterImageSet != nil {
		releaseImage := agentManifests.ClusterImageSet.Spec.ReleaseImage
		pullSecret := agentManifests.GetPullSecretData()
		registriesConf := &mirror.RegistriesConf{}
		dependencies.Get(agentManifests, registriesConf)

		// If we have the image registry location and 'oc' command is available then get from release payload
		ocRelease := NewRelease(&executer.CommonExecuter{},
			Config{MaxTries: OcDefaultTries, RetryDelay: OcDefaultRetryDelay})

		log.Info("Extracting base ISO from release payload")
		baseIsoFileName, err = ocRelease.GetBaseIso(log, releaseImage, pullSecret, archName, registriesConf.MirrorConfig)
		if err == nil {
			log.Debugf("Extracted base ISO image %s from release payload", baseIsoFileName)
			i.File = &asset.File{Filename: baseIsoFileName}
			return nil
		}
		if !errors.Is(err, &exec.Error{}) { // Already warned about missing oc binary
			log.Warning("Failed to extract base ISO from release payload - check registry configuration")
		}
	}

	log.Info("Downloading base ISO")
	isoGetter := newGetIso(GetIsoPluggable)
	baseIsoFileName, err2 := isoGetter.getter()
	if err2 == nil {
		log.Debugf("Using base ISO image %s", baseIsoFileName)
		i.File = &asset.File{Filename: baseIsoFileName}
		return nil
	}
	log.Debugf("Failed to download base ISO: %s", err2)

	return errors.Wrap(err, "failed to get base ISO image")
}

// Files returns the files generated by the asset.
func (i *BaseIso) Files() []*asset.File {

	if i.File != nil {
		return []*asset.File{i.File}
	}
	return []*asset.File{}
}

// Load returns the cached baseIso
func (i *BaseIso) Load(f asset.FileFetcher) (bool, error) {

	if baseIsoFilename == "" {
		return false, nil
	}

	baseIso, err := f.FetchByName(baseIsoFilename)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, errors.Wrap(err, fmt.Sprintf("failed to load %s file", baseIsoFilename))
	}

	i.File = baseIso
	return true, nil
}
