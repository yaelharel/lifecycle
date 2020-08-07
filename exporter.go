package lifecycle

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/buildpacks/imgutil"
	"github.com/pkg/errors"

	"github.com/buildpacks/lifecycle/api"
	"github.com/buildpacks/lifecycle/cmd"
	"github.com/buildpacks/lifecycle/launch"
	"github.com/buildpacks/lifecycle/layers"
)

type Cache interface {
	Name() string
	SetMetadata(metadata CacheMetadata) error
	RetrieveMetadata() (CacheMetadata, error)
	AddLayerFile(tarPath string, sha string) error
	ReuseLayer(sha string) error
	RetrieveLayer(sha string) (io.ReadCloser, error)
	Commit() error
}

type Exporter struct {
	Buildpacks   []Buildpack
	LayerFactory LayerFactory
	Logger       Logger
	PlatformAPI  *api.Version
}

//go:generate mockgen -package testmock -destination testmock/layer_factory.go github.com/buildpacks/lifecycle LayerFactory
type LayerFactory interface {
	DirLayer(id string, dir string) (layers.Layer, error)
	LauncherLayer(path string) (layers.Layer, error)
	ProcessTypesLayer(metadata launch.Metadata) (layers.Layer, error)
	SliceLayers(dir string, slices []layers.Slice) ([]layers.Layer, error)
}

type LauncherConfig struct {
	Path     string
	Metadata LauncherMetadata
}

type ExportOptions struct {
	LayersDir          string
	AppDir             string
	WorkingImage       imgutil.Image
	RunImageRef        string
	OrigMetadata       LayersMetadata
	AdditionalNames    []string
	LauncherConfig     LauncherConfig
	Stack              StackMetadata
	Project            ProjectMetadata
	DefaultProcessType string
}

type ExportReport struct {
	Image ImageReport `toml:"image"`
}

type ImageReport struct {
	Tags    []string `toml:"tags"`
	ImageID string   `toml:"image-id,omitempty"`
	Digest  string   `toml:"digest,omitempty"`
}

func (e *Exporter) Export(opts ExportOptions) (ExportReport, error) {
	var err error

	opts.LayersDir, err = filepath.Abs(opts.LayersDir)
	if err != nil {
		return ExportReport{}, errors.Wrapf(err, "layers dir absolute path")
	}

	opts.AppDir, err = filepath.Abs(opts.AppDir)
	if err != nil {
		return ExportReport{}, errors.Wrapf(err, "app dir absolute path")
	}

	meta := LayersMetadata{}
	meta.RunImage.TopLayer, err = opts.WorkingImage.TopLayer()
	if err != nil {
		return ExportReport{}, errors.Wrap(err, "get run image top layer SHA")
	}

	meta.RunImage.Reference = opts.RunImageRef
	meta.Stack = opts.Stack

	buildMD := &BuildMetadata{}
	if _, err := toml.DecodeFile(launch.GetMetadataFilePath(opts.LayersDir), buildMD); err != nil {
		return ExportReport{}, errors.Wrap(err, "read build metadata")
	}

	// buildpack-provided layers
	if err := e.addBuildpackLayers(opts, &meta); err != nil {
		return ExportReport{}, err
	}

	// launcher layers (launcher binary, launcher config, process symlinks)
	if err := e.addLauncherLayers(opts, buildMD, &meta); err != nil {
		return ExportReport{}, err
	}

	// app layers (split into 1 or more slices)
	if err := e.addAppLayers(opts, buildMD.Slices, &meta); err != nil {
		return ExportReport{}, errors.Wrap(err, "exporting app layers")
	}

	if err := e.setLabels(opts, meta, buildMD); err != nil {
		return ExportReport{}, err
	}

	if err := e.setEnv(opts, buildMD.toLaunchMD()); err != nil {
		return ExportReport{}, err
	}

	entrypoint, err := e.entrypoint(buildMD.toLaunchMD(), opts.DefaultProcessType)
	if err != nil {
		return ExportReport{}, errors.Wrap(err, "determining entrypoint")
	}
	if err = opts.WorkingImage.SetEntrypoint(entrypoint); err != nil {
		return ExportReport{}, errors.Wrap(err, "setting entrypoint")
	}

	if err = opts.WorkingImage.SetCmd(); err != nil { // Note: Command intentionally empty
		return ExportReport{}, errors.Wrap(err, "setting cmd")
	}

	report := ExportReport{}
	report.Image, err = saveImage(opts.WorkingImage, opts.AdditionalNames, e.Logger)
	if err != nil {
		return ExportReport{}, err
	}

	return report, nil
}

func (e *Exporter) addBuildpackLayers(opts ExportOptions, meta *LayersMetadata) error {
	for _, bp := range e.Buildpacks {
		bpDir, err := readBuildpackLayersDir(opts.LayersDir, bp)
		if err != nil {
			return errors.Wrapf(err, "reading layers for buildpack '%s'", bp.ID)
		}
		bpMD := BuildpackLayersMetadata{
			ID:      bp.ID,
			Version: bp.Version,
			Layers:  map[string]BuildpackLayerMetadata{},
			Store:   bpDir.store,
		}
		for _, fsLayer := range bpDir.findLayers(forLaunch) {
			fsLayer := fsLayer
			lmd, err := fsLayer.read()
			if err != nil {
				return errors.Wrapf(err, "reading '%s' metadata", fsLayer.Identifier())
			}

			if fsLayer.hasLocalContents() {
				layer, err := e.LayerFactory.DirLayer(fsLayer.Identifier(), fsLayer.path)
				if err != nil {
					return errors.Wrapf(err, "creating layer")
				}
				origLayerMetadata := opts.OrigMetadata.MetadataForBuildpack(bp.ID).Layers[fsLayer.name()]
				lmd.SHA, err = e.addOrReuseLayer(opts.WorkingImage, layer, origLayerMetadata.SHA)
				if err != nil {
					return err
				}
			} else {
				if lmd.Cache {
					return fmt.Errorf("layer '%s' is cache=true but has no contents", fsLayer.Identifier())
				}
				origLayerMetadata, ok := opts.OrigMetadata.MetadataForBuildpack(bp.ID).Layers[fsLayer.name()]
				if !ok {
					return fmt.Errorf("cannot reuse '%s', previous image has no metadata for layer '%s'", fsLayer.Identifier(), fsLayer.Identifier())
				}

				e.Logger.Infof("Reusing layer '%s'\n", fsLayer.Identifier())
				e.Logger.Debugf("Layer '%s' SHA: %s\n", fsLayer.Identifier(), origLayerMetadata.SHA)
				if err := opts.WorkingImage.ReuseLayer(origLayerMetadata.SHA); err != nil {
					return errors.Wrapf(err, "reusing layer: '%s'", fsLayer.Identifier())
				}
				lmd.SHA = origLayerMetadata.SHA
			}
			bpMD.Layers[fsLayer.name()] = lmd
		}
		meta.Buildpacks = append(meta.Buildpacks, bpMD)

		if malformedLayers := bpDir.findLayers(forMalformed); len(malformedLayers) > 0 {
			ids := make([]string, 0, len(malformedLayers))
			for _, ml := range malformedLayers {
				ids = append(ids, ml.Identifier())
			}
			return fmt.Errorf("failed to parse metadata for layers '%s'", ids)
		}
	}
	return nil
}

func (e *Exporter) addLauncherLayers(opts ExportOptions, buildMD *BuildMetadata, meta *LayersMetadata) error {
	launcherLayer, err := e.LayerFactory.LauncherLayer(opts.LauncherConfig.Path)
	if err != nil {
		return errors.Wrap(err, "creating launcher layers")
	}
	meta.Launcher.SHA, err = e.addOrReuseLayer(opts.WorkingImage, launcherLayer, opts.OrigMetadata.Launcher.SHA)
	if err != nil {
		return errors.Wrap(err, "exporting launcher configLayer")
	}
	configLayer, err := e.LayerFactory.DirLayer("config", filepath.Join(opts.LayersDir, "config"))
	if err != nil {
		return errors.Wrapf(err, "creating layer '%s'", configLayer.ID)
	}
	meta.Config.SHA, err = e.addOrReuseLayer(opts.WorkingImage, configLayer, opts.OrigMetadata.Config.SHA)
	if err != nil {
		return errors.Wrap(err, "exporting config layer")
	}

	if err := e.launcherConfig(opts, buildMD, meta); err != nil {
		return err
	}
	return nil
}

func (e *Exporter) addAppLayers(opts ExportOptions, slices []layers.Slice, meta *LayersMetadata) error {
	// creating app layers (slices + app dir)
	sliceLayers, err := e.LayerFactory.SliceLayers(opts.AppDir, slices)
	if err != nil {
		return errors.Wrap(err, "creating app layers")
	}

	var numberOfReusedLayers int
	for _, slice := range sliceLayers {
		var err error

		found := false
		for _, previous := range opts.OrigMetadata.App {
			if slice.Digest == previous.SHA {
				found = true
				break
			}
		}
		if found {
			err = opts.WorkingImage.ReuseLayer(slice.Digest)
			numberOfReusedLayers++
		} else {
			err = opts.WorkingImage.AddLayerWithDiffID(slice.TarPath, slice.Digest)
		}
		if err != nil {
			return err
		}
		e.Logger.Debugf("Layer '%s' SHA: %s\n", slice.ID, slice.Digest)
		meta.App = append(meta.App, LayerMetadata{SHA: slice.Digest})
	}

	delta := len(sliceLayers) - numberOfReusedLayers
	if numberOfReusedLayers > 0 {
		e.Logger.Infof("Reusing %d/%d app layer(s)\n", numberOfReusedLayers, len(sliceLayers))
	}
	if delta != 0 {
		e.Logger.Infof("Adding %d/%d app layer(s)\n", delta, len(sliceLayers))
	}
	return nil
}

func (e *Exporter) setLabels(opts ExportOptions, meta LayersMetadata, buildMD *BuildMetadata) error {
	data, err := json.Marshal(meta)
	if err != nil {
		return errors.Wrap(err, "marshall metadata")
	}

	e.Logger.Infof("Adding label '%s'", LayerMetadataLabel)
	if err = opts.WorkingImage.SetLabel(LayerMetadataLabel, string(data)); err != nil {
		return errors.Wrap(err, "set app image metadata label")
	}

	buildMD.Launcher = opts.LauncherConfig.Metadata
	buildJSON, err := json.Marshal(buildMD)
	if err != nil {
		return errors.Wrap(err, "parse build metadata")
	}

	e.Logger.Infof("Adding label '%s'", BuildMetadataLabel)
	if err := opts.WorkingImage.SetLabel(BuildMetadataLabel, string(buildJSON)); err != nil {
		return errors.Wrap(err, "set build image metadata label")
	}

	projectJSON, err := json.Marshal(opts.Project)
	if err != nil {
		return errors.Wrap(err, "parse project metadata")
	}

	e.Logger.Infof("Adding label '%s'", ProjectMetadataLabel)
	if err := opts.WorkingImage.SetLabel(ProjectMetadataLabel, string(projectJSON)); err != nil {
		return errors.Wrap(err, "set project metadata label")
	}

	for _, label := range buildMD.Labels {
		e.Logger.Infof("Adding label '%s'", label.Key)
		if err := opts.WorkingImage.SetLabel(label.Key, label.Value); err != nil {
			return errors.Wrapf(err, "set buildpack-provided label '%s'", label.Key)
		}
	}
	return nil
}

func (e *Exporter) setEnv(opts ExportOptions, launchMD launch.Metadata) error {
	if err := opts.WorkingImage.SetEnv(cmd.EnvLayersDir, opts.LayersDir); err != nil {
		return errors.Wrapf(err, "set app image env %s", cmd.EnvLayersDir)
	}

	if err := opts.WorkingImage.SetEnv(cmd.EnvAppDir, opts.AppDir); err != nil {
		return errors.Wrapf(err, "set app image env %s", cmd.EnvAppDir)
	}

	if err := opts.WorkingImage.SetEnv(cmd.EnvPlatformAPI, e.PlatformAPI.String()); err != nil {
		return errors.Wrapf(err, "set app image env %s", cmd.EnvAppDir)
	}

	if err := opts.WorkingImage.SetEnv(cmd.EnvDeprecationMode, cmd.DeprecationModeQuiet); err != nil {
		return errors.Wrapf(err, "set app image env %s", cmd.EnvAppDir)
	}

	if e.supportsMulticallLauncher() {
		path, err := opts.WorkingImage.Env("PATH")
		if err != nil {
			return errors.Wrap(err, "failed to get PATH from app image")
		}
		path = strings.Join([]string{launch.ProcessDir, path}, string(os.PathListSeparator))
		if err := opts.WorkingImage.SetEnv("PATH", path); err != nil {
			return errors.Wrap(err, "set app image env PATH")
		}
	} else if opts.DefaultProcessType != "" {
		if _, ok := launchMD.FindProcessType(opts.DefaultProcessType); !ok {
			return processTypeError(launchMD, opts.DefaultProcessType)
		}
		if err := opts.WorkingImage.SetEnv(cmd.EnvProcessType, opts.DefaultProcessType); err != nil {
			return errors.Wrapf(err, "set app image env %s", cmd.EnvProcessType)
		}
	}
	return nil
}

func (e *Exporter) entrypoint(launchMD launch.Metadata, defaultProcessType string) (string, error) {
	if !e.supportsMulticallLauncher() {
		return launch.LauncherPath, nil
	}
	if defaultProcessType == "" {
		if len(launchMD.Processes) == 1 {
			return launchMD.Processes[0].Path(), nil
		}
		return launch.LauncherPath, nil
	}
	defaultProcess, ok := launchMD.FindProcessType(defaultProcessType)
	if !ok {
		return "", processTypeError(launchMD, defaultProcessType)
	}
	return defaultProcess.Path(), nil
}

// processTypes adds
func (e *Exporter) launcherConfig(opts ExportOptions, buildMD *BuildMetadata, meta *LayersMetadata) error {
	if e.supportsMulticallLauncher() {
		launchMD := launch.Metadata{
			Processes: buildMD.Processes,
		}
		if len(buildMD.Processes) > 0 {
			processTypesLayer, err := e.LayerFactory.ProcessTypesLayer(launchMD)
			if err != nil {
				return errors.Wrapf(err, "creating layer '%s'", processTypesLayer.ID)
			}
			meta.ProcessTypes.SHA, err = e.addOrReuseLayer(opts.WorkingImage, processTypesLayer, opts.OrigMetadata.ProcessTypes.SHA)
			if err != nil {
				return errors.Wrapf(err, "exporting layer '%s'", processTypesLayer.ID)
			}
		}
	}
	return nil
}

func (e *Exporter) supportsMulticallLauncher() bool {
	return e.PlatformAPI.Compare(api.MustParse("0.4")) >= 0
}

func processTypeError(launchMD launch.Metadata, defaultProcessType string) error {
	var typeList []string
	for _, p := range launchMD.Processes {
		typeList = append(typeList, p.Type)
	}
	return fmt.Errorf("default process type '%s' not present in list %+v", defaultProcessType, typeList)
}

func (e *Exporter) Cache(layersDir string, cacheStore Cache) error {
	var err error
	origMeta, err := cacheStore.RetrieveMetadata()
	if err != nil {
		return errors.Wrap(err, "metadata for previous cache")
	}
	meta := CacheMetadata{}

	for _, bp := range e.Buildpacks {
		bpDir, err := readBuildpackLayersDir(layersDir, bp)
		if err != nil {
			return errors.Wrapf(err, "reading layers for buildpack '%s'", bp.ID)
		}

		bpMD := BuildpackLayersMetadata{
			ID:      bp.ID,
			Version: bp.Version,
			Layers:  map[string]BuildpackLayerMetadata{},
		}
		for _, layer := range bpDir.findLayers(forCached) {
			layer := layer
			if !layer.hasLocalContents() {
				return fmt.Errorf("failed to cache layer '%s' because it has no contents", layer.Identifier())
			}
			lmd, err := layer.read()
			if err != nil {
				return errors.Wrapf(err, "reading %q metadata", layer.Identifier())
			}
			origLayerMetadata := origMeta.MetadataForBuildpack(bp.ID).Layers[layer.name()]
			if lmd.SHA, err = e.addOrReuseCacheLayer(cacheStore, &layer, origLayerMetadata.SHA); err != nil {
				return err
			}
			bpMD.Layers[layer.name()] = lmd
		}
		meta.Buildpacks = append(meta.Buildpacks, bpMD)
	}

	if err := cacheStore.SetMetadata(meta); err != nil {
		return errors.Wrap(err, "setting cache metadata")
	}
	if err := cacheStore.Commit(); err != nil {
		return errors.Wrap(err, "committing cache")
	}

	return nil
}
func (e *Exporter) addOrReuseLayer(image imgutil.Image, layer layers.Layer, previousSHA string) (string, error) {
	layer, err := e.LayerFactory.DirLayer(layer.ID, layer.TarPath)
	if err != nil {
		return "", errors.Wrapf(err, "creating layer '%s'", layer.ID)
	}
	if layer.Digest == previousSHA {
		e.Logger.Infof("Reusing layer '%s'\n", layer.ID)
		e.Logger.Debugf("Layer '%s' SHA: %s\n", layer.ID, layer.Digest)
		return layer.Digest, image.ReuseLayer(previousSHA)
	}
	e.Logger.Infof("Adding layer '%s'\n", layer.ID)
	e.Logger.Debugf("Layer '%s' SHA: %s\n", layer.ID, layer.Digest)
	return layer.Digest, image.AddLayerWithDiffID(layer.TarPath, layer.Digest)
}

func (e *Exporter) addOrReuseCacheLayer(cache Cache, layerDir layerDir, previousSHA string) (string, error) {
	layer, err := e.LayerFactory.DirLayer(layerDir.Identifier(), layerDir.Path())
	if err != nil {
		return "", errors.Wrapf(err, "creating layer %q", layer.ID)
	}
	if layer.Digest == previousSHA {
		e.Logger.Infof("Reusing cache layer '%s'\n", layer.ID)
		e.Logger.Debugf("Layer '%s' SHA: %s\n", layer.ID, layer.Digest)
		return layer.Digest, cache.ReuseLayer(previousSHA)
	}
	e.Logger.Infof("Adding cache layer '%s'\n", layer.ID)
	e.Logger.Debugf("Layer '%s' SHA: %s\n", layer.ID, layer.Digest)
	return layer.Digest, cache.AddLayerFile(layer.TarPath, layer.Digest)
}
