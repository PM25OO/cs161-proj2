package client

// CS 161 项目 2

// 只允许使用以下导入！任何额外导入
// 都可能导致自动评分器失败！
// - bytes
// - encoding/hex
// - encoding/json
// - errors
// - fmt
// - github.com/cs161-staff/project2-userlib
// - github.com/google/uuid
// - strconv
// - strings

import (
	"encoding/json"

	userlib "github.com/cs161-staff/project2-userlib"
	"github.com/google/uuid"

	// hex.EncodeToString(...) 可用于将 []byte 转换为 string

	// 用于字符串处理
	"strings"

	// 用于格式化字符串（例如 `fmt.Sprintf`）。
	"fmt"

	// 用于通过 errors.New("...") 创建并返回新的错误信息
	"errors"

	// 可选。
	_ "strconv"
)

// 这个函数有两个用途：展示一些有用的基础能力，
// 以及抑制“导入未使用”的告警。它可以
// 被安全删除！
func someUsefulThings() {

	// 创建一个随机 UUID。
	randomUUID := uuid.New()

	// 以字符串形式打印 UUID。%v 会按默认格式打印值。
	// 参见 https://pkg.go.dev/fmt#hdr-Printing 了解 Golang 的格式化标志。
	userlib.DebugMsg("Random UUID: %v", randomUUID.String())

	// 根据一段字节序列确定性地创建 UUID。
	hash := userlib.Hash([]byte("user-structs/alice"))
	deterministicUUID, err := uuid.FromBytes(hash[:16])
	if err != nil {
		// 正常情况下这里会 `return err`。但这个函数没有返回值，
		// 所以这里只能 panic 终止执行。一定、一定、一定要检查错误！到这个
		// 项目结束时，你的代码里应该有大量 "if err != nil { return err }"。
		// 在你自己的实现中，通常应尽量避免使用 panic。
		panic(errors.New("An error occurred while generating a UUID: " + err.Error()))
	}
	userlib.DebugMsg("Deterministic UUID: %v", deterministicUUID.String())

	// 声明一个 Course 结构体类型，创建实例并将其序列化为 JSON。
	type Course struct {
		name      string
		professor []byte
	}

	course := Course{"CS 161", []byte("Nicholas Weaver")}
	courseBytes, err := json.Marshal(course)
	if err != nil {
		panic(err)
	}

	userlib.DebugMsg("Struct: %v", course)
	userlib.DebugMsg("JSON Data: %v", courseBytes)

	// 生成一对随机公私钥。
	// "_" 表示这里不检查错误分支。
	var pk userlib.PKEEncKey
	var sk userlib.PKEDecKey
	pk, sk, _ = userlib.PKEKeyGen()
	userlib.DebugMsg("PKE Key Pair: (%v, %v)", pk, sk)

	// 这里演示如何使用 HBKDF 从输入密钥派生一个新密钥。
	// 小提示：能生成新密钥的地方尽量都生成新密钥！动态生成通常比
	// 反复思考“密钥复用攻击”的所有可能路径更容易，也更安全。通常也更容易
	// 保存一个主密钥，再从它派生多个用途密钥，而不是
	originalKey := userlib.RandomBytes(16)
	derivedKey, err := userlib.HashKDF(originalKey, []byte("mac-key"))
	if err != nil {
		panic(err)
	}
	userlib.DebugMsg("Original Key: %v", originalKey)
	userlib.DebugMsg("Derived Key: %v", derivedKey)

	// 关于 string 和 []byte 转换的几个小提示：
	// string 转 []byte：使用 []byte("some-string-here")
	// []byte 转 string（用于调试）：使用 fmt.Sprintf("hello world: %s", some_byte_arr)。
	// []byte 转 string（用于哈希表键）：使用 hex.EncodeToString(some_byte_arr)。
	// 如果频繁在 []byte 和 string 间转换，直接对数据做序列化/反序列化通常更方便。
	//
	// 进一步阅读：https://go.dev/blog/strings

	// 这里是一个字符串插值示例！
	_ = fmt.Sprintf("%s_%d", "file", 1)
}

// 这是 User 结构体的类型定义。
// Go 的 struct 类似 Python 或 Java 中的类——可以有属性
// （例如 Username）和方法（例如下方的 StoreFile）。
type User struct {
	Username string

	pkePblicKey   userlib.PKEEncKey // 加密公钥
	PKEPrivateKey userlib.PKEDecKey // 解密私钥

	dsVerifyKey userlib.DSVerifyKey // 认证公钥
	DSSignKey   userlib.DSSignKey   // 签名私钥

	RootKey  []byte
	FileKeys map[string]userlib.UUID

	OwnedFiles  []string
	SharedFiles map[string]string

	// 你可以在这里添加其他属性！但要注意，如果希望这些属性在结构体与 JSON
	// 互相序列化时被包含，字段名首字母必须大写。
	// 相反，如果某个字段希望仅在结构体方法内部可访问，
	// 但不希望被写入 datastore 中保存的序列化结果，
	// 那就可以使用“私有”字段（例如以小写字母开头的字段名）。
}

type File struct {
	Filename string
	Owner    string

	Content []byte
}

func InitUser(username string, password string) (userdataptr *User, err error) {
	_, ok := userlib.KeystoreGet(username)
	if ok {
		return nil, errors.New("user already exists")
	}

	var userdata User
	userdata.Username = username
	userdata.pkePblicKey, userdata.PKEPrivateKey, _ = userlib.PKEKeyGen()
	userdata.DSSignKey, userdata.dsVerifyKey, _ = userlib.DSKeyGen()
	userdata.FileKeys = make(map[string]userlib.UUID)
	userdata.SharedFiles = make(map[string]string)

	// 存储公钥
	userlib.KeystoreSet(username, userdata.pkePblicKey)
	userlib.KeystoreSet(username+"_ds", userdata.dsVerifyKey)

	// 根据password加盐username派生rootKey
	userdata.RootKey = userlib.Argon2Key([]byte(password), []byte(username), 16)

	// 序列化User并加密存储
	userBytes, err := json.Marshal(userdata)
	if err != nil {
		return nil, errors.New("failed to marshal user data")
	}
	cyphertext := userlib.SymEnc(userdata.RootKey, userlib.RandomBytes(16), userBytes)
	uuid, err := uuid.FromBytes([]byte(username))
	if err != nil {
		return nil, errors.New("failed to generate UUID from username")
	}
	userlib.DatastoreSet(uuid, cyphertext)

	return &userdata, nil
}

func GetUser(username string, password string) (userdataptr *User, err error) {
	_, ok := userlib.KeystoreGet(username)
	if !ok {
		return nil, errors.New("user does not exist")
	}

	// 根据password加盐username派生rootKey
	key := userlib.Argon2Key([]byte(password), []byte(username), 16)

	uuid, err := uuid.FromBytes([]byte(username))
	if err != nil {
		return nil, errors.New("failed to generate UUID from username")
	}
	bytes, ok := userlib.DatastoreGet(uuid)
	if !ok {
		return nil, errors.New("failed to retrieve user data")
	}

	// 解密并反序列化User
	plaintext := userlib.SymDec(key, bytes)
	var userdata User
	err = json.Unmarshal(plaintext, &userdata)
	if err != nil {
		return nil, errors.New("failed to unmarshal user data")
	}

	// 验证数据完整性
	if userdata.Username != username {
		return nil, errors.New("data integrity check failed")
	}

	userdataptr = &userdata
	return userdataptr, nil
}

func (userdata *User) StoreFile(filename string, content []byte) (err error) {
	var userfile File
	userfile.Filename = filename
	userfile.Owner = userdata.Username
	userfile.Content = content

	// 序列化并加密文件数据
	contentBytes, err := json.Marshal(userfile)
	if err != nil {
		return err
	}
	ciphertext := userlib.SymEnc(userdata.RootKey, userlib.RandomBytes(16), contentBytes)
	storageKey, err := uuid.FromBytes(userlib.Hash([]byte(filename + userdata.Username))[:16])
	if err != nil {
		return err
	}
	userlib.DatastoreSet(storageKey, ciphertext)
	return nil
}

func (userdata *User) AppendToFile(filename string, content []byte) error {
	return nil
}

func (userdata *User) LoadFile(filename string) (content []byte, err error) {
	storageKey, err := uuid.FromBytes(userlib.Hash([]byte(filename + userdata.Username))[:16])
	if err != nil {
		return nil, err
	}
	response, ok := userlib.DatastoreGet(storageKey)
	if !ok {
		return nil, errors.New(strings.ToTitle("file not found"))
	}

	contentBytes := userlib.SymDec(userdata.RootKey, response)
	var userfile File
	err = json.Unmarshal([]byte(contentBytes), &userfile)
	if err != nil {
		return nil, err
	}
	return userfile.Content, err
}

func (userdata *User) CreateInvitation(filename string, recipientUsername string) (
	invitationPtr uuid.UUID, err error) {
	return
}

func (userdata *User) AcceptInvitation(senderUsername string, invitationPtr uuid.UUID, filename string) error {
	return nil
}

func (userdata *User) RevokeAccess(filename string, recipientUsername string) error {
	return nil
}
