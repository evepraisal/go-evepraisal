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
	err := typeDB.loadData()
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

func (db *TypeDB) loadData() error {
	r, err := zip.OpenReader(db.staticDumpPath())
	if err != nil {
		return err
	}
	defer r.Close()

	var allTypes map[int64]Type
	err = loadDataFromZipFile(r, "sde/fsd/typeIDs.yaml", &allTypes)
	if err != nil {
		return err
	}
	log.Printf("Loaded %d types", len(allTypes))

	var allBlueprints map[int64]Blueprint
	err = loadDataFromZipFile(r, "sde/fsd/blueprints.yaml", &allBlueprints)
	if err != nil {
		return err
	}
	log.Printf("Loaded %d blueprints", len(allBlueprints))

	blueprintsByProductType := make(map[int64][]Blueprint)
	for _, blueprint := range allBlueprints {
		for _, product := range blueprint.Activities.Manufacturing.Products {
			blueprints, ok := blueprintsByProductType[product.TypeID]
			if ok {
				blueprintsByProductType[product.TypeID] = append(blueprints, blueprint)
			} else {
				blueprintsByProductType[product.TypeID] = []Blueprint{blueprint}
			}
		}
	}

	typeMap := make(map[string]typedb.EveType)
	for typeID, t := range allTypes {
		if !t.Published {
			continue
		}

		eveType := typedb.EveType{
			ID:              typeID,
			Name:            t.Name.En,
			Volume:          t.Volume,
			BaseComponenets: resolveBaseComponents(blueprintsByProductType, typeID, 1, 5),
		}

		typeMap[strings.ToLower(t.Name.En)] = eveType
	}

	db.typeMap = typeMap

	return nil
}

func resolveBaseComponents(blueprintsByProductType map[int64][]Blueprint, typeID int64, multiplier int64, left int) []typedb.Component {
	if left == 0 {
		return nil
	}

	blueprints, ok := blueprintsByProductType[typeID]
	if !ok || len(blueprints) == 0 {
		return nil
	}

	bp := blueprints[0]
	var components []typedb.Component
	for _, material := range bp.Activities.Manufacturing.Materials {
		r := resolveBaseComponents(blueprintsByProductType, material.TypeID, material.Quantity*multiplier, left-1)
		if r == nil {
			components = append(components, typedb.Component{Quantity: material.Quantity * multiplier, TypeID: material.TypeID})
		} else {
			components = append(components, r...)
		}
	}
	return components
}

type Type struct {
	Name struct {
		En string
	}
	Published bool
	Volume    float64
}

type Blueprint struct {
	BlueprintTypeID int64 `yaml:"blueprintTypeID"`
	Activities      struct {
		Manufacturing struct {
			Materials []struct {
				Quantity int64
				TypeID   int64 `yaml:"typeID"`
			}
			Products []struct {
				Quantity int64
				TypeID   int64 `yaml:"typeID"`
			}
		}
	}
}

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
