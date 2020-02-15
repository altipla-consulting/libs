package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	flag "github.com/spf13/pflag"
	log "github.com/sirupsen/logrus"
	"golang.org/x/tools/go/packages"

	"libs.altipla.consulting/errors"
)

var (
  std = make(map[string]bool)
)

func main() {
  if err := run(); err != nil {
    log.Fatal(errors.Stack(err))
  }
}

func run() error {
  var flagLocalPrefix string
  flag.StringVarP(&flagLocalPrefix, "local-prefix", "p", "", "Prefix of the local package")
  flag.Parse()

  if flagLocalPrefix == "" {
    return errors.Errorf("-p argument required")
  }

  fn := func(path string, info os.FileInfo, err error) error {
    if err != nil {
      return errors.Trace(err)
    }
    if info.IsDir() {
      return nil
    }
    if filepath.Ext(path) != ".go" {
      return nil
    }

    f, err := os.OpenFile(path, os.O_RDWR, 0)
    if err != nil {
      return errors.Trace(err)
    }
    defer f.Close()

    content, err := ioutil.ReadAll(f)
    if err != nil {
      return errors.Trace(err)
    }
    lines := strings.Split(string(content), "\n")

    var buf bytes.Buffer
    unchanged, err := groupImports(flagLocalPrefix, path, &buf, lines)
    if err != nil {
      return errors.Trace(err)
    }

    if !unchanged && string(content) != buf.String() {
      log.WithField("path", path).Info("Wrong import order detected, fixing in place")

      if _, err := f.Seek(0, io.SeekStart); err != nil {
        return errors.Trace(err)
      }
      if err := f.Truncate(0); err != nil {
        return errors.Trace(err)
      }

      if _, err := io.Copy(f, &buf); err != nil {
        return errors.Trace(err)
      }
    }

    return nil
  }
  for _, root := range flag.Args() {
    if err := filepath.Walk(root, fn); err != nil {
      return errors.Trace(err)
    }
  }

  return nil
}

func isEnd(lines []string, line int) bool {
  return line >= len(lines)
}

type importSpec struct {
  name, path string
}

func groupImports(flagLocalPrefix, filename string, w io.Writer, lines []string) (bool, error) {
  var line int
  var firstLine, lastLine int

  // Read until the import group
  for ; !isEnd(lines, line); line++ {
    if lines[line] == "import (" {
      firstLine = line + 1
      break
    }
    line++
  }
  if isEnd(lines, line) {
    return true, nil
  }
  line++

  var imports []importSpec
  for ; !isEnd(lines, line); line++ {
    if lines[line] == ")" {
      lastLine = line - 1
      break
    }

    trimmed := strings.TrimSpace(lines[line])
    if trimmed == "" {
      continue
    }
    if strings.Contains(trimmed, " ") {
      parts := strings.Split(trimmed, " ")
      if len(parts) != 2 {
        return false, errors.Errorf("unparsable line(%s, %d): %s", filename, line, lines[line])
      }
      path, err := strconv.Unquote(parts[1])
      if err != nil {
        return false, errors.Trace(err)
      }
      imports = append(imports, importSpec{
        name: parts[0],
        path: path,
      })
    } else {
      path, err := strconv.Unquote(trimmed)
      if err != nil {
        return false, errors.Wrapf(err, "line(%d): %s", line, lines[line])
      }
      imports = append(imports, importSpec{path: path})
    }
  }
  if isEnd(lines, line) {
    return true, nil
  }

  if len(std) == 0 {
    pkgs, err := packages.Load(nil, "std")
    if err != nil {
      return false, errors.Trace(err)
    }
    for _, pkg := range pkgs {
      std[pkg.PkgPath] = true
    }
  }

  var system, libs, local []importSpec
  for _, imp := range imports {
    switch {
    case std[imp.path]:
      system = append(system, imp)
    case strings.HasPrefix(imp.path, flagLocalPrefix):
      local = append(local, imp)
    default:
      libs = append(libs, imp)
    }
  }

  for i, line := range lines[:len(lines)-1] {
    if i < firstLine || i > lastLine {
      fmt.Fprintln(w, line)
    } else if i == firstLine {
      printImports(w, system, false)
      printImports(w, libs, len(system) > 0)
      printImports(w, local, len(system) > 0 || len(libs) > 0)
    }
  }

  return false, nil
}

func printImports(w io.Writer, imports []importSpec, separator bool) {
  if separator && len(imports) > 0 {
    fmt.Fprintln(w, "")
  }
  for _, imp := range imports {
    if imp.name != "" {
      fmt.Fprintln(w, "\t"+imp.name+` "`+imp.path+`"`)
    } else {
      fmt.Fprintln(w, "\t"+`"`+imp.path+`"`)
    }
  }
}
