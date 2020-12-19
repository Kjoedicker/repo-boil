package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"gopkg.in/yaml.v2"
)

type hosts struct {
	Hosts []struct {
		Repo struct {
			Name  string `yaml:"name"`
			URL   string `yaml:"url"`
			Token string `yaml:"token"`
		} `yaml:"repo"`
	} `yaml:"hosts"`
}

func getConf() (parsed *hosts) {
	home := os.Getenv("HOME")
	path := fmt.Sprintf("%v/.config/repo-boil/host.yaml", home)

	buf, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("host.yaml - not in path")
		os.Exit(1)
	}

	conf := &hosts{}
	err = yaml.Unmarshal(buf, conf)
	if err != nil {
		fmt.Println("Invalid conf.yaml configuration")
		os.Exit(1)
	}

	return conf
}

func parseHosts(h *hosts) map[string]string {
	parsed := make(map[string]string, len(h.Hosts))

	for _, host := range h.Hosts {
		if host.Repo.Name == "gitea" {
			parsed[host.Repo.Name] = fmt.Sprintf("%v%v", host.Repo.URL, host.Repo.Token)
		}

		if host.Repo.Name == "github" {
			parsed[host.Repo.Name] = fmt.Sprintf("%v %v", host.Repo.Token, host.Repo.URL)
		}
	}

	return parsed
}

func genValues() map[string]string {
	values := make(map[string]string, len(os.Args))
	arguments := []string{"name", "description", "private"}

	for i := 2; i < len(os.Args); i++ {
		values[arguments[i-2]] = os.Args[i]
	}

	return values
}

func genTemplate(host string, cmd string, param string) string {
	url := parseHosts(getConf())
	values := genValues()

	parameter := fmt.Sprintf(
		param,
		values["name"],
		values["description"],
		values["private"])

	command := fmt.Sprintf(
		cmd,
		url[host],
		parameter)

	return command
}

func giteaPost() string {
	cmd := `curl -X POST "%v" -H "accept: application/json" -H "Content-Type: application/json" %v`
	param := `-d "{ \"default_branch\": \"master\", \"name\": \"%v\", \"description\": \"%v\", \"private\": %v}"`

	return genTemplate("gitea", cmd, param)
}

func githubPost() string {
	cmd := "curl  -u %v %v"
	param := `-d "{\"name\": \"%v\", \"description\": \"%v\",  \"private\": %v}"`

	return genTemplate("github", cmd, param)
}

func runcmd(cmd string, shell bool) []byte {
	if shell {
		out, err := exec.Command("bash", "-c", cmd).Output()
		if err != nil {
			log.Fatal(err)
			panic("some error found")
		}
		return out
	}
	out, err := exec.Command(cmd).Output()
	if err != nil {
		log.Fatal(err)
	}
	return out
}

// TODO(#2): How can this be made more dynamic. In a way that repos don't have to be hardcoded, but rather, declared.
func main() {
	if len(os.Args) < 3 {
		fmt.Println("Not enough arguments")
		os.Exit(1)
	}

	var cmd string
	switch os.Args[1] {
	case "gitea":
		cmd = giteaPost()
	case "github":
		cmd = githubPost()
	default:
		fmt.Println("No repo selected")
		os.Exit(2)
	}

	fmt.Println(cmd)
	runcmd(cmd, true)
}
