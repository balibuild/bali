package main

import (
	"time"

	"github.com/balibuild/bali/v2/rpmpack"
)

type Spec struct {
	Name        string            `toml:"name"` //the package name
	Version     string            `toml:"version"`
	Release     string            `toml:"release"`
	Epoch       int64             `toml:"epoch"`
	Arch        string            `toml:"arch,omitempty"`
	BuidTime    time.Time         `toml:"build_time,omitempty"`
	OsName      string            `toml:"os,omitempty"`
	Summary     string            `toml:"summary,omitempty"`
	Description string            `toml:"description,omitempty"`
	Vendor      string            `toml:"vendor,omitempty"`
	Packager    string            `toml:"packager,omitempty"`
	Group       string            `toml:"group,omitempty"`
	URL         string            `toml:"url,omitempty"`
	License     string            `toml:"license,omitempty"`
	Compressor  string            `toml:"compressor,omitempty"` // default gzip
	Prein       string            `toml:"prein,omitempty"`
	Postin      string            `toml:"postin,omitempty"`
	Preun       string            `toml:"preun,omitempty"`
	Postun      string            `toml:"postun,omitempty"`
	Provides    rpmpack.Relations `toml:"provides,omitempty"`
	Obsoletes   rpmpack.Relations `toml:"obsoletes,omitempty"`
	Suggests    rpmpack.Relations `toml:"suggests,omitempty"`
	Recommends  rpmpack.Relations `toml:"recommends,omitempty"`
	Requires    rpmpack.Relations `toml:"requires,omitempty"`
	Conflicts   rpmpack.Relations `toml:"conflicts,omitempty"`
}
