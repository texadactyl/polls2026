package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Show help and then exit to the O/S
func showHelp() {
	name := filepath.Base(os.Args[0])
	fmt.Printf("\nUsage:  %s  [ options ]\n\nwhere\n", name)
	fmt.Printf("\t-h : This display.\n")
	fmt.Printf("\t-l : Load updated polls into the database.\n")
	fmt.Printf("\t-r 1 ST : Given state code ST, report Democrat, GOP, Independent, and undecided results.\n")
	fmt.Printf("\t-r 2 PCT : Run a mock poll of polls based on last Oct polls. The PCT is added to the GOP total if there is sufficient Unknown\n")
	fmt.Printf("\t-v : Verbose console output.\n")
	fmt.Print("\n")
	os.Exit(0)
}

func main() {

	const MyName = "polls2026"
	const PathDirDownloads = "./downloads"
	const PathDirLogs = "./logs"
	const PathVersionCfg = "./VERSION.cfg"
	const Url = "https://electoral-vote.com/evp2026/Senate/senate_polls.txt"

	var args []string
	var versionString string
	var err error
	stateCode := "rubbish!"
	var gopPctIncrement float64
	var verFile *os.File

	// Get the string timestamp and zone for the current time in the default locale.
	var startStamp string
	var startTime = time.Now() // Get the current time.
	if runtime.GOOS == "windows" {
		startStamp = startTime.Format("2006-01-02_150405")
	} else {
		startStamp = startTime.Format("2006-01-02_15:04:05")
	}
	var logFileName = fmt.Sprintf("%s.log", startStamp)

	// Open log file and defer closing.
	openLogFile(PathDirLogs, logFileName)
	defer closeLogFile()

	// Flags
	flagVerbose := false       // -v
	flagLoad := false          // -l
	flagReportByState := false // -r 1 ST
	flagPollOfPolls := false   // -r 2 PCT

	// Positioned in the tree top directory?
	verFile, err = os.Open("VERSION.cfg")
	if err != nil {
		fatalText("You are not positioned in the poll2026 tree top directory")
	} else {
		_ = verFile.Close()
	}

	// Initialise Args to the command-line arguments.
	args = append(args, os.Args[1:]...)

	// If no options specified, just show help.
	if len(args) < 1 {
		showHelp()
	}

	// Parse command line arguments.
	for ii := 0; ii < len(args); ii++ {
		switch args[ii] {

		case "-h":
			showHelp()

		case "-l":
			flagLoad = true

		case "-r":
			ii++
			if ii >= len(args) {
				logError("Missing -r <report type>")
				showHelp()
			}
			arg := args[ii]
			switch arg {
			case "1":
				ii++
				if ii >= len(args) {
					logError("Missing -r 1 <state code>")
					showHelp()
				}
				flagReportByState = true
				stateCode = strings.ToUpper(args[ii])
			case "2":
				ii++
				if ii >= len(args) {
					logError("Missing -r 2 <percentage>")
					showHelp()
				}
				flagPollOfPolls = true
				gopPctIncrement = stringToFloat64(args[ii], "GOP pct increment")
			default:
				logError("Unrecognizable -r argument: %s", args[ii])
				showHelp()
			}

		case "-v":
			flagVerbose = true

		default:
			logError("*** ERROR,main: Unrecognizable argument: %s", args[ii])
			showHelp()
		}

	}

	// Fetch and show the program version.
	versionBytes, err := os.ReadFile(PathVersionCfg)
	if err != nil {
		fatalErr(err, "InitGlobals: ReadFile(%s) failed", PathVersionCfg)
	}
	versionString = cleanerText(string(versionBytes), false)
	versionString = strings.Trim(versionString, " ")
	logger("%s %s", MyName, versionString)

	// Open database.
	DBOpen(flagVerbose)
	defer DBClose()

	// Open the 'logs' directory.
	err = isDirectory(PathDirLogs)
	if err != nil {
		fatalErr(err, "isDirectory(%s) returned an error", PathDirLogs)
	}

	// Open the 'downloads' directory.
	err = isDirectory(PathDirDownloads)
	if err != nil {
		fatalErr(err, "isDirectory(%s) returned an error", PathDirDownloads)
	}

	if flagLoad {
		loader(startStamp, PathDirDownloads, Url)
	}

	if flagReportByState {
		reportByState(stateCode)
	}

	if flagPollOfPolls {
		pollOfPolls("", gopPctIncrement)
	}

	logger("End")

}
