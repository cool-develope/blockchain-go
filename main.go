package main

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type Block struct {
	Index     int
	Timestamp string
	PII       string
	Hash      string
	PrevHash  string
}

var Blockchain []Block

func calculateHash(block Block) string {
	record := string(block.Index) + block.Timestamp + block.PII + block.PrevHash
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
	newBlock.Hash = calculateHash(newBlock)

	return newBlock, nil
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

func handleGetBlockchain(c *gin.Context) {
	c.JSON(200, Blockchain)
}

var router *gin.Engine

func run() error {

	// Set the router as the default one provided by Gin
	router = gin.Default()

	// Define the route for the index page and display the index.html template
	// To start with, we'll use an inline route handler. Later on, we'll create
	// standalone functions that will be used as route handlers.
	router.GET("/", handleGetBlockchain)
	router.POST("/", handleWriteBlock)

	// Start serving the application
	router.Run()

	return nil
}

type Message struct {
	PII string
}

func handleWriteBlock(c *gin.Context) {
	m := new(Message)

	if err := c.Bind(m); err != nil {
		c.AbortWithStatusJSON(400, gin.H{
			"error": "Message Parsing Error",
		})
		return
	}

	newBlock, err := generateBlock(Blockchain[len(Blockchain)-1], m.PII)
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

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		t := time.Now()
		genesisBlock := Block{0, t.String(), "", "", ""}
		spew.Dump(genesisBlock)
		Blockchain = append(Blockchain, genesisBlock)
	}()
	log.Fatal(run())

}
