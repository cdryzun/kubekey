package dns

import (
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/action"
	"github.com/kubesphere/kubekey/pkg/core/module"
	"github.com/kubesphere/kubekey/pkg/core/prepare"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/kubesphere/kubekey/pkg/images"
	"github.com/kubesphere/kubekey/pkg/plugins/dns/templates"
	"path/filepath"
)

type ClusterDNSModule struct {
	common.KubeModule
}

func (c *ClusterDNSModule) Init() {
	c.Name = "ClusterDNSModule"
	c.Desc = "Deploy cluster dns"

	generateCoreDNDSvc := &module.RemoteTask{
		Name:  "GenerateCoreDNSSvc",
		Desc:  "Generate coredns service",
		Hosts: c.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&CoreDNSExist{Not: true},
		},
		Action: &action.Template{
			Template: templates.CorednsService,
			Dst:      filepath.Join(common.KubeConfigDir, templates.CorednsService.Name()),
			Data: util.Data{
				"ClusterIP": c.KubeConf.Cluster.CorednsClusterIP(),
			},
		},
		Parallel: true,
	}

	override := &module.RemoteTask{
		Name:  "OverrideCoreDNSService",
		Desc:  "Override coredns service",
		Hosts: c.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&CoreDNSExist{Not: true},
		},
		Action:   new(OverrideCoreDNS),
		Parallel: true,
	}

	generateNodeLocalDNS := &module.RemoteTask{
		Name:  "GenerateNodeLocalDNS",
		Desc:  "Generate nodelocaldns",
		Hosts: c.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(EnableNodeLocalDNS),
		},
		Action: &action.Template{
			Template: templates.NodeLocalDNSService,
			Dst:      filepath.Join(common.KubeConfigDir, templates.NodeLocalDNSService.Name()),
			Data: util.Data{
				"NodelocaldnsImage": images.GetImage(c.Runtime, c.KubeConf, "k8s-dns-node-cache").ImageName(),
			},
		},
		Parallel: true,
	}

	applyNodeLocalDNS := &module.RemoteTask{
		Name:  "DeployNodeLocalDNS",
		Desc:  "Deploy nodelocaldns",
		Hosts: c.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(EnableNodeLocalDNS),
		},
		Action:   new(DeployNodeLocalDNS),
		Parallel: true,
		Retry:    5,
	}

	generateNodeLocalDNSConfigMap := &module.RemoteTask{
		Name:  "GenerateNodeLocalDNSConfigMap",
		Desc:  "Generate nodelocaldns configmap",
		Hosts: c.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(EnableNodeLocalDNS),
			new(NodeLocalDNSConfigMapNotExist),
		},
		Action:   new(GenerateNodeLocalDNSConfigMap),
		Parallel: true,
	}

	applyNodeLocalDNSConfigMap := &module.RemoteTask{
		Name:  "ApplyNodeLocalDNSConfigMap",
		Desc:  "Apply nodelocaldns configmap",
		Hosts: c.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(EnableNodeLocalDNS),
			new(NodeLocalDNSConfigMapNotExist),
		},
		Action:   new(ApplyNodeLocalDNSConfigMap),
		Parallel: true,
		Retry:    5,
	}

	c.Tasks = []module.Task{
		generateCoreDNDSvc,
		override,
		generateNodeLocalDNS,
		applyNodeLocalDNS,
		generateNodeLocalDNSConfigMap,
		applyNodeLocalDNSConfigMap,
	}
}
