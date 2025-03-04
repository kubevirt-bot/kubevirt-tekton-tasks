package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/kubevirt/kubevirt-tekton-tasks/modules/tests/test/constants"
	"github.com/kubevirt/kubevirt-tekton-tasks/modules/tests/test/framework"
	"github.com/kubevirt/kubevirt-tekton-tasks/modules/tests/test/framework/clients"
	"github.com/kubevirt/kubevirt-tekton-tasks/modules/tests/test/framework/testoptions"
	"github.com/kubevirt/kubevirt-tekton-tasks/modules/tests/test/utils"
	. "github.com/onsi/ginkgo/v2"
	v1 "github.com/openshift/api/template/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/gomega"
)

func TestExit(t *testing.T) {
	RegisterFailHandler(Fail)
	BuildTestSuite()
	suiteConfig, reporterConfig := GinkgoConfiguration()
	reporterConfig.JUnitReport = fmt.Sprintf("../dist/junit_%d.xml", suiteConfig.ParallelProcess)
	RunSpecs(t, "E2E Tests Suite", suiteConfig, reporterConfig)
}

func BuildTestSuite() {
	BeforeSuite(func() {
		err := testoptions.InitTestOptions(framework.TestOptionsInstance)
		noErr(err)
		err = clients.InitClients(framework.ClientsInstance, framework.TestOptionsInstance)
		noErr(err)

		if framework.TestOptionsInstance.EnvScope == constants.OKDEnvScope {
			templateList, err := framework.ClientsInstance.TemplateClient.Templates("openshift").List(context.TODO(), metav1.ListOptions{
				LabelSelector: "template.kubevirt.io/type=base",
			})
			noErr(err)

			framework.TestOptionsInstance.CommonTemplatesVersion = getCommonTemplatesVersion(templateList)
		}
	})
}

func getCommonTemplatesVersion(templateList *v1.TemplateList) string {
	var commonTemplatesVersion []int
	found := false
	requiredTemplate := "fedora-server-tiny"

	for _, template := range templateList.Items {
		if strings.HasPrefix(template.Name, requiredTemplate) {
			found = true
			parts := strings.Split(template.Name, fmt.Sprintf("%v-v", requiredTemplate))
			if len(parts) == 2 {
				nextVersion, err := utils.ConvertStringSliceToInt(strings.Split(parts[1], "."))
				noErr(err)
				if utils.IsBVersionHigher(commonTemplatesVersion, nextVersion) {
					commonTemplatesVersion = nextVersion
				}
			} else {
				// no version suffix
				commonTemplatesVersion = nil
				break
			}
		}
	}

	if len(commonTemplatesVersion) == 0 {
		if found {
			return "" // no version suffix
		}
		Expect(templateList).ShouldNot(BeNil())
		Fail(fmt.Sprintf("Could not compute common templates version. Number of found templates = %v", len(templateList.Items)))
	}

	return fmt.Sprintf("-v%v", utils.JoinIntSlice(commonTemplatesVersion, "."))
}

func noErr(err error) {
	if err != nil {
		Fail(err.Error())
	}
}
