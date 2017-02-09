package config

import (
	"encoding/json"
	"io/ioutil"
	"os/user"
	"path/filepath"

	"github.com/Skarlso/go-furnace/errorhandler"
)

var configPath string

// Configuration is a Configuration object.
type Configuration struct {
	LogLevel   string
	UploadPath string
}

// Monitoring monitoring
type Monitoring struct {
	Enable bool `json:"enabled"`
}

// EC2Config EC2 configuration
type EC2Config struct {
	DryRun       bool       `json:"dry_run"`
	ImageID      string     `json:"image_id"`
	KeyName      string     `json:"key_name"`
	MinCount     int64      `json:"min_count"`
	MaxCount     int64      `json:"max_count"`
	InstanceType string     `json:"instance_type"`
	Monitoring   Monitoring `json:"monitoring"`
}

// IPRange ip range
type IPRange struct {
	// CidrIP cidr_ip
	CidrIP string `json:"cidr_ip"`
}

// IPPermission ip permission
type IPPermission struct {
	IPProtocol string    `json:"ip_protocol"`
	FromPort   string    `json:"from_port"`
	ToPort     string    `json:"to_port"`
	IPRanges   []IPRange `json:"ip_ranges"`
}

// SecurityGroup Security Group
type SecurityGroup struct {
	IPPermissions []IPPermission `json:"ip_permission"`
}

// Path retrieves the main configuration path.
func Path() string {
	// Get configuration path
	usr, err := user.Current()
	errorhandler.CheckError(err)
	return filepath.Join(usr.HomeDir, ".config", "go-furnace")
}

func init() {
	configPath = Path()
}

// LoadEC2Config Loads the EC2 configuration file into the representive struct.
func LoadEC2Config() (ec2Config *EC2Config) {
	dat, err := ioutil.ReadFile(filepath.Join(configPath, "ec2_conf.json"))
	if err != nil {
		errorhandler.CheckError(err)
	}
	ec2Config = &EC2Config{}
	err = json.Unmarshal(dat, &ec2Config)
	if err != nil {
		errorhandler.CheckError(err)
	}
	return ec2Config
}
