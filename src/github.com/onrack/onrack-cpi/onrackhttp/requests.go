package onrackhttp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/onrack/onrack-cpi/config"
)

func UploadFile(c config.Cpi, baseName string, r io.Reader, contentLength int64) (string, error) {
	url := fmt.Sprintf("http://%s:8080/api/common/files/%s", c.ApiServer, baseName)
	body := ioutil.NopCloser(r)
	request, err := http.NewRequest("PUT", url, body)
	if err != nil {
		log.Printf("Error building request to api server: %s", err)
		return "", err
	}
	request.ContentLength = contentLength
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Printf("Error making request to api server: %s", err)
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Unable to read response body")
		return "", err
	}

	if resp.StatusCode != 201 {
		log.Printf("Failed uploading %s with status: %s", baseName, resp.Status)
		return "", fmt.Errorf("Failed uploading %s with status: %s", baseName, resp.Status)
	}

	return string(bodyBytes), nil
}

func DeleteFile(c config.Cpi, uuid string) error {
	url := fmt.Sprintf("http://%s:8080/api/common/files/%s", c.ApiServer, uuid)
	deleteReq, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("error creating delete request %s", err)
	}

	deleteResp, err := http.DefaultClient.Do(deleteReq)
	if err != nil {
		return fmt.Errorf("error performing delete request %s", err)
	}

	if deleteResp.StatusCode != 204 {
		return fmt.Errorf("Failed deleting %s with status: %s", uuid, deleteResp.Status)
	}

	return nil
}

func GetNodes(c config.Cpi) ([]Node, error) {
	nodesURL := fmt.Sprintf("http://%s:8080/api/common/nodes", c.ApiServer)
	resp, err := http.Get(nodesURL)
	if err != nil {
		log.Printf("error fetching nodes %s", err)
		return []Node{}, fmt.Errorf("error fetching nodes %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("error getting nodes %s", err)
		return []Node{}, fmt.Errorf("Failed getting nodes with status: %s", resp.Status)
	}

	nodeBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("error reading node response body %s", err)
		return []Node{}, fmt.Errorf("error reading node response body %s", err)
	}

	var nodes []Node
	err = json.Unmarshal(nodeBytes, &nodes)
	if err != nil {
		log.Printf("error unmarshalling /common/nodes response %s", err)
		return []Node{}, fmt.Errorf("error unmarshalling /common/nodes response %s", err)
	}

	return nodes, nil
}

func GetNodeCatalog(c config.Cpi, nodeID string) (NodeCatalog, error) {
	catalogURL := fmt.Sprintf("http://%s:8080/api/common/nodes/%s/catalogs/ohai", c.ApiServer, nodeID)
	resp, err := http.Get(catalogURL)
	if err != nil {
		log.Printf("error getting catalog %s", err)
		return NodeCatalog{}, fmt.Errorf("error getting catalog %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("error getting nodes %s", err)
		return NodeCatalog{}, fmt.Errorf("Failed getting nodes with status: %s", resp.Status)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("error reading catalog body %s", err)
		return NodeCatalog{}, fmt.Errorf("error reading catalog body %s", err)
	}

	var nodeCatalog NodeCatalog
	err = json.Unmarshal(b, &nodeCatalog)
	if err != nil {
		log.Printf("error unmarshal catalog body %s", err)
		return NodeCatalog{}, fmt.Errorf("error unmarshal catalog body %s", err)
	}

	return nodeCatalog, nil
}

func PublishTask(c config.Cpi, task Task) error {
	body, err := json.Marshal(task)
	if err != nil {
		log.Println("error marshalling createVMWorkflow")
		return errors.New("error marshalling createVMWorkflow")
	}

	url := fmt.Sprintf("http://%s:8080/api/1.1/workflows/tasks", c.ApiServer)
	request, err := http.NewRequest("PUT", url, bytes.NewReader(body))
	if err != nil {
		log.Println("error building publish task request")
		return errors.New("error building publish task request")
	}
	request.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Printf("error sending PUT request to %s", c.ApiServer)
		return fmt.Errorf("error sending PUT request to %s", c.ApiServer)
	}

	if resp.StatusCode != 200 {
		msg, _ := ioutil.ReadAll(resp.Body)
		log.Printf("error publishing task: response code is %d: %s", resp.StatusCode, string(msg))
		return fmt.Errorf("Failed publishing task with status: %s, message: %s", resp.Status, string(msg))
	}

	return nil
}

func RetrieveTasks(c config.Cpi) ([]Task, error) {
	url := fmt.Sprintf("http://%s:8080/api/1.1/workflows/tasks/library", c.ApiServer)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error: %s", err)
		return []Task{}, fmt.Errorf("Error: %s", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("error reading response body: %s\n", err)
		return []Task{}, fmt.Errorf("error reading response body: %s", err)
	}

	if resp.StatusCode != 200 {
		log.Printf("error retrieving tasks: response code is %d: %s\n", resp.StatusCode, string(body))
		return []Task{}, fmt.Errorf("error retrieving tasks: response code is %d: %s", resp.StatusCode, string(body))
	}

	var tasks []Task
	err = json.Unmarshal(body, &tasks)
	if err != nil {
		log.Printf("error unmarshalling response body: %s\n", err)
		return []Task{}, fmt.Errorf("error unmarshalling response body: %s", err)
	}

	return tasks, nil
}

func PublishWorkflow(c config.Cpi, w Workflow) error {
	url := fmt.Sprintf("http://%s:8080/api/1.1/workflows", c.ApiServer)
	body, err := json.Marshal(w)
	if err != nil {
		log.Printf("error marshalling workflow: %s\n", err)
		return fmt.Errorf("error marshalling workflow: %s", err)
	}

	request, err := http.NewRequest("PUT", url, bytes.NewReader(body))
	if err != nil {
		log.Printf("error building http request: %s\n", err)
		return fmt.Errorf("error building http request: %s", err)
	}
	request.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Printf("error sending publishing workflow to %s\n", url)
		return fmt.Errorf("error sending publishing workflow to %s", url)
	}
	defer resp.Body.Close()

	msg, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("error reading response body: %s\n", err)
		return fmt.Errorf("error reading response body: %s", err)
	}

	if resp.StatusCode != 200 {
		log.Printf("error response code is %d: %s\n", resp.StatusCode, string(msg))
		return fmt.Errorf("Failed publishing workflows with status: %s", resp.Status)
	}

	publishedWorkflows, err := RetrieveWorkflows(c)
	if err != nil {
		return err
	}

	var uploadedWorkflow *Workflow
	for i := range publishedWorkflows {
		if publishedWorkflows[i].Name == w.Name {
			uploadedWorkflow = &publishedWorkflows[i]
		}
	}

	if uploadedWorkflow == nil {
		log.Println("workflow was not successfully uploaded to server")
		return errors.New("workflow was not successfully uploaded to server")
	}

	return nil
}

func RetrieveWorkflows(c config.Cpi) ([]Workflow, error) {
	url := fmt.Sprintf("http://%s:8080/api/1.1/workflows/library", c.ApiServer)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error: %s\n", err)
		return []Workflow{}, fmt.Errorf("Error: %s", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("error reading response body: %s\n", err)
		return []Workflow{}, fmt.Errorf("error reading response body: %s", err)
	}

	if resp.StatusCode != 200 {
		log.Printf("error retrieving tasks: response code is %d: %s\n", resp.StatusCode, string(body))
		return []Workflow{}, fmt.Errorf("Failed retrieving workflows with status: %s, message: %s", resp.Status, string(body))
	}

	var workflows []Workflow
	err = json.Unmarshal(body, &workflows)
	if err != nil {
		log.Printf("error unmarshalling response body: %s\n", err)
		return []Workflow{}, fmt.Errorf("error unmarshalling response body: %s", err)
	}

	return workflows, nil
}

// TODO: remove after RunWorkflow completed
func InitiateWorkflow(c config.Cpi, nodeID string, bodyStruct RunWorkflowRequestBody) (err error) {
	url := fmt.Sprintf("http://%s:8080/api/1.1/nodes/%s/workflows/", c.ApiServer, nodeID)
	body, err := json.Marshal(bodyStruct)
	if err != nil {
		log.Printf("error marshalling workflow request body")
		return
	}
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		log.Printf("error building http request to run workflow")
		return
	}
	request.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Printf("error running workflow at url %s", url)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		msg, _ := ioutil.ReadAll(resp.Body)
		log.Printf("error response code is %d: %s\n,with body: ", resp.StatusCode, string(msg), string(body))
		return fmt.Errorf("Failed running workflow at url: %s with status: %s, message: %s", url, resp.Status, string(msg))
	}
	return
}

func KillActiveWorkflowsOnVM(c config.Cpi, nodeID string) (err error) {
	url := fmt.Sprintf("http://%s:8080/api/1.1/nodes/%s/workflows/active", c.ApiServer, nodeID)
	request, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		log.Printf("error building http request to delete active workflows on url %s", url)
		return
	}
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		fmt.Printf("Error deleting active workflows on node at url: %s, msg: %s", url, err)
		return
	}

	if resp.StatusCode != 200 {
		msg, _ := ioutil.ReadAll(resp.Body)
		log.Printf("error response code is %d: %s\n,with body: ", resp.StatusCode, string(msg))
		return fmt.Errorf("Failed deleting active workflows at url: %s with status: %s, message: %s", url, resp.Status, string(msg))
	}
	return
}

func RunWorkflow(c config.Cpi, nodeID string, bodyStruct RunWorkflowRequestBody) (err error) {
	return
}

// func pollActiveWorkflows()
