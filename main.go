package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// defines several variables for parameterizing the protoc command. We can pull
// this out into a toml files in cases where we to vary this per package.
var (
	configPath string
	dryRun     bool
)

func init() {
	flag.StringVar(&configPath, "f", "Protobuild.toml", "override default config location")
	flag.BoolVar(&dryRun, "dryrun", false, "prints commands without running")
}

func main() {
	flag.Parse()

	c, err := readConfig(configPath)
	if err != nil {
		log.Fatalln(err)
	}

	pkgInfos, err := goPkgInfo(flag.Args()...)
	if err != nil {
		log.Fatalln(err)
	}

	gopath, err := gopathSrc()
	if err != nil {
		log.Fatalln(err)
	}

	gopathCurrent, err := gopathCurrent()
	if err != nil {
		log.Fatalln(err)
	}

	// For some reason, the golang protobuf generator makes the god awful
	// decision to output the files relative to the gopath root. It doesn't do
	// this only in the case where you give it ".".
	outputDir := filepath.Join(gopathCurrent, "src")

	for _, pkg := range pkgInfos {
		var includes []string
		includes = append(includes, c.Includes.Before...)

		vendor, err := closestVendorDir(pkg.Dir)
		if err != nil {
			if err != errVendorNotFound {
				log.Fatalln(err)
			}
		}

		if vendor != "" {
			// we also special case the inclusion of gogoproto in the vendor dir.
			// We could parameterize this better if we find it to be a common case.
			var vendoredIncludesResolved []string
			for _, vendoredInclude := range c.Includes.Vendored {
				vendoredIncludesResolved = append(vendoredIncludesResolved,
					filepath.Join(vendor, vendoredInclude))
			}

			includes = append(includes, vendoredIncludesResolved...)
			includes = append(includes, vendor)
		} else if len(c.Includes.Vendored) > 0 {
			log.Println("ignoring vendored includes: vendor directory not found")
		}

		includes = append(includes, gopath)
		includes = append(includes, c.Includes.After...)

		protoc := protocCmd{
			Name:       c.Generator,
			ImportPath: pkg.GoImportPath,
			PackageMap: c.Packages,
			Plugins:    c.Plugins,
			Files:      pkg.ProtoFiles,
			OutputDir:  outputDir,
			Includes:   includes,
		}

		arg, err := protoc.mkcmd()
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Println(arg)

		if dryRun {
			continue
		}

		if err := protoc.run(); err != nil {
			if err, ok := err.(*exec.ExitError); ok {
				if status, ok := err.Sys().(syscall.WaitStatus); ok {
					os.Exit(status.ExitStatus()) // proxy protoc exit status
				}
			}

			log.Fatalln(err)
		}
	}
}

type protoGoPkgInfo struct {
	Dir          string
	GoImportPath string
	ProtoFiles   []string
}

// goPkgInfo hunts down packages with proto files.
func goPkgInfo(golistpath ...string) ([]protoGoPkgInfo, error) {
	args := []string{
		"list", "-e", "-f", "{{.ImportPath}} {{.Dir}}"}
	args = append(args, golistpath...)
	cmd := exec.Command("go", args...)

	p, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var pkgInfos []protoGoPkgInfo
	lines := bytes.Split(p, []byte("\n"))
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		parts := bytes.Fields(line)
		if len(parts) != 2 {
			return nil, fmt.Errorf("bad output from command: %s", p)
		}

		pkgInfo := protoGoPkgInfo{
			Dir:          string(parts[1]),
			GoImportPath: string(parts[0]),
		}

		protoFiles, err := filepath.Glob(filepath.Join(pkgInfo.Dir, "*.proto"))
		if err != nil {
			return nil, err
		}
		if len(protoFiles) == 0 {
			continue // not a proto directory, skip
		}

		pkgInfo.ProtoFiles = protoFiles
		pkgInfos = append(pkgInfos, pkgInfo)
	}

	return pkgInfos, nil
}

// gopathSrc modifies GOPATH elements from env to include the src directory.
func gopathSrc() (string, error) {
	gopathAll := os.Getenv("GOPATH")

	if gopathAll == "" {
		return "", fmt.Errorf("must be run from a gopath")
	}

	var elements []string
	for _, element := range strings.Split(gopathAll, string(filepath.ListSeparator)) {
		elements = append(elements, filepath.Join(element, "src"))
	}

	return strings.Join(elements, string(filepath.ListSeparator)), nil
}

// gopathCurrent provides the top-level gopath for the current generation.
func gopathCurrent() (string, error) {
	gopathAll := os.Getenv("GOPATH")

	if gopathAll == "" {
		return "", fmt.Errorf("must be run from a gopath")
	}

	return strings.Split(gopathAll, string(filepath.ListSeparator))[0], nil
}

var errVendorNotFound = fmt.Errorf("no vendor dir found")

// closestVendorDir walks up from dir until it finds the vendor directory.
func closestVendorDir(dir string) (string, error) {
	dir = filepath.Clean(dir)
	for dir != filepath.VolumeName(dir)+string(filepath.Separator) {
		vendor := filepath.Join(dir, "vendor")
		fi, err := os.Stat(vendor)
		if err != nil {
			if os.IsNotExist(err) {
				// up we go!
				dir = filepath.Dir(dir)
				continue
			}
			return "", err
		}

		if !fi.IsDir() {
			// up we go!
			dir = filepath.Dir(dir)
			continue
		}

		return vendor, nil
	}

	return "", errVendorNotFound
}
