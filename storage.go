package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/sirupsen/logrus"
)

func checkStorageDirectory() error {
	pMounted, err := isPendriveMounted()
	if err != nil {
		return err
	}
	if !pMounted {
		return fmt.Errorf("pendrive is not mounted")
	}
	logrus.Debug("Pendrive is mounted")

	aFullPath := filepath.Join(storageDir, greetingsSubDir, greetingsFileName)
	aExists, err := greetingsExist(aFullPath)
	if err != nil {
		return err
	}
	if !aExists {
		logrus.Debugf("greeetings file doesn't exist. Try to create default: %s", aFullPath)
		if err := createDefaultGreetings(storageDir, greetingsSubDir, greetingsFileName); err != nil {
			return err
		}
	}
	logrus.Debugf("Use greetings file: %s", aFullPath)

	rFullPath := filepath.Join(storageDir, recordingsSubDir)
	rExists, err := recordingsExists(rFullPath)
	if err != nil {
		return err
	}
	if !rExists {
		logrus.Debugf("Recordings directory doesn't exist. We will create it.")
		createRecordingsDir(rFullPath)
	}
	logrus.Debugf("Use recordings dir: %s", rFullPath)
	return nil
}

func isPendriveMounted() (bool, error) {
	cmd := "df"
	dfPattern := regexp.MustCompile(fmt.Sprintf(" %s$", storageDir))
	out, err := exec.Command(cmd).Output()
	if err != nil {
		return false, err
	}
	s := bufio.NewScanner(bytes.NewReader(out))
	for s.Scan() {
		line := s.Text()
		if dfPattern.MatchString(line) {
			return true, nil
		}
	}
	return false, nil
}

func greetingsExist(absolutePath string) (bool, error) {
	if _, err := os.Stat(absolutePath); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, err
	}
}

func createDefaultGreetings(storageDir string, greetingsSubDir string, greetingsFileName string) error {
	aDirPath := filepath.Join(storageDir, greetingsSubDir)
	if _, err := os.Stat(aDirPath); errors.Is(err, os.ErrNotExist) {
		if err := os.Mkdir(aDirPath, 0755); err != nil {
			return err
		}
		logrus.Debugf("Greetings directory (%s) created", aDirPath)
	} else if err != nil {
		return err
	}

	aFullPath := filepath.Join(storageDir, greetingsSubDir, greetingsFileName)
	if _, err := os.Stat(aFullPath); errors.Is(err, os.ErrNotExist) {
		src := greetingsDefaultFile
		dst := aFullPath

		fin, err := os.Open(src)
		if err != nil {
			return err
		}
		defer fin.Close()

		fout, err := os.Create(dst)
		if err != nil {
			return err
		}
		defer fout.Close()

		if _, err = io.Copy(fout, fin); err != nil {
			return err
		}
		logrus.Debugf("Default greetings has been copied (%s -> %s)", src, dst)
	} else if err != nil {
		return err
	}

	return nil
}

func recordingsExists(absolutePath string) (bool, error) {
	if _, err := os.Stat(absolutePath); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, err
	}
}

func createRecordingsDir(absolutePath string) error {
	err := os.Mkdir(absolutePath, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}
