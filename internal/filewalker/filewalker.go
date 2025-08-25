package filewalker

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type Type int

const (
	File Type = 1 << iota
	YAML
	JSON
	Text
	Unknown
	Directory
)

func (t *Type) From(extension string) {
	if strings.HasPrefix(extension, ".") {
		extension = strings.TrimPrefix(extension, ".")
	}

	var v Type
	switch strings.ToLower(extension) {
	case "directory":
		v = Unknown
	case "file":
		v = File
	case "yaml", "yml":
		v = YAML
	case "json", "json5":
		v = JSON
	case "text", "txt":
		v = Text
	case "unknown":
		v = Unknown
	default:
		v = Unknown
	}

	*t = v
}

func (t *Type) String() string {
	if t == nil {
		return "Unknown"
	}

	switch *(t) {
	case Directory:
		return "Directory"
	case File:
		return "File"
	case YAML:
		return "YAML"
	case JSON:
		return "JSON"
	case Text:
		return "Text"
	case Unknown:
		return "Unknown"
	default:
		return "Unknown"
	}
}

func (t *Type) Is() Is {
	if t == nil {
		return &is{t: Unknown}
	}

	return &is{t: *(t)}
}

func Extension(extension string) Type {
	var t Type
	t.From(extension)
	return t
}

// Is defines a set of methods for evaluating and categorizing a Type into distinct file or directory classifications.
// File checks if the provided Type corresponds to a file.
// YAML checks if the provided Type corresponds to YAML format files.
// JSON checks if the provided Type corresponds to JSON format files.
// Text checks if the provided Type corresponds to plaintext files.
// Unknown checks if the provided Type is an unknown file type.
// Directory checks if the provided Type corresponds to a directory.
type Is interface {
	// File evaluates if the given Type corresponds to a regular file.
	File() bool
	// YAML evaluates if the given Type corresponds to YAML format files.
	YAML() bool
	// JSON evaluates if the given Type corresponds to JSON format files.
	JSON() bool
	// Text evaluates if the given Type corresponds to plaintext files.
	Text() bool
	// Unknown evaluates if the given Type corresponds to an unknown file type.
	Unknown() bool
	// Directory evaluates if the given Type corresponds to a directory.
	Directory() bool
}

type is struct{ t Type }

func (i is) File() bool {
	return (i.t & File) == File
}

func (i is) YAML() bool {
	return (i.t & YAML) == YAML
}

func (i is) JSON() bool {
	return (i.t & JSON) == JSON
}

func (i is) Text() bool {
	return (i.t & Text) == Text
}

func (i is) Unknown() bool {
	return (i.t & Unknown) == Unknown
}

func (i is) Directory() bool {
	return (i.t & Directory) == Directory
}

// Runtime interface satisfaction constraint.
var _ Is = (*is)(nil)

// Descriptor represents a file system node with its properties and metadata.
// It contains information about the file's path, name, type (directory/file).
type Descriptor struct {
	Path string // Full path to the file or directory
	Name string // Name of the file or directory
	Type Type   // See [Type] for additional details.
}

func (d *Descriptor) Directory() string {
	if d == nil {
		return ""
	}

	if d.Type.Is().Directory() {
		return d.Path
	}

	return filepath.Dir(d.Path)
}

// Data represents a structure used for locating a target directory, defaulting to "testdata".
//
// The zero value for Data is valid.
type Data struct {
	target []string // target represents the search target -- defaults to "testdata"

	qualified string // the private actually qualified path representing the valid test-related data directory.
}

// New initializes and returns a pointer to a Data instance with default values.
func New(directory ...string) *Data {
	// Although directory is empty, reuse the slice to keep conditional branching more simple.
	if len(directory) == 0 {
		directory = slices.Insert(directory, 0, "testdata")
	}

	// Verify the arguments are valid
	for idx := range directory {
		v := directory[idx]
		if strings.Contains(v, string(filepath.Separator)) {
			slog.Warn("Basename Not Provided - Forcing Overwrite", slog.String("original", v), slog.String("replacement", filepath.Base(v)))

			v = filepath.Base(v)
		}

		directory[idx] = strings.TrimSpace(v)
	}

	return &Data{
		target: directory,
	}
}

// Find - by default the [Data.Find] callable searches for a "testdata" directory
// starting from the current working directory and moving up through parent directories. If not found,
// it attempts to locate it from the Git repository root. Returns the path to the testdata directory
// if found, or an error if not found.
//
// Callers can optionally provide other variadic directories to search for. Note that if arguments are
// present, "testdata" will not be searched for.
//
// "testdata" has special meaning to go and its tooling.
//
//	$ go help test | grep -i "testdata"
//
//	>>> The go tool will ignore a directory named "testdata", making it available
//	>>> to hold ancillary data needed by the tests.
//
// It's recommended to favor "testdata" so then artifacts, packages, or binaries aren't
// bloated by unneeded sources.
func (d *Data) Find() (string, error) {
	if d == nil {
		return "", fmt.Errorf("invalid zero value for data: %w", os.ErrInvalid)
	}

	// Unlikely to conditionally evaluate to true
	if len(d.target) == 0 {
		d.target = slices.Insert(d.target, 0, "testdata")
	}

	// Establish the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	cwd, e := filepath.Abs(cwd)
	if e != nil {
		slog.Error("Unable to Compute Absolute CWD Directory Path", slog.String("error", e.Error()))
	}

	// With error edges checked, establish a logger for the relatively vetted instance
	logger := slog.Default().With(slog.Any("input", d.target), slog.String("cwd", cwd))

	// Optionally establish the git directory
	var gitdirectory string
	if g, e := d.root(cwd); e == nil {
		abs, e := filepath.Abs(g)
		if e != nil {
			logger.Error("Unable to Compute Absolute Git Directory Path", slog.String("error", e.Error()))
		}

		gitdirectory = strings.TrimSpace(abs)
	}

	// Consider a lower logging level than warning
	if gitdirectory == "" {
		logger.Warn("Git Directory Not Found", slog.String("cwd", cwd))
	}

	for idx := range d.target {
		// Purposefully shadow a new instance of logger to prevent overwriting the upper namespace
		logger := logger.With(slog.Int("index", idx), slog.Any("input", d.target))

		literal := strings.TrimSpace(d.target[idx])
		logger = logger.With(slog.String("literal", literal))

		// Still exhaustively check string content
		if literal == "" {
			logger.Warn("Invalid Search Directory")
			continue
		}

		descriptor := filepath.Join(cwd, literal)

		metadata, e := os.Stat(descriptor)

		if e != nil {
			logger.Warn("Error While Evaluating Descriptor", slog.String("error", e.Error()))
			continue
		} else if !(metadata.IsDir()) {
			logger.Error("Invalid Search Directory", slog.String("value", metadata.Name()))
			continue
		}

		if v := d.find(descriptor, literal); v != "" && strings.TrimSpace(v) != "" {
			v = strings.TrimSpace(v)

			slog.Debug("Found Test Data Directory", slog.String("directory", v))

			d.qualified = v
			return v, nil
		}
	}

	slog.Warn("No Test Data Directory Found", slog.String("enumeration", "CWD"))

	if gitdirectory != "" {
		logger = logger.With(slog.String("git", gitdirectory))

		for idx := range d.target {
			// Purposefully shadow a new instance of logger to prevent overwriting the upper namespace
			logger := logger.With(slog.Int("index", idx), slog.Any("input", d.target))

			literal := strings.TrimSpace(d.target[idx])
			logger = logger.With(slog.String("literal", literal))

			// Still exhaustively check string content
			if literal == "" {
				logger.Warn("Invalid Search Directory")
				continue
			}

			descriptor := filepath.Join(gitdirectory, literal)

			metadata, e := os.Stat(descriptor)

			if e != nil {
				logger.Warn("Error While Evaluating Descriptor", slog.String("error", e.Error()))
				continue
			} else if !(metadata.IsDir()) {
				logger.Error("Invalid Search Directory", slog.String("value", metadata.Name()))
				continue
			}

			if v := d.find(descriptor, literal); v != "" && strings.TrimSpace(v) != "" {
				v = strings.TrimSpace(v)

				slog.Debug("Found Test Data Directory", slog.String("directory", v))

				d.qualified = v
				return v, nil
			}
		}
	}

	slog.Error("No Test Data Directory Found")

	return "", os.ErrNotExist
}

// Walk recursively walks through the directory structure starting from the
// provided root path, collecting information about all files and directories encountered.
// Returns a slice of FileNode structures representing the directory tree.
func (d *Data) Walk() ([]Descriptor, error) {
	var nodes []Descriptor

	if d == nil {
		return nodes, fmt.Errorf("invalid zero value for data: %w", os.ErrInvalid)
	} else if d.qualified == "" {
		v, e := d.Find()
		if e != nil {
			return nodes, e
		}

		// Superfluous, but reassigned for brevity
		d.qualified = v
	}

	// With error edges checked, establish a logger
	logger := slog.Default().With(slog.String("input", d.qualified))

	err := filepath.Walk(d.qualified, func(path string, info os.FileInfo, err error) error {
		// Purposefully shadow a new instance of logger to prevent overwriting the upper namespace
		logger := logger.With(slog.String("path", path))

		if err != nil {
			logger.Warn("Error While Walking Directory", slog.String("error", err.Error()))
			return err
		}

		var t Type
		if info.IsDir() {
			t = Directory
		}

		if t.Is().Directory() {
			return nil
		}

		node := Descriptor{
			Path: path,
			Name: info.Name(),
			Type: Extension(filepath.Ext(info.Name())),
		}

		nodes = append(nodes, node)

		return nil
	})

	return nodes, err
}

// find searches for a "testdata" directory starting from the provided
// path and moving up through parent directories until found or root is reached.
// Returns the path to the testdata directory if found, or an empty string if not found.
func (d *Data) find(start, desired string) string {
	current := start
	for {
		v := filepath.Join(current, desired)
		if _, err := os.Stat(v); err == nil {
			return v
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	return ""
}

// root attempts to locate the root directory of the Git repository starting
// from the provided path. Returns the path to the Git repository root if found,
// or an error if not found or if any Git-related operations fail.
func (d *Data) root(start string) (string, error) {
	repo, err := git.PlainOpenWithOptions(start, &git.PlainOpenOptions{
		DetectDotGit:          true,
		EnableDotGitCommonDir: false,
	})
	if err != nil {
		return "", err
	}

	wt, err := repo.Worktree()
	if err != nil {
		return "", err
	}

	return wt.Filesystem.Root(), nil
}
