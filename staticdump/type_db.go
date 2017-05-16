package staticdump

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/evepraisal/go-evepraisal/typedb"

	"gopkg.in/yaml.v2"
)

type TypeDB struct {
	staticDumpURL string
	dir           string

	typeMap map[string]typedb.EveType
}

func NewTypeDB(dir string, staticDumpURL string) (typedb.TypeDB, error) {

	typeDB := &TypeDB{
		typeMap:       make(map[string]typedb.EveType),
		staticDumpURL: staticDumpURL,
		dir:           dir,
	}

	if _, err := os.Stat(typeDB.staticDumpPath()); os.IsNotExist(err) {
		log.Printf("Downloading static dump to %s", typeDB.staticDumpPath())
		err := typeDB.downloadStaticDump()
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	log.Println("Load type data")
	err := typeDB.loadTypeData()
	if err != nil {
		return nil, err
	}

	log.Println("Done loading type data")

	return typeDB, nil
}

func (db *TypeDB) staticDumpPath() string {
	return filepath.Join(db.dir, filepath.Base(db.staticDumpURL))
}

func (db *TypeDB) HasType(typeName string) bool {
	_, ok := db.GetType(typeName)
	return ok
}

func (db *TypeDB) GetType(typeName string) (typedb.EveType, bool) {
	t, ok := db.typeMap[strings.ToLower(typeName)]
	return t, ok
}

func (db *TypeDB) Close() error {
	return nil
}

func (db *TypeDB) downloadStaticDump() error {
	out, err := os.Create(db.staticDumpPath())
	defer out.Close()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("GET", db.staticDumpURL, nil)
	if err != nil {
		return err
	}
	req.Header.Add("User-Agent", "go-evepraisal")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	n, err := io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	log.Printf("Successfully wrote %d bytes to %s", n, db.staticDumpPath())
	return nil
}

func (db *TypeDB) loadTypeData() error {
	r, err := zip.OpenReader(db.staticDumpPath())
	if err != nil {
		return err
	}
	defer r.Close()

	f, err := findZipFile(r.File, "sde/fsd/typeIDs.yaml")
	if err != nil {
		return err
	}

	fr, err := f.Open()
	if err != nil {
		return err
	}

	typeIDContents, err := ioutil.ReadAll(fr)
	if err != nil {
		return err
	}

	var allTypes map[int64]Type

	err = yaml.Unmarshal(typeIDContents, &allTypes)
	if err != nil {
		return err
	}

	typeMap := make(map[string]typedb.EveType)
	for typeID, t := range allTypes {
		typeMap[strings.ToLower(t.Name.En)] = typedb.EveType{
			ID:     typeID,
			Name:   t.Name.En,
			Volume: t.Volume,
		}
	}

	db.typeMap = typeMap

	return nil
}

type Type struct {
	Name struct {
		En string
	}
	Volume float64
}

func findZipFile(files []*zip.File, filename string) (*zip.File, error) {
	for _, f := range files {
		if filename == f.Name {
			return f, nil
		}
	}
	return nil, fmt.Errorf("Could not locate %s in archive", filename)
}
