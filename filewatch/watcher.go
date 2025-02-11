package filewatch

import (
	"log"
	"path/filepath"

	"github.com/aguxez/ffa/models"
	"github.com/fsnotify/fsnotify"
)

// FileWatcher monitors directory changes
type FileWatcher struct {
	stateMgr *models.StateManager
	watcher  *fsnotify.Watcher
}

func NewFileWatcher(paths []string, sm *models.StateManager) (*FileWatcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	for _, path := range paths {
		err = w.Add(path)
		if err != nil {
			return nil, err
		}
	}

	return &FileWatcher{stateMgr: sm, watcher: w}, nil
}

func (fw *FileWatcher) Watch() {
	for {
		select {
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Printf("Modified file: %s", event.Name)
				if filepath.Ext(event.Name) == ".csv" {
					fw.HandleFileChange(event.Name)
				}
			}
		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Error: %v", err)
		}
	}
}

func (fw *FileWatcher) HandleFileChange(path string) {
	if filepath.Base(filepath.Dir(path)) == "foods" {
		foods, err := ParseFoods(path)
		if err != nil {
			log.Printf("Error parsing foods: %v", err)
			return
		}
		fw.stateMgr.UpdateFoods(foods)
	} else if filepath.Base(filepath.Dir(path)) == "targets" {
		targets, err := ParseMacroData(path)
		if err != nil {
			log.Printf("Error parsing targets: %v", err)
			return
		}
		fw.stateMgr.UpdateTargets(targets)
	}
}
