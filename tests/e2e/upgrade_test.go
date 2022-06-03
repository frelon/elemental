/*
Copyright © 2022 SUSE LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package e2e_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/ele-testhelpers/kubectl"
	"github.com/rancher-sandbox/ele-testhelpers/tools"
)

var _ = Describe("E2E - Upgrading node", Label("upgrade"), func() {
	It("Upgrade node", func() {
		By("Checking if VM name is set", func() {
			Expect(vmName).To(Not(BeEmpty()))
		})

		By("Checking if upgrade type is set", func() {
			Expect(upgradeType).To(Not(BeEmpty()))
		})

		By("Triggering Upgrade in Rancher with "+upgradeType, func() {
			upgradeOsYaml := "../assets/upgrade.yaml"

			if upgradeType == "upgradechannel" {
				upgradeChannelFile, err := tools.GetFiles("../..", "rancheros-*.upgradechannel-*.yaml")
				Expect(err).To(Not(HaveOccurred()))
				Expect(upgradeChannelFile).To(Not(BeEmpty()))

				err = kubectl.Apply(clusterNS, upgradeChannelFile[0])
				Expect(err).To(Not(HaveOccurred()))

				// Get ManagedOSVersionChannel name
				name, err := kubectl.Run("get", "ManagedOSVersionChannel",
					"--namespace", clusterNS,
					"-o", "jsonpath={.items[0].metadata.name}")
				Expect(err).To(Not(HaveOccurred()))

				err = tools.Sed("%OS_IMAGE%", "managedOSVersionName: "+name, upgradeOsYaml)
				Expect(err).To(Not(HaveOccurred()))
			}

			if upgradeType == "osimage" {
				err := tools.Sed("%OS_IMAGE%", "osImage: "+osImage, upgradeOsYaml)
				Expect(err).To(Not(HaveOccurred()))
			}

			err := tools.Sed("%CLUSTER_NAME%", clusterName, upgradeOsYaml)
			Expect(err).To(Not(HaveOccurred()))

			err = kubectl.Apply(clusterNS, upgradeOsYaml)
			Expect(err).To(Not(HaveOccurred()))
		})

		By("Checking VM upgrade", func() {
			hostData, err := tools.GetHostNetConfig(".*name='"+vmName+"'.*", netDefaultFileName)
			Expect(err).To(Not(HaveOccurred()))

			client := &tools.Client{
				Host:     string(hostData.IP) + ":22",
				Username: userName,
				Password: userPassword,
			}

			version := strings.Split(osImage, ":")[1]
			Eventually(func() string {
				// Use grep here in case of comment in the file!
				out, _ := client.RunSSH("eval $(grep -v ^# /usr/lib/os-release) && echo ${VERSION_ID}")
				out = strings.Trim(out, "\n")
				return out
			}, "20m", "30s").Should(Equal(version))
		})

		By("Cleaning upgrade orders", func() {
			err := kubectl.DeleteResource(clusterNS, "ManagedOSImage", "default-os-image")
			Expect(err).To(Not(HaveOccurred()))
		})
	})
})
