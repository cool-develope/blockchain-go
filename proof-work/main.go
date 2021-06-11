package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

const difficulty = 1

type Block struct {
	Index      int
	Timestamp  string
	PII        string
	Hash       string
	PrevHash   string
	Difficulty int
	Nonce      string
}

var Blockchain []Block

type Message struct {
	PII string
}

var mutex = &sync.Mutex{}
var router *gin.Engine

func run() error {
	router = gin.Default()

	router.GET("/", handleGetBlockchain)
	router.POST("/", handleWriteBlock)

	router.Run()

	return nil
}

func handleGetBlockchain(c *gin.Context) {
	c.JSON(200, Blockchain)
}

func handleWriteBlock(c *gin.Context) {
	m := new(Message)

	if err := c.Bind(m); err != nil {
		c.AbortWithStatusJSON(400, gin.H{
			"error": "Message Parsing Error",
		})
		return
	}

	mutex.Lock()
	newBlock, err := generateBlock(Blockchain[len(Blockchain)-1], m.PII)
	mutex.Unlock()

	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{
			"error": "New Block Create Error",
		})
		return
	}

	if isBlockValid(newBlock, Blockchain[len(Blockchain)-1]) {
		newBlockchain := append(Blockchain, newBlock)
		replaceChain(newBlockchain)
		spew.Dump(Blockchain)
	}

	c.JSON(200, newBlock)
}

func isBlockValid(newBlock, oldBlock Block) bool {
	if oldBlock.Index+1 != newBlock.Index {
		log.Fatal("Index Matching Error")
		return false
	}

	if oldBlock.Hash != newBlock.PrevHash {
		log.Fatal("PrevHash Matching Error")
		return false
	}

	if calculateHash(newBlock) != newBlock.Hash {
		log.Fatal("Hash Calculation Matching Error")
		return false
	}

	return true
}

func replaceChain(newBlocks []Block) {
	if len(newBlocks) > len(Blockchain) {
		Blockchain = newBlocks
	}
}

func calculateHash(block Block) string {
	record := string(block.Index) + block.Timestamp + block.PII + block.PrevHash + block.Nonce
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

func generateBlock(oldBlock Block, PII string) (Block, error) {
	var newBlock Block

	t := time.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.PII = PII
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Difficulty = difficulty

	for i := 0; ; i++ {
		hex := fmt.Sprintf("%x", i)
		newBlock.Nonce = hex
		if !isHashValid(calculateHash(newBlock), newBlock.Difficulty) {
			fmt.Println(calculateHash(newBlock), " do more work!")
			time.Sleep(time.Second)
			continue
		} else {
			fmt.Println(calculateHash(newBlock), " work done!")
			newBlock.Hash = calculateHash(newBlock)
			break
		}
	}

	return newBlock, nil
}

func isHashValid(hash string, difficulty int) bool {
	prefix := strings.Repeat("0", difficulty)
	return strings.HasPrefix(hash, prefix)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		t := time.Now()
		genesisBlock := Block{}
		genesisBlock = Block{0, t.String(), "", calculateHash(genesisBlock), "", difficulty, ""}
		spew.Dump(genesisBlock)

		mutex.Lock()
		Blockchain = append(Blockchain, genesisBlock)
		mutex.Unlock()
	}()
	log.Fatal(run())

}
