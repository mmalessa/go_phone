package phoneaudio

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

func (pa *PhoneAudio) findRecordingFileName() (string, error) {
	maxDirInt, err := pa.getMaxDirectoryInt()
	if err != nil {
		return "", err
	}

	exists, err := pa.doesDirectoryExistByInt(maxDirInt)
	if err != nil {
		return "", err
	}

	maxFileInt := 0
	if exists {
		maxFileInt, err = pa.getMaxFileInt(maxDirInt)
		if err != nil {
			return "", err
		}
		if maxFileInt == 999 {
			maxFileInt = 0
			maxDirInt++
			pa.createDirectoryById(maxDirInt)
		}
	} else {
		pa.createDirectoryById(maxDirInt)
	}
	maxFileInt++ //new file

	return pa.getAbsoluteFileNameByInts(maxDirInt, maxFileInt), nil
}

func (pa *PhoneAudio) getMaxDirectoryInt() (int, error) {
	directories, err := ioutil.ReadDir(pa.RecordingsDirectory)
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

func (pa *PhoneAudio) doesDirectoryExistByInt(dirInt int) (bool, error) {
	dirAbs := pa.getAbsoluteRecordingsDirectoryByInt(dirInt)
	if _, err := os.Stat(dirAbs); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, err
	}
}

func (pa *PhoneAudio) createDirectoryById(dirInt int) error {
	absDirName := pa.getAbsoluteRecordingsDirectoryByInt(dirInt)
	return os.Mkdir(absDirName, os.ModePerm)
}

func (pa *PhoneAudio) getAbsoluteRecordingsDirectoryByInt(dirInt int) string {
	return filepath.Join(pa.RecordingsDirectory, fmt.Sprintf("%03d", dirInt))
}

func (pa *PhoneAudio) getMaxFileInt(dirInt int) (int, error) {
	absDirName := pa.getAbsoluteRecordingsDirectoryByInt(dirInt)
	files, err := ioutil.ReadDir(absDirName)
	if err != nil {
		return 0, err
	}
	re := regexp.MustCompile(`rec-[0-9]{3}([0-9]{3}).` + recordingsExtension)
	maxFileInt := 0
	for _, file := range files {
		matches := re.FindStringSubmatch(file.Name())
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

func (pa *PhoneAudio) getAbsoluteFileNameByInts(dirInt int, fileInt int) string {
	return filepath.Join(
		pa.getAbsoluteRecordingsDirectoryByInt(dirInt),
		fmt.Sprintf("rec-%03d%03d.%s", dirInt, fileInt, recordingsExtension),
	)
}
