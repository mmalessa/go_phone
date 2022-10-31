package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
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

	aFullPath := filepath.Join(storageDir, "announcement", announcementFileName)
	aExists, err := announcementExists(aFullPath)
	if err != nil {
		return err
	}
	if !aExists {
		return fmt.Errorf("announcement file not found (%s)", aFullPath)
	}
	logrus.Debugf("Use announcement file: %s", aFullPath)

	rFullPath := filepath.Join(storageDir, "recordings")
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

func announcementExists(absolutePath string) (bool, error) {
	if _, err := os.Stat(absolutePath); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, err
	}
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
