package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type CLIConfig struct {
	BaseURL  string `json:"base_url"`
	APIKey   string `json:"api_key"` // #nosec G117
	ProjectID string `json:"project_id"`
}

type CLI struct {
	config *CLIConfig
	client *http.Client
}

func NewCLI() *CLI {
	return &CLI{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *CLI) LoadConfig(path string) error {
	// #nosec G304
	data, err := os.ReadFile(path)
	if err != nil {
		c.config = &CLIConfig{BaseURL: "https://api.limiter.io/v1"}
		return nil
	}
	return json.Unmarshal(data, c.config)
}

func (c *CLI) SaveConfig(path string) error {
	// #nosec G117
	data, err := json.MarshalIndent(c.config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func (c *CLI) request(method, path string, body interface{}) (map[string]interface{}, error) {
	url := c.config.BaseURL + path
	var reqBody io.Reader
	if body != nil {
		data, _ := json.Marshal(body)
		reqBody = bytes.NewReader(data)
	}
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}
	if c.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&result)
	return result, nil
}

func (c *CLI) Login(email, password string) error {
	result, err := c.request("POST", "/auth/login", map[string]string{
		"email":    email,
		"password": password,
	})
	if err != nil {
		return err
	}
	if token, ok := result["access_token"].(string); ok {
		c.config.APIKey = token
		fmt.Println("Logged in successfully")
	}
	return nil
}

func (c *CLI) CreateProject(name string) error {
	result, err := c.request("POST", "/projects", map[string]string{"name": name})
	if err != nil {
		return err
	}
	if id, ok := result["id"].(string); ok {
		c.config.ProjectID = id
		fmt.Printf("Project created: %s\n", id)
	}
	return nil
}

func (c *CLI) CreateKey(name, scope string) error {
	path := fmt.Sprintf("/projects/%s/apikeys", c.config.ProjectID)
	result, err := c.request("POST", path, map[string]string{"name": name, "scope": scope})
	if err != nil {
		return err
	}
	if key, ok := result["key"].(string); ok {
		fmt.Printf("API Key created: %s\n", key)
	}
	return nil
}

func (c *CLI) PushRule(name, route string, maxReq int, windowMs int64) error {
	path := fmt.Sprintf("/projects/%s/rules", c.config.ProjectID)
	result, err := c.request("POST", path, map[string]interface{}{
		"name":     name,
		"route":    route,
		"max_req":  maxReq,
		"window_ms": windowMs,
		"enabled":  true,
	})
	if err != nil {
		return err
	}
	fmt.Printf("Rule created: %+v\n", result)
	return nil
}

func (c *CLI) TestRoute(method, route string) error {
	path := fmt.Sprintf("/projects/%s/gateway", c.config.ProjectID)
	result, err := c.request("POST", path, map[string]string{
		"method": method,
		"route":  route,
	})
	if err != nil {
		return err
	}
	fmt.Printf("Test result: %+v\n", result)
	return nil
}

func (c *CLI) ListProjects() error {
	result, err := c.request("GET", "/projects", nil)
	if err != nil {
		return err
	}
	fmt.Printf("Projects: %+v\n", result)
	return nil
}

func (c *CLI) SetProject(id string) {
	c.config.ProjectID = id
	fmt.Printf("Active project set to: %s\n", id)
}

func RunCLI(args []string) {
	cli := NewCLI()
	home, _ := os.UserHomeDir()
	configPath := home + "/.limiter-config.json"
	_ = cli.LoadConfig(configPath)

	if len(args) < 2 {
		fmt.Println(`Limiter.io CLI
Usage:
  limiter login <email> <password>
  limiter create-project <name>
  limiter create-key <name> [scope]
  limiter push-rule <name> <route> <max_req> <window_ms>
  limiter test-route <method> <route>
  limiter list-projects
  limiter set-project <id>
  limiter help`)
		return
	}

	cmd := args[1]
	switch cmd {
	case "login":
		if len(args) < 4 { fmt.Println("Usage: limiter login <email> <password>"); return }
		if err := cli.Login(args[2], args[3]); err != nil {
			fmt.Fprintf(os.Stderr, "Login failed: %v\n", err)
		}
	case "create-project":
		if len(args) < 3 { fmt.Println("Usage: limiter create-project <name>"); return }
		if err := cli.CreateProject(args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create project: %v\n", err)
		}
	case "create-key":
		if len(args) < 3 { fmt.Println("Usage: limiter create-key <name> [scope]"); return }
		scope := "gateway-only"
		if len(args) >= 4 { scope = args[3] }
		if err := cli.CreateKey(args[2], scope); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create key: %v\n", err)
		}
	case "push-rule":
		if len(args) < 6 { fmt.Println("Usage: limiter push-rule <name> <route> <max_req> <window_ms>"); return }
		maxReq := 100
		var windowMs int64 = 60000
		_, _ = fmt.Sscanf(args[4], "%d", &maxReq)
		_, _ = fmt.Sscanf(args[5], "%d", &windowMs)
		if err := cli.PushRule(args[2], args[3], maxReq, windowMs); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to push rule: %v\n", err)
		}
	case "test-route":
		if len(args) < 4 { fmt.Println("Usage: limiter test-route <method> <route>"); return }
		if err := cli.TestRoute(args[2], args[3]); err != nil {
			fmt.Fprintf(os.Stderr, "Test failed: %v\n", err)
		}
	case "list-projects":
		if err := cli.ListProjects(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to list projects: %v\n", err)
		}
	case "set-project":
		if len(args) < 3 { fmt.Println("Usage: limiter set-project <id>"); return }
		cli.SetProject(args[2])
	default:
		fmt.Println(`Commands: login, create-project, create-key, push-rule, test-route, list-projects, set-project`)
	}

	if err := cli.SaveConfig(configPath); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to save config: %v\n", err)
	}
}

func main() {
	RunCLI(strings.Fields(os.Args[0]))
}
