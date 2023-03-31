package crocgodyl

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type AppServer struct {
	ID            int           `json:"id"`
	ExternalID    string        `json:"external_id"`
	UUID          string        `json:"uuid"`
	Identifier    string        `json:"identifier"`
	Name          string        `json:"name"`
	Description   string        `json:"description"`
	Status        string        `json:"status,omitempty"`
	Suspended     bool          `json:"suspended"`
	Limits        Limits        `json:"limits"`
	FeatureLimits FeatureLimits `json:"feature_limits"`
	User          int           `json:"user"`
	Node          int           `json:"node"`
	Allocation    int           `json:"allocation"`
	Nest          int           `json:"nest"`
	Egg           int           `json:"egg"`
	Container     struct {
		StartupCommand string                 `json:"startup_command"`
		Image          string                 `json:"image"`
		Installed      int                    `json:"installed"`
		Environment    map[string]interface{} `json:"environment"`
	} `json:"container"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

func (s *AppServer) BuildDescriptor() *ServerBuildDescriptor {
	return &ServerBuildDescriptor{
		Allocation:        s.Allocation,
		OOMDisabled:       s.Limits.OOMDisabled,
		Limits:            s.Limits,
		AddAllocations:    []int{},
		RemoveAllocations: []int{},
		FeatureLimits:     s.FeatureLimits,
	}
}

func (s *AppServer) DetailsDescriptor() *ServerDetailsDescriptor {
	return &ServerDetailsDescriptor{
		ExternalID:  s.ExternalID,
		Name:        s.Name,
		User:        s.User,
		Description: s.Description,
	}
}

func (s *AppServer) StartupDescriptor() *ServerStartupDescriptor {
	return &ServerStartupDescriptor{
		Startup:     s.Container.StartupCommand,
		Environment: s.Container.Environment,
		Egg:         s.Egg,
		Image:       s.Container.Image,
	}
}

func (a *Application) GetServers() ([]*AppServer, error) {
	req := a.newRequest("GET", "/servers", nil)
	res, err := a.Http.Do(req)
	if err != nil {
		return nil, err
	}

	buf, err := validate(res)
	if err != nil {
		return nil, err
	}

	var model struct {
		Data []struct {
			Attributes *AppServer `json:"attributes"`
		} `json:"data"`
		PaginationMeta
	}
	if err = json.Unmarshal(buf, &model); err != nil {
		return nil, err
	}

	servers := make([]*AppServer, 0, len(model.Data))
	for _, s := range model.Data {
		servers = append(servers, s.Attributes)
	}

	allServers := servers
	for model.PaginationMeta.Meta.Pagination.CurrentPage <= model.PaginationMeta.Meta.Pagination.TotalPages {
		req = a.newRequest("GET", fmt.Sprintf("/servers?page=%d", model.PaginationMeta.Meta.Pagination.CurrentPage + 1), nil)
		res, err = a.Http.Do(req)
		if err != nil {
			return nil, err
		}

		buf, err = validate(res)
		if err != nil {
			return nil, err
		}

		if err = json.Unmarshal(buf, &model); err != nil {
			return nil, err
		}

		for _, s := range model.Data {
			allServers = append(allServers, s.Attributes)
		}

	}

	return allServers, nil
}

func (a *Application) GetServer(id int) (*AppServer, error) {
	req := a.newRequest("GET", fmt.Sprintf("/servers/%d", id), nil)
	res, err := a.Http.Do(req)
	if err != nil {
		return nil, err
	}

	buf, err := validate(res)
	if err != nil {
		return nil, err
	}

	var model struct {
		Attributes AppServer `json:"attributes"`
	}
	if err = json.Unmarshal(buf, &model); err != nil {
		return nil, err
	}

	return &model.Attributes, nil
}

func (a *Application) GetServerExternal(id string) (*AppServer, error) {
	req := a.newRequest("GET", fmt.Sprintf("/servers/external/%s", id), nil)
	res, err := a.Http.Do(req)
	if err != nil {
		return nil, err
	}

	buf, err := validate(res)
	if err != nil {
		return nil, err
	}

	var model struct {
		Attributes AppServer `json:"attributes"`
	}
	if err = json.Unmarshal(buf, &model); err != nil {
		return nil, err
	}

	return &model.Attributes, nil
}

type AllocationDescriptor struct {
	Default    int   `json:"default"`
	Additional []int `json:"additional,omitempty"`
}

type DeployDescriptor struct {
	Locations   []int    `json:"locations"`
	DedicatedIP bool     `json:"dedicated_ip"`
	PortRange   []string `json:"port_range"`
}

type CreateServerDescriptor struct {
	ExternalID        string                 `json:"external_id,omitempty"`
	Name              string                 `json:"name"`
	Description       string                 `json:"description,omitempty"`
	User              int                    `json:"user"`
	Egg               int                    `json:"egg"`
	DockerImage       string                 `json:"docker_image"`
	Startup           string                 `json:"startup"`
	Environment       map[string]interface{} `json:"environment"`
	SkipScripts       bool                   `json:"skip_scripts,omitempty"`
	OOMDisabled       bool                   `json:"oom_disabled"`
	Limits            *Limits                `json:"limits"`
	FeatureLimtis     FeatureLimits          `json:"feature_limits"`
	Allocation        *AllocationDescriptor  `json:"allocation,omitempty"`
	Deploy            *DeployDescriptor      `json:"deploy,omitempty"`
	StartOnCompletion bool                   `json:"start_on_completion,omitempty"`
}

func (a *Application) CreateServer(fields CreateServerDescriptor) (*AppServer, error) {
	if fields.Allocation == nil && fields.Deploy == nil {
		return nil, errors.New("the allocation object or deploy object must be specified")
	}

	data, _ := json.Marshal(fields)
	body := bytes.Buffer{}
	body.Write(data)

	req := a.newRequest("POST", "/servers", &body)
	res, err := a.Http.Do(req)
	if err != nil {
		return nil, err
	}

	buf, err := validate(res)
	if err != nil {
		return nil, err
	}

	var model struct {
		Attributes AppServer `json:"attributes"`
	}
	if err = json.Unmarshal(buf, &model); err != nil {
		return nil, err
	}

	return &model.Attributes, nil
}

type ServerBuildDescriptor struct {
	Allocation        int           `json:"allocation,omitempty"`
	OOMDisabled       bool          `json:"oom_disabled,omitempty"`
	Limits            Limits        `json:"limits,omitempty"`
	AddAllocations    []int         `json:"add_allocations,omitempty"`
	RemoveAllocations []int         `json:"remove_allocations,omitempty"`
	FeatureLimits     FeatureLimits `json:"feature_limits,omitempty"`
}

func (a *Application) UpdateServerBuild(id int, fields ServerBuildDescriptor) (*AppServer, error) {
	data, _ := json.Marshal(fields)
	if len(data) == 2 {
		return nil, errors.New("no build fields specified")
	}

	body := bytes.Buffer{}
	body.Write(data)

	req := a.newRequest("PATCH", fmt.Sprintf("/servers/%d/build", id), &body)
	res, err := a.Http.Do(req)
	if err != nil {
		return nil, err
	}

	buf, err := validate(res)
	if err != nil {
		return nil, err
	}

	var model struct {
		Attributes AppServer `json:"attributes"`
	}
	if err = json.Unmarshal(buf, &model); err != nil {
		return nil, err
	}

	return &model.Attributes, nil
}

type ServerDetailsDescriptor struct {
	ExternalID  string `json:"external_id,omitempty"`
	Name        string `json:"name,omitempty"`
	User        int    `json:"user,omitempty"`
	Description string `json:"description,omitempty"`
}

func (a *Application) UpdateServerDetails(id int, fields ServerDetailsDescriptor) (*AppServer, error) {
	data, _ := json.Marshal(fields)
	if len(data) == 2 {
		return nil, errors.New("no details fields specified")
	}

	body := bytes.Buffer{}
	body.Write(data)

	req := a.newRequest("PATCH", fmt.Sprintf("/servers/%d/details", id), &body)
	res, err := a.Http.Do(req)
	if err != nil {
		return nil, err
	}

	buf, err := validate(res)
	if err != nil {
		return nil, err
	}

	var model struct {
		Attributes AppServer `json:"attributes"`
	}
	if err = json.Unmarshal(buf, &model); err != nil {
		return nil, err
	}

	return &model.Attributes, nil
}

type ServerStartupDescriptor struct {
	Startup     string                 `json:"startup"`
	Environment map[string]interface{} `json:"environment"`
	Egg         int                    `json:"egg,omitempty"`
	Image       string                 `json:"image"`
	SkipScripts bool                   `json:"skip_scripts"`
}

func (a *Application) UpdateServerStartup(id int, fields ServerStartupDescriptor) (*AppServer, error) {
	data, _ := json.Marshal(fields)
	if len(data) == 2 {
		return nil, errors.New("no startup fields specified")
	}

	body := bytes.Buffer{}
	body.Write(data)

	req := a.newRequest("PATCH", fmt.Sprintf("/servers/%d/startup", id), &body)
	res, err := a.Http.Do(req)
	if err != nil {
		return nil, err
	}

	buf, err := validate(res)
	if err != nil {
		return nil, err
	}

	var model struct {
		Attributes AppServer `json:"attributes"`
	}
	if err = json.Unmarshal(buf, &model); err != nil {
		return nil, err
	}

	return &model.Attributes, nil
}

func (a *Application) SuspendServer(id int) error {
	req := a.newRequest("POST", fmt.Sprintf("/servers/%d/suspend", id), nil)
	res, err := a.Http.Do(req)
	if err != nil {
		return err
	}

	_, err = validate(res)
	return err
}

func (a *Application) UnsuspendServer(id int) error {
	req := a.newRequest("POST", fmt.Sprintf("/servers/%d/unsuspend", id), nil)
	res, err := a.Http.Do(req)
	if err != nil {
		return err
	}

	_, err = validate(res)
	return err
}

func (a *Application) DeleteServer(id int, force bool) error {
	url := fmt.Sprintf("/servers/%d", id)
	if force {
		url += "/force"
	}

	req := a.newRequest("DELETE", url, nil)
	res, err := a.Http.Do(req)
	if err != nil {
		return err
	}

	_, err = validate(res)
	return err
}
