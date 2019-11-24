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
		fmt.Println(chalk.Red.Color("Provide the name of the file containing a list of domains."))
		return
	}

	if len(os.Args) < 3 {
		fmt.Println(chalk.Red.Color("Provide the email with which the certificate will be registered."))
		fmt.Println("Check the hidden '.lego' folder for existing accounts.")
		return
	}

	domainsLinesFile := os.Args[1]
	userEmail := os.Args[2]

	if userEmail == "" || !strings.ContainsRune(userEmail, '@') {
		fmt.Println(chalk.Red.Color("You did not provide a valid email address."))
		return
	}

	f, err := ioutil.ReadFile(domainsLinesFile)
	if err != nil {
		fmt.Printf("ERR: %s \n", err)
		return
	}

	absPath, err := filepath.Abs(domainsLinesFile)
	if err != nil {
		fmt.Printf("ERR: %s \n", err)
		return
	}

	saFileDir, _ := filepath.Split(absPath)
	saFileName := saFileDir + "sa.json"

	saFile, err := os.Open(saFileName)
	if err != nil {
		fmt.Printf("ERR: %v \n", err)
		return
	}
	defer saFile.Close()

	saData := make(map[string]interface{}, 10)
	if err = json.NewDecoder(saFile).Decode(&saData); err != nil {
		fmt.Printf("Err decoding JSON: %v\n", err)
		return
	}

	proj, ok := saData["project_id"].(string)
	if !ok || proj == "" {
		fmt.Printf("SA file is in the wrong format")
		return
	}

	domainsLines := strings.Split(string(f), "\n")

	domains := make([]string, 0, len(domainsLines)*2)
	for _, d := range domainsLines {
		if d == "" {
			continue
		}
		domains = append(domains, d)

		// If this is a sub-domain, we won't also get its "www" sub-domain.
		if strings.Count(d, ".") <= 1 {
			domains = append(domains, "www."+d)
		}

		// You cannot request a certificate for www.example.com and *.example.com.
		if strings.HasPrefix(d, "*.") {
			domains = removeString(domains, "www"+d[1:])
		}
	}

	legoCmd := make([]string, 0, len(domains)+5)
	legoCmd = append(legoCmd, "lego", "--dns=gcloud", "--email="+userEmail, "--key-type=ec256")

	for _, d := range domains {
		legoCmd = append(legoCmd, "-d="+d)
	}

	legoCmd = append(legoCmd, "run")

	legoPath, err := exec.LookPath("lego")
	if err != nil {
		log.Fatal("Could not find lego program path")
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
		return
	}

}

func removeString(slice []string, str string) []string {
	for i := 0; i < len(slice); i++ {
		if slice[i] == str {
			fmt.Printf("Removing %q from slice.\n", str)
			slice = append(slice[:i], slice[i+1:]...)
			i--
		}
	}
	return slice
}
