package file

import (
	log "github.com/go-chassis/openlog"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

var FileRootPath = "/data/kvs"

var NewstKVFile = "newest_version.json"

var MutexMap = make(map[string]*sync.Mutex)
var mutexMapLock = &sync.Mutex{}
var rollbackMutexLock = &sync.Mutex{}
var createDirMutexLock = &sync.Mutex{}

type SchemaDAO struct{}

type FileDoRecord struct {
	filepath string
	content  []byte
}

func GetOrCreateMutex(path string) *sync.Mutex {
	mutexMapLock.Lock()
	mutex, ok := MutexMap[path]
	if !ok {
		mutex = &sync.Mutex{}
		MutexMap[path] = mutex
	}
	mutexMapLock.Unlock()

	return mutex
}

func ExistDir(path string) error {
	_, err := os.ReadDir(path)
	if err != nil {
		// create the dir if not exist
		if os.IsNotExist(err) {
			createDirMutexLock.Lock()
			defer createDirMutexLock.Unlock()
			err = os.MkdirAll(path, fs.ModePerm)
			if err != nil {
				log.Error("failed to make dir: " + path + " " + err.Error())

				return err
			}
			return nil
		}
		log.Error("failed to read dir: " + path + " " + err.Error())
		return err
	}

	return nil
}

func MoveDir(srcDir string, dstDir string) (err error) {
	srcMutex := GetOrCreateMutex(srcDir)
	dstMutex := GetOrCreateMutex(dstDir)
	srcMutex.Lock()
	dstMutex.Lock()
	defer srcMutex.Unlock()
	defer dstMutex.Unlock()

	var movedFiles []string
	files, err := os.ReadDir(srcDir)
	if err != nil {
		log.Error("move schema files failed " + err.Error())
		return err
	}
	for _, file := range files {
		err = ExistDir(dstDir)
		if err != nil {
			log.Error("move schema files failed " + err.Error())
			return err
		}
		srcFile := filepath.Join(srcDir, file.Name())
		dstFile := filepath.Join(dstDir, file.Name())
		err = os.Rename(srcFile, dstFile)
		if err != nil {
			log.Error("move schema files failed " + err.Error())
			break
		}
		movedFiles = append(movedFiles, file.Name())
	}

	if err != nil {
		log.Error("Occur error when move schema files, begain rollback... " + err.Error())
		for _, fileName := range movedFiles {
			srcFile := filepath.Join(srcDir, fileName)
			dstFile := filepath.Join(dstDir, fileName)
			err = os.Rename(dstFile, srcFile)
			if err != nil {
				log.Error("rollback move schema files failed and continue" + err.Error())
			}
		}
	}
	return err
}

func CreateOrUpdateFile(filepath string, content []byte, rollbackOperations *[]FileDoRecord, isRollback bool) error {
	err := ExistDir(path.Dir(filepath))

	if !isRollback {
		mutex := GetOrCreateMutex(path.Dir(filepath))
		mutex.Lock()
		defer mutex.Unlock()
	}

	if err != nil {
		log.Error("failed to build new schema file dir " + filepath + ", " + err.Error())
		return err
	}

	fileExist := true
	_, err = os.Stat(filepath)
	if err != nil {
		fileExist = false
	}

	if fileExist {
		oldcontent, err := os.ReadFile(filepath)
		if err != nil {
			log.Error("failed to read content to file " + filepath + ", " + err.Error())
			return err
		}
		*rollbackOperations = append(*rollbackOperations, FileDoRecord{filepath: filepath, content: oldcontent})
	} else {
		*rollbackOperations = append(*rollbackOperations, FileDoRecord{filepath: filepath, content: nil})
	}

	err = os.WriteFile(filepath, content, 0600)
	if err != nil {
		log.Error("failed to create file " + filepath + ", " + err.Error())
		return err
	}
	return nil
}

func DeleteFile(filepath string, rollbackOperations *[]FileDoRecord) error {
	_, err := os.Stat(filepath)
	if err != nil {
		log.Error("file does not exist when deleting file " + filepath + ", " + err.Error())
		return nil
	}

	oldcontent, err := os.ReadFile(filepath)
	if err != nil {
		log.Error("failed to read content to file " + filepath + ", " + err.Error())
		return err
	}

	*rollbackOperations = append(*rollbackOperations, FileDoRecord{filepath: filepath, content: oldcontent})

	err = os.Remove(filepath)
	if err != nil {
		log.Error("failed to delete file " + filepath + ", " + err.Error())
		return err
	}
	return nil
}

func CleanDir(dir string) error {
	mutex := GetOrCreateMutex(dir)
	mutex.Lock()
	defer delete(MutexMap, dir)
	defer mutex.Unlock()

	rollbackOperations := []FileDoRecord{}
	_, err := os.Stat(dir)
	if err != nil {
		return nil
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		filepath := filepath.Join(dir, file.Name())
		err = DeleteFile(filepath, &rollbackOperations)
		if err != nil {
			break
		}
	}

	if err != nil {
		log.Error("Occur error when create schema files, begain rollback... " + err.Error())
		Rollback(rollbackOperations)
		return err
	}

	err = os.Remove(dir)
	if err != nil {
		log.Error("OOccur error when remove service schema dir, begain rollback... " + err.Error())
		Rollback(rollbackOperations)
		return err
	}

	return nil
}

func ReadFile(filepath string) ([]byte, error) {
	// check the file is empty
	mutex := GetOrCreateMutex(path.Dir(filepath))
	mutex.Lock()
	defer mutex.Unlock()

	content, err := os.ReadFile(filepath)
	if err != nil {
		log.Error("failed to read content to file " + filepath + ", " + err.Error())
		return nil, err
	}
	return content, nil
}

func CountInDomain(dir string) (int, error) {
	mutex := GetOrCreateMutex(dir)
	mutex.Lock()
	defer mutex.Unlock()

	files, err := os.ReadDir(dir)
	if err != nil {
		log.Error("failed to read directory " + dir + ", " + err.Error())
		return 0, err
	}

	count := 0
	for _, projectFolder := range files {
		if projectFolder.IsDir() {
			kvs, err := os.ReadDir(path.Join(dir, projectFolder.Name()))
			if err != nil {
				log.Error("failed to read directory " + dir + ", " + err.Error())
				return 0, err
			}
			for _, kv := range kvs {
				if kv.IsDir() {
					count++
				}
			}
		}
	}
	// count kv numbers
	return count, nil
}

func ReadAllKvsFromProjectFolder(dir string) ([][]byte, error) {
	var kvs [][]byte

	kvDir, err := os.ReadDir(dir)
	if err != nil {
		log.Error("failed to read directory " + dir + ", " + err.Error())
		return nil, err
	}

	for _, file := range kvDir {
		if file.IsDir() {
			filepath := path.Join(dir, file.Name(), NewstKVFile)
			content, err := ReadFile(filepath)
			if err != nil {
				log.Error("failed to read content to file " + filepath + ", " + err.Error())
				return nil, err
			}
			kvs = append(kvs, content)
		}
	}
	return kvs, nil
}

func ReadAllFiles(dir string) ([]string, [][]byte, error) {
	mutex := GetOrCreateMutex(dir)
	mutex.Lock()
	defer mutex.Unlock()

	files := []string{}
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.Contains(path, NewstKVFile) {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	var contentArray [][]byte

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			log.Error("failed to read content from schema file " + file + ", " + err.Error())
			return nil, nil, err
		}
		contentArray = append(contentArray, content)
	}
	return files, contentArray, nil
}

func Rollback(rollbackOperations []FileDoRecord) {
	rollbackMutexLock.Lock()
	defer rollbackMutexLock.Unlock()

	var err error
	for _, fileOperation := range rollbackOperations {
		if fileOperation.content == nil {
			err = DeleteFile(fileOperation.filepath, &[]FileDoRecord{})
		} else {
			err = CreateOrUpdateFile(fileOperation.filepath, fileOperation.content, &[]FileDoRecord{}, true)
		}
		if err != nil {
			log.Error("Occur error when rolling back schema files:  " + err.Error())
		}
	}
}
