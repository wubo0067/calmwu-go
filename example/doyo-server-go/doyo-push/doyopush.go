/*
 * @Author: ternence
 * @Date: 2018-11-06 17:01:39
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-11-24 11:37:13
 */

package main

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/urfave/cli"

	base "doyo-server-go/doyo-base-go"

	uuid "github.com/satori/go.uuid"
)

/*
 * @Author: calmwu
 * @Date: 2018-10-12 10:51:24
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-10-12 17:23:13
 */

var (
	appFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "videodir, d",
			Value: "",
			Usage: "视频文件目录",
		},
		cli.StringFlag{
			Name:  "roomid, i",
			Value: "",
			Usage: "房间id",
		},
		cli.StringFlag{
			Name:  "env, e",
			Value: "test",
			Usage: "环境 test/product",
		},
		cli.StringFlag{
			Name:  "country, c",
			Value: "",
			Usage: "国家 th/id/vn/others",
		},
	}

	exitFlag    = false
	logger      *log.Logger
	ffMpegeProc *os.Process
)

const pushStreamUrlTemp = "rtmp://video-center-sg.alivecdn.com/{appName}/{streamName}?vhost={domain}&auth_key={timestamp}-{rand}-{userId}-{hashValue}"
const sourceValue = "/{appName}/{streamName}-{timestamp}-{rand}-{userId}-{privateKey}"

func getPushStreamParams(env string, roomID string) (appName string, overTime int, streamName string, userID string, privateKey string, domain string) {
	if env == "test" {
		/*
			appName = "doyo_test_v1"
			overTime = 21600
			streamName = "100026"
			userID = "im100026"
			privateKey = "reWwlQL6AH"
			domain = "tlive.doyo.tv"
		*/

		appName = "doyo_test_v1"
		overTime = 21600
		streamName = roomID
		userID = fmt.Sprintf("im%s", roomID)
		privateKey = "reWwlQL6AH"
		domain = "tlive.doyo.tv"

	} else {
		appName = "doyo_v1"
		overTime = 21600
		streamName = roomID
		userID = fmt.Sprintf("im%s", roomID)
		privateKey = "wNQK8dE66W"
		domain = "live.doyo.tv"
	}
	return
}

func execFFmpeg(mp4 string, pushURL string) (string, error) {
	//函数返回一个*Cmd，用于使用给出的参数执行name指定的程序
	pushCmd := fmt.Sprintf("/usr/bin/ffmpeg -re -i %s -vcodec copy -acodec copy -f flv \"%s\"",
		mp4, pushURL)
	logger.Println(pushCmd)

	var outPut bytes.Buffer
	cmd := exec.Command("bash", "-c", pushCmd)
	cmd.Stdout = &outPut
	cmd.Stderr = &outPut
	err := cmd.Start()
	if err != nil {
		logger.Printf("cmd start failed! reason:%s", err.Error())
		return "nil", err
	}
	ffMpegeProc = cmd.Process
	logger.Printf("ffmpeg process pid[%d]", ffMpegeProc.Pid)
	cmd.Wait()

	return string(outPut.Bytes()), err
}

func generatePushUrl(env string, roomID string) string {
	//生成推流地址
	appName, overTime, streamName, userID, privateKey, domian := getPushStreamParams(env, roomID)

	uid, _ := uuid.NewV4()
	randID := strings.Replace(uid.String(), "-", "", -1)
	timeStamp := strconv.FormatInt((time.Now().Unix() + int64(overTime)), 10)

	value := strings.Replace(sourceValue, "{appName}", appName, -1)
	value = strings.Replace(value, "{streamName}", streamName, -1)
	value = strings.Replace(value, "{timestamp}", timeStamp, -1)
	value = strings.Replace(value, "{rand}", randID, -1)
	value = strings.Replace(value, "{userId}", userID, -1)
	value = strings.Replace(value, "{privateKey}", privateKey, -1)

	hashValue := fmt.Sprintf("%x", md5.Sum([]byte(value)))

	pushURL := strings.Replace(pushStreamUrlTemp, "{appName}", appName, -1)
	pushURL = strings.Replace(pushURL, "{streamName}", streamName, -1)
	pushURL = strings.Replace(pushURL, "{domain}", domian, -1)
	pushURL = strings.Replace(pushURL, "{timestamp}", timeStamp, -1)
	pushURL = strings.Replace(pushURL, "{rand}", randID, -1)
	pushURL = strings.Replace(pushURL, "{userId}", userID, -1)
	pushURL = strings.Replace(pushURL, "{hashValue}", hashValue, -1)

	logger.Printf("roomID[%s] env[%s] appName[%s] streamName[%s] userID[%s] privateKey[%s] domain[%s] randID[%s] timeStamp[%s] value[%s] hashValue[%s] pushURL[%s]\n",
		roomID, env, appName, streamName, userID, privateKey, domian, randID, timeStamp, value, hashValue, pushURL)

	return pushURL
}

var mp4Lst []string

func doPush(videoDir string, roomID string, env string, country string) {
	// 判断目录
	err := base.CheckDir(videoDir)
	if err != nil {
		logger.Printf("videoDir[%s] is invalid!\n", videoDir)
		os.Exit(-1)
	}

	// 读取目录中的文件
	err = filepath.Walk(videoDir, findMp4DirFile)
	if err != nil {
		logger.Printf("Get mp4 file from %s failed! reason:%s\n", videoDir, err.Error())
		os.Exit(-1)
	}
	if len(mp4Lst) == 0 {
		logger.Printf("The directory[%s] is empty\n", videoDir)
		os.Exit(-1)
	}
	logger.Printf("videoDir[%s] roomID[%s] env[%s]\n", videoDir, roomID, env)

	sort.SliceStable(mp4Lst, func(l, r int) bool {
		lName := mp4Lst[l]
		rName := mp4Lst[r]

		//logger.Printf("lName[%s] rName[%s]", lName, rName)

		lNames := strings.Split(lName, "/")
		lFileName := lNames[len(lNames)-1]

		rNames := strings.Split(rName, "/")
		rFileName := rNames[len(rNames)-1]

		//logger.Printf("lFileName[%s] rFileName[%s]", lFileName, rFileName)

		lNumName := strings.Split(lFileName, ".")[0]
		rNumName := strings.Split(rFileName, ".")[0]

		lNum, _ := strconv.Atoi(strings.Split(lNumName, ".")[0])
		rNum, _ := strconv.Atoi(strings.Split(rNumName, ".")[0])

		return lNum < rNum
	})

	for !exitFlag {
		// 循环文件
		logger.Printf("mp4lst:%+v\n", mp4Lst)

		for index, mp4 := range mp4Lst {
			logger.Printf("-----------mp4 index:%d %s\n", index, mp4)
			pushURL := generatePushUrl(env, roomID)
			output, err := execFFmpeg(mp4, pushURL)
			if err != nil {
				logger.Printf("ffmpeg failed: %s\n", err.Error())
			} else {
				logger.Printf("ffmpeg play %s end! output:%s\n", mp4, output)
			}

			time.Sleep(2 * time.Second)
			if exitFlag {
				break
			}
		}
	}
	return
}

func findMp4DirFile(path string, info os.FileInfo, err error) error {
	ok, err := filepath.Match("*.mp4", info.Name())
	if ok {
		if info.IsDir() {
			return nil
		}
		mp4Lst = append(mp4Lst, path)
		return nil
	}
	return err
}

func main() {
	logger = base.NewSimpleLog(nil)

	err := loadConfig("./config.json")
	if err != nil {
		logger.Printf("load config.json failed! reason:%s", err.Error())
		return
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)

	app := cli.NewApp()
	app.Name = "doyopush"
	app.Usage = "doyopush"
	app.Flags = appFlags

	app.Action = func(c *cli.Context) error {
		videoDir := c.String("videodir")
		roomID := c.String("roomid")
		env := c.String("env")
		country := strings.ToUpper(c.string("country"))
		doPush(videoDir, roomID, env, country)
		return nil
	}

	go func() {
		for {
			select {
			case sig := <-sigChan:
				switch sig {
				case syscall.SIGINT:
					fallthrough
				case syscall.SIGTERM:
					if ffMpegeProc != nil {
						// 杀死推流ffmpeg命令
						logger.Printf("kill ffmpeg:%d", ffMpegeProc.Pid)
						ffMpegeProc.Kill()
					}
					exitFlag = true
				case syscall.SIGUSR1:
					logger.Printf("reload config.json")
					loadConfig("./config.json")
				}

			}
		}

	}()

	app.Run(os.Args)

	time.Sleep(time.Second)

	logger.Println("doyopush exit!")
}
