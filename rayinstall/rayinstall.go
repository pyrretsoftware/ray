package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

//go:embed version
var Version string

//go:embed rays
var Raysbinary []byte

func SizeToString(size int) string {
	units := []string{"B", "KB", "MB", "GB", "TB", "PB"}

	sI := 0
	for size >= 1000 && sI < len(units) {
		size /= 1000
		sI++
	}

	return strconv.Itoa(size) + units[sI]
}

var purpleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
func purple(str string) string {
	return purpleStyle.Render(str)
}

var Grey = lipgloss.NewStyle().Foreground(lipgloss.Color("#808080"))

var fileEnding string
func main() {
	Version = strings.ReplaceAll(Version, "\n", "")
	Version = strings.ReplaceAll(Version, "\r", "")
	forceFlag := flag.Bool("force", false, "forces update even if ray server appears to be currently running.")
	flag.Parse()


	fmt.Println(purple(                                                   
`                                             ▄▄ ▄▄ 
                  ▀▀               ██        ██ ██ 
████▄  ▀▀█▄ ██ ██ ██  ████▄ ▄█▀▀▀ ▀██▀▀ ▀▀█▄ ██ ██ 	
██ ▀▀ ▄█▀██ ██▄██ ██  ██ ██ ▀███▄  ██  ▄█▀██ ██ ██ 
██    ▀█▄██  ▀██▀ ██▄ ██ ██ ▄▄▄█▀  ██  ▀█▄██ ██ ██ 
              ██                                   
            ▀▀▀                                    `))
	fmt.Println("Welcome to ray server's installation tool, rayinstall!")
	fmt.Println("Ray " + purple(Version) + " has been embedded inside of this tool. Total embedded size is " + purple(SizeToString(len(Raysbinary))))
	fmt.Println()

	fmt.Println("This installer collects some telemetry about your installation such as OS, Processor architecture, version, etc for me to know how to prioritize backwards compatability and get statistics on installations. Once ray server is installed it will never contact the internet without you telling it to.")
	fmt.Print(purple("Continue?") + " (y/n) ")
	continueStr := ""
	_, err := fmt.Scan(&continueStr)
	if err != nil || (continueStr != "y" && continueStr != "yes")  {
		fmt.Println("Aborting...")
		os.Exit(0)
	}

	fileEnding = ""
	if (runtime.GOOS == "windows") {
		fileEnding = ".exe"
	}
	
	installLocation := "/usr/bin"
	if runtime.GOOS == "windows" {
		dir, err := os.UserCacheDir()
		if err != nil {
			log.Fatal(err)
		}
		dir = filepath.Join(dir, "rays")

		os.MkdirAll(dir, 0600)
		if err != nil {
			log.Fatal(err)
		}
		installLocation = dir
	}
	fmt.Println("Enviroument is", purple(installLocation))

	_, err = os.Stat(path.Join(installLocation, "rays" + fileEnding))
	alreadyInstalled := err == nil
	installText := "Install"
	if alreadyInstalled {
		installText = "Update"
	}

	fmt.Println()
	fmt.Println("What would you like to do?")
	fmt.Println(Grey.Render("←/→ - move • ↵ - select"))

	boxes := tea.NewProgram(box{
		items: []string{
			"📦\n" + installText,
			"🔧\nRepair",
			"🧹\nUninstall",
			"💾\nExport",
		},
		itemsAvailable: []bool{
			true,
			alreadyInstalled,
			alreadyInstalled,
			true,
		},
	})
	var boxResultRaw tea.Model
    if boxResultRaw, err = boxes.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
    }
	boxResult := boxResultRaw.(box).active
	if boxResult == -1 {
		os.Exit(0)
	}
	
	fmt.Println("---------------------------------------------------------------")
	fmt.Println("Rayinstall ready on " + purple(runtime.GOOS + "/" + runtime.GOARCH) + " version " + purple(Version))

	fmt.Print("Starting process: ")
	switch boxResult {
	case 0:
		Install(installLocation, *forceFlag, false)
		Collect("installation")
	case 1:
		Install(installLocation, *forceFlag, true)
		Collect("installation")
	case 2:
		Collect("uninstallation")
		Uninstall(installLocation, *forceFlag)
	case 3:
		if err := os.WriteFile("./export-rays" + fileEnding, Raysbinary, 0600); err != nil {
			fmt.Println("Export error: " + err.Error())
		} else {
			fmt.Println("Export OK")
		}
	} 

	fmt.Println("Press enter to exit...")
	_inp := ""
	fmt.Scan(&_inp)
}