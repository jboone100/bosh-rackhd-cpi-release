package workflows

import "github.com/rackhd/rackhd-cpi/rackhdapi"

var provisionNodeTemplate = []byte(`{
  "friendlyName": "Provision Node",
  "implementsTask": "Task.Base.Linux.Commands",
  "injectableName": "Task.BOSH.Provision.Node",
  "options": {
    "agentSettingsFile": null,
    "agentSettingsMd5Uri": "{{ api.files }}/md5/{{ options.agentSettingsFile }}/latest",
    "agentSettingsPath": null,
    "agentSettingsUri": "{{ api.files }}/{{ options.agentSettingsFile }}/latest",
    "commands": [
      "curl --retry 3 {{ options.stemcellUri }} -o {{ options.downloadDir }}/{{ options.stemcellFile }}",
      "curl --retry 3 {{ options.agentSettingsUri }} -o {{ options.downloadDir }}/{{ options.agentSettingsFile }}",
      "curl {{ options.stemcellFileMd5Uri }} | tr -d '\"' > /opt/downloads/stemcellFileExpectedMd5",
      "curl {{ options.agentSettingsMd5Uri }} | tr -d '\"' > /opt/downloads/agentSettingsExpectedMd5",
      "md5sum {{ options.downloadDir }}/{{ options.stemcellFile }} | cut -d' ' -f1 > /opt/downloads/stemcellFileCalculatedMd5",
      "md5sum {{ options.downloadDir }}/{{ options.agentSettingsFile }} | cut -d' ' -f1 > /opt/downloads/agentSettingsCalculatedMd5",
      "test $(cat /opt/downloads/stemcellFileCalculatedMd5) = $(cat /opt/downloads/stemcellFileExpectedMd5)",
      "test $(cat /opt/downloads/agentSettingsCalculatedMd5) = $(cat /opt/downloads/agentSettingsExpectedMd5)",
      "sudo umount {{ options.device }} || true",
      "sudo tar --to-stdout -xvf {{ options.downloadDir }}/{{ options.stemcellFile }} | sudo dd of={{ options.device }}",
      "sudo sfdisk -R {{ options.device }}",
      "sudo mount {{ options.device }}1 /mnt",
      "sudo cp {{ options.downloadDir }}/{{ options.agentSettingsFile }} /mnt/{{ options.agentSettingsPath }}",
      "sudo sync"
    ],
    "device": "/dev/sda",
    "downloadDir": "/opt/downloads",
    "stemcellFile": null,
    "stemcellFileMd5Uri": "{{ api.files }}/md5/{{ options.stemcellFile }}/latest",
    "stemcellUri": "{{ api.files }}/{{ options.stemcellFile }}/latest"
  },
  "properties": {}
}`)

type provisionNodeOptions struct {
	AgentSettingsFile   *string  `json:"agentSettingsFile"`
	AgentSettingsMd5Uri string   `json:"agentSettingsMd5Uri"`
	AgentSettingsPath   *string  `json:"agentSettingsPath"`
	AgentSettingsURI    string   `json:"agentSettingsUri"`
	Commands            []string `json:"commands"`
	Device              string   `json:"device"`
	DownloadDir         string   `json:"downloadDir"`
	StemcellFileMd5Uri  string   `json:"stemcellFileMd5Uri"`
	StemcellFile        *string  `json:"stemcellFile"`
	StemcellURI         string   `json:"stemcellUri"`
}

type provisionNodeTask struct {
	*rackhdapi.TaskStub
	*rackhdapi.PropertyContainer
	*provisionNodeOptionsContainer
}

type provisionNodeOptionsContainer struct {
	Options provisionNodeOptions `json:"options"`
}