package controler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
)

type ServiceController interface {
	RefreshStatus() []string
	Status(name string) string
	Restart(name string) error
	Run(name string) error
	Stop(name string) error
}

type Conf struct {
	Services []string `json:"services"`
}

func NewServiceController() ServiceController {
	jFile, err := ioutil.ReadFile(CFG_PATH)
	if err != nil {
		panic(err)
	}
	var data Conf
	err = json.Unmarshal(jFile, &data)
	if err != nil {
		panic(err)
	}

	fmt.Printf("[MONITOR]: Append srv list: %s", &data.Services)
	return &data
}

func (cfg *Conf) RefreshStatus() []string {
	res := make([]string, len(cfg.Services))
	for _, srv := range cfg.Services {
		result := cfg.Status(srv)
		fmt.Println(result)
		res = append(res, result)
	}
	return res
}

func (cfg *Conf) Status(name string) string {
	result, err := control(name, STAT)
	if err != nil {
		fmt.Printf("[ERROR]: Status for srv: %s", name)
		panic(err)
	}
	return result
}

func (cfg *Conf) Restart(name string) error {
	_, err := control(name, RESTART)
	if err != nil {
		fmt.Printf("[ERROR]: Restart srv: %s", name)
		return err
	}
	return nil
}

func (cfg *Conf) Run(name string) error {
	_, err := control(name, RUN)
	if err != nil {
		fmt.Printf("[ERROR]: Run srv: %s", name)
		return err
	}
	return nil
}

func (cfg *Conf) Stop(name string) error {
	_, err := control(name, STOP)
	if err != nil {
		fmt.Printf("[ERROR]: Run srv: %s", name)
		return err
	}
	return nil
}

func control(name string, operation string) (string, error) {
	stdout := &bytes.Buffer{}
	cmd := exec.Command(COMMAND, name, operation)
	cmd.Stdout = io.Writer(stdout)
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	res := stdout.String()

	return res, nil
}
