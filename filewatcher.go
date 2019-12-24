package filewatcher

import (
	"io/ioutil"
	golog "log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"time"

	"github.com/broothie/filewatcher/pkg/safemap"
	"github.com/gobwas/glob"
)

var (
	emptyStruct struct{}
	cmdSplitter = regexp.MustCompile(`\s`)
	log         = golog.New(os.Stdout, "", 0)
)

type FileWatcher struct {
	Root           string
	Glob           glob.Glob
	RunFrequency   time.Duration
	CheckFrequency time.Duration
	RunFunc        func()

	files       safemap.SafeMap
	directories safemap.SafeMap
	cmdChan     chan struct{}
}

func New(cmd, fileGlob, root string) (FileWatcher, error) {
	glob, err := glob.Compile(fileGlob)
	if err != nil {
		return FileWatcher{}, err
	}

	return FileWatcher{
		Root:           root,
		Glob:           glob,
		RunFrequency:   10 * time.Millisecond,
		CheckFrequency: 10 * time.Millisecond,
		RunFunc:        runCmd(cmd),
		files:          safemap.New(),
		directories:    safemap.New(),
		cmdChan:        make(chan struct{}),
	}, nil
}

func (f FileWatcher) Start() {
	go f.watchDir(f.Root)

	for {
		<-f.cmdChan
		go func() {
			for len(f.cmdChan) > 0 {
				<-f.cmdChan
			}
		}()

		go f.RunFunc()
		time.Sleep(f.RunFrequency)
	}
}

func (f FileWatcher) watchDir(dirname string) {
	f.directories.Set(dirname, emptyStruct)

	for {
		elements, err := ioutil.ReadDir(dirname)
		if err != nil {
			if os.IsNotExist(err) {
				f.directories.Remove(dirname)
				return
			}

			log.Printf("list dir error: %v\n", err)
		}

		for _, element := range elements {
			name := element.Name()
			rootPath := path.Join(dirname, name)

			if element.IsDir() {
				if name != "." && name != ".." && !f.directories.HasKey(rootPath) {
					go f.watchDir(rootPath)
				}
			} else {
				if f.Glob.Match(name) && !f.files.HasKey(rootPath) {
					go f.watchFile(rootPath)
				}
			}
		}

		time.Sleep(f.CheckFrequency)
	}
}

func (f FileWatcher) watchFile(filename string) {
	f.files.Set(filename, time.Now())

	for {
		info, err := os.Stat(filename)
		if err != nil {
			if os.IsNotExist(err) {
				f.files.Remove(filename)
				return
			}

			log.Printf("file stat error: %v\n", err)
		}

		if lastChecked, ok := f.files.Get(filename); ok && info.ModTime().After(lastChecked.(time.Time)) {
			f.files.Set(filename, info.ModTime())
			go func() { f.cmdChan <- emptyStruct }()
		}

		time.Sleep(f.CheckFrequency)
	}
}

func runCmd(cmd string) func() {
	cmdTokens := cmdSplitter.Split(cmd, -1)

	return func() {
		if out, err := exec.Command(cmdTokens[0], cmdTokens[1:]...).Output(); err != nil {
			log.Printf("exec error: %v\n", err)
		} else {
			log.Print(string(out))
		}
	}
}
