package controler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sukauto/templates"
	"sync"
)

type Access interface {
	Login(username string, password string) (err error)
}

type ServiceController interface {
	RefreshStatus() AllStatuses
	Status(name string) ServiceStatus
	Restart(name string) error
	Run(name string) error
	Stop(name string) error
	Enable(name string) error  // enable autostart
	Disable(name string) error // disable autostart
	Create(service NewService) error
	Update(name string) error
	Attach(name string) error // attach exists service
	Forget(name string) error // forget about service
	Log(name string) (string, error)
	Groups() []string
	// Create new group
	Group(name string) error
	// Remove group
	Ungroup(name string) error
	Members(groupName string) []string
	Join(groupName string, serviceName string) error
	Leave(groupName string, serviceName string) error
	Events() <-chan Event
}

type AccessServiceController interface {
	ServiceController
	Access
}

type Conf struct {
	Services   []string            `json:"services,omitempty"`
	GroupsList map[string][]string `json:"groups,omitempty"`
	Global     bool                `json:"global"` // as a system-wide services, otherwise - user based
	Users      map[string]string   `json:"users"`  // no users means no login
	location   string              `json:"-"`      // config file location
	event      chan Event
	updCmd     string
	lock       sync.RWMutex
}

func NewServiceControllerByPath(location string, updcmd string) AccessServiceController {
	jFile, err := ioutil.ReadFile(location)
	if os.IsNotExist(err) {
		// create default
		cfg := &Conf{
			Users:    map[string]string{"root": "root"},
			location: location,
			updCmd:   updcmd,
			event:    make(chan Event),
		}
		err = cfg.save()
		if err != nil {
			panic(err)
		}
		return cfg
	}
	if err != nil {
		panic(err)
	}
	var data Conf
	err = json.Unmarshal(jFile, &data)
	if err != nil {
		panic(err)
	}
	data.location = location
	data.updCmd = updcmd
	data.event = make(chan Event)
	fmt.Printf("[MONITOR]: Append srv list: %s", &data.Services)
	return &data
}

func NewServiceController(cmdUpd string) AccessServiceController {
	return NewServiceControllerByPath(CFG_PATH, cmdUpd)
}

func (cfg *Conf) Events() <-chan Event {
	return cfg.event
}

func (cfg *Conf) Groups() []string {
	cfg.lock.RLock()
	defer cfg.lock.RUnlock()
	var ans []string
	for name := range cfg.GroupsList {
		ans = append(ans, name)
	}
	sort.Strings(ans)
	return ans
}

func (cfg *Conf) Ungroup(name string) error {
	cfg.lock.Lock()
	defer cfg.lock.Unlock()
	delete(cfg.GroupsList, name)
	return cfg.saveUnsafe()
}
func (cfg *Conf) Group(name string) error {
	cfg.lock.Lock()
	defer cfg.lock.Unlock()
	name = strings.TrimSpace(name)
	for gname := range cfg.GroupsList {
		if strings.EqualFold(gname, name) {
			return errors.New("group already exists")
		}
	}
	if cfg.GroupsList == nil {
		cfg.GroupsList = make(map[string][]string)
	}
	cfg.GroupsList[name] = make([]string, 0)
	return cfg.saveUnsafe()
}

func (cfg *Conf) Members(groupName string) []string {
	cfg.lock.RLock()
	defer cfg.lock.RUnlock()
	return cfg.GroupsList[groupName]
}

func (cfg *Conf) Join(groupName string, serviceName string) error {
	cfg.lock.Lock()
	defer cfg.lock.Unlock()
	if cfg.GroupsList == nil {
		cfg.GroupsList = make(map[string][]string)
	}
	for _, item := range cfg.GroupsList[groupName] {
		if item == serviceName {
			return nil
		}
	}
	if !cfg.isServiceExists(serviceName) {
		return errors.New("service not exists")
	}
	cfg.GroupsList[groupName] = append(cfg.GroupsList[groupName], serviceName)
	return cfg.saveUnsafe()
}

func (cfg *Conf) Leave(groupName string, serviceName string) error {
	cfg.lock.Lock()
	defer cfg.lock.Unlock()
	for i, srv := range cfg.GroupsList[groupName] {
		if srv == serviceName {
			ar := cfg.GroupsList[groupName]
			cfg.GroupsList[groupName] = append(ar[:i], ar[i+1:]...)
			return cfg.saveUnsafe()
		}
	}
	return nil
}

func (cfg *Conf) RefreshStatus() AllStatuses {
	res := make([]ServiceStatus, 0)
	for _, srv := range cfg.Services {
		result := cfg.Status(srv)
		res = append(res, result)
	}
	return AllStatuses{Services: res}
}

func (cfg *Conf) Status(name string) ServiceStatus {
	result, err := controlQueryField(name, FieldStatus, !cfg.Global)
	if err != nil {
		fmt.Printf("[ERROR]: Status for srv: %s", name)
		return ServiceStatus{Status: StateUnknown, Name: name}
	}
	return ServiceStatus{Status: result, Name: name}
}

func (cfg *Conf) Restart(name string) error {
	_, err := control(name, RESTART, !cfg.Global)
	if err != nil {
		fmt.Printf("[ERROR]: Restart srv: %s", name)
		return err
	} else {
		cfg.event <- Event{Type: EventRestarted, Name: name}
	}
	return nil
}

func (cfg *Conf) Run(name string) error {
	_, err := control(name, RUN, !cfg.Global)
	if err != nil {
		fmt.Printf("[ERROR]: Run srv: %s", name)
		return err
	} else {
		cfg.event <- Event{Type: EventStarted, Name: name}
	}
	return nil
}

func (cfg *Conf) Stop(name string) error {
	_, err := control(name, STOP, !cfg.Global)
	if err != nil {
		fmt.Printf("[ERROR]: Run srv: %s", name)
		return err
	} else {
		cfg.event <- Event{Type: EventStopped, Name: name}
	}
	return nil
}

func (cfg *Conf) Update(name string) error {
	var err error
	preUpdInfo := cfg.Status(name)

	err = cfg.Stop(name)
	if err != nil {
		fmt.Printf("[ERROR]: Stop srv on upd: %s", name)
		return err
	}

	_, err = updater(name, cfg.updCmd, !cfg.Global)
	if err != nil {
		fmt.Printf("[ERROR]: Update srv: %s", name)
		return err
	}

	if preUpdInfo.Status == "running" {
		err = cfg.Run(name)
		if err != nil {
			fmt.Printf("[ERROR]: Start srv on upd: %s", name)
			return err
		}
	}
	cfg.event <- Event{Type: EventUpdated, Name: name}
	return nil
}

func (cfg *Conf) isServiceExists(name string) bool {
	for _, srv := range cfg.Services {
		if srv == name {
			return true
		}
	}
	return false
}

func updater(name string, updcmd string, user bool) (string, error) {
	stdout := &bytes.Buffer{}
	srvWorkDir, _ := controlQueryField(name, WORKDIR, user)
	// remove 'WorkingDirectory=' from string
	srvWorkDir = strings.TrimSpace(srvWorkDir)
	if len(srvWorkDir) > 0 && srvWorkDir[0] == '!' {
		srvWorkDir = srvWorkDir[1:]
	}

	cmd := exec.Command(SHELL, "-c", updcmd)
	cmd.Stdout = io.Writer(stdout)
	cmd.Stderr = os.Stderr
	cmd.Dir = srvWorkDir
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	res := stdout.String()
	return res, nil

}

func (cfg *Conf) Create(service NewService) error {
	cfg.lock.Lock()
	defer cfg.lock.Unlock()
	// resolve working directory
	workingDir, err := filepath.Abs(service.WorkingDirectory)
	if err != nil {
		return err
	}
	service.WorkingDirectory = workingDir
	// generate unit file
	data := &bytes.Buffer{}
	err = templates.ServiceUnitTemplate.Execute(data, service)
	if err != nil {
		return err
	}
	// detect location for unit file
	var location = LocationGlobal
	if !cfg.Global {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		location = filepath.Join(home, LocationUser)
	}
	// ensure that target directory exists
	err = os.MkdirAll(location, 0755)
	if err != nil {
		return err
	}
	unitFile := filepath.Join(location, service.Name+".service")
	// save unit file
	err = ioutil.WriteFile(unitFile, data.Bytes(), 0755)
	if err != nil {
		return err
	}
	// install (enable)
	err = cfg.Enable(service.Name)
	if err != nil {
		return err
	}
	// save to config
	// TODO: maybe save full information
	cfg.Services = append(cfg.Services, service.Name)
	err = cfg.saveUnsafe()
	if err != nil {
		return err
	}
	cfg.event <- Event{Type: EventCreated, Name: service.Name}
	return nil
}

func (cfg *Conf) Attach(name string) error {
	cfg.lock.Lock()
	defer cfg.lock.Unlock()
	cfg.Services = append(cfg.Services, name)
	err := cfg.saveUnsafe()
	if err != nil {
		return err
	}
	cfg.event <- Event{Type: EventCreated, Name: name}
	return nil
}

func (cfg *Conf) Enable(name string) error {
	_, err := control(name, CmdEnable, !cfg.Global)
	if err == nil {
		cfg.event <- Event{Type: EventEnabled, Name: name}
	}
	return err
}

func (cfg *Conf) Disable(name string) error {
	_, err := control(name, CmdDisable, !cfg.Global)
	if err == nil {
		cfg.event <- Event{Type: EventDisabled, Name: name}
	}
	return err
}

func (cfg *Conf) Login(username string, password string) (err error) {
	if len(cfg.Users) == 0 {
		return nil
	}
	expected, ok := cfg.Users[username]
	if !ok || expected != password {
		return errors.New("invalid user or password")
	}
	return nil
}

func (cfg *Conf) Log(name string) (string, error) {
	return journal(name, !cfg.Global)
}

func (cfg *Conf) Forget(name string) error {
	cfg.lock.Lock()
	defer cfg.lock.Unlock()
	for i, srv := range cfg.Services {
		if srv == name {
			cfg.Services = append(cfg.Services[:i], cfg.Services[i+1:]...)
			break
		}
	}
	// remove from groups
	for group, services := range cfg.GroupsList {
		for i, srv := range services {
			if srv == name {
				cfg.GroupsList[group] = append(services[:i], services[i+1:]...)
				break
			}
		}
	}
	err := cfg.saveUnsafe()
	if err != nil {
		return err
	}
	cfg.event <- Event{Type: EventRemoved, Name: name}
	return nil
}

func (cfg *Conf) save() error {
	cfg.lock.RLock()
	defer cfg.lock.RUnlock()
	return cfg.saveUnsafe()
}

func (cfg *Conf) saveUnsafe() error {
	data, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(cfg.location, data, 0755)
}

func control(name string, operation string, user bool) (string, error) {
	stdout := &bytes.Buffer{}
	var args []string
	if user {
		args = append(args, ModeUser)
	}
	args = append(args, operation, name)
	cmd := exec.Command(COMMAND, args...)
	cmd.Stdout = io.Writer(stdout)
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	res := stdout.String()

	return res, nil
}

func journal(name string, user bool) (string, error) {
	stdout := &bytes.Buffer{}
	var args = []string{ModeMergeJournals, ModeNoPages, ModeQuite, ModeLimit, strconv.Itoa(LogLimit)}
	if user {
		args = append(args, ModeUserUnit)
	} else {
		args = append(args, ModeSystemUnit)
	}
	args = append(args, name)
	cmd := exec.Command(JournalCommand, args...)
	cmd.Stdout = io.Writer(stdout)
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	res := strings.TrimSpace(stdout.String())

	return res, nil
}

func controlQueryField(name string, field string, user bool) (string, error) {
	stdout := &bytes.Buffer{}
	var args []string
	if user {
		args = append(args, ModeUser)
	}
	args = append(args, CmdShow, "-p", field, "--value", name)
	cmd := exec.Command(COMMAND, args...)
	cmd.Stdout = io.Writer(stdout)
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	res := strings.TrimSpace(stdout.String())

	return res, nil
}
