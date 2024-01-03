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

var mutexMap = make(map[string]*sync.Mutex)
var mutexLock = &sync.Mutex{}

var FileRootPath = "./filepath"

type SchemaDAO struct{}

func ExistDir(path string) error {
	_, err := os.ReadDir(path)
	if err != nil {
		// create the dir if not exist
		err = os.MkdirAll(path, fs.ModePerm)
		if err != nil {
			log.Error("failed to makr dir: " + path + " " + err.Error())
		}
	}
	return err
}

func MoveDir(srcDir string, dstDir string) (err error) {
	var movedFiles []string
	files, err := os.ReadDir(srcDir)
	if err != nil {
		log.Error("move schema files failed " + err.Error())
	}
	for _, file := range files {
		ExistDir(dstDir)
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
			os.Rename(dstFile, srcFile)
		}
	}
	return err
}

func CreateOrUpdateFile(filepath string, content []byte, rollbackOperations []FileDoRecord) error {
	err := ExistDir(path.Dir(filepath))
	if err != nil {
		log.Error("failed to build new schema file dir " + filepath + ", " + err.Error())
		return err
	}

	var fileExist = true
	_, err = os.Stat(filepath)
	if err != nil {
		fileExist = false
	}

	if fileExist {
		oldcontent, err := ReadFile(filepath)
		if err != nil {
			log.Error("failed to read content to file " + filepath + ", " + err.Error())
			return err
		}
		rollbackOperations = append(rollbackOperations, FileDoRecord{filepath: filepath, content: oldcontent})
	} else {
		rollbackOperations = append(rollbackOperations, FileDoRecord{filepath: filepath, content: nil})
	}

	err = os.WriteFile(filepath, content, 0666)
	if err != nil {
		log.Error("failed to create file " + filepath + ", " + err.Error())
		return err
	}

	return nil
}

func DeleteFile(filepath string, rollbackOperations []FileDoRecord) error {
	_, err := os.Stat(filepath)
	if err != nil {
		log.Error("file does not exist when deleting file " + filepath + ", " + err.Error())
		return nil
	}

	oldcontent, err := ReadFile(filepath)
	if err != nil {
		log.Error("failed to read content to file " + filepath + ", " + err.Error())
		return err
	}

	rollbackOperations = append(rollbackOperations, FileDoRecord{filepath: filepath, content: oldcontent})

	err = os.Remove(filepath)
	if err != nil {
		log.Error("failed to delete file " + filepath + ", " + err.Error())
		return err
	}

	return nil
}

func CleanDir(dir string) error {
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
		err = DeleteFile(filepath, rollbackOperations)
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
	content, err := os.ReadFile(filepath)
	if err != nil {
		log.Error("failed to read content to file " + filepath + ", " + err.Error())
		return nil, err
	}
	return content, nil
}

func Count(dir string) (int, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Error("failed to read directory " + dir + ", " + err.Error())
		return 0, err
	}

	count := 0
	for _, file := range files {
		if file.IsDir() {
			count++
		}
	}

	return count, nil
}

func ReadAllKvsFromProjectFolder(dir string) ([][]byte, error) {
	kvDir, err := os.ReadDir(dir)
	var kvs [][]byte
	if err != nil {
		log.Error("failed to read directory " + dir + ", " + err.Error())
		return nil, err
	}

	for _, file := range kvDir {
		if file.IsDir() {
			filepath := path.Join(dir, file.Name(), "newest_version.json")
			content, err := os.ReadFile(filepath)
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
	files := []string{}
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.Contains(path, "newest_version.json") {
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
	var err error
	for _, fileOperation := range rollbackOperations {
		if fileOperation.content == nil {
			err = DeleteFile(fileOperation.filepath, []FileDoRecord{})
		} else {
			err = CreateOrUpdateFile(fileOperation.filepath, fileOperation.content, []FileDoRecord{})
		}
		if err != nil {
			log.Error("Occur error when rolling back schema files:  " + err.Error())
		}
	}
}

type FileDoRecord struct {
	filepath string
	content  []byte
}
