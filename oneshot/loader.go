package oneshot

import (
    "io/ioutil"
    "os"
    "path/filepath"
    "strings"

    "github.com/pkg/errors"
    "github.com/docker/cli/cli/compose/schema"
    "github.com/docker/cli/cli/compose/types"
    "github.com/docker/cli/cli/compose/loader"
)

func loadComposefile(filenames []string) (*types.Config, error) {
    details, err := getConfigDetails(filenames);
    if err != nil {
        return nil, errors.Wrap(err, "Error parsing config")
    }

    config, err := loader.Load(details)

    if err != nil {
        return nil, errors.Wrap(err, "Error resolving configs")
    }

    return config, nil
}

func getConfigDetails(filenames []string) (types.ConfigDetails, error) {
    var details types.ConfigDetails

    absPath, err := filepath.Abs(filenames[0])
    if err != nil {
        return details, errors.Wrap(err,
            "Could not get absolute path to the compose file workdir")
    }
    details.WorkingDir = filepath.Dir(absPath)

    configs, err := loadConfigFiles(filenames);
    if err != nil {
        return details, errors.Wrap(err, "Could not load config files")
    }
    details.ConfigFiles = configs

    details.Version = schema.Version(details.ConfigFiles[0].Config)
    details.Environment, err = buildEnvironment(os.Environ())
    return details, err
}

// These three functions are taken from loader/loader.go

func buildEnvironment(env []string) (map[string]string, error) {
    result := make(map[string]string, len(env))
    for _, s := range env {
        // if value is empty, s is like "K=", not "K".
        if !strings.Contains(s, "=") {
            return result, errors.Errorf("unexpected environment %q", s)
        }
        kv := strings.SplitN(s, "=", 2)
        result[kv[0]] = kv[1]
    }
    return result, nil
}

func loadConfigFiles(filenames []string) ([]types.ConfigFile, error) {
    var configFiles []types.ConfigFile

    for _, filename := range filenames {
        configFile, err := loadConfigFile(filename)
        if err != nil {
            return configFiles, errors.Wrap(err,
                "Could not load config file " + filename)
        }
        configFiles = append(configFiles, *configFile)
    }

    return configFiles, nil
}

func loadConfigFile(filename string) (*types.ConfigFile, error) {
    var bytes []byte
    var err error

    bytes, err = ioutil.ReadFile(filename)

    if err != nil {
        return nil, errors.Wrap(err, "Could not read file " + filename)
    }

    config, err := loader.ParseYAML(bytes)
    if err != nil {
        return nil, errors.Wrap(err, "Could not parse YAML " + filename)
    }

    return &types.ConfigFile{
        Filename: filename,
        Config:   config,
    }, nil
}
