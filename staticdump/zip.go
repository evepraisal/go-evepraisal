package staticdump

import (
	"archive/zip"
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

func findZipFile(files []*zip.File, filename string) (*zip.File, error) {
	for _, f := range files {
		if filename == f.Name {
			return f, nil
		}
	}
	return nil, fmt.Errorf("Could not locate %s in archive", filename)
}

func loadDataFromZipFile(r *zip.ReadCloser, filename string, res interface{}) error {
	f, err := findZipFile(r.File, filename)
	if err != nil {
		return err
	}

	fr, err := f.Open()
	if err != nil {
		return err
	}

	contents, err := ioutil.ReadAll(fr)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(contents, res)
	if err != nil {
		return err
	}

	return nil
}
