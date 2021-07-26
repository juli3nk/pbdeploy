package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	gogit "github.com/go-git/go-git"
	"github.com/juli3nk/go-git"
	"github.com/juli3nk/go-utils"
	"github.com/juli3nk/go-utils/filedir"
	"github.com/juli3nk/pbdeploy/pkg/config"
	pfile "github.com/juli3nk/pbdeploy/pkg/file"
	"github.com/juli3nk/pbdeploy/pkg/repository"
	log "github.com/sirupsen/logrus"
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
	log.Debugf("PACKAGE_VERSION = %s", version)

	if string(version[0]) != "v" {
		log.Fatalf("package version (%s) is not prefixed with a \"v\"", version)
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
		log.Debugf("Changed directory to %s (%s)", *dstPath, cwd())

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
			log.Debugf("Cloned repository '%s'", pack.Name)

			// Change dir
			if err := os.Chdir(pack.Name); err != nil {
				log.Fatal(err)
			}
			log.Debugf("Changed directory (%s)", cwd())

			if err := g.Open(); err != nil {
				log.Fatal(err)
			}
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
			log.Debugf("%s => %v", file, status.Worktree)

			if status.Worktree == gogit.Untracked || status.Worktree == gogit.Modified {
				if err := g.Add(file); err != nil {
					log.Fatal(err)
				}
				stagedCount += 1

				log.Debugf("Added file (%s) contents to the index", file)
			}
			if status.Worktree == gogit.Deleted {
				if err := g.Remove(file); err != nil {
					log.Fatal(err)
				}
				stagedCount += 1

				log.Debugf("Removed file (%s) contents from the index", file)
			}
		}

		if stagedCount == 0 {
			log.Debug("No file(s) indexed")
		} else {
			// git commit
			if err = g.Commit(commitMsg); err != nil {
				log.Fatal(err)
			}
			log.Debug("Recorded changes to the repository")

			// git push commit
			if err = g.Push("origin", "", false); err != nil {
				log.Fatal(err)
			}
			log.Debug("Pushed to remote repository")

			if pack.CreateTag {
				// git tag
				if err = g.CreateTag(version, fmt.Sprintf("Release %s", version)); err != nil {
					log.Fatal(err)
				}
				log.Debugf("Created tag (%s)", version)

				// git push tag
				if err = g.Push("origin", version, false); err != nil {
					log.Fatal(err)
				}
				log.Debug("Pushed tag to remote repository")
			}
		}
	}
}

func cwd() string {
	dir, _ := os.Getwd()

	return dir
}
