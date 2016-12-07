package config

import (
	"io/ioutil"

	"github.com/craigmonson/colonize/util"
	"gopkg.in/yaml.v2"
)

// Find the main config file: .colonize.yaml will ALWAYS be in the root of the
// project.
const (
	valsFileExt = ".tfvars"
)

type ColonizeConfig struct {
	// Inputs
	Environment string
	OriginPath  string
	TmplName    string
	TmplPath    string
	CfgPath     string
	RootPath    string

	// Generated
	TmplRelPaths         []string
	WalkablePaths        []string
	WalkableValPaths     []string
	CombinedValsFilePath string

	// Read in from config
	Autogenerate_Comment      string
	Combined_Vals_File        string
	Combined_Vars_File        string
	Combined_Tf_File          string
	Vars_File_Env_Post_String string
	Vals_File_Env_Post_String string
	Templates_Dir             string
	Environments_Dir          string
}

type LoadConfigInput struct {
	// The environment
	Environment string
	// origin path where the command is run (typically cwd)
	OriginPath string
	// name for this template ie: vpc
	TmplName string
	// the difference between the cfg path and the root path.
	TmplPath string
	// path to config file
	CfgPath string
	// the root of the project (dir where config.yaml is)
	RootPath string
}

func LoadConfig(input *LoadConfigInput) (*ColonizeConfig, error) {
	conf := ColonizeConfig{
		Environment: input.Environment,
		OriginPath:  input.OriginPath,
		TmplName:    input.TmplName,
		TmplPath:    input.TmplPath,
		CfgPath:     input.CfgPath,
		RootPath:    input.RootPath,
	}

	contents, err := ioutil.ReadFile(input.CfgPath)
	if err != nil {
		return &conf, err
	}

	err = yaml.Unmarshal(contents, &conf)

	conf.initialize()

	return &conf, err
}

func LoadConfigInTree(path string, env string) (*ColonizeConfig, error) {
	cfgPath, err := util.FindCfgPath(path)
	if err != nil {
		return &ColonizeConfig{}, err
	}

	tmplName := util.GetBasename(path)
	rootPath := util.GetDir(cfgPath)
	tmplPath := util.GetTmplRelPath(path, rootPath)

	return LoadConfig(&LoadConfigInput{
		Environment: env,
		OriginPath:  path,
		TmplName:    tmplName,
		TmplPath:    tmplPath,
		CfgPath:     cfgPath,
		RootPath:    rootPath,
	})
}

func (c *ColonizeConfig) GetEnvValPath() string {
	return util.PathJoin(c.Environments_Dir, c.Environment+valsFileExt)
}

func (c *ColonizeConfig) initialize() {
	c.TmplRelPaths = util.GetTreePaths(c.TmplPath)

	// this will represent the root path in our relative paths.
	c.WalkablePaths = util.PrependPathToPaths(
		append([]string{""}, c.TmplRelPaths...),
		c.RootPath,
	)

	c.WalkableValPaths = util.AppendPathToPaths(
		c.WalkablePaths,
		c.GetEnvValPath(),
	)
	c.CombinedValsFilePath = util.PathJoin(c.OriginPath, c.Combined_Vals_File)
}