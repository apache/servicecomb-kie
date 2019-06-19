package v1_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/apache/servicecomb-kie/server/resource/v1"
	"github.com/emicklei/go-restful"
	"net/http"
)

var _ = Describe("Common", func() {
	Describe("set query combination", func() {
		Context("valid param", func() {
			r, err := http.NewRequest("GET",
				"/kv?q=app:mall+service:payment&q=app:mall+service:payment+version:1.0.0",
				nil)
			It("should not return err ", func() {
				Expect(err).Should(BeNil())
			})
			c, err := ReadLabelCombinations(restful.NewRequest(r))
			It("should not return err ", func() {
				Expect(err).Should(BeNil())
			})
			It("should has 2 combinations", func() {
				Expect(len(c)).Should(Equal(2))
			})

		})
	})
})
