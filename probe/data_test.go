package probe

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/megaease/easeprobe/global"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

var testResults = []Result{
	{
		Name:             "Probe Result - HTTP",
		Endpoint:         "https://www.example.com",
		StartTime:        time.Now().UTC(),
		StartTimestamp:   time.Now().UnixMilli(),
		RoundTripTime:    478783 * time.Microsecond,
		Status:           StatusUp,
		PreStatus:        StatusDown,
		Message:          "Success (http): HTTP Status Code is 200",
		LatestDownTime:   time.Time{},
		RecoveryDuration: 200 * time.Second,
		Stat: Stat{
			Since: time.Now().UTC(),
			Total: 100,
			Status: map[Status]int64{
				StatusUp:   70,
				StatusDown: 30,
			},
			UpTime:   70 * time.Minute,
			DownTime: 30 * time.Minute,
		},
	},
	{
		Name:             "Probe Result - Host",
		Endpoint:         "CPU: 0.15, Mem: 0.10, Disk: 0.90",
		StartTime:        time.Now().UTC(),
		StartTimestamp:   time.Now().UnixMilli(),
		RoundTripTime:    283455 * time.Millisecond,
		Status:           StatusDown,
		PreStatus:        StatusUp,
		Message:          "Error (host/server): CPU Busy! | Memory Shortage! ( CPU: 37.10% - Memory: 49.45% - Disk: 64.00% )",
		LatestDownTime:   time.Time{},
		RecoveryDuration: 5 * time.Minute,
		Stat: Stat{
			Since: time.Now().UTC(),
			Total: 300,
			Status: map[Status]int64{
				StatusInit: 1,
				StatusUp:   270,
				StatusDown: 30,
			},
			UpTime:   270 * time.Minute,
			DownTime: 30 * time.Minute,
		},
	},
}

func newDataFile(file string) error {
	makeAll(file)
	SetResultsData(testResults)
	return SaveDataToFile(file)
}

func newDataFileWithOutMeta(file string) error {
	makeAll(file)
	SetResultsData(testResults)
	buf, err := yaml.Marshal(resultData)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(file, []byte(buf), 0644); err != nil {
		return err
	}
	return nil
}

func isDataFileExisted(file string) bool {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return false
	}
	return true
}

func removeDataFile(file string) {
	os.Remove(file)
}

func makeAll(file string) {
	dir, _ := filepath.Split(file)
	os.MkdirAll(dir, 0755)
}

func removeAll(file string) {
	os.RemoveAll(file)
}

func checkData(t *testing.T) {
	for _, r := range testResults {
		assert.Equal(t, r, *resultData[r.Name])
	}
}

func TestNewDataFile(t *testing.T) {
	//disable data file
	newDataFile("-")
	assert.False(t, isDataFileExisted("-"))

	//default data file
	file := global.DefaultDataFile
	newDataFile(file)
	assert.True(t, isDataFileExisted(file))
	removeAll("data/")

	//default data file
	file = "data.yaml"
	newDataFile(file)
	assert.True(t, isDataFileExisted(file))
	removeAll(file)

	//custom data file
	file = "x/y/z/mydata.yaml"
	makeAll(file)
	newDataFile(file)
	assert.True(t, isDataFileExisted(file))
	removeAll("x/")
}

func TestLoadDataFile(t *testing.T) {
	// no data file
	file := "data.yaml"
	err := LoadDataFromFile(file)
	assert.True(t, err != nil)

	// create data file
	newDataFile(file)
	if err := LoadDataFromFile(file); err != nil {
		t.Fatal(err)
	}

	assert.True(t, isDataFileExisted(metaData.backup))
	removeAll(metaData.backup)
	checkData(t)
}

func numOfBackup(file string) int {
	files, _ := filepath.Glob(file + "-*")
	return len(files)
}

func TestCleanDataFile(t *testing.T) {

	file := "data/data.yaml"
	assert.Equal(t, 0, numOfBackup(file))

	// create data file with backups
	n := 5
	for i := 0; i < 5; i++ {
		newDataFile(file)
		if err := LoadDataFromFile(file); err != nil {
			t.Fatal(err)
		}
	}
	assert.Equal(t, n, numOfBackup(file))

	n = 3
	CleanDataFile(file, n)
	assert.Equal(t, n, numOfBackup(file))

	n = 0
	CleanDataFile(file, n)
	assert.Equal(t, n, numOfBackup(file))

	// clean data file
	removeAll("data/")
}

func TestMetaData(t *testing.T) {
	// no meta
	file := "data/data.yaml"
	newDataFileWithOutMeta(file)
	if err := LoadDataFromFile(file); err != nil {
		t.Error(err)
	}
	assert.Equal(t, metaData.Name, global.DefaultProg)
	assert.Equal(t, metaData.Ver, global.Ver)

	// with meta
	SetMetaData("myprog", "1.0.0")
	SaveDataToFile(file)
	if err := LoadDataFromFile(file); err != nil {
		t.Error(err)
	}
	assert.Equal(t, metaData.Name, "myprog")
	assert.Equal(t, metaData.Ver, global.Ver)

	removeAll("data/")
}
