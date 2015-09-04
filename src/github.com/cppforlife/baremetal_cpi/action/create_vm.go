package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bwcvm "github.com/cppforlife/baremetal_cpi/vm"
	httpclient "github.com/cppforlife/baremetal_cpi/utils/httpclient"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"strings"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"time"
	"net/http"
	"errors"
)

type CreateVM struct {
	APIServer string
	agentOptions bwcvm.AgentOptions
	logger boshlog.Logger
	logTag string
}

type VMCloudProperties struct{}

type Environment map[string]interface{}

func NewCreateVM(APIServer string, agentOptions bwcvm.AgentOptions, logger boshlog.Logger) CreateVM {
	return CreateVM{
		APIServer: APIServer,
		agentOptions: agentOptions,
		logger: logger,
		logTag: "create_vm",
	}
}

func (a CreateVM) Run(agentID string, stemcellCID StemcellCID, _ VMCloudProperties, networks Networks, _ []DiskCID, env Environment) (VMCID, error) {
	client := httpclient.NewHTTPClient(httpclient.DefaultClient, a.logger)
	machineID, err := a.getMachineID()
	if err != nil {
		bosherr.WrapError(err, "Error getting machine ID")
	}

    macAddress, err := a.getMacAddress(machineID)
	if err != nil {
		bosherr.WrapError(err, "Error getting mac address")
	}

    //create agent and network env
	vmNetworks := networks.AsVMNetworks()
	vmEnv := bwcvm.Environment(env)
	agentEnv := bwcvm.NewAgentEnvForVM(agentID,machineID, vmNetworks, vmEnv, a.agentOptions, macAddress)
	jsonBytes, err := json.Marshal(agentEnv)
	if err != nil {
		bosherr.WrapError(err, "Marshalling agent env")
	}

	a.logger.Info(a.logTag, "Agent Json data '%s'", string(jsonBytes))

    //call api to provision the machine
	url := fmt.Sprintf("http://%s:8080/api/common/nodes/%s/workflows", a.APIServer, machineID)
	path := "/var/vcap/bosh/baremetal-cpi-agent-env.json"

	jsonStr := fmt.Sprintf("{\"name\":\"Graph.CF.CreateVM\",\"options\":{\"defaults\": {\"file\": \"%s\",\"path\": \"%s\", \"env\": \"%s\"}}}", stemcellCID, path, strings.Replace(string(jsonBytes), "\"", "\\\"", -1))
	a.logger.Info(a.logTag, "Json string is '%s'", jsonStr)

	resp, err := client.Post(url, strings.NewReader(jsonStr))
	if err != nil {
		bosherr.WrapError(err, "Error uploading stemcell")
	}

	if (resp.StatusCode != http.StatusCreated) {
		return VMCID(""), errors.New("Error creating nodes")
	}

	//succeeded in creating workflow, so wait for the workflow to finish before returning
	for a.isWorkFlowActive(machineID) {
		time.Sleep(30 * time.Second)
	}

	return VMCID(machineID), nil
}

func (a CreateVM) getMachineID() (string, error) {
	var machineID string //initialized to be ""
	nodesUrl := fmt.Sprintf("http://%s:8080/api/common/nodes", a.APIServer)
	resp, err := http.Get(nodesUrl)
	if err != nil {
		return machineID, errors.New("Error getting response")
	}

	defer resp.Body.Close()
	var nodes Nodes
	err = a.readResponseToJson(resp, &nodes)
	if err != nil {
		return machineID, err
	}

	a.logger.Info(a.logTag, "Selecting machine '%s'", nodes[14].Id)

	return nodes[14].Id, nil

}

func (a CreateVM) getMacAddress(machineID string) (string, error){
	//call api for mac address
	nodeInfoUrl := fmt.Sprintf("http://%s:8080/api/common/nodes/%s", a.APIServer, machineID)
	resp, err := http.Get(nodeInfoUrl)
	if (err != nil) {
		return "", errors.New("Cannot get url")
	}
	defer resp.Body.Close()

	var nodeInfo NodeInfo
	err = a.readResponseToJson(resp, &nodeInfo)
	if err != nil {
		return "", err
	}

	return nodeInfo.Identifiers[0], nil
}

func (a CreateVM) isWorkFlowActive(machineID string) bool {
	a.logger.Info(a.logTag, "Checking workflow actively")
	workFlowUrl := fmt.Sprintf("http://%s:8080/api/common/nodes/%s/workflows/active", a.APIServer, machineID)
	resp, err := http.Get(workFlowUrl)
	defer resp.Body.Close()

	var workflow Workflow
	err = a.readResponseToJson(resp, &workflow)
	if err != nil {
		return false
	}

	if workflow.Status != nil {
		return true
	}

	return false
}

func (a CreateVM) readResponseToJson(resp *http.Response, jsonObject interface{}) error {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.New("Cannot read body")
	}


	err = json.Unmarshal(body, jsonObject)
	if err != nil {
		return errors.New("Cannot unmarshall the body")
	}

	return nil
}

type Workflow struct {
	Status *string `json:"_status"`
}

type NodeInfo struct {
	Identifiers []string `json:"identifiers"`
}

type Nodes []struct {
	Id string `json:"id"`
}