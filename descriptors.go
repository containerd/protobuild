package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
)

type descriptorSet struct {
	merged      descriptor.FileDescriptorSet
	seen        map[string]struct{}
	ignoreFiles map[string]struct{}
	descProto   string
	protoPath   string
}

func newDescriptorSet(ignoreFiles []string, d, p string) *descriptorSet {
	ifm := make(map[string]struct{}, len(ignoreFiles))
	for _, ignore := range ignoreFiles {
		ifm[ignore] = struct{}{}
	}
	return &descriptorSet{
		seen:        make(map[string]struct{}),
		ignoreFiles: ifm,
		descProto:   d,
		protoPath:   p,
	}
}

func (d *descriptorSet) add(descs ...*descriptor.FileDescriptorProto) {
	for _, file := range descs {
		name := file.GetName()
		if _, ok := d.seen[name]; ok {
			continue
		}

		if _, ok := d.ignoreFiles[name]; ok {
			continue
		}

		// TODO(stevvooe): If we want to filter certain fields in the descriptor,
		// this is the place to do it. May be necessary if certain fields are
		// noisy, such as option fields.
		d.merged.File = append(d.merged.File, file)
		d.seen[name] = struct{}{}
	}
}

// stabilize outputs the merged protobuf descriptor set into the provided writer.
//
// This is equivalent to the following command:
//
// cat merged.pb | protoc --decode google.protobuf.FileDescriptorSet /path/to/google/protobuf/descriptor.proto
func (d *descriptorSet) marshalTo(w io.Writer) error {
	p, err := proto.Marshal(&d.merged)
	if err != nil {
		return err
	}

	args := []string{
		"protoc",
		"--decode",
		"google.protobuf.FileDescriptorSet",
		"--proto_path",
		d.protoPath,
		d.descProto,
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = bytes.NewReader(p)
	cmd.Stdout = w
	cmd.Stderr = os.Stderr

	if !quiet {
		fmt.Println(strings.Join(args, " "))
	}
	return cmd.Run()
}

func readDesc(path string) (*descriptor.FileDescriptorSet, error) {
	var desc descriptor.FileDescriptorSet

	p, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := proto.Unmarshal(p, &desc); err != nil {
		log.Fatalln(err)
	}

	return &desc, nil
}
