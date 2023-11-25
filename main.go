package main

import (
	"embed"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	gittools "github.com/sebastianrau/deployer/pkg/gitTools"
	ostools "github.com/sebastianrau/deployer/pkg/osTools"
	"github.com/sebastianrau/deployer/pkg/steps"
)

var (
	version = "1.0.1"
)

//go:embed keys/*
var keyFiles embed.FS

func main() {

	const (
		// TODO change folder name

		configTemplateFile string = "./temp/kw-bem-missions/kw_bem_make.template.yaml"
		configDataFile     string = "./temp/kw-bem-missions/kw_bem_make.data.yaml"
	)

	var (
		verbose      = flag.Bool("v", false, "verbose output")
		versionPrint = flag.Bool("version", false, "print version")
	)

	var (
		err   error
		check bool
	)

	flag.Parse()

	if *versionPrint {
		fmt.Printf("KW Mission Builder %s\n", version)
		os.Exit(0)

	}
	fmt.Printf("KW Mission Builder\n")
	fmt.Println()

	check, _ = CheckKnownHosts("github.com")
	if !check {
		fmt.Println("github.com is no present in you known hosts file!\nPlease call\n     ssh -T git@github.com \nto add it to knows hosts")
		os.Exit(2)
	}
	defer removeTemp()

	id_BasicMissions, err := keyFiles.ReadFile("keys/id_BasicMissions")
	if err != nil {
		fmt.Println("No Keyfile found for id_basic_missions")
		os.Exit(3)
	}

	id_BasicScripts, err := keyFiles.ReadFile("keys/id_BasicScripts")
	checkError(err, 1)

	var writer io.Writer

	if *verbose {
		writer = os.Stdout
	} else {
		writer = nil
	}

	_, _, err = gittools.UpdateKeyBytes(
		// TODO change folder name

		"git@github.com:kdo-wildsau/kw-bem-basic-scripts.git",
		"./temp/kw-bem-basic-scripts",
		id_BasicScripts,
		"",
		writer,
	)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(4)
	} else {
		fmt.Printf("%-*s %s\n", 80, "Update kw-bem-basic-scripts", "OK")
	}

	_, _, err = gittools.UpdateKeyBytes(
		// TODO change folder name
		"git@github.com:kdo-wildsau/kw-bem-missions.git",
		"./temp/kw-bem-missions",
		id_BasicMissions,
		"",
		writer,
	)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(5)
	} else {
		fmt.Printf("%-*s %s\n", 80, "Update kw-bem-missions", "OK")
	}

	// read and unmarshal template / no keyfile needed
	st, err := steps.UnmarshalConfigTemplate(configTemplateFile, configDataFile)

	if err != nil {
		removeTemp()
		fmt.Println(err)
		os.Exit(1)
	}

	start := time.Now()
	st.Exceute(os.Stdout, *verbose)
	stop := time.Now()
	elapsed := stop.Sub(start)

	fmt.Printf("Time: %s\n", elapsed.Round(time.Millisecond))
	removeTemp()
	os.Exit(0)

}

func removeTemp() {
	err := ostools.Delete(
		"./temp/",
		nil,
	)
	if err != nil {
		panic(err.Error())
	} else {
		fmt.Printf("%-*s %s\n", 80, "Removing tempfolder", "OK")
	}
}

func CheckKnownHosts(url string) (bool, error) {

	homeDir, _ := os.UserHomeDir()
	knownHosts := homeDir + "/.ssh/known_hosts"
	bytes, err := os.ReadFile(knownHosts)
	if err != nil {
		return false, err
	}

	return strings.Contains(string(bytes), url), nil
}

func checkError(err error, returnCode int) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(returnCode)
	}
}
