package controler

import (
	"os"
	"os/exec"
	"testing"
)

const testData = "test-data"

func TestConf_Create(t *testing.T) {

	controller := NewServiceControllerByPath(testData + "/config.json")
	err := controller.Create(NewService{
		Name:             "test-gm",
		Command:          resolveBin("nc") + " -l 9000",
		WorkingDirectory: testData,
	})
	if err != nil {
		t.Error("create service", err)
		return
	}
	status := controller.Status("test-gm")
	if status.Status != "dead" {
		t.Error("mismatch status:", status.Status)
		return
	}

	err = controller.Run("test-gm")
	if err != nil {
		t.Error("run service", err)
		return
	}

	status = controller.Status("test-gm")
	if status.Status != "running" {
		t.Error("mismatch status:", status.Status)
		return
	}

	err = controller.Stop("test-gm")
	if err != nil {
		t.Error("stop service", err)
		return
	}

	status = controller.Status("test-gm")
	if status.Status != "dead" {
		t.Error("mismatch status:", status.Status)
		return
	}

	err = controller.Disable("test-gm")
	if err != nil {
		t.Error("disable service", err)
		return
	}
}

func resolveBin(bin string) string {
	cmd, err := exec.LookPath(bin)
	if err != nil {
		panic(err)
	}
	return cmd
}

func init() {
	os.RemoveAll(testData)
	os.MkdirAll(testData, 0755)
}
