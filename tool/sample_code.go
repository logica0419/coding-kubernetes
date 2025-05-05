package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

func getLangsWithoutEN() ([]string, error) {
	langMap := []string{}

	files, err := os.ReadDir("docs")
	if err != nil {
		return nil, errors.WithStack(err)
	}

	for _, file := range files {
		if file.IsDir() && file.Name() != "en" && file.Name() != "public" {
			langMap = append(langMap, file.Name())
		}
	}

	return langMap, nil
}

func listSampleCodeFolders(prefix string) (map[string][]string, error) {
	sourceMap := make(map[string][]string)

	err := filepath.WalkDir(prefix, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") {
			return errors.WithStack(err)
		}

		pathSlice := strings.Split(path, "/")
		chapter := pathSlice[2]
		section := strings.Join(pathSlice[3:], "/")

		if _, ok := sourceMap[chapter]; !ok {
			sourceMap[chapter] = make([]string, 0)
		}

		sourceMap[chapter] = append(sourceMap[chapter], section)

		return nil
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return sourceMap, nil
}

func copySampleCodeFiles() error {
	sourceMap, err := listSampleCodeFolders("docs/en")
	if err != nil {
		return err
	}

	langMap, err := getLangsWithoutEN()
	if err != nil {
		return err
	}

	for _, lang := range langMap {
		for chapter, sections := range sourceMap {
			for _, section := range sections {
				srcPath := fmt.Sprintf("docs/en/%s/%s", chapter, section)
				dstPath := fmt.Sprintf("docs/%s/%s/%s", lang, chapter, section)

				const dirPermission = 0o755
				if err = os.MkdirAll(filepath.Dir(dstPath), dirPermission); err != nil {
					return errors.WithStack(err)
				}

				src, err := os.ReadFile(filepath.Clean(srcPath))
				if err != nil {
					return errors.WithStack(err)
				}

				const srcPermission = 0o644
				if err = os.WriteFile(dstPath, src, srcPermission); err != nil {
					return errors.WithStack(err)
				}
			}
		}
	}

	return nil
}

func generateSampleDiffFiles() error {
	langMap, err := getLangsWithoutEN()
	if err != nil {
		return err
	}

	langMap = append(langMap, "en")

	for _, lang := range langMap {
		sourceMap, err := listSampleCodeFolders("docs/" + lang)
		if err != nil {
			return err
		}

		for chapter, sections := range sourceMap {
			for i, section := range sections {
				if i == 0 {
					continue
				}

				path1 := fmt.Sprintf("docs/%s/%s/%s", lang, chapter, sections[i-1])
				path2 := fmt.Sprintf("docs/%s/%s/%s", lang, chapter, section)

				src1, err := os.ReadFile(filepath.Clean(path1))
				if err != nil {
					return errors.WithStack(err)
				}

				src2, err := os.ReadFile(filepath.Clean(path2))
				if err != nil {
					return errors.WithStack(err)
				}

				log.Printf("Generating diff between %s and %s\n", path1, path2)

				a, b, c := DiffLinesToRunes(string(src1), string(src2))
				diffs := DiffMainRunes(a, b)
				diffs = DiffCharsToLines(diffs, c)

				diffPath := fmt.Sprintf("docs/%s/%s/%s.diff", lang, chapter, section)
				if err = exportDiffs(diffPath, diffs); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func exportDiffs(filePath string, diffs []Diff) error {
	file, err := os.Create(filepath.Clean(filePath))
	if err != nil {
		return errors.WithStack(err)
	}

	defer func() { _ = file.Close() }()

	for _, diff := range diffs {
		switch diff.Type {
		case DiffEqual:
			_, err = fmt.Fprintf(file, "%s", diff.Text)
			if err != nil {
				return errors.WithStack(err)
			}

		case DiffDelete:
			texts := strings.Split(diff.Text, "\n")
			for _, text := range texts[:len(texts)-1] {
				_, err = fmt.Fprintf(file, "%s // [!code --]\n", text)
				if err != nil {
					return errors.WithStack(err)
				}
			}

		case DiffInsert:
			texts := strings.Split(diff.Text, "\n")
			for _, text := range texts[:len(texts)-1] {
				_, err = fmt.Fprintf(file, "%s // [!code ++]\n", text)
				if err != nil {
					return errors.WithStack(err)
				}
			}
		}
	}

	return nil
}
