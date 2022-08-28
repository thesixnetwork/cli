package chainconfig

import (
	"io"

	"gopkg.in/yaml.v2"

	"github.com/ignite-hq/cli/ignite/chainconfig/config"
	v1 "github.com/ignite-hq/cli/ignite/chainconfig/v1"
)

// Build time check for the latest config version type.
// This is required to be sure that conversion to latest
// doesn't break when a new config version is added without
// updating the references to the previous version.
var _ = Versions[LatestVersion].(*v1.Config)

// ConvertLatest converts a config to the latest version.
func ConvertLatest(c config.Converter) (_ *v1.Config, err error) {
	for c.GetVersion() < LatestVersion {
		c, err = c.ConvertNext()
		if err != nil {
			return nil, err
		}
	}

	// Cast to the latest version type.
	// This is safe because there is a build time check that makes sure
	// the type for the latest config version is the right one here.
	return c.(*v1.Config), nil
}

// MigrateLatest migrates a config file to the latest version.
func MigrateLatest(configFile io.ReadWriteSeeker) error {
	// Parse the config file and convert it to the latest version
	cfg, err := Parse(configFile)
	if err != nil {
		return err
	}

	// Position at the beginning of the file before writing the new version
	if _, err := configFile.Seek(0, 0); err != nil {
		return err
	}

	return yaml.NewEncoder(configFile).Encode(cfg)
}
