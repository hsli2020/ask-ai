package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io"
	"io/ioutil"
	"log"
	"time"
)

// Contact 联系人结构体
type Contact struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	Address string `json:"address"`
}

// Contacts 联系人数组
type Contacts []Contact

// FileHeader 文件头结构
type FileHeader struct {
	MagicNumber [4]byte  // 魔数，标识文件类型
	Version     uint32   // 文件格式版本
	Timestamp   int64    // 文件创建时间戳
	DataLength  uint32   // 数据部分长度
	Checksum    uint32   // 数据部分的校验码
	Reserved    [16]byte // 保留字段
}

// 加密密钥
var encryptionKey = []byte("12345678901234567890123456789012") // 32字节，AES-256

// 计算数据校验码
func calculateChecksum(data []byte) uint32 {
	return crc32.ChecksumIEEE(data)
}

// encrypt 加密数据
func encrypt(plaintext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	ciphertext := aesGCM.Seal(nil, nonce, plaintext, nil)
	result := append(nonce, ciphertext...)
	return result, nil
}

// decrypt 解密数据
func decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < 12 {
		return nil, fmt.Errorf("密文太短")
	}

	nonce := ciphertext[:12]
	encryptedData := ciphertext[12:]

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plaintext, err := aesGCM.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// SaveContactsToFile 将联系人加密后保存到文件（包含文件头）
func SaveContactsToFile(contacts Contacts, filename string) error {
	// 将联系人转换为JSON格式
	jsonData, err := json.Marshal(contacts)
	if err != nil {
		return fmt.Errorf("序列化联系人失败: %v", err)
	}

	// 加密数据
	encryptedData, err := encrypt(jsonData, encryptionKey)
	if err != nil {
		return fmt.Errorf("加密数据失败: %v", err)
	}

	// 计算数据部分的校验码
	checksum := calculateChecksum(encryptedData)

	// 创建文件头
	header := FileHeader{
		MagicNumber: [4]byte{'C', 'T', 'C', 'T'}, // Contact
		Version:     1,
		Timestamp:   time.Now().Unix(),
		DataLength:  uint32(len(encryptedData)),
		Checksum:    checksum,
	}

	// 将文件头序列化为字节
	headerBytes := make([]byte, 32) // 文件头固定32字节
	copy(headerBytes[0:4], header.MagicNumber[:])
	binary.LittleEndian.PutUint32(headerBytes[4:8], header.Version)
	binary.LittleEndian.PutUint64(headerBytes[8:16], uint64(header.Timestamp))
	binary.LittleEndian.PutUint32(headerBytes[16:20], header.DataLength)
	binary.LittleEndian.PutUint32(headerBytes[20:24], header.Checksum)
	copy(headerBytes[24:32], header.Reserved[:])

	// 合并文件头和数据
	fileData := append(headerBytes, encryptedData...)

	// 保存到文件
	err = ioutil.WriteFile(filename, fileData, 0644)
	if err != nil {
		return fmt.Errorf("保存文件失败: %v", err)
	}

	fmt.Printf("成功保存 %d 个联系人到文件 %s\n", len(contacts), filename)
	fmt.Printf("文件信息 - 版本: %d, 创建时间: %s, 数据长度: %d 字节, 校验码: 0x%08X\n",
		header.Version,
		time.Unix(header.Timestamp, 0).Format("2006-01-02 15:04:05"),
		header.DataLength,
		header.Checksum)

	return nil
}

// LoadContactsFromFile 从加密文件中读取联系人（包含文件头校验）
func LoadContactsFromFile(filename string) (Contacts, error) {
	// 读取文件
	fileData, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %v", err)
	}

	// 检查文件最小长度（文件头32字节 + 数据）
	if len(fileData) < 32 {
		return nil, fmt.Errorf("文件太短，可能已损坏")
	}

	// 解析文件头
	var header FileHeader
	copy(header.MagicNumber[:], fileData[0:4])
	header.Version = binary.LittleEndian.Uint32(fileData[4:8])
	header.Timestamp = int64(binary.LittleEndian.Uint64(fileData[8:16]))
	header.DataLength = binary.LittleEndian.Uint32(fileData[16:20])
	header.Checksum = binary.LittleEndian.Uint32(fileData[20:24])
	copy(header.Reserved[:], fileData[24:32])

	// 验证魔数
	expectedMagic := [4]byte{'C', 'T', 'C', 'T'}
	if header.MagicNumber != expectedMagic {
		return nil, fmt.Errorf("文件格式不正确，魔数不匹配")
	}

	// 验证数据长度
	if len(fileData) != 32+int(header.DataLength) {
		return nil, fmt.Errorf("文件长度不匹配，可能已损坏")
	}

	// 提取数据部分
	dataBytes := fileData[32:]

	// 验证校验码
	calculatedChecksum := calculateChecksum(dataBytes)
	if calculatedChecksum != header.Checksum {
		return nil, fmt.Errorf("校验码不匹配，文件可能已损坏或被篡改 (期望: 0x%08X, 实际: 0x%08X)",
			header.Checksum, calculatedChecksum)
	}

	fmt.Printf("文件头验证通过 - 版本: %d, 创建时间: %s, 数据长度: %d 字节\n",
		header.Version,
		time.Unix(header.Timestamp, 0).Format("2006-01-02 15:04:05"),
		header.DataLength)

	// 解密数据
	jsonData, err := decrypt(dataBytes, encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("解密数据失败: %v", err)
	}

	// 解析JSON数据
	var contacts Contacts
	err = json.Unmarshal(jsonData, &contacts)
	if err != nil {
		return nil, fmt.Errorf("解析JSON数据失败: %v", err)
	}

	fmt.Printf("成功从文件 %s 读取 %d 个联系人\n", filename, len(contacts))
	return contacts, nil
}

// PrintContacts 打印联系人信息
func PrintContacts(contacts Contacts) {
	fmt.Println("\n联系人列表:")
	fmt.Println("========================================")
	for i, contact := range contacts {
		fmt.Printf("联系人 %d:\n", i+1)
		fmt.Printf("  姓名: %s\n", contact.Name)
		fmt.Printf("  电话: %s\n", contact.Phone)
		fmt.Printf("  邮箱: %s\n", contact.Email)
		fmt.Printf("  地址: %s\n", contact.Address)
		fmt.Println("  ----------------------------------------")
	}
}

// 模拟文件损坏的测试函数
func testFileCorruption(filename string) {
	fmt.Println("\n=== 测试文件损坏检测 ===")

	// 读取原文件
	fileData, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Printf("读取文件失败: %v", err)
		return
	}

	// 修改数据部分的一个字节来模拟损坏
	if len(fileData) > 40 {
		fileData[40] = fileData[40] ^ 0xFF // 翻转一个字节
	}

	// 保存损坏的文件
	corruptedFilename := "contacts_corrupted.dat"
	err = ioutil.WriteFile(corruptedFilename, fileData, 0644)
	if err != nil {
		log.Printf("保存损坏文件失败: %v", err)
		return
	}

	// 尝试读取损坏的文件
	_, err = LoadContactsFromFile(corruptedFilename)
	if err != nil {
		fmt.Printf("✅ 成功检测到文件损坏: %v\n", err)
	} else {
		fmt.Println("❌ 未能检测到文件损坏")
	}
}

func main() {
	// 创建示例联系人数据
	contacts := Contacts{
		{
			Name:    "张三",
			Phone:   "13800138001",
			Email:   "zhangsan@example.com",
			Address: "北京市朝阳区某某街道1号",
		},
		{
			Name:    "李四",
			Phone:   "13900139002",
			Email:   "lisi@example.com",
			Address: "上海市浦东新区某某路2号",
		},
		{
			Name:    "王五",
			Phone:   "13700137003",
			Email:   "wangwu@example.com",
			Address: "广州市天河区某某大道3号",
		},
		{
			Name:    "赵六",
			Phone:   "13600136004",
			Email:   "zhaoliu@example.com",
			Address: "深圳市南山区某某科技园4号",
		},
	}

	// 打印原始联系人信息
	fmt.Println("原始联系人数据:")
	PrintContacts(contacts)

	// 保存联系人到加密文件（包含文件头）
	filename := "contacts.dat"
	err := SaveContactsToFile(contacts, filename)
	if err != nil {
		log.Fatalf("保存联系人失败: %v", err)
	}

	// 从加密文件读取联系人（包含文件头校验）
	loadedContacts, err := LoadContactsFromFile(filename)
	if err != nil {
		log.Fatalf("读取联系人失败: %v", err)
	}

	// 打印读取的联系人信息
	fmt.Println("\n从文件读取的联系人数据:")
	PrintContacts(loadedContacts)

	// 验证数据一致性
	if len(contacts) == len(loadedContacts) {
		fmt.Println("\n✅ 数据保存和读取成功！")
	} else {
		fmt.Println("\n❌ 数据不一致！")
	}

	// 测试文件损坏检测
	testFileCorruption(filename)

	// 显示文件的SHA256哈希值（用于验证文件完整性）
	fileData, _ := ioutil.ReadFile(filename)
	if len(fileData) > 0 {
		hash := sha256.Sum256(fileData)
		fmt.Printf("\n文件SHA256哈希: %x\n", hash)
	}
}
