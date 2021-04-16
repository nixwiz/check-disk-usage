package main

import (
	"fmt"
	"strings"
    "strconv"

	human "github.com/dustin/go-humanize"
	"github.com/sensu-community/sensu-plugin-sdk/sensu"
	"github.com/sensu/sensu-go/types"
	"github.com/shirou/gopsutil/disk"
)

// Config represents the check plugin config.
type Config struct {
	sensu.PluginConfig
	IncludeFSType   []string
	ExcludeFSType   []string
	IncludeFSPath   []string
	ExcludeFSPath   []string
	Warning         float64
	Critical        float64
	IncludePseudo   bool
	IncludeReadOnly bool
	FailOnError     bool
    InodeMode       bool
}

var (
	plugin = Config{
		PluginConfig: sensu.PluginConfig{
			Name:     "check-disk-usage",
			Short:    "Cross platform disk usage check for Sensu",
			Keyspace: "sensu.io/plugins/check-disk-usage/config",
		},
	}

	options = []*sensu.PluginConfigOption{
		{
			Path:      "include-fs-type",
			Env:       "",
			Argument:  "include-fs-type",
			Shorthand: "i",
			Default:   []string{},
			Usage:     "Comma separated list of file system types to check",
			Value:     &plugin.IncludeFSType,
		},
		{
			Path:      "exclude-fs-type",
			Env:       "",
			Argument:  "exclude-fs-type",
			Shorthand: "e",
			Default:   []string{},
			Usage:     "Comma separated list of file system types to exclude from checking",
			Value:     &plugin.ExcludeFSType,
		},
		{
			Path:      "include-fs-path",
			Env:       "",
			Argument:  "include-fs-path",
			Shorthand: "I",
			Default:   []string{},
			Usage:     "Comma separated list of file system paths to check",
			Value:     &plugin.IncludeFSPath,
		},
		{
			Path:      "exclude-fs-path",
			Env:       "",
			Argument:  "exclude-fs-path",
			Shorthand: "E",
			Default:   []string{},
			Usage:     "Comma separated list of file system paths to exclude from checking",
			Value:     &plugin.ExcludeFSPath,
		},
		{
			Path:      "warning",
			Env:       "",
			Argument:  "warning",
			Shorthand: "w",
			Default:   float64(85),
			Usage:     "Warning threshold for file system usage",
			Value:     &plugin.Warning,
		},
		{
			Path:      "critical",
			Env:       "",
			Argument:  "critical",
			Shorthand: "c",
			Default:   float64(95),
			Usage:     "Critical threshold for file system usage",
			Value:     &plugin.Critical,
		},
		{
			Path:      "include-pseudo-fs",
			Env:       "",
			Argument:  "include-pseudo-fs",
			Shorthand: "p",
			Default:   false,
			Usage:     "Include pseudo-filesystems (e.g. tmpfs) (default false)",
			Value:     &plugin.IncludePseudo,
		},
		{
			Path:      "fail-on-error",
			Env:       "",
			Argument:  "fail-on-error",
			Shorthand: "f",
			Default:   false,
			Usage:     "Fail and exit on errors getting file system usage (e.g. permission denied) (default false)",
			Value:     &plugin.FailOnError,
		},
		{
			Path:      "include-read-only",
			Env:       "",
			Argument:  "include-read-only",
			Shorthand: "r",
			Default:   false,
			Usage:     "Include read-only filesystems (default false)",
			Value:     &plugin.IncludeReadOnly,
		},
		{
			Path:      "inode-mode",
			Env:       "",
			Argument:  "inode-mode",
			Shorthand: "j",
			Default:   false,
			Usage:     "Inode mode, check usage of inode (default false)",
			Value:     &plugin.InodeMode,
		},

	}
)

func main() {
	check := sensu.NewGoCheck(&plugin.PluginConfig, options, checkArgs, executeCheck, false)
	check.Execute()
}

func checkArgs(event *types.Event) (int, error) {
	if len(plugin.IncludeFSType) > 0 && len(plugin.ExcludeFSType) > 0 {
		return sensu.CheckStateCritical, fmt.Errorf("--include-fs-type and --exclude-fs-type are mutually exclusive")
	}
	if len(plugin.IncludeFSPath) > 0 && len(plugin.ExcludeFSPath) > 0 {
		return sensu.CheckStateCritical, fmt.Errorf("--include-fs-path and --exclude-fs-path are mutually exclusive")
	}
	if plugin.Warning >= plugin.Critical {
		return sensu.CheckStateCritical, fmt.Errorf("--warning value can not be greater than or equal to --critical value")
	}
	return sensu.CheckStateOK, nil
}

func executeCheck(event *types.Event) (int, error) {
	var (
		criticals int
		warnings  int
	)

	parts, err := disk.Partitions(plugin.IncludePseudo)
	if err != nil {
		return sensu.CheckStateCritical, fmt.Errorf("Failed to get partitions, error: %v", err)
	}

	for _, p := range parts {
		// Ignore excluded (or non-included) file system types
		if !isValidFSType(p.Fstype) {
			continue
		}

		// Ignore excluded (or non-included) file systems
		if !isValidFSPath(p.Mountpoint) {
			continue
		}

		// Ignore read-only file systems?
		if !plugin.IncludeReadOnly && isReadOnly(p.Opts) {
			continue
		}

		device := p.Mountpoint
		s, err := disk.Usage(device)
		if err != nil {
			if plugin.FailOnError {
				return sensu.CheckStateCritical, fmt.Errorf("Failed to get disk usage for %s, error: %v", device, err)
			}
			fmt.Printf("%s  UNKNOWN: %s - error: %v\n", plugin.PluginConfig.Name, device, err)
			continue
		}

		// Ignore empty file systems
		if s.Total == 0 {
			continue
		}

        if plugin.InodeMode {

			// implement magic factor for larger file systems?
			fmt.Printf("%s ", plugin.PluginConfig.Name)
			if s.InodesUsedPercent >= plugin.Critical {
				criticals++
				fmt.Printf("CRITICAL: ")
			} else if s.InodesUsedPercent >= plugin.Warning {
				warnings++
				fmt.Printf(" WARNING: ")
			} else {
				fmt.Printf("      OK: ")
			}
			fmt.Printf("%s %.2f%% - Total: %s, Used: %s, Free: %s\n", p.Mountpoint, s.InodesUsedPercent, strconv.FormatUint(s.InodesTotal, 10), strconv.FormatUint(s.InodesUsed, 10), strconv.FormatUint(s.InodesFree, 10))
		} else {

			// implement magic factor for larger file systems?
			fmt.Printf("%s ", plugin.PluginConfig.Name)
			if s.UsedPercent >= plugin.Critical {
				criticals++
				fmt.Printf("CRITICAL: ")
			} else if s.UsedPercent >= plugin.Warning {
				warnings++
				fmt.Printf(" WARNING: ")
			} else {
				fmt.Printf("      OK: ")
			}
			fmt.Printf("%s %.2f%% - Total: %s, Used: %s, Free: %s\n", p.Mountpoint, s.UsedPercent, human.Bytes(s.Total), human.Bytes(s.Used), human.Bytes(s.Free))
		}

    }

	if criticals > 0 {
		return sensu.CheckStateCritical, nil
	} else if warnings > 0 {
		return sensu.CheckStateWarning, nil
	}

	return sensu.CheckStateOK, nil
}

func isValidFSType(fsType string) bool {
	// should i account for case insensitive searches for windows (ntfs vs. NTFS)?
	if len(plugin.IncludeFSType) > 0 && contains(plugin.IncludeFSType, fsType) {
		return true
	} else if len(plugin.IncludeFSType) > 0 {
		return false
	} else if len(plugin.ExcludeFSType) > 0 && contains(plugin.ExcludeFSType, fsType) {
		return false
	}

	// either not in exclude list or neither list is specified
	return true
}

func isValidFSPath(fsPath string) bool {
	// should i account for case insensitive searches for windows (c: vs. C:)?
	if len(plugin.IncludeFSPath) > 0 && contains(plugin.IncludeFSPath, fsPath) {
		return true
	} else if len(plugin.IncludeFSPath) > 0 {
		return false
	} else if len(plugin.ExcludeFSPath) > 0 && contains(plugin.ExcludeFSPath, fsPath) {
		return false
	}

	// either not in exclude list or neither list is specified
	return true
}

func isReadOnly(mountOpts string) bool {
	mOpts := strings.Split(mountOpts, ",")
	// "ro" should cover Linux, macOS, and Windows, "read-only" is reportd by mount(8)
	// on macOS so check for it, just in case
	if contains(mOpts, "ro") || contains(mOpts, "read-only") {
		return true
	}
	return false
}

func contains(a []string, s string) bool {
	for _, v := range a {
		if v == s {
			return true
		}
	}
	return false
}
