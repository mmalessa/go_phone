package filemanager

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

type FileManager struct {
	RecordingsDirectory string
	RecordingsExtention string
}

func (fm *FileManager) FindRecordingFileName() (string, error) {
	maxDirInt, err := fm.getMaxDirectoryInt()
	if err != nil {
		return "", err
	}

	exists, err := fm.doesDirectoryExistByInt(maxDirInt)
	if err != nil {
		return "", err
	}

	maxFileInt := 0
	if exists {
		maxFileInt, err = fm.getMaxFileInt(maxDirInt)
		if err != nil {
			return "", err
		}
		if maxFileInt == 999 {
			maxFileInt = 0
			maxDirInt++
			fm.createDirectoryById(maxDirInt)
		}
	} else {
		fm.createDirectoryById(maxDirInt)
	}
	maxFileInt++ //new file

	return fm.getAbsoluteFileNameByInts(maxDirInt, maxFileInt), nil
}

func (fm *FileManager) getMaxDirectoryInt() (int, error) {
	directories, err := ioutil.ReadDir(fm.RecordingsDirectory)
	if err != nil {
		return 0, err
	}
	maxDirInt := 0
	re := regexp.MustCompile("^[0-9]{3}$")
	for _, file := range directories {
		if !re.MatchString(file.Name()) {
			continue
		}
		dirInt, err := strconv.Atoi(file.Name())
		if err != nil {
			return 0, err
		}
		if dirInt > maxDirInt {
			maxDirInt = dirInt
		}
	}
	return maxDirInt, nil
}

func (fm *FileManager) doesDirectoryExistByInt(dirInt int) (bool, error) {
	dirAbs := fm.getAbsoluteRecordingsDirectoryByInt(dirInt)
	if _, err := os.Stat(dirAbs); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, err
	}
}

func (fm *FileManager) createDirectoryById(dirInt int) error {
	absDirName := fm.getAbsoluteRecordingsDirectoryByInt(dirInt)
	return os.Mkdir(absDirName, os.ModePerm)
}

func (fm *FileManager) getAbsoluteRecordingsDirectoryByInt(dirInt int) string {
	return filepath.Join(fm.RecordingsDirectory, fmt.Sprintf("%03d", dirInt))
}

func (fm *FileManager) getMaxFileInt(dirInt int) (int, error) {
	absDirName := fm.getAbsoluteRecordingsDirectoryByInt(dirInt)
	files, err := ioutil.ReadDir(absDirName)
	if err != nil {
		return 0, err
	}
	re := regexp.MustCompile(`rec-[0-9]{3}([0-9]{3}).` + fm.RecordingsExtention)
	maxFileInt := 0
	for _, file := range files {
		matches := re.FindStringSubmatch(file.Name())
		if len(matches) < 2 {
			continue
		}
		fileInt, err := strconv.Atoi(matches[1])
		if err != nil {
			return 0, err
		}
		if fileInt > maxFileInt {
			maxFileInt = fileInt
		}
	}
	return maxFileInt, nil
}

func (fm *FileManager) getAbsoluteFileNameByInts(dirInt int, fileInt int) string {
	return filepath.Join(
		fm.getAbsoluteRecordingsDirectoryByInt(dirInt),
		fmt.Sprintf("rec-%03d%03d.%s", dirInt, fileInt, fm.RecordingsExtention),
	)
}
