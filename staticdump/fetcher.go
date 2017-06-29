package staticdump

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/evepraisal/go-evepraisal/bolt"
	"github.com/evepraisal/go-evepraisal/typedb"
	"github.com/sethgrid/pester"
)

type StaticFetcher struct {
	dbPath   string
	callback func(typeDB typedb.TypeDB)
	client   *pester.Client

	stop chan bool
	wg   *sync.WaitGroup
}

var volumeGroupOverrides = map[int64]float64{
	// Ships
	25:   2500,     // Frigate
	26:   10000,    // Cruiser
	27:   50000,    // Battleship
	28:   10000,    // Industrial
	30:   13000000, // Titan
	31:   500,      // Shuttle
	237:  2500,     // Rookie ship
	324:  2500,     // Assault Frigate
	358:  10000,    // Heavy Assault Cruiser
	380:  10000,    // Deep Space Transport
	419:  15000,    // Combat Battlecruiser
	420:  5000,     // Destroyer
	463:  3750,     // Mining Barge
	485:  1300000,  // Dreadnought
	513:  1300000,  // Freighter
	540:  15000,    // Command Ship
	541:  5000,     // Interdictor
	543:  3750,     // Exhumer
	547:  1300000,  // Carrier
	659:  13000000, // Supercarrier
	830:  2500,     // Covert Ops
	831:  2500,     // Interceptor
	832:  10000,    // Logistics
	833:  10000,    // Force Recon Ship
	834:  2500,     // Stealth Bomber
	883:  1300000,  // Capital Industrial Ship
	893:  2500,     // Electronic Attack Ship
	894:  10000,    // Heavy Interdiction Cruiser
	898:  50000,    // Black Ops
	900:  50000,    // Marauder
	902:  1300000,  // Jump Freighter
	906:  10000,    // Combat Recon Ship
	941:  500000,   // Industrial Command Ship
	963:  10000,    // Strategic Cruiser
	1022: 500,      // Prototype Exploration Ship
	1201: 15000,    // Attack Battlecruiser
	1202: 10000,    // Blockade Runner
	1283: 2500,     // Expedition Frigate
	1305: 5000,     // Tactical Destroyer
	1527: 2500,     // Logistics Frigate
	1534: 5000,     // Command Destroyer
	1538: 1300000,  // Force Auxiliary

	// Modules
	600:  1000,
	771:  1000,
	772:  1000,
	773:  1000,
	774:  1000,
	775:  1000,
	776:  1000,
	777:  1000,
	778:  1000,
	910:  1000,
	1052: 1000,
	1063: 1000,
	2240: 1000,
	2241: 1000,
	2242: 1000,
	2243: 1000,
	2244: 1000,
	2245: 1000,
	2246: 1000,
	2247: 1000,
	2250: 1000,
	2251: 1000,
	2249: 2000,
	2267: 2000,
	2268: 2000,
	2269: 2000,
	2270: 2000,
	2276: 2000,
}

var volumeItemOverrides = map[int64]float64{
	41249: 1000,
	41250: 1000,
	41251: 1000,
	41252: 1000,
	41253: 1000,
	41254: 1000,
	41255: 1000,
	41236: 1000,
	41238: 1000,
	41239: 1000,
	41240: 1000,
	41241: 1000,
	41411: 1000,
	24283: 1000,
	41414: 1000,
	41415: 1000,
	40715: 2000,
	40716: 2000,
	40717: 2000,
	40718: 2000,
}

func NewStaticFetcher(client *pester.Client, dbPath string, callback func(typeDB typedb.TypeDB)) (*StaticFetcher, error) {
	fetcher := &StaticFetcher{
		dbPath:   dbPath,
		callback: callback,
		client:   client,

		stop: make(chan bool),
		wg:   &sync.WaitGroup{},
	}

	err := fetcher.RunOnce()
	if err != nil {
		return nil, err
	}

	fetcher.wg.Add(1)
	go func() {
		defer fetcher.wg.Done()
		for {

			select {
			case <-time.After(6 * time.Hour):
			case <-fetcher.stop:
				return
			}

			err := fetcher.RunOnce()
			if err != nil {
				log.Printf("WARNING: Fetcher failed to run: %s", err)
			}

			log.Println("FETCHER RAN")
		}
	}()

	return fetcher, nil
}

func (f *StaticFetcher) RunOnce() error {
	staticDumpURL, err := FindLastStaticDumpURL(f.client)
	if err != nil {
		return err
	}
	//
	staticDumpURLBase := filepath.Base(staticDumpURL)
	typedbPath := filepath.Join(f.dbPath, "types-"+strings.TrimSuffix(staticDumpURLBase, filepath.Ext(staticDumpURLBase)))
	if _, err := os.Stat(typedbPath); os.IsNotExist(err) {
		err := f.loadTypes(typedbPath, staticDumpURL)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	log.Println("Done loading types", staticDumpURLBase)

	typeDB, err := bolt.NewTypeDB(typedbPath, false)
	if err != nil {
		return err
	}
	log.Println("done making new typedb")

	f.callback(typeDB)
	return nil
}

func (f *StaticFetcher) Close() error {
	close(f.stop)
	f.wg.Wait()
	return nil
}

func (f *StaticFetcher) loadTypes(staticCacheFile string, staticDumpURL string) error {

	// TODO: I need to find a reliable source for this information..... CCP????
	// typeVolumes, err := downloadTypeVolumes(f.client)
	// if err != nil {
	// 	return err
	// }

	// avoid re-downloading the entire static dump if we already have it
	cachepath := staticCacheFile + ".zip"
	if _, err := os.Stat(cachepath); os.IsNotExist(err) {
		log.Printf("Downloading static dump to %s", cachepath)
		err := downloadTypes(f.client, staticDumpURL, cachepath)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	types, err := loadtypes(cachepath)
	if err != nil {
		return err
	}

	typeDB, err := bolt.NewTypeDB(staticCacheFile, true)
	if err != nil {
		return err
	}
	finished := false
	defer func() {
		if finished == true {
			typeDB.Close()
		} else {
			log.Println("Deleting new typedb because it was stopped before finishing")
			err := typeDB.Delete()
			if err != nil {
				log.Printf("Error deleting old typedb: %s", err)
			}
		}
	}()

	for i, t := range types {
		if i%1000 == 0 {
			select {
			case <-f.stop:
				return nil
			default:
			}
		}

		volume, ok := volumeGroupOverrides[t.GroupID]
		if ok {
			t.PackagedVolume = volume
		} else {
			volume, ok := volumeItemOverrides[t.ID]
			if ok {
				t.PackagedVolume = volume
			}
		}

		err = typeDB.PutType(t)
		if err != nil {
			return err
		}
	}
	finished = true
	log.Println("Finished typedb fetch")
	return nil
}
