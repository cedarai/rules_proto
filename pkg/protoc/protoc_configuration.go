package protoc

import (
	"path"
	"sort"
	"strings"
)

// ProtocConfiguration represents the complete configuration and source
// mappings.
type ProtocConfiguration struct {
	// The config for the p
	LanguageConfig *LanguageConfig
	// the workspace relative path of the BUILD file where this rule is being
	// generated.
	Rel string
	// the prefix for the rule (e.g. 'java')
	Prefix string
	// the library thar holds the proto files
	Library ProtoLibrary
	// the configuration for the plugins
	Plugins []*PluginConfiguration
	// The merged set of Source files for the compilations
	Outputs []string
	// The merged set of imports for the compilations
	Imports []string
	// The generated source mappings
	Mappings map[string]string
}

func newProtocConfiguration(lc *LanguageConfig, workDir, rel, prefix string, lib ProtoLibrary, plugins []*PluginConfiguration) *ProtocConfiguration {
	srcs, mappings := mergeSources(workDir, rel, plugins)
	imports := mergeImports(plugins)

	return &ProtocConfiguration{
		LanguageConfig: lc,
		Rel:            rel,
		Prefix:         prefix,
		Library:        lib,
		Plugins:        plugins,
		Outputs:        srcs,
		Imports:        imports,
		Mappings:       mappings,
	}
}

func (c *ProtocConfiguration) GetPluginConfiguration(implementationName string) *PluginConfiguration {
	for _, plugin := range c.Plugins {
		if plugin.Config.Implementation == implementationName {
			return plugin
		}
	}
	return nil
}

func (c *ProtocConfiguration) GetPluginOutputs(implementationName string) []string {
	plugin := c.GetPluginConfiguration(implementationName)
	if plugin == nil {
		return nil
	}
	return plugin.Outputs
}

// mergeSources computes the source files that are generated by the rule and any
// necessary mappings.
func mergeSources(workDir, rel string, plugins []*PluginConfiguration) ([]string, map[string]string) {
	srcs := make([]string, 0)
	mappings := make(map[string]string)

	for _, plugin := range plugins {

		// if plugin provided mappings for us, use those preferentially
		if len(plugin.Mappings) > 0 {
			srcs = append(srcs, plugin.Outputs...)

			for k, v := range plugin.Mappings {
				mappings[k] = v
			}
			continue
		}

		// otherwise, fallback to baseline method
		for _, filename := range plugin.Outputs {
			dir := path.Dir(filename)
			if dir == "." && rel == "" {
				dir = rel
			}
			if dir == rel {
				// no mapping required, just add to the srcs list
				srcs = append(srcs, strings.TrimPrefix(filename, rel+"/"))
			} else {
				// add the basename only to the srcs list and add a mapping.
				base := path.Base(filename)
				mappings[base] = filename
				srcs = append(srcs, base)
			}
		}
	}

	// if this is being built in an external workspace the workDir will be an abs path like
	// /private/var/tmp/_bazel_foo/452e264843978a138d8e9cb8305e394a/external/proto_googleapis.
	// all mappings will then be of the form {BIN_DIR}/external/{WORKSPACE_NAME}/...
	if workDir != "" {
		parts := strings.Split(workDir, "/")
		n := len(parts)
		if n >= 2 && parts[n-2] == "external" {
			for k, v := range mappings {
				mappings[k] = path.Join("external", parts[n-1], v)
			}
		}
	}

	return srcs, mappings
}

// mergeImports computes the merged list of imports for the list of plugins.
func mergeImports(plugins []*PluginConfiguration) []string {
	imports := make([]string, 0)

	for _, plugin := range plugins {
		imports = append(imports, plugin.Imports...)
	}

	sort.Strings(imports)

	return imports
}
