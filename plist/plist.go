package plist

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
	"text/template"
)

type Setting struct {
	Version, FileName, MappingJson string
}

type Swappable struct {
	HIDKeyboardModifierMappingSrc, HIDKeyboardModifierMappingDst uint64
}

type MapObj struct {
	UserKeyMapping []Swappable
}

const (
	plistLocation = "/Library/LaunchAgents/"
	fileName      = "com.keyboard.hidutil-remap.plist"
)

const (
	CTRLKEY = 0x7000000e0
	RCMD    = 0x7000000e7
)

const (
	initTmpl         = `{"UserKeyMapping":{{.List}}}`
	swappableArgTmpl = `{"HIDKeyboardModifierMappingSrc":{{.HIDKeyboardModifierMappingSrc | hex}},"HIDKeyboardModifierMappingDst":{{.HIDKeyboardModifierMappingDst|hex}}}`
)

var possibleKeys = [2]string{"HIDKeyboardModifierMappingDst", "HIDKeyboardModifierMappingSrc"}

func (ent *Swappable) isComplete() bool {
	return ent.HIDKeyboardModifierMappingDst > 0 &&
		ent.HIDKeyboardModifierMappingSrc > 0
}

func (ent *Swappable) setField(fieldName, value string) {
	val := reflect.ValueOf(ent).Elem()
	field := val.FieldByName(fieldName)

	if field.IsValid() && field.CanSet() {
		if num, err := strconv.ParseUint(value, 10, 64); err == nil { // strconv.Atoi(value); err == nil {
			field.SetUint(num)
		}
	}
}

func (set *Setting) updateSetting(conf []Swappable) {
	set.Version = "1.0"
	set.FileName = fileName
	set.MappingJson = createUserKeyMapping(conf)
}

var funcMap = template.FuncMap{
	"hex": func(n uint64) string {
		return fmt.Sprintf("0x%x", n)
	},
}

type ListMap struct {
	List []string
}

func createUserKeyMapping(conf []Swappable) string {
	fmt.Println(" arguments length", len(conf))
	userMap, err := template.New("userMap").Parse(initTmpl)
	if err != nil {
		panic(err)
	}

	argList := make([]string, len(conf))
	entryTmpl, err := template.New("entry").Funcs(funcMap).Parse(swappableArgTmpl)
	if err != nil {
		panic(err)
	}

	for _, entry := range conf {
		var entryText strings.Builder
		entryTmpl.Execute(&entryText, entry)
		argList = append(argList, entryText.String())
	}

	var templateResult strings.Builder
	userMap.Execute(&templateResult, ListMap{argList})
	return templateResult.String()
}

func GetFileLocation() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	fullPath := path.Join(homeDir, plistLocation, fileName)
	return fullPath
}

func UseTemplateToCreatePlist() (filename string, er error) {
	list := []Swappable{
		{RCMD, CTRLKEY},
	}
	settings := Setting{}
	settings.updateSetting(list)

	templ, err := template.New("plist").Parse(fileTemplate)
	if err != nil {
		return "", err
	}

	fullPath := GetFileLocation()

	file, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	err = templ.Execute(file, settings)
	if err != nil {
		return "", err
	}
	return fullPath, nil
}

func MappingIsApplied(ls []Swappable) bool {
	if len(ls) == 0 {
		return false
	}
	for _, swp := range ls {
		if swp.HIDKeyboardModifierMappingSrc == RCMD && swp.HIDKeyboardModifierMappingDst == CTRLKEY {
			return true
		}
	}
	return false
}

func isValidKey(k string) bool {
	for _, okKey := range possibleKeys {
		if k == okKey {
			return true
		}
	}
	return false
}

func GetMapping(output string) []Swappable {
	var res []Swappable
	entry := Swappable{}
	lines := strings.Split(output, "\n")
	for _, line := range lines {

		if entry.isComplete() {
			res = append(res, entry)
			entry = Swappable{}
		}

		if strings.Contains(line, "=") {
			keyVal := strings.Split(line, "=")
			if len(keyVal) == 2 {
				key := strings.TrimSpace(keyVal[0])
				val := strings.TrimSpace(keyVal[1])
				val = strings.Replace(val, ";", "", 1)
				if isValidKey(key) {
					entry.setField(key, val)
				}
			}
		}
	}
	return res
}

const fileTemplate = `
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="{{.Version}}">
  <dict>
    <key>Label</key>
    <string>{{.FileName}}</string>

    <!-- Path to the shell script or directly to the command -->
    <key>ProgramArguments</key>
    <array>
      <string>/usr/bin/env</string>
      <string>bash</string>
      <string>-c</string>
      <string>hidutil property --set '{{.MappingJson}}'
      </string>
    </array>

    <key>RunAtLoad</key>
    <true/>
    <key>StartInterval</key>
    <integer>300</integer> <!-- Optional: Runs every 5 minutes in case the change is lost -->
  </dict>
</plist>
`

/*
* cbase command line execution
 */
