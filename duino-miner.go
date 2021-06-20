package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

var (
	username = "12345"
)

const (
	IP   = "51.15.127.80"
	Port = "2811"

	Algorithm     = " (DUCO-S1) "
	MinerVersion  = "1.0.0"
	RigIdentifier = "12345-duino-miner"
)

type IMasterServerClient interface {
	InitServerConnection()
	CheckServerVersion() string
	CheckServerMessage() string
	//CheckServerJobXXHASH() []string
	CheckServerJob() []string
	CheckSolvedHash(string, int64)
}

type MasterServerClient struct {
	Client IMasterServerClient
}

type MasterServer struct {
	Sock net.Conn
}

func (ms *MasterServer) InitServerConnection() {
	if ms.Sock != nil {
		ms.Sock.Close()
	}
	d := net.Dialer{Timeout: 30 * time.Second}
	conn, err := d.Dial("tcp", IP+":"+Port)
	if err != nil {
		fmt.Println("[ConnectToServer-Err]", err)
		return
	}
	ms.Sock = conn
}

func (ms MasterServer) CheckServerVersion() string {
	readbuf := make([]byte, 3)
	_, err := ms.Sock.Read(readbuf)
	if err != nil {
		fmt.Println("[CheckServerVersion-Err]", err)
		return ""
	}

	return string(readbuf)
}

func (ms MasterServer) CheckServerMessage() string {
	data := "MOTD"
	ms.Sock.Write([]byte(data))

	readbuf := make([]byte, 1024)
	_, err := ms.Sock.Read(readbuf)
	if err != nil {
		fmt.Println("[CheckServerMessage-Err]", err)
		return ""
	}

	return string(readbuf)
}

func (ms MasterServer) CheckSolvedHash(hash string, hashrate int64) {
	hashrateStr := strconv.FormatInt(hashrate, 10)
	data := hash + "," + hashrateStr + "," + "github@Lutwidse duino-miner" + Algorithm + "v" + MinerVersion + "," + RigIdentifier
	ms.Sock.Write([]byte(data))

	readbuf := make([]byte, 64)
	_, err := ms.Sock.Read(readbuf)
	if err != nil {
		fmt.Println("[CheckSolvedHash-Err]", err)
	}
	readbuf = bytes.Trim(readbuf, "\x00")
	readbuf = bytes.Trim(readbuf, "\x0a") // Please refer to comment of CheckServerJob.
	feedback := string(readbuf)           // Its Weird but sometime response contains lastBlockHash and expectedHash.

	switch feedback {
	case "GOOD":
		fmt.Println("[GOOD]")
	case "BLOCK":
		fmt.Println("[BLOCK]")
	case "BAD":
		fmt.Println("[BAD]")
	default:
		fmt.Printf("[Feedback Error] %s\n", feedback)
	}
}

/*
func (ms MasterServer) CheckServerJobXXHASH() []string {
	data := "JOBXX" + "," + username + "," + "NET"
	ms.Sock.Write([]byte(data))

	readbuf := make([]byte, 128)
	_, err := ms.Sock.Read(readbuf)
	if err != nil {
		fmt.Println("[Sock-ERR]", err)
	}

	return strings.Split(string(readbuf), ",")
}
*/

func (ms MasterServer) CheckServerJob() []string {
	data := "JOB" + "," + username + "," + "NET"
	ms.Sock.Write([]byte(data))

	readbuf := make([]byte, 128)
	_, err := ms.Sock.Read(readbuf)
	if err != nil {
		fmt.Println("[CheckServerJob-Err]", err)
		return []string{}
	}

	readbuf = bytes.Trim(readbuf, "\x00")
	readbuf = bytes.Trim(readbuf, "\x0a") // This should be 88 bytes without new line feed. Im not sure about this weird response.
	job := string(readbuf)

	return strings.Split(job, ",")
}

func ducos1(lastBlockHash string, expectedHash string, difficulty int, efficiency int) (string, int64) {
	start := time.Now()
	for ducos1res := 0; ducos1res <= difficulty*100; ducos1res++ {
		h := sha1.New()
		h.Write([]byte(lastBlockHash + strconv.Itoa(ducos1res)))
		s := h.Sum(nil)
		ducos1 := hex.EncodeToString(s[:])

		if ducos1 == expectedHash {
			end := time.Since(start).Milliseconds()
			if end >= 800 {
				fmt.Printf("[Hashrate] %d %s\n", end/1000, "MH/s")
			} else {
				fmt.Printf("[Hashrate] %d %s\n", end, "KH/s")
			}
			return ducos1, end
		}
	}
	return "", -1
}

/*
func ducos1xxh(lastBlockHash string, expectedHash string, difficulty int, efficiency int) uint64 {
	for ducos1xxres := 0; ducos1xxres <= (100*difficulty); ducos1xxres++ {
		if ducos1xxres%1000000 == 0 && float32(100-efficiency*100) < 100 {
			time.Sleep(time.Duration(float32(efficiency)))
		}

		d := xxhash.New()
		d.WriteString(string(lastBlockHash) + strconv.Itoa(ducos1xxres))
		ducos1xx := d.Sum64()

		ducos1xxString := strconv.FormatUint(ducos1xx, 10)
		// Someone please implement hexdigest lol
		if ducos1xxString == expectedHash {
			return uint64(ducos1xxres)
		}
	}
	return 0
}
*/

func Mining() {

	fmt.Println("Connecting to server")
	ms := MasterServerClient{Client: &MasterServer{Sock: nil}}
	ms.Client.InitServerConnection()
	fmt.Println("Connected")

	/*
		version := ms.Client.CheckServerVersion()
		motd := ms.Client.CheckServerMessage()
		fmt.Printf("[Server version] %s\n", version)
		fmt.Printf("[Server motd] %s\n", motd)
	*/

	fmt.Println("MINiNG aWaYY")
	for {
		// TODO: Fix connecting problem. but I guess its server side issue. so there is nothing I can do.
		job := ms.Client.CheckServerJob()

		if len(job) != 3 {
			time.Sleep((1 * time.Second))
			continue
		}

		lastBlockHash := job[0]
		expectedHash := job[1]
		difficulty, _ := strconv.Atoi(job[2])
		efficiency := 65
		/* TODO:
		Calculate efficiency
		Threading
		*/

		ducos1res, hashrate := ducos1(lastBlockHash, expectedHash, difficulty, efficiency)
		if hashrate < 0 {
			continue
		}

		ms.Client.CheckSolvedHash(ducos1res, hashrate)
	}
}

func main() {
	Mining()
}
