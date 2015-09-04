package action

import (
	"fmt"
	"encoding/json"
	"io/ioutil"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	httpclient "github.com/cppforlife/baremetal_cpi/utils/httpclient"
	"errors"
)

type HasVM struct {
	APIServer string
	logger boshlog.Logger
	logTag string
}

func NewHasVM(APIServer string, logger boshlog.Logger) HasVM {
	return HasVM{
		APIServer: APIServer,
		logger: logger,
		logTag: "has-vm",
	}
}

func (a HasVM) Run(vmCID VMCID) (bool, error) {
	client := httpclient.NewHTTPClient(httpclient.DefaultClient, a.logger)
	resp, err := client.Get(fmt.Sprintf("http://%s:8080/api/common/nodes/%s", a.APIServer, vmCID))

	if (err != nil) {
		//maybe better/diff error handling
		return false, errors.New("Error Getting node")
	}

	a.logger.Info(a.logTag, "The response status is '%s'", resp.Status)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, errors.New("Error getting response body")
	}
	defer resp.Body.Close()

	var node Node
	err = json.Unmarshal(body, &node)
	if err != nil {
		return false, errors.New("Unmarshalling Node Metadata")
	}

	if node.Reserved != nil && *(node.Reserved) == "true" {
		a.logger.Info(a.logTag, "The node's reserve status is '%s'", *(node.Reserved))
		return true, nil
	}

	return false, nil
}

type Node struct {
	Reserved *string `json:"reserved"`
}