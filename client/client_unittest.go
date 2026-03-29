package client

///////////////////////////////////////////////////
//                                               //
// 本文件中的内容不会被评分！！！               //
//                                               //
///////////////////////////////////////////////////

// 在这个单元测试文件中，你可以为自己的实现编写白盒测试。
// 这和 client_test.go 里的黑盒集成测试不同，
// 因为这里可以使用你实现中的内部细节。

// 例如，在这个单元测试文件里，你可以访问自己定义的结构体字段和辅助方法；
// 但在集成测试（client_test.go）中，你只能访问
// 所有实现都共有的 8 个函数（StoreFile、LoadFile 等）。

// 在这个单元测试文件中，你可以直接写 InitUser；而在
// 集成测试（client_test.go）中则要写 client.InitUser。也就是说，这里不再需要前缀 "client."。

import (
	"testing"

	userlib "github.com/cs161-staff/project2-userlib"

	_ "encoding/hex"

	_ "errors"

	. "github.com/onsi/ginkgo/v2"

	. "github.com/onsi/gomega"

	_ "strconv"

	_ "strings"
)

func TestSetupAndExecution(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Client Unit Tests")
}

var _ = Describe("Client Unit Tests", func() {

	BeforeEach(func() {
		userlib.DatastoreClear()
		userlib.KeystoreClear()
	})

	Describe("Unit Tests", func() {
		Specify("Basic Test: Check that the Username field is set for a new user", func() {
			userlib.DebugMsg("Initializing user Alice.")
			// 注意：在集成测试（client_test.go）里这里应写
			// client.InitUser；但在这里（client_unittests.go）可直接写 InitUser。
			alice, err := InitUser("alice", "password")
			Expect(err).To(BeNil())

			// 注意：这里可以访问 User 结构体的 Username 字段。
			// 但在集成测试（client_test.go）中不能访问结构体字段，
			// 因为并不是所有实现都一定有 username 这个字段。
			Expect(alice.Username).To(Equal("alice"))
		})
	})
})
