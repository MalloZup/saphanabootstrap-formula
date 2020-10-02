package main

import (
	"os"
	"os/exec"

	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
)

type saphanaboostrapFormula struct {
	Local struct {
		Hana struct {
			InstallPackages   bool   `yaml:"install_packages"`
			SaptuneSolution   string `yaml:"saptune_solution"`
			SoftwarePath      string `yaml:"software_path"`
			HanaArchiveFile   string `yaml:"hana_archive_file"`
			HanaExtractDir    string `yaml:"hana_extract_dir"`
			SapcarExeFile     string `yaml:"sapcar_exe_file"`
			HaEnabled         bool   `yaml:"ha_enabled"`
			MonitoringEnabled bool   `yaml:"monitoring_enabled"`
			Nodes             []struct {
				Host     string `yaml:"host"`
				Sid      string `yaml:"sid"`
				Instance int    `yaml:"instance"`
				Password string `yaml:"password"`
				Install  struct {
					SoftwarePath       string `yaml:"software_path"`
					RootUser           string `yaml:"root_user"`
					RootPassword       string `yaml:"root_password"`
					HdbPwdFile         string `yaml:"hdb_pwd_file"`
					SystemUserPassword string `yaml:"system_user_password"`
					SapadmPassword     string `yaml:"sapadm_password"`
				} `yaml:"install,omitempty"`
				Primary struct {
					Name    string `yaml:"name"`
					Userkey struct {
						KeyName      string `yaml:"key_name"`
						Environment  string `yaml:"environment"`
						UserName     string `yaml:"user_name"`
						UserPassword string `yaml:"user_password"`
						Database     string `yaml:"database"`
					} `yaml:"userkey"`
					Backup struct {
						KeyName      string `yaml:"key_name"`
						UserName     string `yaml:"user_name"`
						UserPassword string `yaml:"user_password"`
						Database     string `yaml:"database"`
						File         string `yaml:"file"`
					} `yaml:"backup"`
				} `yaml:"primary,omitempty"`
				Exporter struct {
					ExpositionPort int    `yaml:"exposition_port"`
					MultiTenant    bool   `yaml:"multi_tenant"`
					User           string `yaml:"user"`
					Password       string `yaml:"password"`
					Timeout        int    `yaml:"timeout"`
				} `yaml:"exporter,omitempty"`
				SaptuneSolution string `yaml:"saptune_solution,omitempty"`
				Secondary       struct {
					Name            string `yaml:"name"`
					RemoteHost      string `yaml:"remote_host"`
					RemoteInstance  string `yaml:"remote_instance"`
					ReplicationMode string `yaml:"replication_mode"`
					OperationMode   string `yaml:"operation_mode"`
					PrimaryTimeout  int    `yaml:"primary_timeout"`
				} `yaml:"secondary,omitempty"`
				ScenarioType            string `yaml:"scenario_type,omitempty"`
				CostOptimizedParameters struct {
					GlobalAllocationLimit string `yaml:"global_allocation_limit"`
					PreloadColumnTables   bool   `yaml:"preload_column_tables"`
				} `yaml:"cost_optimized_parameters,omitempty"`
			} `yaml:"nodes"`
		} `yaml:"hana"`
	}
}

const (
	//	# these variables are formula specific:
	formulaLog  = "/var/log/salt-hana-formula.log"
	formulaName = "hana"
	// this is where the pillar are located
	formulaConfig = "/usr/share/salt-formulas/config/hana"
	formulaPillar = "/usr/share/salt-formulas/config/hana/pillar/hana/hana.sls"
)

var (
	// global --help flag
	helpFlag *bool
)

func validatePillar() {
	var c saphanaboostrapFormula
	// convert jinja to yaml
	// https://docs.saltstack.com/en/latest/ref/modules/all/salt.modules.slsutil.html#salt.modules.slsutil.renderer
	cmd := exec.Command("/usr/bin/salt-call", "--local", "slsutil.renderer", "default_renderer=jinja", formulaPillar)
	stdout, err := cmd.Output()
	if err != nil {
		log.Error(err)
	}

	log.Info("[PREFLIGHT]: pillar rendered... converting to yaml")
	err = yaml.Unmarshal(stdout, &c)
	if err != nil {
		log.Fatalf("Formula Pillar data is not valid!: %v", err)
	}
	log.Printf("--- t:\n%v\n\n", c)
	log.Info("[PREFLIGHT]: pillar valid!")
}

func init() {
	flag.CommandLine.SortFlags = false
	flag.StringP("loglevel", "l", "error", "salt log level")
	helpFlag = flag.BoolP("help", "h", false, "show this help message")
}

func main() {

	flag.Parse()

	switch {
	case *helpFlag:
		flag.Usage()
		os.Exit(0)
	}

	log.Infof("[PREFLIGHT]: validating pillar of salt formula %s", formulaName)

	// copy grains to config of formula
	_, err := exec.Command("/usr/bin/cp", "/etc/salt/grains", formulaConfig).Output()
	if err != nil {
		log.Errorf("error while rendindering pillar jinja file %s", err)
	}

	// render and validate via static types pillars
	validatePillar()

	// run formula
	log.Infof("[FORMULA]: formula %s will be executed. Please wait..", formulaName)
	formulaOut, err := exec.Command("/usr/bin/salt-call", "--local", "--log-level=info",
		"--retcode-passthrough", "--force-color", "--config="+formulaConfig, "state.apply", formulaName).CombinedOutput()

	if err != nil {
		log.Errorf("error while executing salt formula %s, %s", err, formulaOut)
	}

}
