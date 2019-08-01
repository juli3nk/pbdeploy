package main

import (
	"flag"
	"os"
	"strconv"

	"github.com/juli3nk/go-git"
	"github.com/juli3nk/go-utils"
	"github.com/juli3nk/go-utils/filedir"
	"github.com/juli3nk/pbdeploy/pkg/config"
	pfile "github.com/juli3nk/pbdeploy/pkg/file"
	"github.com/juli3nk/pbdeploy/pkg/repository"
	log "github.com/sirupsen/logrus"
	gitv4 "gopkg.in/src-d/go-git.v4"
)

var (
	configFile = flag.String("conf", ".pbdeploy.yml", "Path to config file")
	debug = flag.Bool("debug", false, "Enable debug logging")
	dstPath = flag.String("path", "/tmp/workspace", "Path where to clone git repositories")
)

func main() {
	flag.Parse()

	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	c, err := config.New(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	version, err := utils.GetEnv("PACKAGE_VERSION")
	if err != nil {
		log.Fatal(err)
	}

	authorName, err := utils.GetEnv("GIT_AUTHOR_NAME")
	if err != nil {
		log.Fatal(err)
	}

	authorEmail, err := utils.GetEnv("GIT_AUTHOR_EMAIL")
	if err != nil {
		log.Fatal(err)
	}

	token, err := utils.GetEnv("GIT_PROVIDER_TOKEN")
	if err != nil {
		log.Fatal(err)
	}

	if filedir.DirExists(*dstPath) {
		log.Fatalf("Destination directory %s exists", *dstPath)
	}

	if err := filedir.CreateDirIfNotExist(*dstPath, true, 0775); err != nil {
		log.Fatal(err)
	}
	log.Debugf("Created directory %s", *dstPath)

	for _, pack := range c.Packages {
		var commitMsg string

		if err := os.Chdir(*dstPath); err != nil {
			log.Fatal(err)
		}
		log.Debugf("Changed directory to %s", *dstPath)

		config := make(map[string]string)
		config["username"] = c.Repository.Username
		config["token"] = token

		r, err := repository.NewDriver(c.Repository.Provider, config)
		if err != nil {
			log.Fatal(err)
		}

		repoExists, err := r.Exists(pack.Org, pack.Name)
		if err != nil {
			log.Println(err)
		}
		log.Debugf("Checked if repository %s exist => %t", pack.Name, repoExists)

		g, err := git.New(r.GetURL(pack.Org, pack.Name))
		if err != nil {
			log.Fatal(err)
		}

		if err := g.SetConfigUser(authorName, authorEmail); err != nil {
			log.Fatal(err)
		}

		if err := g.SetAuth(c.Repository.Username, "token", token); err != nil {
			log.Fatal(err)
		}

		if !repoExists {
			commitMsg = "Initial commit"

			// Create remote repository
			if err := r.Create(pack.Org, pack.Name, true); err != nil {
				log.Fatal(err)
			}
			log.Debugf("Created repository %s/%s", pack.Org, pack.Name)

			// Create local directory
			if err := filedir.CreateDirIfNotExist(pack.Name, false, 0775); err != nil {
				log.Fatal(err)
			}
			log.Debug("Created local directory")

			// Change dir
			if err := os.Chdir(pack.Name); err != nil {
				log.Fatal(err)
			}
			log.Debug("Changed directory")

			// git init
			if err := g.Init(); err != nil {
				log.Fatal(err)
			}
			log.Debug("Created an empty Git repository")

			// git remote add
			if err = g.RemoteAdd("origin"); err != nil {
				log.Fatal(err)
			}
			log.Debugf("Added a remote named \"origin\" for the repository at %s", g.URL)
		} else {
			commitMsg = "Update"

			// git clone
			if err = g.Clone(pack.Name); err != nil {
				log.Fatal(err)
			}
			log.Debug("Cloned repository")

			// Change dir
			if err := os.Chdir(pack.Name); err != nil {
				log.Fatal(err)
			}
			log.Debug("Changed directory")
		}

		// Get files
		conf := make(map[string][]string)
		conf["org"] = []string{pack.Org}
		conf["name"] = []string{pack.Name}
		conf["version"] = []string{version}
		conf["description"] = []string{c.Description}
		conf["private"] = []string{strconv.FormatBool(pack.Private)}
		conf["src_path"] = []string{pack.SrcPath}
		conf["dst_path"] = []string{*dstPath}

		for k, v := range pack.Options {
			conf[k] = v
		}

		f, err := pfile.NewDriver(pack.Type, conf)
		if err != nil {
			log.Fatal(err)
		}

		// Delete all files from repo
		if err = f.DeleteFiles(); err != nil {
			log.Fatal(err)
		}

		// Add common files
		if err = f.CreateOrUpdateFiles(); err != nil {
			log.Fatal(err)
		}

		// Copy generated files to local repo
		if err = f.CopyFiles(); err != nil {
			log.Fatal(err)
		}
		log.Debug("Copied generated file(s) to local repository")

		// git status
		statusFiles, err := g.Status()
		if err != nil {
			log.Fatal(err)
		}

		// git add / rm
		stagedCount := 0

		for file, status := range statusFiles {
			if status.Worktree == gitv4.Untracked || status.Worktree == gitv4.Modified {
				if err := g.Add(file); err != nil {
					log.Fatal(err)
				}
				stagedCount += 1

				log.Debugf("Added file (%s) contents to the index", file)
			}
			if status.Worktree == gitv4.Deleted {
				if err := g.Remove(file); err != nil {
					log.Fatal(err)
				}
				stagedCount += 1

				log.Debugf("Removed file (%s) contents from the index", file)
			}
		}

		if stagedCount == 0 {
			log.Debug("No file(s) indexed")
			continue
		}

		// git commit
		if err = g.Commit(commitMsg); err != nil {
			log.Fatal(err)
		}
		log.Debug("Recorded changes to the repository")

		// git push
		if err = g.Push("origin"); err != nil {
			log.Fatal(err)
		}
		log.Debug("Pushed to remote repository")
	}
}
