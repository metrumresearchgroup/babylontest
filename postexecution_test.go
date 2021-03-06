package bbitest

import (
	"bufio"
	"bytes"
	"context"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"
)

//Will need to set a custom env for execution
const postExecutionScriptString string = `#!/bin/bash

env > ${BBI_ROOT_EXECUTION_DIR}/working/${BBI_SCENARIO}/${BBI_MODEL_FILENAME}.out
`

func generatePostWorkEnvsString(content map[string]string) (string, error){
	//"--additional_post_work_envs=\"SCENARIO=" + v.identifier + "\",ROOT_EXECUTION_DIR=\"" + ROOT_EXECUTION_DIR + "\"",

	const templateString string = `--additional_post_work_envs="{{ range $key, $value := . }}{{$key}}={{$value}},{{end}}"`

	t, err := template.New("pieces").Parse(templateString)

	if err != nil {
		return "", err
	}

	outBuffer := new(bytes.Buffer)

	err = t.Execute(outBuffer,content)

	if err != nil {
		return "", err
	}

	stringResult := outBuffer.String()

	stringResult = strings.Replace(stringResult,`,"`,`"`,1)

	return stringResult, nil
}

func TestKVPExpansion(t *testing.T){
	mapdata := make(map[string]string)
	mapdata["SCENARIO"] = "240"
	mapdata["ROOT_EXECUTION_DIR"] = "/data/one"

	generated, err := generatePostWorkEnvsString(mapdata)

	require.Nil(t,err)
	require.NotNil(t,generated)
}

func TestPostExecutionSucceeds(t *testing.T){

	//Skip the test if the flag isn't enabled
	if ! FeatureEnabled("POST_EXECUTION"){
		t.Skip("Post execution not enabled as far as testing is concerned")
	}

	Scenarios := InitializeScenarios([]string{
		"240",
		"acop",
		"ctl_test",
		"metrum_std",
	})

	ioutil.WriteFile(filepath.Join(ROOT_EXECUTION_DIR,"post.sh"),[]byte(postExecutionScriptString),0755)



	for _, v := range Scenarios {
		v.Prepare(context.Background())

		arguments := []string{
			"-d",
			"nonmem",
			"--nm_version",
			os.Getenv("NMVERSION"),
			"run",
			"local",
			"--overwrite=true",
			"--post_work_executable",
			filepath.Join(ROOT_EXECUTION_DIR,"post.sh"),
			"--additional_post_work_envs=\"BBI_ROOT_EXECUTION_DIR=" + ROOT_EXECUTION_DIR  + " BBI_SCENARIO=" + v.identifier + "\"" ,
		}

		//Do the actual execution
		for _, m := range v.models {
			t.Run(v.identifier + "_post_execution",func(t *testing.T){
				output, err := m.Execute(v,arguments...)
				require.Nil(t,err)

				nmd := NonMemTestingDetails{
					t:         t,
					OutputDir: filepath.Join(v.Workpath,m.identifier),
					Model:     m,
					Output:    output,
				}

				AssertNonMemCompleted(nmd)
				AssertNonMemCreatedOutputFiles(nmd)

				exists, err := afero.Exists(afero.NewOsFs(),filepath.Join(ROOT_EXECUTION_DIR,"working",v.identifier,m.identifier + ".out") )

				require.Nil(t,err)
				require.True(t,exists)

				//Does the file contain the expected Details:
				//SCENARIO (Additional provided value)
				file, _ := os.Open(filepath.Join(ROOT_EXECUTION_DIR,"working",v.identifier, m.identifier + ".out"))
				defer file.Close()

				var lines []string

				scanner := bufio.NewScanner(file)
				//scanner.Split(bufio.ScanLines)

				for scanner.Scan() {
					lines = append(lines,scanner.Text())
				}


				require.True(t, doesOutputFileContainKeyWithValue(lines,"BBI_MODEL",m.filename))
				require.True(t, doesOutputFileContainKeyWithValue(lines, "BBI_MODEL_FILENAME", m.identifier))
				require.True(t, doesOutputFileContainKeyWithValue(lines, "BBI_MODEL_EXT", strings.Replace(m.extension,".","",1)))
				require.True(t, doesOutputFileContainKeyWithValue(lines, "BBI_SUCCESSFUL", "true"))
				require.True(t, doesOutputFileContainKeyWithValue(lines, "BBI_ERROR", ""))

			})

		}
	}


	//Test a scenario for the first scenario where we force failure. Model is deleted (not found)
	t.Run("verify_failure_results", func(t *testing.T){

		var lines []string

		scenario := Scenarios[0]
		scenario.Prepare(context.Background())



		arguments := []string{
			"nonmem",
			"--nm_version",
			os.Getenv("NMVERSION"),
			"run",
			"local",
			"--post_work_executable",
			filepath.Join(ROOT_EXECUTION_DIR,"post.sh"),
			"--overwrite=false",
			//`--additional_post_work_envs "BBI_SCENARIO=` + scenario.identifier + ` BBI_ROOT_EXECUTION_DIR=` + ROOT_EXECUTION_DIR  + `"`,
			//"--additional_post_work_envs BBI_ROOT_EXECUTION_DIR=" + ROOT_EXECUTION_DIR,
		}

		//Removing the model won't do anything. Execute with overwrite = false?
		for _, v := range scenario.models {
			os.Setenv("BBI_ADDITIONAL_POST_WORK_ENVS",`BBI_SCENARIO=` + scenario.identifier + ` BBI_ROOT_EXECUTION_DIR=` + ROOT_EXECUTION_DIR)
			os.Remove(filepath.Join(scenario.Workpath,v.identifier + ".out"))
			output, err := v.Execute(scenario, arguments...)

			//Does the file contain the expected Details:
			//SCENARIO (Additional provided value)
			file, _ := os.Open(filepath.Join(ROOT_EXECUTION_DIR,"working",scenario.identifier, v.identifier + ".out"))
			defer file.Close()

			scanner := bufio.NewScanner(file)
			//scanner.Split(bufio.ScanLines)

			for scanner.Scan() {
				lines = append(lines,scanner.Text())
			}



			require.NotNil(t,err)
			require.Error(t,err)

			require.True(t, doesOutputFileContainKeyWithValue(lines, "BBI_SUCCESSFUL", "false"))
			if err != nil {
				require.True(t, doesExecutionOutputContainErrorString(err.Error(), output))
			}

		}
	})

}


func doesOutputFileContainKeyWithValue(lines []string, key string, value string) bool {

	for _, v := range lines {
		if strings.Contains(v,key+"=") {
			components := strings.Split(v,"=")
			return components[0] == key && components[1] == value
		}
	}

	return false
}

func doesExecutionOutputContainErrorString(line string, output string) bool {

	lines := strings.Split(output, "\n")

	for _, v := range lines {
		if strings.Contains(line, v){
			//We have a match
			return true
		}
	}

	return false
}




