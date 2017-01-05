package config

import (
	"bufio"
	"io"
	"io/ioutil"
	"os"

	"github.com/craigmonson/colonize/util"
	"gopkg.in/yaml.v2"
)

type ConfigFile struct {
	Environments_Dir            string "environments_dir"
	Base_Environment_Ext        string "base_environment_ext"
	Autogenerate_Comment        string "autogenerate_comment"
	Combined_Vals_File          string "combined_vals_file"
	Combined_Vars_File          string "combined_vars_file"
	Combined_Derived_Vals_File  string "combined_derived_vals_file"
	Combined_Derived_Vars_File  string "combined_derived_vars_file"
	Combined_Tf_File            string "combined_tf_file"
	Combined_Remote_Config_File string "combined_remote_config_file"
	Remote_Config_File          string "remote_config_file"
	Derived_File                string "derived_file"
	Vals_File_Env_Post_String   string "vals_file_env_post_string"
	Branch_Order_File           string "branch_order_file"
}

var ConfigFileDefaults = ConfigFile{
	"env",
	"default",
	"This file generated by colonize",
	"_combined.tfvars",
	"_combined_variables.tf",
	"_combined_derived.tfvars",
	"_combined_derived.tf",
	"_combined.tf",
	"_remote_setup.sh",
	"remote_setup.sh",
	"derived.tfvars",
	".tfvars",
	"build_order.txt",
}

type Config struct {
	// Inputs
	Environment string
	OriginPath  string
	TmplName    string
	TmplPath    string
	CfgPath     string
	RootPath    string

	// Generated
	TmplRelPaths                []string
	WalkablePaths               []string
	WalkableValPaths            []string
	CombinedValsFilePath        string
	CombinedVarsFilePath        string
	WalkableTfPaths             []string
	CombinedTfFilePath          string
	WalkableDerivedPaths        []string
	CombinedDerivedValsFilePath string
	CombinedDerivedVarsFilePath string
	CombinedRemoteFilePath      string
	RemoteFilePath              string

	// Read in from config
	ConfigFile ConfigFile
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

const ConfigFileComment = "## Generated by Colonize init\n---\n"

func LoadConfig(input *LoadConfigInput) (*Config, error) {
	conf := Config{
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

	err = yaml.Unmarshal(contents, &conf.ConfigFile)

	conf.initialize()

	return &conf, err
}

func LoadConfigInTree(path string, env string) (*Config, error) {
	cfgPath, err := util.FindCfgPath(path)
	if err != nil {
		return &Config{}, err
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

func (c *Config) GetEnvValPath() string {
	return util.PathJoin(
		c.ConfigFile.Environments_Dir,
		c.Environment+c.ConfigFile.Vals_File_Env_Post_String,
	)
}

func (c *Config) GetEnvTfPath() string {
	return c.ConfigFile.Environments_Dir
}

func (c *Config) GetEnvDerivedPath() string {
	return util.PathJoin(c.ConfigFile.Environments_Dir, c.ConfigFile.Derived_File)
}

func (c *Config) IsBranch() bool {
	// if build_order.txt exists, then it's a branch, if not, we expect it to be
	// a leaf.
	branchPath := util.PathJoin(c.OriginPath, c.ConfigFile.Branch_Order_File)
	if _, err := os.Stat(branchPath); os.IsNotExist(err) {
		return false
	}

	return true
}

func (c *Config) IsNotBranch() bool {
	return !c.IsBranch()
}

func (c *Config) IsLeaf() bool {
	return !c.IsBranch()
}

func (c *Config) IsNotLeaf() bool {
	return !c.IsLeaf()
}

func (c *Config) initialize() {
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
	c.CombinedValsFilePath = util.PathJoin(c.OriginPath, c.ConfigFile.Combined_Vals_File)
	c.CombinedVarsFilePath = util.PathJoin(c.OriginPath, c.ConfigFile.Combined_Vars_File)

	c.WalkableTfPaths = util.AppendPathToPaths(
		c.WalkablePaths,
		c.GetEnvTfPath(),
	)
	c.CombinedTfFilePath = util.PathJoin(c.OriginPath, c.ConfigFile.Combined_Tf_File)

	c.WalkableDerivedPaths = util.AppendPathToPaths(
		c.WalkablePaths,
		c.GetEnvDerivedPath(),
	)
	c.CombinedDerivedValsFilePath = util.PathJoin(c.OriginPath, c.ConfigFile.Combined_Derived_Vals_File)
	c.CombinedDerivedVarsFilePath = util.PathJoin(c.OriginPath, c.ConfigFile.Combined_Derived_Vars_File)

	c.CombinedRemoteFilePath = util.PathJoin(c.OriginPath, c.ConfigFile.Combined_Remote_Config_File)
	c.RemoteFilePath = util.PathJoin(
		util.PathJoin(c.RootPath, c.ConfigFile.Environments_Dir),
		c.ConfigFile.Remote_Config_File,
	)
}

func (c *ConfigFile) ToYaml(buf io.Writer) error {
	output, err := yaml.Marshal(c)

	w := bufio.NewWriter(buf)
	_, err = w.WriteString(ConfigFileComment)
	if err != nil {
		return err
	}

	_, err = w.Write(output)
	if err != nil {
		return err
	}

	w.Flush()

	return nil
}

func (c *ConfigFile) WriteToFile(filename string) error {

	// create/open file
	f, err := os.Create(filename)
	if err != nil {
		return err
	}

	defer f.Close()

	return c.ToYaml(f)
}
