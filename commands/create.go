package commands

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Skarlso/go-furnace/config"
	"github.com/Skarlso/go-furnace/utils"
	"github.com/Yitsushi/go-commander"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/fatih/color"
)

// Create command.
type Create struct {
}

// CFClient abstraction for cloudFormation client.
type CFClient struct {
	Client cloudformationiface.CloudFormationAPI
}

// Execute defines what this command does.
func (c *Create) Execute(opts *commander.CommandHelper) {
	stackname := opts.Arg(0)
	if len(stackname) < 1 {
		stackname = config.STACKNAME
	}

	config := config.LoadCFStackConfig()
	log.Println("Creating cloud formation session.")
	sess := session.New(&aws.Config{Region: aws.String("eu-central-1")})
	cfClient := cloudformation.New(sess, nil)
	client := CFClient{cfClient}
	createStack(stackname, config, &client)
}

// createStack will create a full stack and encapsulate the functionality of
// the create command.
func createStack(stackname string, config []byte, cfClient *CFClient) {
	validateParams := &cloudformation.ValidateTemplateInput{
		TemplateBody: aws.String(string(config)),
	}
	log.Println("Validating template.")
	validResp, err := cfClient.Client.ValidateTemplate(validateParams)
	log.Println("Response from validate:", validResp)
	utils.CheckError(err)
	var stackParameters []*cloudformation.Parameter
	keyName := color.New(color.FgWhite, color.Bold).SprintFunc()
	defaultValue := color.New(color.FgHiBlack, color.Italic).SprintFunc()
	log.Println("Gathering parameters.")
	for _, v := range validResp.Parameters {
		var param cloudformation.Parameter
		fmt.Printf("%s - '%s'(%s):", *v.Description, keyName(*v.ParameterKey), defaultValue(*v.DefaultValue))
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		param.SetParameterKey(*v.ParameterKey)
		text = strings.Trim(text, "\n")
		if len(text) > 0 {
			param.SetParameterValue(*aws.String(text))
		} else {
			param.SetParameterValue(*v.DefaultValue)
		}
		stackParameters = append(stackParameters, &param)
	}
	stackInputParams := &cloudformation.CreateStackInput{
		StackName:    aws.String(stackname),
		Parameters:   stackParameters,
		TemplateBody: aws.String(string(config)),
	}
	log.Println("Creating Stack with name: ", keyName(stackname))
	resp, err := cfClient.Client.CreateStack(stackInputParams)
	utils.CheckError(err)
	describeStackInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackname),
	}
	log.Println("Create stack response: ", resp.GoString())
	utils.WaitForFunctionWithStatusOutput("CREATE_COMPLETE", func() {
		cfClient.Client.WaitUntilStackCreateComplete(describeStackInput)
	})
	descResp, err := cfClient.Client.DescribeStacks(&cloudformation.DescribeStacksInput{StackName: aws.String(stackname)})
	utils.CheckError(err)
	fmt.Println()
	var red = color.New(color.FgRed).SprintFunc()
	if len(descResp.Stacks) > 0 {
		log.Println("Stack state is: ", red(*descResp.Stacks[0].StackStatus))
	}
}

// NewCreate Creates a new Create command.
func NewCreate(appName string) *commander.CommandWrapper {
	return &commander.CommandWrapper{
		Handler: &Create{},
		Help: &commander.CommandDescriptor{
			Name:             "create",
			ShortDescription: "Create a stack",
			LongDescription:  `Create a stack on which to deploy code to later on. By default FurnaceStack is used as name.`,
			Arguments:        "name",
			Examples:         []string{"create", "create MyStackName"},
		},
	}
}
