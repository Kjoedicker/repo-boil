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

	buf, err := ioutil.ReadFile("host.yaml")
	if err != nil {
		panic("host.yaml - not in path")
	}

	conf := &hosts{}
	err = yaml.Unmarshal(buf, conf)
	if err != nil {
		panic("getCred()")
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

func giteaPost(values map[string]string) string {
	url := parseHosts(getConf())

	param := fmt.Sprintf(
		`-d "{ \"default_branch\": \"master\", \"description\": \"%v\", \"name\": \"%v\", \"private\": %v}"`,
		values["description"],
		values["name"],
		values["private"])

	cmd := fmt.Sprintf(
		`curl -X POST "%v" -H "accept: application/json" -H "Content-Type: application/json" %v`,
		url["gitea"],
		param)

	return cmd
}

// TODO(#1): add on the fly github support
func githubPost(values map[string]string) string {
	url := parseHosts(getConf())

	param := fmt.Sprintf(
		`-d "{\"name\": \"%v\", \"description\": \"%v\",  \"private\": %v}"`,
		values["name"],
		values["description"],
		values["private"])

	cmd := fmt.Sprintf(
		"curl  -u %v %v",
		url["github"],
		param)

	return cmd
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

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Not enough arguments")
		os.Exit(1)
	}

	values := make(map[string]string, len(os.Args))
	arguments := []string{"name", "description", "private"}

	for i := 2; i < len(os.Args); i++ {
		values[arguments[i-2]] = os.Args[i]
	}

	var cmd string
	switch os.Args[1] {
	case "gitea":
		cmd = giteaPost(values)
	case "github":
		cmd = githubPost(values)
	default:
		fmt.Println("No repo selected")
		os.Exit(2)
	}
	fmt.Println(cmd)
	runcmd(cmd, true)
}
