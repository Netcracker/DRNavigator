/*
Copyright 2025.

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

package v3

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	qubershiporgv3 "github.com/netcracker/drnavigator/site-manager/api/v3"
)

var _ = Describe("SiteManager Webhook", func() {
	Context("When creating or updating SiteManager under Validating Webhook", func() {
		It("Should deny creation if service name is duplicated", func() {
			By("simulating an invalid creation scenario")
			obj1 := &qubershiporgv3.SiteManager{
				ObjectMeta: v1.ObjectMeta{
					Name:      "service-1",
					Namespace: "default",
				},
			}
			Expect(k8sClient.Create(context.Background(), obj1)).To(Succeed())

			obj2 := &qubershiporgv3.SiteManager{
				ObjectMeta: v1.ObjectMeta{
					Name:      "service-2",
					Namespace: "default",
				},
				Spec: qubershiporgv3.SiteManagerSpec{
					SiteManager: qubershiporgv3.SiteManagerOptions{
						Alias: ptr.To("service-1.default"),
					},
				},
			}
			err := k8sClient.Create(context.Background(), obj2)
			Expect(err).To(Not(BeNil()))
			Expect(err.Error()).To(ContainSubstring("this name is used for another service"))
		})
	})
})
