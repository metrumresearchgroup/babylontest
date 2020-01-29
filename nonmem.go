package babylontest

import (
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"strings"
	"testing"
)

type NonMemTestingDetails struct {
	t *testing.T
	OutputDir string
	Model Model
}

func AssertNonMemCompleted(details NonMemTestingDetails){
	nmlines, err := fileLines(filepath.Join(details.OutputDir,details.Model.identifier + ".lst"))

	assert.Nil(details.t,err)
	assert.NotNil(details.t,nmlines)
	assert.NotEmpty(details.t,nmlines)
	//Make sure that nonmem shows it finished and generated files
	assert.Contains(details.t,strings.Join(nmlines,"\n"),"finaloutput")
	//Make sure that nonmem records a stop time
	assert.Contains(details.t,strings.Join(nmlines,"\n"),"Stop Time:")
}

func AssertNonMemCreatedOutputFiles( details NonMemTestingDetails){
	fs := afero.NewOsFs()
	expected := []string{
		".xml",
		".cpu",
		".grd",
	}

	for _, v := range expected {
		ok, _ := afero.Exists(fs,filepath.Join(details.OutputDir,details.Model.identifier + v))
		assert.True(details.t,ok,"Unable to locate expected file %s",v)
	}
}

func AssertContainsBBIScript(t *testing.T, outputPath string, m Model){

}