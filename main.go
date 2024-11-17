package main

import (
	"fmt"
	"keyswap/plist"
	"log"
	"os/exec"
	"runtime"
	"strings"
)

func main() {
	os := runtime.GOOS

	if os != "darwin" {
		log.Fatal("This program should be run on a darwin OS (MacOs) machine")
	}

	cmd := exec.Command("hidutil", "property", "--get", "UserKeyMapping")

	output, err := cmd.Output()
	if err != nil {
		log.Fatal("error getting mapping key value")
	}

	mapping := plist.GetMapping(string(output))
	fileLoc := ""
	if !plist.MappingIsApplied(mapping) {
		fileLoc, err := plist.UseTemplateToCreatePlist()
		if err != nil {
			panic(err)
		}
		fmt.Println("Add launch Agent to ", fileLoc)
		cmd = exec.Command("launchctl", "load", fileLoc)
		output, err := cmd.Output()
		if err != nil {
			panic(err)
		}
		fmt.Println(string(output))
		cmd = exec.Command("bash", "-c", "launchctl list | grep remap")
		output, err = cmd.CombinedOutput()
		if err != nil {
			fmt.Println("error checking $> launchtl list")
			fmt.Println(err)
			panic(err)
		}
		fmt.Printf("Execution result %q\n", strings.TrimSpace(string(output)))
	} else {
		fmt.Printf("The mapping is already applied\n")
		fileLoc = plist.GetFileLocation()
	}
	fmt.Println("to unload/disable the launch agent run the following commands")
	fmt.Printf("\n\nlaunchctl unload %s\n", fileLoc)
	fmt.Printf("%s\n", `hidutil property --set '{"UserKeyMapping":[]}'`)
}
