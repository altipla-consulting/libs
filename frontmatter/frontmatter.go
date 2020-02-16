package frontmatter

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/golang/protobuf/proto"
	"gopkg.in/yaml.v2"

	"libs.altipla.consulting/errors"
)

// Unmarshal reads the input and parses the frontmatter into data returning the
// content. If data is a proto.Message it will use Protobuf Text, otherwise it will
// use YAML to read it.
func Unmarshal(rd io.Reader, data interface{}) (string, error) {
	r := bufio.NewReader(rd)

	for {
		line, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return "", nil
			}

			return "", errors.Wrapf(err, "cannot read file line")
		}
		line = strings.TrimSpace(line)
		if line == "---" {
			break
		} else if line == "" {
			continue
		}

		return "", errors.Errorf("unknown file header line: %s", line)
	}

	var header []string
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return "", errors.Wrapf(err, "cannot read header")
		}
		line = strings.TrimSpace(line)
		if line == "---" {
			break
		}

		header = append(header, line+"\n")
	}

	content, err := ioutil.ReadAll(r)
	if err != nil {
		return "", errors.Wrapf(err, "cannot read content")
	}

	msg, ok := data.(proto.Message)
	if ok {
		if err := proto.UnmarshalText(strings.Join(header, ""), msg); err != nil {
			return "", errors.Wrapf(err, "cannot decode proto header")
		}
	} else {
		if err := yaml.Unmarshal([]byte(strings.Join(header, "")), data); err != nil {
			return "", errors.Wrapf(err, "cannot decode yaml header")
		}
	}

	return string(content[1:]), nil
}

// Marshal writes a frontmatter file with data in the header and content in
// the body. If data is a proto.Message it will use Protobuf Text, otherwise
// it will use a simple YAML serialization.
func Marshal(w io.Writer, data interface{}, content string) error {
	var header string
	msg, ok := data.(proto.Message)
	if ok {
		header = proto.MarshalTextString(msg)
	} else {
		out, err := yaml.Marshal(data)
		if err != nil {
			return errors.Wrapf(err, "cannot encode proto header")
		}
		header = string(out)
	}

	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	if _, err := fmt.Fprintf(w, "---\n%s---\n\n%s", header, content); err != nil {
		return errors.Wrapf(err, "cannot encode yaml header")
	}

	return nil
}
