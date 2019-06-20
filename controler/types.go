package controler

type ServiceStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type AllStatuses struct {
	Services []ServiceStatus `json:"services"`
}

type NewService struct {
	Name             string            `json:"name" form:"name" bind:"command"`
	Command          string            `json:"command" form:"command" bind:"command"`
	WorkingDirectory string            `json:"work_dir" form:"work_dir" bind:"work_dir"`
	Environment      map[string]string `json:"environment" form:"environment" bind:"environment"`
}
