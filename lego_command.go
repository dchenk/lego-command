package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ttacon/chalk"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println(chalk.Red.Color("Provide the name of the file with a list of domains."))
		return
	}

	if len(os.Args) < 3 {
		fmt.Println(chalk.Red.Color("Provide the email with which the certificate will be registered."))
		fmt.Println("Check the hidden '.lego' folder for existing accounts.")
		return
	}

	domainsListFileName := os.Args[1]
	userEmail := os.Args[2]

	if userEmail == "" || !strings.ContainsRune(userEmail, '@') {
		fmt.Println(chalk.Red.Color("You did not provide a valid email address."))
		return
	}

	f, err := ioutil.ReadFile(domainsListFileName)
	if err != nil {
		fmt.Printf("ERR: %s \n", err)
		return
	}

	saFileDir, _ := filepath.Split(domainsListFileName)
	saFileName := saFileDir + "sa.json"

	saFile, err := os.Open(saFileName)
	if err != nil {
		fmt.Printf("ERR: %v \n", err)
		os.Exit(1)
	}
	defer saFile.Close()

	saData := make(map[string]interface{}, 10)
	if err = json.NewDecoder(saFile).Decode(&saData); err != nil {
		fmt.Printf("Err decoding JSON: %v\n", err)
		os.Exit(1)
	}

	proj, ok := saData["project_id"].(string)
	if !ok || proj == "" {
		fmt.Printf("SA file is in the wrong format")
		os.Exit(1)
	}

	domainsList := strings.Split(string(f), "\n")

	legoCmd := make([]string, 0, len(domainsList)*2+5)
	legoCmd = append(legoCmd, "lego")

	legoCmd = append(legoCmd, "--dns=gcloud", "--email="+userEmail, "--key-type=ec256")

	for _, d := range domainsList {
		if d == "" {
			continue
		}
		legoCmd = append(legoCmd, "-d="+d, "-d=www."+d)
	}

	legoCmd = append(legoCmd, "run")

	legoPath, err := exec.LookPath("lego")
	if err != nil {
		log.Fatal("could not find lego path")
	}

	cmd := &exec.Cmd{
		Path: legoPath,
		Args: legoCmd,
		Env: []string{
			"GCE_PROJECT=" + proj,
			"GCE_SERVICE_ACCOUNT_FILE=" + saFileName,
		},
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	if err = cmd.Run(); err != nil {
		fmt.Println(chalk.Red.Color(fmt.Sprintf("ERR running: %v\n", err)))
		os.Exit(1)
	}

}
